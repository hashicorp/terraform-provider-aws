// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rekognition_stream_processor", name="Stream Processor")
func newResourceStreamProcessor(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceStreamProcessor{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameStreamProcessor = "Stream Processor"
)

type resourceStreamProcessor struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceStreamProcessor) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_stream_processor"
}

func (r *resourceStreamProcessor) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	kmsKeyIdRegex := regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9:_/+=,@.-]$`)
	nameRegex := regexache.MustCompile(`[a-zA-Z0-9_.\-]+`)
	collectionIdRegex := regexache.MustCompile(`[a-zA-Z0-9_.\-]+`)
	s3bucketRegex := regexache.MustCompile(`[0-9A-Za-z\.\-_]*`)
	kinesisStreamArnRegex := regexache.MustCompile(`(^arn:([a-z\d-]+):kinesis:([a-z\d-]+):\d{12}:.+$)`)           // lintignore:AWSAT005
	kinesisVideoStreamArnRegex := regexache.MustCompile(`(^arn:([a-z\d-]+):kinesisvideo:([a-z\d-]+):\d{12}:.+$)`) // lintignore:AWSAT005
	snsArnRegex := regexache.MustCompile(`(^arn:aws:sns:.*:\w{12}:.+$)`)                                          // lintignore:AWSAT005
	roleArnRegex := regexache.MustCompile(`arn:aws:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+`)                     // lintignore:AWSAT005

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrKMSKeyID: schema.StringAttribute{
				Description: "The identifier for your AWS Key Management Service key (AWS KMS key). You can supply the Amazon Resource Name (ARN) of your KMS key, the ID of your KMS key, an alias for your KMS key, or an alias ARN.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(kmsKeyIdRegex, "must conform to: ^[A-Za-z0-9][A-Za-z0-9:_/+=,@.-]$"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Description: "An identifier you assign to the stream processor.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(128),
					stringvalidator.RegexMatches(nameRegex, "must conform to: [a-zA-Z0-9_.\\-]+"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				Description: "The Amazon Resource Number (ARN) of the IAM role that allows access to the stream processor.",
				// CustomType:  fwtypes.ARNType,
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(roleArnRegex, "must conform to: arn:aws:iam::\\d{12}:role/?[a-zA-Z_0-9+=,.@\\-_/]+"), // lintignore:AWSAT005
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"data_sharing_preference": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[dataSharingPreferenceModel](ctx),
				Description: "Shows whether you are sharing data with Rekognition to improve model performance.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
					objectvalidator.AlsoRequires(path.MatchRelative().AtName("opt_in")),
				},
				Attributes: map[string]schema.Attribute{
					"opt_in": schema.BoolAttribute{
						Description: "Do you want to share data with Rekognition to improve model performance.",
						Optional:    true,
					},
				},
			},
			"input": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[inputModel](ctx),
				Description: "Information about the source streaming video.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"kinesis_video_stream": schema.SingleNestedBlock{
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						CustomType:  fwtypes.NewObjectTypeOf[kinesisVideoStreamInputModel](ctx),
						Description: "Kinesis video stream stream that provides the source streaming video for a Amazon Rekognition Video stream processor.",
						Attributes: map[string]schema.Attribute{
							names.AttrARN: schema.StringAttribute{
								CustomType:  fwtypes.ARNType,
								Description: "ARN of the Kinesis video stream stream that streams the source video.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.All(
										stringvalidator.RegexMatches(kinesisVideoStreamArnRegex, "must conform to: (^arn:([a-z\\d-]+):kinesisvideo:([a-z\\d-]+):\\d{12}:.+$)"), // lintignore:AWSAT005
									),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"notification_channel": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[notificationChannelModel](ctx),
				Description: "The Amazon Simple Notification Service topic to which Amazon Rekognition publishes the object detection results and completion status of a video analysis operation.",
				Attributes: map[string]schema.Attribute{
					names.AttrSNSTopicARN: schema.StringAttribute{
						Description: "The Amazon Resource Number (ARN) of the Amazon Amazon Simple Notification Service topic to which Amazon Rekognition posts the completion status.",
						CustomType:  fwtypes.ARNType,
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(snsArnRegex, "must conform to: (^arn:aws:sns:.*:\\w{12}:.+$)"), // lintignore:AWSAT005
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"regions_of_interest": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[regionOfInterestModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						objectvalidator.AtLeastOneOf(path.MatchRelative().AtName("bounding_box"), path.MatchRelative().AtName("polygon")),
					},
					Blocks: map[string]schema.Block{
						"bounding_box": schema.SingleNestedBlock{
							CustomType:  fwtypes.NewObjectTypeOf[boundingBoxModel](ctx),
							Description: "The box representing a region of interest on screen.",
							Validators: []validator.Object{
								objectvalidator.AlsoRequires(
									path.MatchRelative().AtName("height"),
									path.MatchRelative().AtName("left"),
									path.MatchRelative().AtName("top"),
									path.MatchRelative().AtName("width"),
								),
								objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("polygon")),
							},
							Attributes: map[string]schema.Attribute{
								"height": schema.Float64Attribute{
									Optional:    true,
									Description: "Height of the bounding box as a ratio of the overall image height.",
									Validators: []validator.Float64{
										float64validator.Between(0.0, 1.0),
									},
								},
								"left": schema.Float64Attribute{
									Description: "Left coordinate of the bounding box as a ratio of overall image width.",
									Optional:    true,
									Validators: []validator.Float64{
										float64validator.Between(0.0, 1.0),
									},
								},
								"top": schema.Float64Attribute{
									Description: "Top coordinate of the bounding box as a ratio of overall image height.",
									Optional:    true,
									Validators: []validator.Float64{
										float64validator.Between(0.0, 1.0),
									},
								},
								"width": schema.Float64Attribute{
									Description: "Width of the bounding box as a ratio of the overall image width.",
									Optional:    true,
									Validators: []validator.Float64{
										float64validator.Between(0.0, 1.0),
									},
								},
							},
						},
						"polygon": schema.ListNestedBlock{
							CustomType:  fwtypes.NewListNestedObjectTypeOf[polygonModel](ctx),
							Description: "Specifies a shape made of 3 to 10 Point objects that define a region of interest.",
							Validators: []validator.List{
								listvalidator.SizeBetween(3, 10),
								listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("bounding_box")),
							},
							NestedObject: schema.NestedBlockObject{
								CustomType: fwtypes.NewObjectTypeOf[polygonModel](ctx),
								Validators: []validator.Object{
									objectvalidator.AlsoRequires(
										path.MatchRelative().AtName("x"),
										path.MatchRelative().AtName("y"),
									)},
								Attributes: map[string]schema.Attribute{
									"x": schema.Float64Attribute{
										Description: "The value of the X coordinate for a point on a Polygon.",
										Optional:    true,
										Validators: []validator.Float64{
											float64validator.Between(0.0, 1.0),
										},
									},
									"y": schema.Float64Attribute{
										Description: "The value of the Y coordinate for a point on a Polygon.",
										Optional:    true,
										Validators: []validator.Float64{
											float64validator.Between(0.0, 1.0),
										},
									},
								},
							},
						},
					},
				},
			},
			"output": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[outputModel](ctx),
				Description: "Kinesis data stream stream or Amazon S3 bucket location to which Amazon Rekognition Video puts the analysis results.",
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(
						path.MatchRelative().AtName("kinesis_data_stream"),
						path.MatchRelative().AtName("s3_destination"),
					),
				},
				Blocks: map[string]schema.Block{
					"kinesis_data_stream": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[kinesisDataStreamModel](ctx),
						Description: "The Amazon Kinesis Data Streams stream to which the Amazon Rekognition stream processor streams the analysis results.",
						Validators: []validator.Object{
							objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("s3_destination")),
						},
						Attributes: map[string]schema.Attribute{
							names.AttrARN: schema.StringAttribute{
								CustomType:  fwtypes.ARNType,
								Description: "ARN of the output Amazon Kinesis Data Streams stream.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(kinesisStreamArnRegex, "must conform to: (^arn:([a-z\\d-]+):kinesis:([a-z\\d-]+):\\d{12}:.+$)"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"s3_destination": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[s3DestinationModel](ctx),
						Description: "The Amazon S3 bucket location to which Amazon Rekognition publishes the detailed inference results of a video analysis operation.",
						Validators: []validator.Object{
							objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("kinesis_data_stream")),
						},
						Attributes: map[string]schema.Attribute{
							names.AttrBucket: schema.StringAttribute{
								Description: "The name of the Amazon S3 bucket you want to associate with the streaming video project.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(3, 255),
									stringvalidator.RegexMatches(s3bucketRegex, "must conform to: [0-9A-Za-z\\.\\-_]*"),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"key_prefix": schema.StringAttribute{
								Description: "The prefix value of the location within the bucket that you want the information to be published to.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1024),
								},
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"settings": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[settingsModel](ctx),
				Description: "Input parameters used in a streaming video analyzed by a stream processor.",
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(
						path.MatchRelative().AtName("connected_home"),
						path.MatchRelative().AtName("face_search"),
					),
				},
				Blocks: map[string]schema.Block{
					"connected_home": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[connectedHomeModel](ctx),
						Description: "Label detection settings to use on a streaming video.",
						Validators: []validator.Object{
							objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("face_search")),
						},
						Attributes: map[string]schema.Attribute{
							"labels": schema.ListAttribute{
								Description: "Specifies what you want to detect in the video, such as people, packages, or pets.",
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
									listvalidator.ValueStringsAre(stringvalidator.OneOf(connectedHomeLabels()...)),
								},
							},
							"min_confidence": schema.Float64Attribute{
								Description: "The minimum confidence required to label an object in the video.",
								Validators: []validator.Float64{
									float64validator.Between(0.0, 100.0), //nolint:mnd
								},
								Default:  float64default.StaticFloat64(50), //nolint:mnd
								Computed: true,
								Optional: true,
							},
						},
					},
					"face_search": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[faceSearchModel](ctx),
						Description: "Face search settings to use on a streaming video.",
						Validators: []validator.Object{
							objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("connected_home")),
							objectvalidator.AlsoRequires(
								path.MatchRelative().AtName("collection_id"),
							),
						},
						Attributes: map[string]schema.Attribute{
							"collection_id": schema.StringAttribute{
								Description: "The ID of a collection that contains faces that you want to search for.",
								Validators: []validator.String{
									stringvalidator.LengthAtMost(2048),
									stringvalidator.RegexMatches(collectionIdRegex, "must conform to: [a-zA-Z0-9_.\\-]+"),
								},
								Optional: true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"face_match_threshold": schema.Float64Attribute{
								Description: "Minimum face match confidence score that must be met to return a result for a recognized face.",
								Validators: []validator.Float64{
									float64validator.Between(0.0, 100.0), //nolint:mnd
								},
								PlanModifiers: []planmodifier.Float64{
									float64planmodifier.RequiresReplace(),
								},
								Default:  float64default.StaticFloat64(80), //nolint:mnd
								Computed: true,
								Optional: true,
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

func (r *resourceStreamProcessor) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan resourceStreamProcessorDataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.CreateStreamProcessorInput{}
	in.Tags = getTagsIn(ctx)

	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateStreamProcessor(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.StreamProcessorArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameStreamProcessor, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = fwflex.StringToFramework(ctx, out.StreamProcessorArn)
	plan.ID = plan.Name

	if plan.DataSharingPreference.IsNull() {
		dataSharing, diag := fwtypes.NewObjectValueOf(ctx, &dataSharingPreferenceModel{OptIn: basetypes.NewBoolValue(false)})
		resp.Diagnostics.Append(diag...)
		plan.DataSharingPreference = dataSharing
		resp.Diagnostics.Append(req.Plan.Set(ctx, &plan)...)
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitStreamProcessorCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForCreation, ResNameStreamProcessor, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, created, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStreamProcessor) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceStreamProcessorDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findStreamProcessorByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceStreamProcessor) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan, state resourceStreamProcessorDataModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DataSharingPreference.Equal(state.DataSharingPreference) ||
		!plan.Settings.Equal(state.Settings) ||
		!plan.RegionsOfInterest.Equal(state.RegionsOfInterest) {
		in := &rekognition.UpdateStreamProcessorInput{
			Name:               plan.Name.ValueStringPointer(),
			ParametersToDelete: []awstypes.StreamProcessorParameterToDelete{},
		}

		if !plan.DataSharingPreference.Equal(state.DataSharingPreference) {
			dspPlan, dspState := unwrapObjectValueOf(ctx, resp.Diagnostics, plan.DataSharingPreference, state.DataSharingPreference)
			if resp.Diagnostics.HasError() {
				return
			}

			if !dspPlan.OptIn.Equal(dspState.OptIn) {
				in.DataSharingPreferenceForUpdate = &awstypes.StreamProcessorDataSharingPreference{
					OptIn: dspPlan.OptIn.ValueBool(),
				}
			}
		}

		if !plan.Settings.Equal(state.Settings) {
			in.SettingsForUpdate = &awstypes.StreamProcessorSettingsForUpdate{
				ConnectedHomeForUpdate: &awstypes.ConnectedHomeSettingsForUpdate{},
			}

			settingsPlan, settingsState := unwrapObjectValueOf(ctx, resp.Diagnostics, plan.Settings, state.Settings)
			if resp.Diagnostics.HasError() {
				return
			}

			connectedHomePlan, connectedHomeState := unwrapObjectValueOf(ctx, resp.Diagnostics, settingsPlan.ConnectedHome, settingsState.ConnectedHome)
			if resp.Diagnostics.HasError() {
				return
			}

			if !connectedHomePlan.MinConfidence.Equal(connectedHomeState.MinConfidence) { // nosemgrep:ci.semgrep.migrate.aws-api-context
				if !connectedHomePlan.MinConfidence.IsNull() && connectedHomeState.MinConfidence.IsNull() { // nosemgrep:ci.semgrep.migrate.aws-api-context
					in.ParametersToDelete = append(in.ParametersToDelete, awstypes.StreamProcessorParameterToDeleteConnectedHomeMinConfidence)
				}

				if !connectedHomePlan.MinConfidence.IsNull() { // nosemgrep:ci.semgrep.migrate.aws-api-context
					in.SettingsForUpdate.ConnectedHomeForUpdate.MinConfidence = aws.Float32(float32(connectedHomePlan.MinConfidence.ValueFloat64())) // nosemgrep:ci.semgrep.migrate.aws-api-context
				}
			}

			if !connectedHomePlan.Labels.Equal(connectedHomeState.Labels) { // nosemgrep:ci.semgrep.migrate.aws-api-context
				in.SettingsForUpdate.ConnectedHomeForUpdate.Labels = fwflex.ExpandFrameworkStringValueList(ctx, connectedHomePlan.Labels)
			}
		}

		if plan.RegionsOfInterest.IsNull() && !state.RegionsOfInterest.IsNull() {
			in.ParametersToDelete = append(in.ParametersToDelete, awstypes.StreamProcessorParameterToDeleteRegionsOfInterest)
		}

		if !plan.RegionsOfInterest.Equal(state.RegionsOfInterest) {
			planRegions, diags := plan.RegionsOfInterest.ToSlice(ctx)
			resp.Diagnostics.Append(diags...)

			plannedRegions := make([]awstypes.RegionOfInterest, len(planRegions))

			for i := 0; i < len(planRegions); i++ {
				planRegion := planRegions[i]
				plannedRegions[i] = awstypes.RegionOfInterest{}

				if !planRegion.BoundingBox.IsNull() {
					boundingBox, diags := planRegion.BoundingBox.ToPtr(ctx)
					resp.Diagnostics.Append(diags...)

					plannedRegions[i].BoundingBox = &awstypes.BoundingBox{
						Top:    aws.Float32(float32(boundingBox.Top.ValueFloat64())),
						Left:   aws.Float32(float32(boundingBox.Left.ValueFloat64())),
						Height: aws.Float32(float32(boundingBox.Height.ValueFloat64())),
						Width:  aws.Float32(float32(boundingBox.Width.ValueFloat64())),
					}
				}

				if !planRegion.Polygon.IsNull() {
					polygons, diags := planRegion.Polygon.ToSlice(ctx)
					resp.Diagnostics.Append(diags...)

					plannedPolygons := make([]awstypes.Point, len(polygons))

					for i := 0; i < len(polygons); i++ {
						polygon := polygons[i]
						plannedPolygons[i] = awstypes.Point{
							X: aws.Float32(float32(polygon.X.ValueFloat64())),
							Y: aws.Float32(float32(polygon.Y.ValueFloat64())),
						}
					}
					plannedRegions[i].Polygon = plannedPolygons
				}
			}
			in.RegionsOfInterestForUpdate = plannedRegions
		}

		_, err := conn.UpdateStreamProcessor(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Rekognition, create.ErrActionUpdating, ResNameStreamProcessor, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	updated, err := waitStreamProcessorUpdated(ctx, conn, plan.Name.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForUpdate, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, updated, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceStreamProcessor) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceStreamProcessorDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &rekognition.DeleteStreamProcessorInput{
		Name: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteStreamProcessor(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameStreamProcessor, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitStreamProcessorDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForDeletion, ResNameStreamProcessor, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceStreamProcessor) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceStreamProcessor) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitStreamProcessorCreated(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.StreamProcessorStatusStopped),
		Refresh:                   statusStreamProcessor(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rekognition.DescribeStreamProcessorOutput); ok {
		return out, err
	}

	return nil, err
}

func waitStreamProcessorUpdated(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StreamProcessorStatusUpdating),
		Target:                    enum.Slice(awstypes.StreamProcessorStatusStopped),
		Refresh:                   statusStreamProcessor(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rekognition.DescribeStreamProcessorOutput); ok {
		return out, err
	}

	return nil, err
}

func waitStreamProcessorDeleted(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.StreamProcessorStatusStopped,
			awstypes.StreamProcessorStatusStarting,
			awstypes.StreamProcessorStatusRunning,
			awstypes.StreamProcessorStatusFailed,
			awstypes.StreamProcessorStatusStopping,
			awstypes.StreamProcessorStatusUpdating,
		),
		Target:  []string{},
		Refresh: statusStreamProcessor(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rekognition.DescribeStreamProcessorOutput); ok {
		return out, err
	}

	return nil, err
}

func statusStreamProcessor(ctx context.Context, conn *rekognition.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findStreamProcessorByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findStreamProcessorByID(ctx context.Context, conn *rekognition.Client, name string) (*rekognition.DescribeStreamProcessorOutput, error) {
	in := &rekognition.DescribeStreamProcessorInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribeStreamProcessor(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func unwrapObjectValueOf[T any](ctx context.Context, diagnostics diag.Diagnostics, plan fwtypes.ObjectValueOf[T], state fwtypes.ObjectValueOf[T]) (*T, *T) {
	ptrPlan, diags := plan.ToPtr(ctx)
	diagnostics.Append(diags...)

	ptrState, diags := state.ToPtr(ctx)
	diagnostics.Append(diags...)

	return ptrPlan, ptrState
}

type resourceStreamProcessorDataModel struct {
	ARN                   types.String                                           `tfsdk:"arn"`
	DataSharingPreference fwtypes.ObjectValueOf[dataSharingPreferenceModel]      `tfsdk:"data_sharing_preference"`
	ID                    types.String                                           `tfsdk:"id"`
	Input                 fwtypes.ObjectValueOf[inputModel]                      `tfsdk:"input"`
	KmsKeyId              types.String                                           `tfsdk:"kms_key_id"`
	NotificationChannel   fwtypes.ObjectValueOf[notificationChannelModel]        `tfsdk:"notification_channel"`
	Name                  types.String                                           `tfsdk:"name"`
	Output                fwtypes.ObjectValueOf[outputModel]                     `tfsdk:"output"`
	RegionsOfInterest     fwtypes.ListNestedObjectValueOf[regionOfInterestModel] `tfsdk:"regions_of_interest"`
	RoleARN               types.String                                           `tfsdk:"role_arn"` //TODO ARN types?
	Settings              fwtypes.ObjectValueOf[settingsModel]                   `tfsdk:"settings"`
	Tags                  types.Map                                              `tfsdk:"tags"`
	TagsAll               types.Map                                              `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                         `tfsdk:"timeouts"`
}

