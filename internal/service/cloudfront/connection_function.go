// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @FrameworkResource("aws_cloudfront_connection_function", name="Connection Function")
// @Tags(identifierAttribute="arn")
func newResourceConnectionFunction(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceConnectionFunction{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameConnectionFunction = "Connection Function"
)

type resourceConnectionFunction struct {
	framework.ResourceWithModel[resourceConnectionFunctionModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceConnectionFunction) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"code": schema.StringAttribute{
				Required: true,
			},
			names.AttrComment: schema.StringAttribute{
				Optional: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"live_stage_etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"publish": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"runtime": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.FunctionRuntime](),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"key_value_store_associations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueStoreAssociationsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"items": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[keyValueStoreAssociationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"key_value_store_arn": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceConnectionFunction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceConnectionFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := cloudfront.CreateConnectionFunctionInput{
		Name:                   fwflex.StringFromFramework(ctx, data.Name),
		ConnectionFunctionCode: []byte(data.Code.ValueString()),
		ConnectionFunctionConfig: &awstypes.FunctionConfig{
			Comment: fwflex.StringFromFramework(ctx, data.Comment),
			Runtime: awstypes.FunctionRuntime(data.Runtime.ValueString()),
		},
	}

	// Add key value store associations if provided
	if !data.KeyValueStoreAssociations.IsNull() {
		kvStoreAssocList, d := data.KeyValueStoreAssociations.ToSlice(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(kvStoreAssocList) > 0 {
			kvStoreAssoc := kvStoreAssocList[0]
			if !kvStoreAssoc.Items.IsNull() {
				itemsList, d := kvStoreAssoc.Items.ToSlice(ctx)
				smerr.AddEnrich(ctx, &resp.Diagnostics, d)
				if resp.Diagnostics.HasError() {
					return
				}

				if len(itemsList) > 0 {
					var associations []awstypes.KeyValueStoreAssociation
					for _, item := range itemsList {
						associations = append(associations, awstypes.KeyValueStoreAssociation{
							KeyValueStoreARN: item.KeyValueStoreARN.ValueStringPointer(),
						})
					}
					input.ConnectionFunctionConfig.KeyValueStoreAssociations = &awstypes.KeyValueStoreAssociations{
						Quantity: aws.Int32(int32(len(associations))),
						Items:    associations,
					}
				}
			}
		}
	}

	// Additional fields.
	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	out, err := conn.CreateConnectionFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, out.ConnectionFunctionSummary.Id)
	data.Etag = fwflex.StringToFramework(ctx, out.ETag)

	// Publish the function if requested
	if data.Publish.ValueBool() {
		publishInput := &cloudfront.PublishConnectionFunctionInput{
			Id:      data.ID.ValueStringPointer(),
			IfMatch: data.Etag.ValueStringPointer(),
		}

		_, err := conn.PublishConnectionFunction(ctx, publishInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
			return
		}
	}

	// Read the function to get the current state including live_stage_etag
	err = r.readConnectionFunction(ctx, conn, data.ID.ValueString(), &data)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

// readConnectionFunction reads the connection function from both development and live stages
// and populates the provided model with the complete state information.
func (r *resourceConnectionFunction) readConnectionFunction(ctx context.Context, conn *cloudfront.Client, id string, model *resourceConnectionFunctionModel) error {
	// Read DEVELOPMENT stage first to get most attributes
	outputDF, err := findConnectionFunctionByTwoPartKey(ctx, conn, id, awstypes.FunctionStageDevelopment)
	if err != nil {
		return err
	}

	model.Etag = fwflex.StringToFramework(ctx, outputDF.ETag)

	// Populate the resource model from DescribeConnectionFunctionOutput
	populateConnectionFunctionModel(ctx, outputDF.ConnectionFunctionSummary, outputDF.ConnectionFunctionSummary.ConnectionFunctionConfig, model)

	// Use GetConnectionFunction to get the code from development stage
	inputGF := cloudfront.GetConnectionFunctionInput{
		Identifier: aws.String(id),
		Stage:      awstypes.FunctionStageDevelopment,
	}
	outputGF, err := conn.GetConnectionFunction(ctx, &inputGF)
	if err != nil {
		return err
	}

	model.Code = types.StringValue(string(outputGF.ConnectionFunctionCode))

	// Read LIVE stage to get live_stage_etag
	outputLive, err := findConnectionFunctionByTwoPartKey(ctx, conn, id, awstypes.FunctionStageLive)
	if retry.NotFound(err) {
		model.LiveStageEtag = types.StringValue("")
	} else if err != nil {
		return err
	} else {
		model.LiveStageEtag = fwflex.StringToFramework(ctx, outputLive.ETag)
	}

	return nil
}

func (r *resourceConnectionFunction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state resourceConnectionFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.readConnectionFunction(ctx, conn, state.ID.ValueString(), &state)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceConnectionFunction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var plan, state resourceConnectionFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudfront.UpdateConnectionFunctionInput{
		Id:                     fwflex.StringFromFramework(ctx, state.ID),
		IfMatch:                fwflex.StringFromFramework(ctx, state.Etag),
		ConnectionFunctionCode: []byte(plan.Code.ValueString()),
		ConnectionFunctionConfig: &awstypes.FunctionConfig{
			Comment: fwflex.StringFromFramework(ctx, plan.Comment),
			Runtime: awstypes.FunctionRuntime(plan.Runtime.ValueString()),
		},
	}

	// Add key value store associations if provided
	if !plan.KeyValueStoreAssociations.IsNull() {
		kvStoreAssocList, d := plan.KeyValueStoreAssociations.ToSlice(ctx)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(kvStoreAssocList) > 0 {
			kvStoreAssoc := kvStoreAssocList[0]
			if !kvStoreAssoc.Items.IsNull() {
				itemsList, d := kvStoreAssoc.Items.ToSlice(ctx)
				smerr.AddEnrich(ctx, &resp.Diagnostics, d)
				if resp.Diagnostics.HasError() {
					return
				}

				if len(itemsList) > 0 {
					var associations []awstypes.KeyValueStoreAssociation
					for _, item := range itemsList {
						associations = append(associations, awstypes.KeyValueStoreAssociation{
							KeyValueStoreARN: item.KeyValueStoreARN.ValueStringPointer(),
						})
					}
					input.ConnectionFunctionConfig.KeyValueStoreAssociations = &awstypes.KeyValueStoreAssociations{
						Quantity: aws.Int32(int32(len(associations))),
						Items:    associations,
					}
				}
			}
		}
	}

	out, err := conn.UpdateConnectionFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}
	if out == nil || out.ConnectionFunctionSummary == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
		return
	}

	plan.Etag = fwflex.StringToFramework(ctx, out.ETag)

	// Publish the function if requested
	if plan.Publish.ValueBool() {
		publishInput := &cloudfront.PublishConnectionFunctionInput{
			Id:      plan.ID.ValueStringPointer(),
			IfMatch: plan.Etag.ValueStringPointer(),
		}

		_, err := conn.PublishConnectionFunction(ctx, publishInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	// Read the function to get the current state including live_stage_etag
	err = r.readConnectionFunction(ctx, conn, plan.ID.ValueString(), &plan)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceConnectionFunction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudFrontClient(ctx)

	var state resourceConnectionFunctionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudfront.DeleteConnectionFunctionInput{
		Id:      state.ID.ValueStringPointer(),
		IfMatch: state.Etag.ValueStringPointer(),
	}

	_, err := conn.DeleteConnectionFunction(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFound](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func findConnectionFunctionByTwoPartKey(ctx context.Context, conn *cloudfront.Client, id string, stage awstypes.FunctionStage) (*cloudfront.DescribeConnectionFunctionOutput, error) {
	input := &cloudfront.DescribeConnectionFunctionInput{
		Identifier: aws.String(id),
		Stage:      stage,
	}

	output, err := conn.DescribeConnectionFunction(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFound](err) {
			return nil, smarterr.NewError(&sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if output == nil || output.ConnectionFunctionSummary == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(input))
	}

	return output, nil
}

type resourceConnectionFunctionModel struct {
	ARN                       types.String                                                    `tfsdk:"arn"`
	Code                      types.String                                                    `tfsdk:"code"`
	Comment                   types.String                                                    `tfsdk:"comment"`
	Etag                      types.String                                                    `tfsdk:"etag"`
	ID                        types.String                                                    `tfsdk:"id"`
	KeyValueStoreAssociations fwtypes.ListNestedObjectValueOf[keyValueStoreAssociationsModel] `tfsdk:"key_value_store_associations"`
	LiveStageEtag             types.String                                                    `tfsdk:"live_stage_etag"`
	Name                      types.String                                                    `tfsdk:"name"`
	Publish                   types.Bool                                                      `tfsdk:"publish"`
	Runtime                   types.String                                                    `tfsdk:"runtime"`
	Status                    types.String                                                    `tfsdk:"status"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                  `tfsdk:"timeouts"`
}

type keyValueStoreAssociationsModel struct {
	Items fwtypes.ListNestedObjectValueOf[keyValueStoreAssociationModel] `tfsdk:"items"`
}

type keyValueStoreAssociationModel struct {
	KeyValueStoreARN types.String `tfsdk:"key_value_store_arn"`
}

// populateConnectionFunctionModel populates the resource model from ConnectionFunctionSummary and FunctionConfig
func populateConnectionFunctionModel(ctx context.Context, summary *awstypes.ConnectionFunctionSummary, config *awstypes.FunctionConfig, model *resourceConnectionFunctionModel) {
	if summary != nil {
		model.ARN = fwflex.StringToFramework(ctx, summary.ConnectionFunctionArn)
		model.ID = fwflex.StringToFramework(ctx, summary.Id)
		model.Name = fwflex.StringToFramework(ctx, summary.Name)
		model.Status = fwflex.StringToFramework(ctx, summary.Status)
	}

	if config != nil {
		model.Comment = fwflex.StringToFramework(ctx, config.Comment)
		model.Runtime = fwflex.StringValueToFramework(ctx, string(config.Runtime))

		// Handle KeyValueStoreAssociations
		if config.KeyValueStoreAssociations != nil && len(config.KeyValueStoreAssociations.Items) > 0 {
			var items []keyValueStoreAssociationModel
			for _, item := range config.KeyValueStoreAssociations.Items {
				items = append(items, keyValueStoreAssociationModel{
					KeyValueStoreARN: fwflex.StringToFramework(ctx, item.KeyValueStoreARN),
				})
			}

			itemsValue, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, items)
			if diags.HasError() {
				return
			}
			kvStoreAssoc := keyValueStoreAssociationsModel{
				Items: itemsValue,
			}
			kvStoreAssocValue, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []keyValueStoreAssociationsModel{kvStoreAssoc})
			if diags.HasError() {
				return
			}
			model.KeyValueStoreAssociations = kvStoreAssocValue
		}
	}
}
