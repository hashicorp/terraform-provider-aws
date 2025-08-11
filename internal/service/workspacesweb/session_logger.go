// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspacesweb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspacesweb_session_logger", name="Session Logger")
// @Tags(identifierAttribute="session_logger_arn")
// @Testing(tagsTest=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.SessionLogger")
// @Testing(importStateIdAttribute="session_logger_arn")
func newSessionLoggerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &sessionLoggerResource{}, nil
}

type sessionLoggerResource struct {
	framework.ResourceWithModel[sessionLoggerResourceModel]
}

func (r *sessionLoggerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_encryption_context": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"associated_portal_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_managed_key": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDisplayName: schema.StringAttribute{
				Optional: true,
			},
			"event_filter": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventFilterModel](ctx),
				Required:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"session_logger_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"log_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3LogConfigurationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket": schema.StringAttribute{
										Required: true,
									},
									"bucket_owner": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
									"folder_structure": schema.StringAttribute{
										Required: true,
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
									},
									"log_file_format": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *sessionLoggerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data sessionLoggerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	var input workspacesweb.CreateSessionLoggerInput

	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateSessionLogger(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb Session Logger", err.Error())
		return
	}

	data.SessionLoggerARN = fwflex.StringToFramework(ctx, output.SessionLoggerArn)

	// Get the session logger details to populate other fields
	sessionLogger, err := findSessionLoggerByARN(ctx, conn, data.SessionLoggerARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Session Logger (%s)", data.SessionLoggerARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, sessionLogger, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *sessionLoggerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data sessionLoggerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findSessionLoggerByARN(ctx, conn, data.SessionLoggerARN.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Session Logger (%s)", data.SessionLoggerARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *sessionLoggerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new sessionLoggerResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	if !new.DisplayName.Equal(old.DisplayName) ||
		!new.EventFilter.Equal(old.EventFilter) ||
		!new.CustomerManagedKey.Equal(old.CustomerManagedKey) ||
		!new.AdditionalEncryptionContext.Equal(old.AdditionalEncryptionContext) ||
		!new.LogConfiguration.Equal(old.LogConfiguration) {
		conn := r.Meta().WorkSpacesWebClient(ctx)

		var input workspacesweb.UpdateSessionLoggerInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateSessionLogger(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb Session Logger (%s)", old.SessionLoggerARN.ValueString()), err.Error())
			return
		}

		// Use new as base and flatten the response into it
		response.Diagnostics.Append(fwflex.Flatten(ctx, output.SessionLogger, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)

}

func (r *sessionLoggerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data sessionLoggerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteSessionLoggerInput{
		SessionLoggerArn: data.SessionLoggerARN.ValueStringPointer(),
	}
	_, err := conn.DeleteSessionLogger(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Session Logger (%s)", data.SessionLoggerARN.ValueString()), err.Error())
		return
	}
}

func (r *sessionLoggerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("session_logger_arn"), request, response)
}

func findSessionLoggerByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.SessionLogger, error) {
	input := workspacesweb.GetSessionLoggerInput{
		SessionLoggerArn: &arn,
	}
	output, err := conn.GetSessionLogger(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SessionLogger == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SessionLogger, nil
}

type sessionLoggerResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext fwtypes.MapOfString                                    `tfsdk:"additional_encryption_context"`
	AssociatedPortalARNs        fwtypes.ListOfString                                   `tfsdk:"associated_portal_arns"`
	CustomerManagedKey          fwtypes.ARN                                            `tfsdk:"customer_managed_key"`
	DisplayName                 types.String                                           `tfsdk:"display_name"`
	EventFilter                 fwtypes.ListNestedObjectValueOf[eventFilterModel]      `tfsdk:"event_filter"`
	LogConfiguration            fwtypes.ListNestedObjectValueOf[logConfigurationModel] `tfsdk:"log_configuration"`
	SessionLoggerARN            types.String                                           `tfsdk:"session_logger_arn"`
	Tags                        tftags.Map                                             `tfsdk:"tags"`
	TagsAll                     tftags.Map                                             `tfsdk:"tags_all"`
}

type logConfigurationModel struct {
	S3 fwtypes.ListNestedObjectValueOf[s3LogConfigurationModel] `tfsdk:"s3"`
}

type eventFilterModel struct {
	All     types.Object        `tfsdk:"all"`
	Include fwtypes.SetOfString `tfsdk:"include"`
}

var (
	_ fwflex.Expander  = eventFilterModel{}
	_ fwflex.Flattener = &eventFilterModel{}
)

func (m eventFilterModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	if !m.All.IsNull() {
		return &awstypes.EventFilterMemberAll{Value: awstypes.Unit{}}, nil
	}
	if !m.Include.IsNull() {
		var events []awstypes.Event
		for _, event := range m.Include.Elements() {
			events = append(events, awstypes.Event(event.(types.String).ValueString()))
		}
		return &awstypes.EventFilterMemberInclude{Value: events}, nil
	}
	return nil, nil
}

func (m *eventFilterModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	switch val := v.(type) {
	case awstypes.EventFilterMemberAll:
		m.All = types.ObjectValueMust(map[string]attr.Type{}, map[string]attr.Value{})
		m.Include = fwtypes.NewSetValueOfNull[types.String](ctx)
	case awstypes.EventFilterMemberInclude:
		m.All = types.ObjectNull(map[string]attr.Type{})
		var events []string
		for _, event := range val.Value {
			events = append(events, string(event))
		}
		var elements []attr.Value
		for _, event := range events {
			elements = append(elements, types.StringValue(event))
		}
		m.Include, _ = fwtypes.NewSetValueOf[types.String](ctx, elements)
	}
	return nil
}

type s3LogConfigurationModel struct {
	Bucket          types.String `tfsdk:"bucket"`
	BucketOwner     types.String `tfsdk:"bucket_owner"`
	FolderStructure types.String `tfsdk:"folder_structure"`
	KeyPrefix       types.String `tfsdk:"key_prefix"`
	LogFileFormat   types.String `tfsdk:"log_file_format"`
}
