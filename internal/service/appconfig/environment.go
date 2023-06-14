package appconfig

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
// @Tags(identifierAttribute="arn")
func newResourceEnvironment(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnvironment{}
	r.SetMigratedFromPluginSDK(true)

	return r, nil
}

type resourceEnvironment struct {
	framework.ResourceWithConfigure
}

func (r *resourceEnvironment) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appconfig_environment"
}

func (r *resourceEnvironment) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]{4,7}$`),
						"value must contain 4-7 lowercase letters or numbers",
					),
				},
			},
			"arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"environment_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"monitor": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeAtMost(5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alarm_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
						"alarm_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(20, 2048),
							},
						},
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourceEnvironment) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().AppConfigConn(ctx)

	var plan resourceEnvironmentData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	appId := plan.ApplicationID.ValueString()

	var monitors []monitorData
	response.Diagnostics.Append(plan.Monitors.ElementsAs(ctx, &monitors, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &appconfig.CreateEnvironmentInput{
		Name:          aws.String(plan.Name.ValueString()),
		ApplicationId: aws.String(appId),
		Tags:          getTagsIn(ctx),
		Monitors:      expandMonitors(monitors),
	}

	if !(plan.Description.IsNull() || plan.Description.IsUnknown()) {
		input.Description = aws.String(plan.Description.ValueString())
	}

	environment, err := conn.CreateEnvironmentWithContext(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("creating AppConfig Environment for Application (%s)", appId),
			err.Error(),
		)
	}
	if environment == nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("creating AppConfig Environment for Application (%s)", appId),
			"empty response",
		)
	}

	state := plan

	response.Diagnostics.Append(state.refreshFromCreateOutput(ctx, r.Meta(), environment)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceEnvironment) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().AppConfigConn(ctx)

	var state resourceEnvironmentData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	envID := state.EnvironmentID.ValueString()

	input := &appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(appID),
		EnvironmentId: aws.String(envID),
	}

	output, err := conn.GetEnvironmentWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("reading AppConfig Environment (%s) for Application (%s)", envID, appID),
			err.Error(),
		)
	}

	response.Diagnostics.Append(state.refreshFromGetOutput(ctx, r.Meta(), output)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceEnvironment) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().AppConfigConn(ctx)

	var state resourceEnvironmentData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceEnvironmentData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) ||
		!plan.Name.Equal(state.Name) ||
		!plan.Monitors.Equal(state.Monitors) {
		appID := state.ApplicationID.ValueString()
		envID := state.EnvironmentID.ValueString()

		updateInput := &appconfig.UpdateEnvironmentInput{
			EnvironmentId: aws.String(envID),
			ApplicationId: aws.String(appID),
		}

		if !plan.Description.Equal(state.Description) {
			updateInput.Description = aws.String(plan.Description.ValueString())
		}

		if !plan.Name.Equal(state.Name) {
			updateInput.Name = aws.String(plan.Name.ValueString())
		}

		if !plan.Monitors.Equal(state.Monitors) {
			var monitors []monitorData
			response.Diagnostics.Append(plan.Monitors.ElementsAs(ctx, &monitors, false)...)
			if response.Diagnostics.HasError() {
				return
			}
			updateInput.Monitors = expandMonitors(monitors)
		}

		output, err := conn.UpdateEnvironmentWithContext(ctx, updateInput)
		if err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf("updating AppConfig Environment (%s) for Application (%s)", envID, appID),
				err.Error(),
			)
		}

		response.Diagnostics.Append(plan.refreshFromUpdateOutput(ctx, r.Meta(), output)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceEnvironment) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().AppConfigConn(ctx)

	var state resourceEnvironmentData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	envID := state.EnvironmentID.ValueString()

	input := &appconfig.DeleteEnvironmentInput{
		ApplicationId: aws.String(appID),
		EnvironmentId: aws.String(envID),
	}

	tflog.Debug(ctx, "Deleting AppConfig Environment", map[string]any{
		"application_id": appID,
		"environment_id": envID,
	})

	_, err := conn.DeleteEnvironmentWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("deleting AppConfig Environment (%s) for Application (%s)", envID, appID),
			err.Error(),
		)
	}
}

func (r *resourceEnvironment) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ":")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "EnvironmentID:ApplicationID"`, request.ID))
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("environment_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("application_id"), parts[1])...)
}

