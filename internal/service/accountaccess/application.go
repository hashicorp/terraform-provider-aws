// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_accountaccess_application", name="Application")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/accountaccess;accountaccess.GetApplicationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstances")
// @Testing(identityRegionOverrideTest=false)
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
func newApplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &applicationResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type applicationResource struct {
	framework.ResourceWithModel[applicationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *applicationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:                     framework.ARNAttributeComputedOnly(),
			"identity_center_application_arn": framework.ARNAttributeComputedOnly(),
			"identity_center_instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tenant_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *applicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan applicationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	identityCenterInstanceARN := fwflex.StringValueFromFramework(ctx, plan.IdentityCenterInstanceARN)
	input := &accountaccess.CreateApplicationInput{
		IdentitySource: &awstypes.IdentitySourceMemberIdentityCenter{
			Value: awstypes.IdentityCenter{
				InstanceArn: aws.String(identityCenterInstanceARN),
			},
		},
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateApplication(ctx, input)
	// AlreadyCreatedException means an Application already exists for this
	// IdC instance. Surface a clear, importable error rather than a generic
	// "ConflictException" so users know the recovery path.
	if errs.IsA[*awstypes.AlreadyCreatedException](err) {
		response.Diagnostics.AddError(
			"creating Account Access Application: an application already exists for this Identity Center instance",
			"AccountAccess allows only one Application per Identity Center instance. "+
				"To manage the existing Application with Terraform, run:\n\n"+
				"  terraform import aws_accountaccess_application.<name> <applicationArn>\n",
		)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, identityCenterInstanceARN)
		return
	}

	plan.ARN = fwflex.StringToFramework(ctx, output.ApplicationArn)

	app, err := waitApplicationCreated(ctx, conn, plan.ARN.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Set ARN so the resource is tracked even if the waiter timed out;
		// next refresh will see the actual state.
		response.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.ARN)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, app, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *applicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state applicationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	app, err := findApplicationByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, app, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *applicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state applicationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := accountaccess.DeleteApplicationInput{
		ApplicationArn: aws.String(arn),
	}
	_, err := conn.DeleteApplication(ctx, &input)
	if isResourceNotFoundError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitApplicationDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *applicationResource) flatten(ctx context.Context, app *accountaccess.GetApplicationOutput, model *applicationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	model.TenantID = fwflex.StringToFramework(ctx, app.TenantId)

	if details, ok := app.IdentitySource.(*awstypes.IdentitySourceDetailsMemberIdentityCenter); ok {
		model.IdentityCenterApplicationARN = fwflex.StringToFramework(ctx, details.Value.ApplicationArn)
		model.IdentityCenterInstanceARN = fwflex.StringToFrameworkARN(ctx, details.Value.InstanceArn)
	}

	setTagsOut(ctx, app.Tags)

	return diags
}

func findApplicationByARN(ctx context.Context, conn *accountaccess.Client, arn string) (*accountaccess.GetApplicationOutput, error) {
	input := accountaccess.GetApplicationInput{
		ApplicationArn: aws.String(arn),
	}
	return findApplication(ctx, conn, &input)
}

func findApplication(ctx context.Context, conn *accountaccess.Client, input *accountaccess.GetApplicationInput) (*accountaccess.GetApplicationOutput, error) {
	output, err := conn.GetApplication(ctx, input)
	if isResourceNotFoundError(err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	return output, nil
}

// statusApplication is the StateRefreshFunc used by the create/delete waiters.
func statusApplication(conn *accountaccess.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findApplicationByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return output, string(output.Status), nil
	}
}

func waitApplicationCreated(ctx context.Context, conn *accountaccess.Client, arn string, timeout time.Duration) (*accountaccess.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusCreateInProgress),
		Target:  enum.Slice(awstypes.StatusActive),
		Refresh: statusApplication(conn, arn),
		Timeout: timeout,
	}

	out, err := stateConf.WaitForStateContext(ctx)
	if app, ok := out.(*accountaccess.GetApplicationOutput); ok {
		if v := app.Error; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", v.Code, aws.ToString(v.Message)))
		}
		return app, err
	}
	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *accountaccess.Client, arn string, timeout time.Duration) (*accountaccess.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusActive, awstypes.StatusDeleteInProgress, awstypes.StatusCreateFailed),
		Target:  []string{},
		Refresh: statusApplication(conn, arn),
		Timeout: timeout,
	}
	out, err := stateConf.WaitForStateContext(ctx)
	if app, ok := out.(*accountaccess.GetApplicationOutput); ok {
		if v := app.Error; v != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", v.Code, aws.ToString(v.Message)))
		}
		return app, err
	}
	return nil, err
}

// applicationResourceModel is the Terraform state representation of an
// Account Access Application.
type applicationResourceModel struct {
	framework.WithRegionModel
	ARN                          types.String   `tfsdk:"arn"`
	IdentityCenterApplicationARN types.String   `tfsdk:"identity_center_application_arn"`
	IdentityCenterInstanceARN    fwtypes.ARN    `tfsdk:"identity_center_instance_arn"`
	Tags                         tftags.Map     `tfsdk:"tags"`
	TagsAll                      tftags.Map     `tfsdk:"tags_all"`
	TenantID                     types.String   `tfsdk:"tenant_id"`
	Timeouts                     timeouts.Value `tfsdk:"timeouts"`
}
