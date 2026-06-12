// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dsql

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	clusterPolicyAttrBypassPolicyLockoutSafetyCheck = "bypass_policy_lockout_safety_check"
	clusterPolicyAttrPolicyVersion                  = "policy_version"
)

var clusterPolicyIdentifierRegex = regexp.MustCompile(`^[a-z0-9]{26}$`)

type clusterPolicyResourceModel struct {
	framework.WithRegionModel
	BypassPolicyLockoutSafetyCheck types.Bool        `tfsdk:"bypass_policy_lockout_safety_check"`
	ID                             types.String      `tfsdk:"id"`
	Identifier                     types.String      `tfsdk:"identifier"`
	Policy                         fwtypes.IAMPolicy `tfsdk:"policy"`
	PolicyVersion                  types.String      `tfsdk:"policy_version"`
	Timeouts                       timeouts.Value    `tfsdk:"timeouts"`
}

type clusterPolicyResource struct {
	framework.ResourceWithModel[clusterPolicyResourceModel]
	framework.WithTimeouts
}

// @FrameworkResource("aws_dsql_cluster_policy", name="Cluster Policy")
// @Testing(importIgnore="bypass_policy_lockout_safety_check;policy")
func newClusterPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &clusterPolicyResource{}
	r.SetDefaultCreateTimeout(1 * time.Minute)
	r.SetDefaultUpdateTimeout(1 * time.Minute)
	r.SetDefaultDeleteTimeout(1 * time.Minute)

	return r, nil
}

func (cpr *clusterPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			clusterPolicyAttrBypassPolicyLockoutSafetyCheck: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrIdentifier: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(clusterPolicyIdentifierRegex, "must be a valid Aurora DSQL cluster identifier"),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20480),
				},
			},
			clusterPolicyAttrPolicyVersion: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func putClusterPolicyVersion(output *dsql.PutClusterPolicyOutput) (string, error) {
	if output == nil || output.PolicyVersion == nil {
		return "", tfresource.NewEmptyResultError()
	}

	return aws.ToString(output.PolicyVersion), nil
}

func syncClusterPolicyAfterPut(ctx context.Context, conn *dsql.Client, data *clusterPolicyResourceModel, policyVersion, operationName string, timeout time.Duration) error {
	identifier := data.Identifier.ValueString()

	data.ID = data.Identifier
	data.PolicyVersion = types.StringValue(policyVersion)

	clusterPolicy, err := waitClusterPolicyUpdated(ctx, conn, identifier, policyVersion, data.Policy.ValueString(), timeout)
	if err != nil {
		return fmt.Errorf("waiting for Aurora DSQL Cluster Policy (%s) %s: %w", identifier, operationName, err)
	}

	policyToSet, err := verify.PolicyToSet(data.Policy.ValueString(), aws.ToString(clusterPolicy.Policy))
	if err != nil {
		return fmt.Errorf("setting Aurora DSQL Cluster Policy (%s): %w", identifier, err)
	}

	data.Policy = fwtypes.IAMPolicyValue(policyToSet)
	data.PolicyVersion = types.StringPointerValue(clusterPolicy.PolicyVersion)

	return nil
}

