// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package appsync

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_appsync_channel_namespace", name="Channel Namespace")
// @Tags(identifierAttribute="channel_namespace_arn")
// @Testing(importStateIdAttribute="name")
// @Testing(importStateIdFunc=testAccChannelNamespaceImportStateID)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appsync/types;awstypes;awstypes.ChannelNamespace")
func newChannelNamespaceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &channelNamespaceResource{}

	return r, nil
}

type channelNamespaceResource struct {
	framework.ResourceWithModel[channelNamespaceResourceModel]
}

func (r *channelNamespaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"channel_namespace_arn": framework.ARNAttributeComputedOnly(),
			"code_handlers": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32768),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(regexache.MustCompile(`^([A-Za-z0-9](?:[A-Za-z0-9\-]{0,48}[A-Za-z0-9])?)$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"handler_configs": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[handlerConfigsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"on_publish":   handlerConfigBlock(ctx),
						"on_subscribe": handlerConfigBlock(ctx),
					},
				},
			},
			"publish_auth_mode": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auth_type": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
						},
					},
				},
			},
			"subscribe_auth_mode": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auth_type": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
						},
					},
				},
			},
		},
	}
}

func handlerConfigBlock(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[handlerConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"behavior": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.HandlerBehavior](),
				},
			},
			Blocks: map[string]schema.Block{
				"integration": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[integrationModel](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"data_source_name": schema.StringAttribute{
								Required: true,
							},
						},
						Blocks: map[string]schema.Block{
							"lambda_config": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaConfigModel](ctx),
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"invoke_type": schema.StringAttribute{
											CustomType: fwtypes.StringEnumType[awstypes.InvokeType](),
											Optional:   true,
										},
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

func (r *channelNamespaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data channelNamespaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input appsync.CreateChannelNamespaceInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateChannelNamespace(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set values for unknowns.
	data.ChannelNamespaceARN = fwflex.StringToFramework(ctx, output.ChannelNamespace.ChannelNamespaceArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data), smerr.ID, name)
}

func (r *channelNamespaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data channelNamespaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	apiID, name := fwflex.StringValueFromFramework(ctx, data.ApiID), fwflex.StringValueFromFramework(ctx, data.Name)
	output, err := findChannelNamespaceByTwoPartKey(ctx, conn, apiID, name)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	if inttypes.IsZero(output.HandlerConfigs) {
		output.HandlerConfigs = nil
	}

	// Set attributes for import.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data), smerr.ID, name)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, name)
}

func (r *channelNamespaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old channelNamespaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, new.Name)
	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d, smerr.ID, name)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input appsync.UpdateChannelNamespaceInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input), smerr.ID, name)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateChannelNamespace(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new), smerr.ID, name)
}

func (r *channelNamespaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data channelNamespaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	apiID, name := fwflex.StringValueFromFramework(ctx, data.ApiID), fwflex.StringValueFromFramework(ctx, data.Name)
	input := appsync.DeleteChannelNamespaceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	}
	_, err := conn.DeleteChannelNamespace(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *channelNamespaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		channelNamespaceIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, channelNamespaceIDParts, true)

	if err != nil {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("api_id"), parts[0]))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrName), parts[1]))
}

func findChannelNamespaceByTwoPartKey(ctx context.Context, conn *appsync.Client, apiID, name string) (*awstypes.ChannelNamespace, error) {
	input := appsync.GetChannelNamespaceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	}

	return findChannelNamespace(ctx, conn, &input)
}

func findChannelNamespace(ctx context.Context, conn *appsync.Client, input *appsync.GetChannelNamespaceInput) (*awstypes.ChannelNamespace, error) {
	output, err := conn.GetChannelNamespace(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.ChannelNamespace == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.ChannelNamespace, nil
}

type channelNamespaceResourceModel struct {
	framework.WithRegionModel
	ApiID               types.String                                         `tfsdk:"api_id"`
	ChannelNamespaceARN types.String                                         `tfsdk:"channel_namespace_arn"`
	CodeHandlers        types.String                                         `tfsdk:"code_handlers"`
	HandlerConfigs      fwtypes.ListNestedObjectValueOf[handlerConfigsModel] `tfsdk:"handler_configs"`
	Name                types.String                                         `tfsdk:"name"`
	PublishAuthModes    fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"publish_auth_mode"`
	SubscribeAuthModes  fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"subscribe_auth_mode"`
	Tags                tftags.Map                                           `tfsdk:"tags"`
	TagsAll             tftags.Map                                           `tfsdk:"tags_all"`
}

type handlerConfigsModel struct {
	OnPublish   fwtypes.ListNestedObjectValueOf[handlerConfigModel] `tfsdk:"on_publish"`
	OnSubscribe fwtypes.ListNestedObjectValueOf[handlerConfigModel] `tfsdk:"on_subscribe"`
}

type handlerConfigModel struct {
	Behavior    fwtypes.StringEnum[awstypes.HandlerBehavior]      `tfsdk:"behavior"`
	Integration fwtypes.ListNestedObjectValueOf[integrationModel] `tfsdk:"integration"`
}

type integrationModel struct {
	DataSourceName types.String                                       `tfsdk:"data_source_name"`
	LambdaConfig   fwtypes.ListNestedObjectValueOf[lambdaConfigModel] `tfsdk:"lambda_config"`
}

type lambdaConfigModel struct {
	InvokeType fwtypes.StringEnum[awstypes.InvokeType] `tfsdk:"invoke_type"`
}
