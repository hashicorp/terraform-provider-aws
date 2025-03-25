// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_appconfig_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func newEnvironmentResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &environmentResource{}

	return r, nil
}

type environmentResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *environmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrApplicationID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9a-z]{4,7}$`),
						"value must contain 4-7 lowercase letters or numbers",
					),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""), // Needed for backwards compatibility with SDK resource
			},
			"environment_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"monitor": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[monitorModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtMost(5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alarm_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"alarm_role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
		},
	}
}

func (r *environmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppConfigClient(ctx)

	name := data.Name.ValueString()
	input := appconfig.CreateEnvironmentInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateEnvironment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating AppConfig Environment (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringValueToFramework(ctx, environmentARN(ctx, r.Meta(), aws.ToString(output.ApplicationId), aws.ToString(output.Id)))
	data.EnvironmentID = fwflex.StringToFramework(ctx, output.Id)
	data.State = fwflex.StringValueToFramework(ctx, output.State)
	data.setID(ctx)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *environmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(ctx); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().AppConfigClient(ctx)

	output, err := findEnvironmentByTwoPartKey(ctx, conn, data.ApplicationID.ValueString(), data.EnvironmentID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AppConfig Environment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.ARN = fwflex.StringValueToFramework(ctx, environmentARN(ctx, r.Meta(), aws.ToString(output.ApplicationId), aws.ToString(output.Id)))
	data.Description = fwflex.StringToFrameworkLegacy(ctx, output.Description)
	data.setID(ctx) // output.Id will have overwritten ID.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *environmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppConfigClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.Monitors.Equal(old.Monitors) ||
		!new.Name.Equal(old.Name) {
		input := appconfig.UpdateEnvironmentInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		if input.Monitors == nil {
			input.Monitors = make([]awstypes.Monitor, 0)
		}

		output, err := conn.UpdateEnvironment(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating AppConfig Environment (%s)", new.ID.ValueString()), err.Error())

			return
		}

		new.State = fwflex.StringValueToFramework(ctx, output.State)
	} else {
		new.State = old.State
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppConfigClient(ctx)

	tflog.Debug(ctx, "Deleting AppConfig Environment", map[string]any{
		names.AttrApplicationID: data.ApplicationID.ValueString(),
		"environment_id":        data.EnvironmentID.ValueString(),
	})
	input := appconfig.DeleteEnvironmentInput{
		ApplicationId: fwflex.StringFromFramework(ctx, data.ApplicationID),
		EnvironmentId: fwflex.StringFromFramework(ctx, data.EnvironmentID),
	}
	_, err := conn.DeleteEnvironment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting AppConfig Environment (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

const environmentResourceIDSeparator = ":"

func environmentCreateResourceID(environmentID, applicationID string) string {
	parts := []string{environmentID, applicationID}
	id := strings.Join(parts, environmentResourceIDSeparator)

	return id
}

func environmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, environmentResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EnvironmentID%[2]sApplicationID", id, environmentResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findEnvironmentByTwoPartKey(ctx context.Context, conn *appconfig.Client, applicationID, environmentID string) (*appconfig.GetEnvironmentOutput, error) {
	input := appconfig.GetEnvironmentInput{
		ApplicationId: aws.String(applicationID),
		EnvironmentId: aws.String(environmentID),
	}

	return findEnvironment(ctx, conn, &input)
}

func findEnvironment(ctx context.Context, conn *appconfig.Client, input *appconfig.GetEnvironmentInput) (*appconfig.GetEnvironmentOutput, error) {
	output, err := conn.GetEnvironment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type environmentResourceModel struct {
	ApplicationID types.String                                 `tfsdk:"application_id"`
	ARN           types.String                                 `tfsdk:"arn"`
	Description   types.String                                 `tfsdk:"description"`
	EnvironmentID types.String                                 `tfsdk:"environment_id"`
	ID            types.String                                 `tfsdk:"id"`
	Monitors      fwtypes.SetNestedObjectValueOf[monitorModel] `tfsdk:"monitor"`
	Name          types.String                                 `tfsdk:"name"`
	State         types.String                                 `tfsdk:"state"`
	Tags          tftags.Map                                   `tfsdk:"tags"`
	TagsAll       tftags.Map                                   `tfsdk:"tags_all"`
}

func (model *environmentResourceModel) InitFromID(ctx context.Context) error {
	environmentID, applicationID, err := environmentParseResourceID(model.ID.ValueString())

	if err != nil {
		return err
	}

	model.ApplicationID = fwflex.StringValueToFramework(ctx, applicationID)
	model.EnvironmentID = fwflex.StringValueToFramework(ctx, environmentID)

	return nil
}

func (model *environmentResourceModel) setID(ctx context.Context) {
	model.ID = fwflex.StringValueToFramework(ctx, environmentCreateResourceID(model.EnvironmentID.ValueString(), model.ApplicationID.ValueString()))
}

type monitorModel struct {
	AlarmARN     fwtypes.ARN `tfsdk:"alarm_arn"`
	AlarmRoleARN fwtypes.ARN `tfsdk:"alarm_role_arn"`
}

func environmentARN(ctx context.Context, c *conns.AWSClient, applicationID, environmentID string) string {
	return c.RegionalARN(ctx, "appconfig", "application/"+applicationID+"/environment/"+environmentID)
}
