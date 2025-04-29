// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func streamProcessorSchema0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"stream_processor_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"data_sharing_preference": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSharingPreferenceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"opt_in": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			"input": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inputModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"kinesis_video_stream": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisVideoStreamInputModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"notification_channel": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[notificationChannelModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSNSTopicARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			"regions_of_interest": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[regionOfInterestModelV0](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"bounding_box": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[boundingBoxModel](ctx),
							Attributes: map[string]schema.Attribute{
								"height": schema.Float64Attribute{
									Optional: true,
								},
								"left": schema.Float64Attribute{
									Optional: true,
								},
								"top": schema.Float64Attribute{
									Optional: true,
								},
								"width": schema.Float64Attribute{
									Optional: true,
								},
							},
						},
						"polygon": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[polygonModel](ctx),
							NestedObject: schema.NestedBlockObject{
								CustomType: fwtypes.NewObjectTypeOf[polygonModel](ctx),
								Attributes: map[string]schema.Attribute{
									"x": schema.Float64Attribute{
										Optional: true,
									},
									"y": schema.Float64Attribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"output": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[outputModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"kinesis_data_stream": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisDataStreamModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"s3_destination": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3DestinationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucket: schema.StringAttribute{
										Optional: true,
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[settingsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"connected_home": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[connectedHomeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"labels": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Optional:   true,
									},
									"min_confidence": schema.Float64Attribute{
										Computed: true,
										Optional: true,
									},
								},
							},
						},
						"face_search": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[faceSearchModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"collection_id": schema.StringAttribute{
										Required: true,
									},
									"face_match_threshold": schema.Float64Attribute{
										Default:  float64default.StaticFloat64(faceMatchThresholdDefault),
										Optional: true,
										Computed: true,
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

type regionOfInterestModelV0 struct {
	BoundingBox fwtypes.ObjectValueOf[boundingBoxModel]       `tfsdk:"bounding_box"`
	Polygon     fwtypes.ListNestedObjectValueOf[polygonModel] `tfsdk:"polygon"`
}

type resourceStreamProcessorDataModelV0 struct {
	ARN                   types.String                                                `tfsdk:"arn"`
	DataSharingPreference fwtypes.ListNestedObjectValueOf[dataSharingPreferenceModel] `tfsdk:"data_sharing_preference"`
	Input                 fwtypes.ListNestedObjectValueOf[inputModel]                 `tfsdk:"input"`
	KmsKeyId              types.String                                                `tfsdk:"kms_key_id"`
	NotificationChannel   fwtypes.ListNestedObjectValueOf[notificationChannelModel]   `tfsdk:"notification_channel"`
	Name                  types.String                                                `tfsdk:"name"`
	Output                fwtypes.ListNestedObjectValueOf[outputModel]                `tfsdk:"output"`
	RegionsOfInterest     fwtypes.ListNestedObjectValueOf[regionOfInterestModelV0]    `tfsdk:"regions_of_interest"`
	RoleARN               fwtypes.ARN                                                 `tfsdk:"role_arn"`
	Settings              fwtypes.ListNestedObjectValueOf[settingsModel]              `tfsdk:"settings"`
	StreamProcessorARN    fwtypes.ARN                                                 `tfsdk:"stream_processor_arn"`
	Tags                  tftags.Map                                                  `tfsdk:"tags"`
	TagsAll               tftags.Map                                                  `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                              `tfsdk:"timeouts"`
}

func upgradeStreamProcessorStateV0toV1(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var streamProcessorDataV0 resourceStreamProcessorDataModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &streamProcessorDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	streamProcessorDataV1 := streamProcessorResourceModel{
		ARN:                   streamProcessorDataV0.ARN,
		DataSharingPreference: streamProcessorDataV0.DataSharingPreference,
		Input:                 streamProcessorDataV0.Input,
		KmsKeyId:              streamProcessorDataV0.KmsKeyId,
		NotificationChannel:   streamProcessorDataV0.NotificationChannel,
		Name:                  streamProcessorDataV0.Name,
		Output:                streamProcessorDataV0.Output,
		RoleARN:               streamProcessorDataV0.RoleARN,
		RegionsOfInterest:     upgradeRegionsOfInterestStateFromV0(ctx, streamProcessorDataV0.RegionsOfInterest, &response.Diagnostics),
		Settings:              streamProcessorDataV0.Settings,
		StreamProcessorARN:    streamProcessorDataV0.StreamProcessorARN,
		Tags:                  streamProcessorDataV0.Tags,
		TagsAll:               streamProcessorDataV0.TagsAll,
		Timeouts:              streamProcessorDataV0.Timeouts,
	}

	response.Diagnostics.Append(response.State.Set(ctx, streamProcessorDataV1)...)
}

func upgradeRegionsOfInterestStateFromV0(ctx context.Context, old fwtypes.ListNestedObjectValueOf[regionOfInterestModelV0], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[regionOfInterestModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[regionOfInterestModel](ctx)
	}

	var oldElems []regionOfInterestModelV0
	diags.Append(old.ElementsAs(ctx, &oldElems, false)...)

	newElems := make([]regionOfInterestModel, len(oldElems))
	for i, oldElem := range oldElems {
		newRegion := regionOfInterestModel{
			BoundingBox: upgradeBoundingBoxModelStateFromV0(ctx, oldElem.BoundingBox, diags),
			Polygon:     oldElem.Polygon,
		}
		newElems[i] = newRegion
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newElems)
	diags.Append(d...)

	return result
}

func upgradeBoundingBoxModelStateFromV0(ctx context.Context, old fwtypes.ObjectValueOf[boundingBoxModel], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[boundingBoxModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[boundingBoxModel](ctx)
	}

	var oldObj boundingBoxModel
	diags.Append(old.As(ctx, &oldObj, basetypes.ObjectAsOptions{})...)

	newObj := []boundingBoxModel{
		{
			Height: oldObj.Height,
			Left:   oldObj.Left,
			Top:    oldObj.Top,
			Width:  oldObj.Width,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newObj)
	diags.Append(d...)

	return result
}
