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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
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
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
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
			"identity_source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[identitySourceModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("identity_center"),
						),
					},
					Blocks: map[string]schema.Block{
						"identity_center": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"application_arn": framework.ARNAttributeComputedOnly(),
									"instance_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
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

	var input accountaccess.CreateApplicationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	// Setting tags on CreateApplication does not work. Apply tags after creation.
	// input.Tags = getTagsIn(ctx)

	output, err := conn.CreateApplication(ctx, &input)
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
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	arn := aws.ToString(output.ApplicationArn)
	app, err := waitApplicationCreated(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Set ARN so the resource is tracked even if the waiter timed out;
		// next refresh will see the actual state.
		response.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		if err := createTags(ctx, conn, arn, tags); err != nil {
			response.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	// Set values for unknowns.
	plan.ApplicationARN = fwflex.StringValueToFramework(ctx, arn)
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

	arn := fwflex.StringValueFromFramework(ctx, state.ApplicationARN)
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

	arn := fwflex.StringValueFromFramework(ctx, state.ApplicationARN)
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

func (r *applicationResource) flatten(ctx context.Context, app *accountaccess.GetApplicationOutput, model *applicationResourceModel) diag.Diagnostics { //nolint:unparam
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, app, model)...)
	// The tags returned from GetApplication are not to be trusted.
	// setTagsOut(ctx, app.Tags)
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
	ApplicationARN types.String                                         `tfsdk:"arn"`
	IdentitySource fwtypes.ListNestedObjectValueOf[identitySourceModel] `tfsdk:"identity_source"`
	Tags           tftags.Map                                           `tfsdk:"tags"`
	TagsAll        tftags.Map                                           `tfsdk:"tags_all"`
	TenantID       types.String                                         `tfsdk:"tenant_id"`
	Timeouts       timeouts.Value                                       `tfsdk:"timeouts"`
}

type identitySourceModel struct {
	IdentityCenter fwtypes.ListNestedObjectValueOf[identityCenterModel] `tfsdk:"identity_center"`
}

var (
	_ fwflex.Expander  = identitySourceModel{}
	_ fwflex.Flattener = &identitySourceModel{}
)

func (m *identitySourceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IdentitySourceDetailsMemberIdentityCenter:
		var model identityCenterModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.IdentityCenter, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("identitySourceModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m identitySourceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.IdentityCenter.IsNull():
		model, d := m.IdentityCenter.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.IdentitySourceMemberIdentityCenter
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

type identityCenterModel struct {
	ApplicationARN types.String `tfsdk:"application_arn"`
	InstanceARN    fwtypes.ARN  `tfsdk:"instance_arn"`
}