func (r *resourceEnvironment) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceEnvironmentData struct {
	ApplicationID types.String `tfsdk:"application_id"`
	ARN           types.String `tfsdk:"arn"`
	Description   types.String `tfsdk:"description"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ID            types.String `tfsdk:"id"`
	Monitors      types.Set    `tfsdk:"monitor"`
	Name          types.String `tfsdk:"name"`
	State         types.String `tfsdk:"state"`
	Tags          types.Map    `tfsdk:"tags"`
	TagsAll       types.Map    `tfsdk:"tags_all"`
}

func (d *resourceEnvironmentData) refreshFromCreateOutput(ctx context.Context, meta *conns.AWSClient, out *appconfig.CreateEnvironmentOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	appID := aws.StringValue(out.ApplicationId)
	envID := aws.StringValue(out.Id)

	d.ApplicationID = types.StringValue(appID)
	d.ARN = types.StringValue(environmentARN(meta, aws.StringValue(out.ApplicationId), aws.StringValue(out.Id)).String())
	d.Description = flex.StringToFrameworkLegacy(ctx, out.Description)
	d.EnvironmentID = types.StringValue(envID)
	d.ID = types.StringValue(fmt.Sprintf("%s:%s", envID, appID))
	d.Monitors = flattenMonitors(ctx, out.Monitors, &diags)
	d.Name = flex.StringToFramework(ctx, out.Name)
	d.State = flex.StringToFramework(ctx, out.State)

	return diags
}

func (d *resourceEnvironmentData) refreshFromGetOutput(ctx context.Context, meta *conns.AWSClient, out *appconfig.GetEnvironmentOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	appID := aws.StringValue(out.ApplicationId)
	envID := aws.StringValue(out.Id)

	d.ApplicationID = types.StringValue(appID)
	d.ARN = types.StringValue(environmentARN(meta, aws.StringValue(out.ApplicationId), aws.StringValue(out.Id)).String())
	d.Description = flex.StringToFrameworkLegacy(ctx, out.Description)
	d.EnvironmentID = types.StringValue(envID)
	d.ID = types.StringValue(fmt.Sprintf("%s:%s", envID, appID))
	d.Monitors = flattenMonitors(ctx, out.Monitors, &diags)
	d.Name = flex.StringToFramework(ctx, out.Name)
	d.State = flex.StringToFramework(ctx, out.State)

	return diags
}

func (d *resourceEnvironmentData) refreshFromUpdateOutput(ctx context.Context, meta *conns.AWSClient, out *appconfig.UpdateEnvironmentOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	appID := aws.StringValue(out.ApplicationId)
	envID := aws.StringValue(out.Id)

	d.ApplicationID = types.StringValue(appID)
	d.ARN = types.StringValue(environmentARN(meta, aws.StringValue(out.ApplicationId), aws.StringValue(out.Id)).String())
	d.Description = flex.StringToFrameworkLegacy(ctx, out.Description)
	d.EnvironmentID = types.StringValue(envID)
	d.ID = types.StringValue(fmt.Sprintf("%s:%s", envID, appID))
	d.Monitors = flattenMonitors(ctx, out.Monitors, &diags)
	d.Name = flex.StringToFramework(ctx, out.Name)
	d.State = flex.StringToFramework(ctx, out.State)

	return diags
}

func environmentARN(meta *conns.AWSClient, appID, envID string) arn.ARN {
	return arn.ARN{
		AccountID: meta.AccountID,
		Partition: meta.Partition,
		Region:    meta.Region,
		Resource:  fmt.Sprintf("application/%s/environment/%s", appID, envID),
		Service:   "appconfig",
	}
}

func expandMonitors(l []monitorData) []*appconfig.Monitor {
	monitors := make([]*appconfig.Monitor, len(l))
	for i, item := range l {
		monitors[i] = item.expand()
	}
	return monitors
}

func flattenMonitors(ctx context.Context, apiObjects []*appconfig.Monitor, diags *diag.Diagnostics) types.Set {
	monitorDataTypes := framework.AttributeTypesMust[monitorData](ctx)
	elemType := types.ObjectType{AttrTypes: monitorDataTypes}

	if len(apiObjects) == 0 {
		return types.SetValueMust(elemType, []attr.Value{})
	}

	values := make([]attr.Value, len(apiObjects))
	for i, o := range apiObjects {
		values[i] = flattenMonitorData(ctx, o, diags).value(ctx, diags)
	}

	result, d := types.SetValueFrom(ctx, elemType, values)
	diags.Append(d...)

	return result
}

type monitorData struct {
	AlarmARN     fwtypes.ARN `tfsdk:"alarm_arn"`
	AlarmRoleARN fwtypes.ARN `tfsdk:"alarm_role_arn"`
}

func (m monitorData) expand() *appconfig.Monitor {
	result := &appconfig.Monitor{
		AlarmArn: aws.String(m.AlarmARN.ValueARN().String()),
	}

	if !m.AlarmRoleARN.IsNull() {
		result.AlarmRoleArn = aws.String(m.AlarmRoleARN.ValueARN().String())
	}

	return result
}

func flattenMonitorData(ctx context.Context, apiObject *appconfig.Monitor, diags *diag.Diagnostics) monitorData {
	return monitorData{
		AlarmARN:     flex.StringToFrameworkARN(ctx, apiObject.AlarmArn, diags),
		AlarmRoleARN: flex.StringToFrameworkARN(ctx, apiObject.AlarmRoleArn, diags),
	}
}

func (m monitorData) value(ctx context.Context, diags *diag.Diagnostics) types.Object {
	monitorDataTypes := framework.AttributeTypesMust[monitorData](ctx)

	obj, d := types.ObjectValueFrom(ctx, monitorDataTypes, m)
	diags.Append(d...)

	return obj
}
