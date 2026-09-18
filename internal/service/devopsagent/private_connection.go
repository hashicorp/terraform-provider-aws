// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"context"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_devopsagent_private_connection", name="Private Connection")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name")
// @ArnFormat("private-connection/{name}", attribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/devopsagent;devopsagent.DescribePrivateConnectionOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="randomPrivateConnectionName(t)")
// @Testing(importStateIdAttribute="name")
// @Testing(importIgnore="resource_configuration_id")
// @Testing(identityTest=false)
func newPrivateConnectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &privateConnectionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNamePrivateConnection = "Private Connection"
)

type privateConnectionResource struct {
	framework.ResourceWithModel[privateConnectionResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *privateConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCertificate: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"host_address": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrMode: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(awstypes.PrivateConnectionTypeSelfManaged),
						string(awstypes.PrivateConnectionTypeServiceManaged),
					),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_configuration_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					resourceConfigurationIDPlanModifier{},
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrSubnetIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *privateConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var plan privateConnectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	input := devopsagent.CreatePrivateConnectionInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	connType := awstypes.PrivateConnectionType(plan.Mode.ValueString())

	switch connType {
	case awstypes.PrivateConnectionTypeSelfManaged:
		selfManaged := &awstypes.PrivateConnectionModeMemberSelfManaged{
			Value: awstypes.SelfManagedInput{
				ResourceConfigurationId: plan.ResourceConfigurationID.ValueStringPointer(),
			},
		}
		if !plan.Certificate.IsNull() && !plan.Certificate.IsUnknown() {
			selfManaged.Value.Certificate = plan.Certificate.ValueStringPointer()
		}
		input.Mode = selfManaged

	case awstypes.PrivateConnectionTypeServiceManaged:
		serviceManaged := &awstypes.PrivateConnectionModeMemberServiceManaged{
			Value: awstypes.ServiceManagedInput{
				HostAddress: plan.HostAddress.ValueStringPointer(),
				VpcId:       plan.VpcID.ValueStringPointer(),
			},
		}
		if !plan.SubnetIDs.IsNull() && !plan.SubnetIDs.IsUnknown() {
			serviceManaged.Value.SubnetIds = flex.ExpandFrameworkStringValueSet(ctx, plan.SubnetIDs)
		}
		if !plan.Certificate.IsNull() && !plan.Certificate.IsUnknown() {
			serviceManaged.Value.Certificate = plan.Certificate.ValueStringPointer()
		}
		input.Mode = serviceManaged
	}

	// Retry on transient "Failed to create resource association" errors
	// that occur when the resource gateway dependency isn't fully ready.
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, createTimeout, func(ctx context.Context) (*devopsagent.CreatePrivateConnectionOutput, error) {
		return conn.CreatePrivateConnection(ctx, &input)
	}, "ValidationException", "Failed to create resource association")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	// Wait for creation to complete
	out, err := waitPrivateConnectionCreated(ctx, conn, name, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set computed attributes from the describe response
	r.flatten(ctx, out, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *privateConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var state privateConnectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPrivateConnectionByName(ctx, conn, state.Name.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.ValueString())
		return
	}

	r.flatten(ctx, out, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *privateConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var plan, state privateConnectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Only the certificate can be updated in-place.
	if !plan.Certificate.Equal(state.Certificate) {
		input := devopsagent.UpdatePrivateConnectionCertificateInput{
			Name:        plan.Name.ValueStringPointer(),
			Certificate: plan.Certificate.ValueStringPointer(),
		}

		_, err := conn.UpdatePrivateConnectionCertificate(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
			return
		}

		// Wait for update to complete
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitPrivateConnectionUpdated(ctx, conn, plan.Name.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
			return
		}
	}

	// Always read back to populate computed attributes (e.g., status).
	out, err := findPrivateConnectionByName(ctx, conn, plan.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	r.flatten(ctx, out, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *privateConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsAgentClient(ctx)

	var state privateConnectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := devopsagent.DeletePrivateConnectionInput{
		Name: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeletePrivateConnection(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPrivateConnectionDeleted(ctx, conn, state.Name.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.ValueString())
		return
	}
}

// flatten maps the DescribePrivateConnection output to the resource model.
func (r *privateConnectionResource) flatten(ctx context.Context, out *devopsagent.DescribePrivateConnectionOutput, model *privateConnectionResourceModel, diags *diag.Diagnostics) {
	if out == nil {
		diags.AddError("flattening Private Connection", "nil output")
		return
	}

	model.Name = flex.StringToFramework(ctx, out.Name)
	model.Status = types.StringValue(string(out.Status))
	model.Mode = types.StringValue(string(out.Type))

	// Construct the ARN since the API does not return it directly.
	arn := r.Meta().RegionalARN(ctx, "aidevops", "private-connection/"+aws.ToString(out.Name))
	model.ARN = types.StringValue(arn)

	if out.ResourceConfigurationId != nil {
		apiVal := aws.ToString(out.ResourceConfigurationId)
		// Preserve the user-provided value if it refers to the same resource
		// (short ID vs ARN). Only overwrite if genuinely different or if model is empty.
		if model.ResourceConfigurationID.IsNull() || !isEquivalentResourceConfigurationID(apiVal, model.ResourceConfigurationID.ValueString()) {
			model.ResourceConfigurationID = types.StringValue(apiVal)
		}
	}
	if out.HostAddress != nil {
		model.HostAddress = flex.StringToFramework(ctx, out.HostAddress)
	}
	if out.VpcId != nil {
		model.VpcID = flex.StringToFramework(ctx, out.VpcId)
	}

	setTagsOut(ctx, out.Tags)
}

// Waiters

func waitPrivateConnectionCreated(ctx context.Context, conn *devopsagent.Client, name string, timeout time.Duration) (*devopsagent.DescribePrivateConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PrivateConnectionStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.PrivateConnectionStatusActive),
		Refresh:                   statusPrivateConnection(conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*devopsagent.DescribePrivateConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitPrivateConnectionUpdated(ctx context.Context, conn *devopsagent.Client, name string, timeout time.Duration) (*devopsagent.DescribePrivateConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PrivateConnectionStatusCreateInProgress),
		Target:                    enum.Slice(awstypes.PrivateConnectionStatusActive),
		Refresh:                   statusPrivateConnection(conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*devopsagent.DescribePrivateConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitPrivateConnectionDeleted(ctx context.Context, conn *devopsagent.Client, name string, timeout time.Duration) (*devopsagent.DescribePrivateConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PrivateConnectionStatusDeleteInProgress, awstypes.PrivateConnectionStatusActive),
		Target:  []string{},
		Refresh: statusPrivateConnection(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*devopsagent.DescribePrivateConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

// Status function

func statusPrivateConnection(conn *devopsagent.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPrivateConnectionByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

// Finder

func findPrivateConnectionByName(ctx context.Context, conn *devopsagent.Client, name string) (*devopsagent.DescribePrivateConnectionOutput, error) {
	input := devopsagent.DescribePrivateConnectionInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribePrivateConnection(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

// resourceConfigurationIDPlanModifier handles the short-ID vs ARN equivalence
// for resource_configuration_id. The API accepts both forms but always returns
// the ARN. This modifier suppresses replacement when the config value (short ID)
// is contained within the state value (ARN), and forces replacement otherwise.
type resourceConfigurationIDPlanModifier struct{}

func (m resourceConfigurationIDPlanModifier) Description(_ context.Context) string {
	return "Handles short-ID vs ARN equivalence for resource_configuration_id. Forces replacement when the underlying resource configuration changes."
}

func (m resourceConfigurationIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m resourceConfigurationIDPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the resource is being created or destroyed, nothing to do.
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		if !req.PlanValue.IsNull() && !req.PlanValue.Equal(req.StateValue) {
			resp.RequiresReplace = true
		}
		return
	}

	planVal := req.PlanValue.ValueString()
	stateVal := req.StateValue.ValueString()

	// If values are identical, no change needed.
	if planVal == stateVal {
		return
	}

	// The API always returns the ARN form:
	//   arn:aws:vpc-lattice:<region>:<account>:resourceconfiguration/<id>
	// The user may pass either the short ID (rcfg-xxx) or the full ARN.
	// Extract the ID suffix from the ARN and compare exactly.
	if isEquivalentResourceConfigurationID(stateVal, planVal) {
		// Same resource, just different format. Use the state value (ARN).
		resp.PlanValue = req.StateValue
		return
	}

	// Genuinely different resource configuration — force replacement.
	resp.RequiresReplace = true
}

// isEquivalentResourceConfigurationID returns true if the two values refer to
// the same VPC Lattice resource configuration. It handles the case where one
// value is the full ARN and the other is the short ID (the part after the last "/").
func isEquivalentResourceConfigurationID(a, b string) bool {
	idFromA := a
	if idx := strings.LastIndex(a, "/"); idx >= 0 {
		idFromA = a[idx+1:]
	}
	idFromB := b
	if idx := strings.LastIndex(b, "/"); idx >= 0 {
		idFromB = b[idx+1:]
	}
	return idFromA == idFromB
}

// Data model

type privateConnectionResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String        `tfsdk:"arn"`
	Certificate             types.String        `tfsdk:"certificate"`
	HostAddress             types.String        `tfsdk:"host_address"`
	Mode                    types.String        `tfsdk:"mode"`
	Name                    types.String        `tfsdk:"name"`
	ResourceConfigurationID types.String        `tfsdk:"resource_configuration_id"`
	Status                  types.String        `tfsdk:"status"`
	SubnetIDs               fwtypes.SetOfString `tfsdk:"subnet_ids"`
	Tags                    tftags.Map          `tfsdk:"tags"`
	TagsAll                 tftags.Map          `tfsdk:"tags_all"`
	Timeouts                timeouts.Value      `tfsdk:"timeouts"`
	VpcID                   types.String        `tfsdk:"vpc_id"`
}