type dataSharingPreferenceModel struct {
	OptIn types.Bool `tfsdk:"opt_in"`
}

type inputModel struct {
	KinesisVideoStream fwtypes.ObjectValueOf[kinesisVideoStreamInputModel] `tfsdk:"kinesis_video_stream"`
}

type kinesisVideoStreamInputModel struct {
	ARN types.String `tfsdk:"arn"`
}

type notificationChannelModel struct {
	SNSTopicArn fwtypes.ARN `tfsdk:"sns_topic_arn"`
}

type outputModel struct {
	KinesisDataStream fwtypes.ObjectValueOf[kinesisDataStreamModel] `tfsdk:"kinesis_data_stream"`
	S3Destination     fwtypes.ObjectValueOf[s3DestinationModel]     `tfsdk:"s3_destination"`
}

type kinesisDataStreamModel struct {
	ARN types.String `tfsdk:"arn"`
}

type s3DestinationModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	KeyPrefix types.String `tfsdk:"key_prefix"`
}

type regionOfInterestModel struct {
	BoundingBox fwtypes.ObjectValueOf[boundingBoxModel]       `tfsdk:"bounding_box"`
	Polygon     fwtypes.ListNestedObjectValueOf[polygonModel] `tfsdk:"polygon"`
}

type boundingBoxModel struct {
	Height types.Float64 `tfsdk:"height"`
	Left   types.Float64 `tfsdk:"left"`
	Top    types.Float64 `tfsdk:"top"`
	Width  types.Float64 `tfsdk:"width"`
}