func findClusterPolicyByID(ctx context.Context, conn *dsql.Client, identifier string) (*dsql.GetClusterPolicyOutput, error) {
	input := &dsql.GetClusterPolicyInput{
		Identifier: aws.String(identifier),
	}

	output, err := conn.GetClusterPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil || output.PolicyVersion == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func waitClusterPolicyUpdated(ctx context.Context, conn *dsql.Client, identifier, expectedPolicyVersion, expectedPolicy string, timeout time.Duration) (*dsql.GetClusterPolicyOutput, error) {
	var policyOutput *dsql.GetClusterPolicyOutput

	err := tfresource.WaitUntil(ctx, timeout, func(ctx context.Context) (bool, error) {
		var err error
		policyOutput, err = findClusterPolicyByID(ctx, conn, identifier)

		if retry.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if aws.ToString(policyOutput.PolicyVersion) != expectedPolicyVersion {
			return false, nil
		}

		return verify.PolicyStringsEquivalent(expectedPolicy, aws.ToString(policyOutput.Policy)), nil
	}, tfresource.WaitOpts{
		MinTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	if policyOutput == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return policyOutput, nil
}

func (cpr *clusterPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data clusterPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := cpr.Meta().DSQLClient(ctx)
	identifier := data.Identifier.ValueString()
	input := &dsql.PutClusterPolicyInput{
		BypassPolicyLockoutSafetyCheck: data.BypassPolicyLockoutSafetyCheck.ValueBool(),
		ClientToken:                    aws.String(create.UniqueId(ctx)),
		Identifier:                     aws.String(identifier),
		Policy:                         data.Policy.ValueStringPointer(),
	}

	output, err := conn.PutClusterPolicy(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Aurora DSQL Cluster Policy (%s)", identifier), err.Error())

		return
	}

	policyVersion, err := putClusterPolicyVersion(output)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Aurora DSQL Cluster Policy (%s)", identifier), err.Error())

		return
	}

	if err := syncClusterPolicyAfterPut(ctx, conn, &data, policyVersion, "create", cpr.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Aurora DSQL Cluster Policy (%s)", identifier), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (cpr *clusterPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data clusterPolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := cpr.Meta().DSQLClient(ctx)
	output, err := findClusterPolicyByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Aurora DSQL Cluster Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	policyToSet, err := verify.PolicyToSet(data.Policy.ValueString(), aws.ToString(output.Policy))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("setting Aurora DSQL Cluster Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.Identifier = data.ID
	data.Policy = fwtypes.IAMPolicyValue(policyToSet)
	data.PolicyVersion = types.StringPointerValue(output.PolicyVersion)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (cpr *clusterPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new clusterPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := cpr.Meta().DSQLClient(ctx)
	identifier := new.Identifier.ValueString()
	input := &dsql.PutClusterPolicyInput{
		BypassPolicyLockoutSafetyCheck: new.BypassPolicyLockoutSafetyCheck.ValueBool(),
		ClientToken:                    aws.String(create.UniqueId(ctx)),
		ExpectedPolicyVersion:          old.PolicyVersion.ValueStringPointer(),
		Identifier:                     aws.String(identifier),
		Policy:                         new.Policy.ValueStringPointer(),
	}

	output, err := conn.PutClusterPolicy(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Aurora DSQL Cluster Policy (%s)", identifier), err.Error())

		return
	}

	policyVersion, err := putClusterPolicyVersion(output)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Aurora DSQL Cluster Policy (%s)", identifier), err.Error())

		return
	}

	if err := syncClusterPolicyAfterPut(ctx, conn, &new, policyVersion, "update", cpr.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Aurora DSQL Cluster Policy (%s)", identifier), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func waitClusterPolicyDeleted(ctx context.Context, conn *dsql.Client, identifier string, timeout time.Duration) error {
	return tfresource.WaitUntil(ctx, timeout, func(ctx context.Context) (bool, error) {
		_, err := findClusterPolicyByID(ctx, conn, identifier)

		if retry.NotFound(err) {
			return true, nil
		}

		if err != nil {
			return false, err
		}

		return false, nil
	}, tfresource.WaitOpts{
		MinTimeout: 5 * time.Second,
	})
}

func (cpr *clusterPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data clusterPolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := cpr.Meta().DSQLClient(ctx)
	input := &dsql.DeleteClusterPolicyInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		Identifier:  data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteClusterPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Aurora DSQL Cluster Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if err := waitClusterPolicyDeleted(ctx, conn, data.ID.ValueString(), cpr.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Aurora DSQL Cluster Policy (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (cpr *clusterPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrIdentifier), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(clusterPolicyAttrBypassPolicyLockoutSafetyCheck), false)...)
}
