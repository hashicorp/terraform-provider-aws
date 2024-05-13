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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
}

func (r *resourceStreamProcessor) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_stream_processor"
}

func (r *resourceStreamProcessor) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	kmsKeyIdRegex := regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9:_/+=,@.-]{0,2048}$`)
	nameRegex := regexache.MustCompile(`[a-zA-Z0-9_.\-]+`)
	collectionIdRegex := regexache.MustCompile(`[a-zA-Z0-9_.\-]+`)
	kinesisStreamArnRegex := regexache.MustCompile(`(^arn:([a-z\d-]+):kinesis:([a-z\d-]+):\d{12}:.+$)`)
	s3bucketRegex := regexache.MustCompile(`[0-9A-Za-z\.\-_]*`)
	snsArnRegex := regexache.MustCompile(`(^arn:aws:sns:.*:\w{12}:.+$)`)
	roleArnRegex := regexache.MustCompile(`arn:aws:iam::\d{12}:role/?[a-zA-Z_0-9+=,.@\-_/]+`)

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"kms_key_id": schema.StringAttribute{
				Description: "The identifier for your AWS Key Management Service key (AWS KMS key). You can supply the Amazon Resource Name (ARN) of your KMS key, the ID of your KMS key, an alias for your KMS key, or an alias ARN.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(kmsKeyIdRegex, "must conform to: ^[A-Za-z0-9][A-Za-z0-9:_/+=,@.-]{0,2048}$"),
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
			"role_arn": schema.StringAttribute{
				Description: "The Amazon Resource Number (ARN) of the IAM role that allows access to the stream processor.",
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(roleArnRegex, "must conform to: arn:aws:iam::\\d{12}:role/?[a-zA-Z_0-9+=,.@\\-_/]+"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"data_sharing_preference": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[dataSharingPreferenceModel](ctx),
				Description: "Shows whether you are sharing data with Rekognition to improve model performance.",
				Attributes: map[string]schema.Attribute{
					"opt_in": schema.BoolAttribute{
						Description: "Do you want to share data with Rekognition to improve model performance.",
						Optional:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"input": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[inputModel](ctx),
				Description: "Information about the source streaming video.",
				Blocks: map[string]schema.Block{
					"kinesis_video_stream": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[kinesisVideoStreamInputModel](ctx),
						Description: "Kinesis video stream stream that provides the source streaming video for a Amazon Rekognition Video stream processor.",
						Attributes: map[string]schema.Attribute{
							"kinesis_video_stream_arn": schema.StringAttribute{
								CustomType:  fwtypes.ARNType,
								Description: "ARN of the Kinesis video stream stream that streams the source video.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(kinesisStreamArnRegex, "must conform to: (^arn:([a-z\\d-]+):kinesisvideo:([a-z\\d-]+):\\d{12}:.+$)"),
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
					"sns_topic_arn": schema.StringAttribute{
						Description: "The Amazon Resource Number (ARN) of the Amazon Amazon Simple Notification Service topic to which Amazon Rekognition posts the completion status.",
						CustomType:  fwtypes.ARNType,
						Required:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(snsArnRegex, "must conform to: (^arn:aws:sns:.*:\\w{12}:.+$)"),
						},
					},
				},
			},
			"regions_of_interest": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[regionOfInterestModel](ctx),
				Description: "Specifies locations in the frames where Amazon Rekognition checks for objects or people. You can specify up to 10 regions of interest, and each region has either a polygon or a bounding box.",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"region": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[regionOfInterestModel](ctx),
							Blocks: map[string]schema.Block{
								"bounding_box": schema.SingleNestedBlock{
									CustomType:  fwtypes.NewObjectTypeOf[boundingBoxModel](ctx),
									Description: "The box representing a region of interest on screen.",
									Attributes: map[string]schema.Attribute{
										"height": schema.Float64Attribute{
											Optional: true,
											Validators: []validator.Float64{
												float64validator.Between(0.0, 1.0),
											},
										},
										"left": schema.Float64Attribute{
											Optional: true,
											Validators: []validator.Float64{
												float64validator.Between(0.0, 1.0),
											},
										},
										"top": schema.Float64Attribute{
											Optional: true,
											Validators: []validator.Float64{
												float64validator.Between(0.0, 1.0),
											},
										},
										"width": schema.Float64Attribute{
											Optional: true,
											Validators: []validator.Float64{
												float64validator.Between(0.0, 1.0),
											},
										},
									},
								},
								"polygon": schema.SingleNestedBlock{
									CustomType:  fwtypes.NewObjectTypeOf[polygonModel](ctx),
									Description: "Specifies a shape made up of up to 10 Point objects to define a region of interest.",
									Attributes: map[string]schema.Attribute{
										"x": schema.Float64Attribute{
											Optional: true,
											Validators: []validator.Float64{
												float64validator.Between(0.0, 1.0),
											},
										},
										"y": schema.Float64Attribute{
											Optional: true,
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
			},
			"output": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[outputModel](ctx),
				Description: "Kinesis data stream stream or Amazon S3 bucket location to which Amazon Rekognition Video puts the analysis results.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"kinesis_data_stream": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[kinesisDataStreamModel](ctx),
						Description: "The Amazon Kinesis Data Streams stream to which the Amazon Rekognition stream processor streams the analysis results.",
						Attributes: map[string]schema.Attribute{
							"arn": schema.StringAttribute{
								CustomType:  fwtypes.ARNType,
								Description: "ARN of the output Amazon Kinesis Data Streams stream.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.RegexMatches(kinesisStreamArnRegex, "must conform to: (^arn:([a-z\\d-]+):kinesis:([a-z\\d-]+):\\d{12}:.+$)"),
								},
							},
						},
					},
					"s3_destination": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[s3DestinationModel](ctx),
						Description: "The Amazon S3 bucket location to which Amazon Rekognition publishes the detailed inference results of a video analysis operation.",
						Attributes: map[string]schema.Attribute{
							names.AttrBucket: schema.StringAttribute{
								Description: "The name of the Amazon S3 bucket you want to associate with the streaming video project.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(3, 255),
									stringvalidator.RegexMatches(s3bucketRegex, "must conform to: [0-9A-Za-z\\.\\-_]*"),
								},
							},
							"key_prefix": schema.StringAttribute{
								Description: "The prefix value of the location within the bucket that you want the information to be published to.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1024),
								},
							},
						},
					},
				},
			},
			"settings": schema.SingleNestedBlock{
				CustomType:  fwtypes.NewObjectTypeOf[settingsModel](ctx),
				Description: "Input parameters used in a streaming video analyzed by a stream processor.",
				Blocks: map[string]schema.Block{
					"connected_home": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[connectedHomeModel](ctx),
						Description: "Label detection settings to use on a streaming video.",
						Attributes: map[string]schema.Attribute{
							"labels": schema.ListAttribute{
								Description: "Specifies what you want to detect in the video, such as people, packages, or pets.",
								ElementType: fwtypes.StringEnumType[labelSettings](),
								Required:    true,
							},
							"min_confidence": schema.Int64Attribute{
								Description: "The minimum confidence required to label an object in the video.",
								Validators:  []validator.Int64{int64validator.Between(0, 100)},
								Optional:    true,
							},
						},
					},
					"face_search": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[faceSearchModel](ctx),
						Description: "Face search settings to use on a streaming video.",
						Attributes: map[string]schema.Attribute{
							"collection_id": schema.StringAttribute{
								Description: "The ID of a collection that contains faces that you want to search for.",
								Validators: []validator.String{
									stringvalidator.LengthAtMost(2048),
									stringvalidator.RegexMatches(collectionIdRegex, "must conform to: [a-zA-Z0-9_.\\-]+"),
								},
								Optional: true,
							},
							"face_match_threshold": schema.Int64Attribute{
								Description: "Minimum face match confidence score that must be met to return a result for a recognized face.",
								Validators:  []validator.Int64{int64validator.Between(0, 100)},
								Optional:    true,
							},
						},
					},
				},
			},
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

	plan.ID = plan.ARN

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitStreamProcessorCreated(ctx, conn, plan.Name.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForCreation, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceStreamProcessor) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceStreamProcessorDataModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findStreamProcessorByID(ctx, conn, state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, state.Name.String(), err),
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

	if !plan.Name.Equal(state.Name) {

		in := &rekognition.UpdateStreamProcessorInput{
			Name: aws.String(plan.Name.ValueString()),
		}

		// TIP: -- 4. Call the AWS modify/update function
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
	_, err := waitStreamProcessorUpdated(ctx, conn, plan.Name.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForUpdate, ResNameStreamProcessor, plan.Name.String(), err),
			err.Error(),
		)
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
		Name: aws.String(state.Name.ValueString()),
	}

	_, err := conn.DeleteStreamProcessor(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameStreamProcessor, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitStreamProcessorDeleted(ctx, conn, state.Name.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForDeletion, ResNameStreamProcessor, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceStreamProcessor) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitStreamProcessorCreated(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target: []string{
			string(awstypes.StreamProcessorStatusStarting),
			string(awstypes.StreamProcessorStatusRunning),
			string(awstypes.StreamProcessorStatusFailed)},
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
		Pending: []string{string(awstypes.StreamProcessorStatusUpdating)},
		Target: []string{
			string(awstypes.StreamProcessorStatusStarting),
			string(awstypes.StreamProcessorStatusRunning),
			string(awstypes.StreamProcessorStatusFailed),
		},
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
		Pending: []string{
			string(awstypes.StreamProcessorStatusStopped),
			string(awstypes.StreamProcessorStatusStarting),
			string(awstypes.StreamProcessorStatusRunning),
			string(awstypes.StreamProcessorStatusFailed),
			string(awstypes.StreamProcessorStatusStopping),
			string(awstypes.StreamProcessorStatusUpdating),
		},
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
	RoleARN               fwtypes.ARN                                            `tfsdk:"role_arn"`
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
	BoundingBox fwtypes.ObjectValueOf[boundingBoxModel] `tfsdk:"bounding_box"`
	Polygon     fwtypes.ObjectValueOf[polygonModel]     `tfsdk:"polygon"`
}

type boundingBoxModel struct {
	Height types.Number `tfsdk:"height"`
	Left   types.Number `tfsdk:"left"`
	Top    types.Number `tfsdk:"top"`
	Width  types.Number `tfsdk:"width"`
}

type polygonModel struct {
	X types.Number `tfsdk:"x"`
	Y types.Number `tfsdk:"y"`
}

type settingsModel struct {
	ConnectedHome fwtypes.ObjectValueOf[connectedHomeModel] `tfsdk:"connected_home"`
	FaceSearch    fwtypes.ObjectValueOf[faceSearchModel]    `tfsdk:"face_search"`
}

type connectedHomeModel struct {
	Labels        types.List   `tfsdk:"labels"`
	MinConfidence types.Number `tfsdk:"min_confidence"`
}

type faceSearchModel struct {
	CollectionId       types.String `tfsdk:"collection_id"`
	FaceMatchThreshold types.Number `tfsdk:"face_match_threshold"`
}

/*
- AWS SDK doesn't have a CreateStreamProcessorInput.StreamProcessorSettings.ConnectedHomeSettings.Labels enum available as of 5/13/24

- see docs https://docs.aws.amazon.com/rekognition/latest/APIReference/API_ConnectedHomeSettings.html#API_ConnectedHomeSettings_Contents
*/
type labelSettings string

func (labelSettings) Values() []labelSettings {
	return []labelSettings{
		"PERSON",
		"PET",
		"PACKAGE",
		"ALL",
	}
}
