// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/accountaccess;accountaccess.GetApplicationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(identityRegionOverrideTest=false)
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
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
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identity_center_application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identity_center_instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Status](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				// Keep the prior value on tag-only updates (transparent tagging
				// uses a no-op Update that doesn't refresh computed fields);
				// the next refresh picks up the new server timestamp. Without
				// this, an in-place tag update leaves updated_at unknown after
				// apply, which the framework rejects.
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

	input := &accountaccess.CreateApplicationInput{
		IdentitySource: &awstypes.IdentitySourceMemberIdentityCenter{
			Value: awstypes.IdentityCenter{
				InstanceArn: fwflex.StringFromFramework(ctx, plan.IdentityCenterInstanceARN),
			},
		},
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateApplication(ctx, input)
	if err != nil {
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
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.IdentityCenterInstanceARN.ValueString())
		return
	}
	if output == nil || output.ApplicationArn == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.IdentityCenterInstanceARN.ValueString())
		return
	}

	plan.ARN = fwflex.StringToFramework(ctx, output.ApplicationArn)
	plan.ID = plan.ARN

	app, err := waitApplicationCreated(ctx, conn, plan.ARN.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Set ID so the resource is tracked even if the waiter timed out;
		// next refresh will see the actual state.
		response.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID)
		response.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.ARN)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}

	flattenApplication(ctx, app, &plan)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *applicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state applicationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	app, err := FindApplicationByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	flattenApplication(ctx, app, &state)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *applicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state applicationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AccountAccessClient(ctx)

	_, err := conn.DeleteApplication(ctx, &accountaccess.DeleteApplicationInput{
		ApplicationArn: state.ARN.ValueStringPointer(),
	})
	if isNotFoundError(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}

	if _, err := waitApplicationDeleted(ctx, conn, state.ARN.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
}

// flattenApplication copies fields from the SDK GetApplication response onto
// the Terraform model. Handles the IdentitySource union by extracting the
// IdentityCenter member.
func flattenApplication(ctx context.Context, app *accountaccess.GetApplicationOutput, model *applicationResourceModel) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	model.Status = fwtypes.StringEnumValue(app.Status)
	model.TenantID = fwflex.StringToFramework(ctx, app.TenantId)
	model.CreatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(app.CreatedAt))
	model.UpdatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(app.UpdatedAt))

	if details, ok := app.IdentitySource.(*awstypes.IdentitySourceDetailsMemberIdentityCenter); ok {
		if details.Value.InstanceArn != nil {
			model.IdentityCenterInstanceARN = fwtypes.ARNValue(aws.ToString(details.Value.InstanceArn))
		}
		if details.Value.ApplicationArn != nil {
			model.IdentityCenterApplicationARN = fwtypes.ARNValue(aws.ToString(details.Value.ApplicationArn))
		} else {
			model.IdentityCenterApplicationARN = fwtypes.ARNNull()
		}
	}
}

// FindApplicationByARN returns the Application identified by ARN, or
// retry.NotFoundError when the API returns ResourceNotFoundException.
func FindApplicationByARN(ctx context.Context, conn *accountaccess.Client, arn string) (*accountaccess.GetApplicationOutput, error) {
	output, err := conn.GetApplication(ctx, &accountaccess.GetApplicationInput{
		ApplicationArn: aws.String(arn),
	})
	if isNotFoundError(err) {
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
		output, err := FindApplicationByARN(ctx, conn, arn)
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
		if app.Status == awstypes.StatusCreateFailed {
			return app, errors.New("application create failed (status CREATE_FAILED)")
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
		if app.Status == awstypes.StatusDeleteFailed {
			return app, errors.New("application delete failed (status DELETE_FAILED)")
		}
		return app, err
	}
	return nil, err
}

// applicationResourceModel is the Terraform state representation of an
// Account Access Application.
type applicationResourceModel struct {
	framework.WithRegionModel
	ARN                          types.String                        `tfsdk:"arn"`
	CreatedAt                    timetypes.RFC3339                   `tfsdk:"created_at"`
	ID                           types.String                        `tfsdk:"id"`
	IdentityCenterApplicationARN fwtypes.ARN                         `tfsdk:"identity_center_application_arn"`
	IdentityCenterInstanceARN    fwtypes.ARN                         `tfsdk:"identity_center_instance_arn"`
	Status                       fwtypes.StringEnum[awstypes.Status] `tfsdk:"status"`
	Tags                         tftags.Map                          `tfsdk:"tags"`
	TagsAll                      tftags.Map                          `tfsdk:"tags_all"`
	TenantID                     types.String                        `tfsdk:"tenant_id"`
	Timeouts                     timeouts.Value                      `tfsdk:"timeouts"`
	UpdatedAt                    timetypes.RFC3339                   `tfsdk:"updated_at"`
}