type polygonModel struct {
	X types.Float64 `tfsdk:"x"`
	Y types.Float64 `tfsdk:"y"`
}

type settingsModel struct {
	ConnectedHome fwtypes.ObjectValueOf[connectedHomeModel] `tfsdk:"connected_home"`
	FaceSearch    fwtypes.ObjectValueOf[faceSearchModel]    `tfsdk:"face_search"`
}

type connectedHomeModel struct {
	Labels        fwtypes.ListValueOf[types.String] `tfsdk:"labels"`
	MinConfidence types.Float64                     `tfsdk:"min_confidence"`
}

type faceSearchModel struct {
	CollectionId       types.String  `tfsdk:"collection_id"`
	FaceMatchThreshold types.Float64 `tfsdk:"face_match_threshold"`
}

const (
	person_label  = "PERSON"
	pet_label     = "PET"
	package_label = "PACKAGE"
	all_label     = "ALL"
)

/*
- AWS SDK doesn't have a CreateStreamProcessorInput.StreamProcessorSettings.ConnectedHomeSettings.Labels enum available as of 5/13/24

- see docs https://docs.aws.amazon.com/rekognition/latest/APIReference/API_ConnectedHomeSettings.html#API_ConnectedHomeSettings_Contents
*/
func connectedHomeLabels() []string {
	return []string{
		person_label,
		pet_label,
		package_label,
		all_label,
	}
}
