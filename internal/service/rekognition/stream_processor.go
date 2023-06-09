package rekognition

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func boundingBoxSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"height": {
					Type:         schema.TypeFloat,
					Optional:     true,
					ValidateFunc: validation.FloatBetween(0, 1),
				},
				"left": {
					Type:         schema.TypeFloat,
					Optional:     true,
					ValidateFunc: validation.FloatBetween(0, 1),
				},
				"top": {
					Type:         schema.TypeFloat,
					Optional:     true,
					ValidateFunc: validation.FloatBetween(0, 1),
				},
				"width": {
					Type:         schema.TypeFloat,
					Optional:     true,
					ValidateFunc: validation.FloatBetween(0, 1),
				},
			},
		},
	}
}

func polygonSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"point": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 10,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"x": {
								Type:         schema.TypeFloat,
								Required:     true,
								ValidateFunc: validation.FloatBetween(0, 1),
							},
							"y": {
								Type:         schema.TypeFloat,
								Required:     true,
								ValidateFunc: validation.FloatBetween(0, 1),
							},
						},
					},
				},
			},
		},
	}
}

func connectedHomeSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"labels": {
					Type:     schema.TypeSet,
					Required: true,
					MinItems: 1,
					MaxItems: 128,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{
							"ALL",
							"PERSON",
							"PET",
							"PACKAGE",
						}, false),
					},
				},
				"min_confidence": {
					Type:         schema.TypeFloat,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.FloatBetween(0, 100),
				},
			},
		},
	}
}

func faceSearchSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"collection_id": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"face_match_threshold": {
					Type:         schema.TypeFloat,
					Optional:     true,
					Default:      80,
					ValidateFunc: validation.FloatBetween(0, 100),
				},
			},
		},
	}
}

func kinesisVideoStreamSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func kinesisDataStreamSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func s3DestinationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"bucket": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(3, 255),
				},
				"key_prefix": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 1024),
				},
			},
		},
	}
}

// @SDKResource("aws_rekognition_stream_processor", name="Stream Processor")
// @Tags(identifierAttribute="arn")
func ResourceStreamProcessor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStreamProcessorCreate,
		ReadWithoutTimeout:   resourceStreamProcessorRead,
		UpdateWithoutTimeout: resourceStreamProcessorUpdate,
		DeleteWithoutTimeout: resourceStreamProcessorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_sharing_preference": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"opt_in": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"input": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_video_stream": kinesisVideoStreamSchema(),
					},
				},
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"notification_channel": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sns_topic_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"output": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_data_stream": kinesisDataStreamSchema(),
						"s3_destination":      s3DestinationSchema(),
					},
				},
			},
			"regions_of_interest": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bounding_box": boundingBoxSchema(),
						"polygon":      polygonSchema(),
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connected_home": connectedHomeSchema(),
						"face_search":    faceSearchSchema(),
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameStreamProcessor = "Stream Processor"
)

func resourceStreamProcessorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()

	name := d.Get("name").(string)

	in := &rekognition.CreateStreamProcessorInput{
		Input:    extractInput(d.Get("input").([]interface{})),
		Name:     aws.String(name),
		Output:   extractOutput(d.Get("output").([]interface{})),
		RoleArn:  aws.String(d.Get("role_arn").(string)),
		Settings: extractSettings(d.Get("settings").([]interface{})),
		Tags:     GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("data_sharing_preference"); ok {
		in.DataSharingPreference = extractDataSharingPreference(v.([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		in.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_channel"); ok {
		in.NotificationChannel = extractNotificationChannel(v.([]interface{}))
	}

	if v, ok := d.GetOk("regions_of_interest"); ok {
		in.RegionsOfInterest = extractRegionsOfInterest(v.(*schema.Set))
	}

	out, err := conn.CreateStreamProcessor(ctx, in)
	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionCreating, ResNameStreamProcessor, d.Get("name").(string), err)
	}

	if out == nil || out.StreamProcessorArn == nil {
		return create.DiagError(names.Rekognition, create.ErrActionCreating, ResNameStreamProcessor, d.Get("name").(string), errors.New("empty output"))
	}

	arn := aws.ToString(out.StreamProcessorArn)
	d.SetId(arn[strings.LastIndex(arn, "/")+1:])

	return resourceStreamProcessorRead(ctx, d, meta)
}

func resourceStreamProcessorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()
	out, err := findStreamProcessorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Rekognition StreamProcessor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionReading, ResNameStreamProcessor, d.Id(), err)
	}

	d.Set("arn", out.StreamProcessorArn)
	d.Set("name", out.Name)
	d.Set("role_arn", out.RoleArn)

	if out.KmsKeyId != nil {
		if err := d.Set("kms_key_id", out.KmsKeyId); err != nil {
			return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
		}
	}

	if out.DataSharingPreference != nil {
		if err := d.Set("data_sharing_preference", flattenDataSharingPreference(out.DataSharingPreference)); err != nil {
			return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
		}
	}

	if err := d.Set("input", flattenInput(out.Input)); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
	}

	if out.NotificationChannel != nil {
		if err := d.Set("notification_channel", flattenNotificationChannel(out.NotificationChannel)); err != nil {
			return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
		}
	}

	if err := d.Set("output", flattenOutput(out.Output)); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
	}

	if out.RegionsOfInterest != nil {
		if err := d.Set("regions_of_interest", flattenRegionsOfInterest(out.RegionsOfInterest)); err != nil {
			return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
		}
	}

	if err := d.Set("settings", flattenSettings(out.Settings)); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameStreamProcessor, d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionReading, ResNameCollection, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameCollection, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameCollection, d.Id(), err)
	}

	return nil
}

func resourceStreamProcessorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()

	update := false

	in := &rekognition.UpdateStreamProcessorInput{
		Name: aws.String(d.Id()),
	}

	if d.HasChangesExcept("tags", "tags_all") {

		if d.HasChanges("data_sharing_preference") {
			in.DataSharingPreferenceForUpdate = extractDataSharingPreference(d.Get("data_sharing_preference").([]interface{}))
			if in.DataSharingPreferenceForUpdate != nil {
				update = true
			}
		}

		if d.HasChanges("regions_of_interest") {
			in.RegionsOfInterestForUpdate = extractRegionsOfInterest(d.Get("regions_of_interest").(*schema.Set))
			if in.RegionsOfInterestForUpdate == nil {
				in.ParametersToDelete = append(in.ParametersToDelete, types.StreamProcessorParameterToDeleteRegionsOfInterest)
			}
			update = true
		}

		if d.HasChanges("settings.0.connected_home") {
			in.SettingsForUpdate = &types.StreamProcessorSettingsForUpdate{
				ConnectedHomeForUpdate: extractConnectedHomeForUpdate(d.Get("settings.0.connected_home").([]interface{})),
			}
			if in.SettingsForUpdate.ConnectedHomeForUpdate == nil {
				in.ParametersToDelete = append(in.ParametersToDelete, types.StreamProcessorParameterToDeleteConnectedHomeMinConfidence)
			}
			update = true
		}
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating Rekognition StreamProcessor (%s): %#v", d.Id(), in)
	_, err := conn.UpdateStreamProcessor(ctx, in)
	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionUpdating, ResNameStreamProcessor, d.Id(), err)
	}

	if _, err := waitStreamProcessorUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionWaitingForUpdate, ResNameStreamProcessor, d.Id(), err)
	}

	return resourceStreamProcessorRead(ctx, d, meta)
}

func resourceStreamProcessorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()

	log.Printf("[INFO] Deleting Rekognition StreamProcessor %s", d.Id())

	_, err := conn.DeleteStreamProcessor(ctx, &rekognition.DeleteStreamProcessorInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.Rekognition, create.ErrActionDeleting, ResNameStreamProcessor, d.Id(), err)
	}
	return nil
}

const (
	statusStopped  = "STOPPED"
	statusUpdating = "UPDATING"
)

func waitStreamProcessorUpdated(ctx context.Context, conn *rekognition.Client, id string, timeout time.Duration) (*rekognition.DescribeStreamProcessorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusUpdating},
		Target:                    []string{statusStopped},
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

func statusStreamProcessor(ctx context.Context, conn *rekognition.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findStreamProcessorByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findStreamProcessorByID(ctx context.Context, conn *rekognition.Client, id string) (*rekognition.DescribeStreamProcessorOutput, error) {
	in := &rekognition.DescribeStreamProcessorInput{
		Name: aws.String(id),
	}
	out, err := conn.DescribeStreamProcessor(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
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

func flattenDataSharingPreference(apiObject *types.StreamProcessorDataSharingPreference) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	tfMap["opt_in"] = apiObject.OptIn

	return []map[string]interface{}{tfMap}
}

func flattenInputKinesisVideoStream(apiObject *types.KinesisVideoStream) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.Arn; v != nil {
		tfMap["arn"] = v
	}
	return []map[string]interface{}{tfMap}
}

func flattenInput(apiObject *types.StreamProcessorInput) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.KinesisVideoStream; v != nil {
		tfMap["kinesis_video_stream"] = flattenInputKinesisVideoStream(v)
	}
	return []map[string]interface{}{tfMap}
}

func flattenNotificationChannel(apiObject *types.StreamProcessorNotificationChannel) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.SNSTopicArn; v != nil {
		tfMap["sns_topic_arn"] = v
	}
	return []map[string]interface{}{tfMap}
}

func flattenOutputKinesisDataStream(apiObject *types.KinesisDataStream) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.Arn; v != nil {
		tfMap["arn"] = v
	}
	return []map[string]interface{}{tfMap}
}

func flattenOutputS3Destination(apiObject *types.S3Destination) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.Bucket; v != nil {
		tfMap["bucket"] = v
	}
	if v := apiObject.KeyPrefix; v != nil {
		tfMap["key_prefix"] = v
	}
	return []map[string]interface{}{tfMap}
}

func flattenOutput(apiObject *types.StreamProcessorOutput) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.KinesisDataStream; v != nil {
		tfMap["kinesis_data_stream"] = flattenOutputKinesisDataStream(v)
	}
	if v := apiObject.S3Destination; v != nil {
		tfMap["s3_destination"] = flattenOutputS3Destination(v)
	}
	return []map[string]interface{}{tfMap}
}

func flattenRegionsOfInterest(apiObjects []types.RegionOfInterest) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}
	var tfList []interface{}
	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}
		if v := apiObject.BoundingBox; v != nil {
			tfMap["bounding_box"] = flattenRegionsOfInterestBoundingBox(v)
		}
		if v := apiObject.Polygon; v != nil {
			tfMap["polygon"] = flattenRegionsOfInterestPolygon(v)
		}
		tfList = append(tfList, tfMap)
	}
	return tfList
}

func flattenRegionsOfInterestBoundingBox(apiObject *types.BoundingBox) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.Height; v != nil {
		tfMap["height"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	if v := apiObject.Left; v != nil {
		tfMap["left"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	if v := apiObject.Top; v != nil {
		tfMap["top"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	if v := apiObject.Width; v != nil {
		tfMap["width"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	return []map[string]interface{}{tfMap}
}

func flattenRegionsOfInterestPoint(apiObject types.Point) map[string]interface{} {
	tfMap := map[string]interface{}{}
	if v := apiObject.X; v != nil {
		tfMap["x"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	if v := apiObject.Y; v != nil {
		tfMap["y"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	return tfMap
}

func flattenRegionsOfInterestPolygon(apiObjects []types.Point) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}
	polygon := make([]map[string]interface{}, 1)
	tfList := make([]interface{}, 0)
	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenRegionsOfInterestPoint(apiObject))
	}
	polygon[0] = map[string]interface{}{
		"point": tfList,
	}
	return polygon
}

func flattenSettingsConnectedHome(apiObject *types.ConnectedHomeSettings) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.Labels; len(v) != 0 {
		tfMap["labels"] = flex.FlattenStringValueSet(v)
	}
	if v := apiObject.MinConfidence; v != nil {
		tfMap["min_confidence"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	return []map[string]interface{}{tfMap}
}

func flattenSettingsFaceSearch(apiObject *types.FaceSearchSettings) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.CollectionId; v != nil {
		tfMap["collection_id"] = v
	}
	if v := apiObject.FaceMatchThreshold; v != nil {
		tfMap["face_match_threshold"] = roundFloat64(float64(aws.ToFloat32(v)))
	}
	return []map[string]interface{}{tfMap}
}

func flattenSettings(apiObject *types.StreamProcessorSettings) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	tfMap := map[string]interface{}{}
	if v := apiObject.ConnectedHome; v != nil {
		tfMap["connected_home"] = flattenSettingsConnectedHome(v)
	}
	if v := apiObject.FaceSearch; v != nil {
		tfMap["face_search"] = flattenSettingsFaceSearch(v)
	}
	return []map[string]interface{}{tfMap}
}

func roundFloat64(v float64) float64 {
	s := fmt.Sprintf("%.7f", v)
	rounded, _ := strconv.ParseFloat(s, 64)
	return rounded
}

func extractInput(tfList []interface{}) *types.StreamProcessorInput {
	if len(tfList) == 0 {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	result := &types.StreamProcessorInput{}
	if v, ok := tfMap["kinesis_video_stream"].([]interface{}); ok && len(v) > 0 {

		result.KinesisVideoStream = &types.KinesisVideoStream{
			Arn: extractArn(v),
		}
	}
	return result
}

func extractArn(tfList []interface{}) *string {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		return aws.String(v)
	} else {
		return nil
	}
}

func extractDataSharingPreference(tfList []interface{}) *types.StreamProcessorDataSharingPreference {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.StreamProcessorDataSharingPreference{}

	if v, ok := tfMap["opt_in"].(bool); ok {
		result.OptIn = v
	}
	return result
}

func extractNotificationChannel(tfList []interface{}) *types.StreamProcessorNotificationChannel {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.StreamProcessorNotificationChannel{}

	if v, ok := tfMap["sns_topic_arn"].(string); ok && v != "" {
		result.SNSTopicArn = aws.String(v)
	}
	return result
}

func extractOutput(tfList []interface{}) *types.StreamProcessorOutput {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.StreamProcessorOutput{}

	if v, ok := tfMap["kinesis_data_stream"].([]interface{}); ok && len(v) > 0 {
		result.KinesisDataStream = &types.KinesisDataStream{
			Arn: extractArn(v),
		}
	}
	if v, ok := tfMap["s3_destination"].([]interface{}); ok && len(v) > 0 {
		result.S3Destination = extractS3Destination(v)
	}
	return result
}

func extractS3Destination(tfList []interface{}) *types.S3Destination {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3Destination{}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}
	if v, ok := tfMap["key_prefix"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}
	return result
}

func extractRegionsOfInterest(tfSet *schema.Set) []types.RegionOfInterest {
	if tfSet.Len() == 0 {
		return nil
	}

	var results []types.RegionOfInterest

	for _, r := range tfSet.List() {
		m := r.(map[string]interface{})
		result := types.RegionOfInterest{}
		if v, ok := m["bounding_box"].([]interface{}); ok && len(v) > 0 {
			result.BoundingBox = extractRegionsOfInterestBoundingBox(v)
		}
		if v, ok := m["polygon"].([]interface{}); ok && len(v) > 0 {
			result.Polygon = extractRegionsOfInterestPolygon(v)
		}

		results = append(results, result)
	}

	return results
}

func extractRegionsOfInterestBoundingBox(tfList []interface{}) *types.BoundingBox {
	if len(tfList) == 0 {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	result := &types.BoundingBox{}
	if v, ok := tfMap["height"].(float64); ok {
		result.Height = aws.Float32(float32(v))
	}
	if v, ok := tfMap["left"].(float64); ok {
		result.Left = aws.Float32(float32(v))
	}
	if v, ok := tfMap["top"].(float64); ok {
		result.Top = aws.Float32(float32(v))
	}
	if v, ok := tfMap["width"].(float64); ok {
		result.Width = aws.Float32(float32(v))
	}
	return result
}

func extractRegionsOfInterestPoint(tfMap map[string]interface{}) types.Point {
	result := types.Point{}
	if v, ok := tfMap["x"].(float64); ok {
		result.X = aws.Float32(float32(v))
	}
	if v, ok := tfMap["y"].(float64); ok {
		result.Y = aws.Float32(float32(v))
	}
	return result
}

func extractRegionsOfInterestPolygon(tfList []interface{}) []types.Point {
	if len(tfList) == 0 {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})
	polygon := make([]types.Point, 0)

	if tfList, ok := tfMap["point"].([]interface{}); ok {
		for _, tfMap := range tfList {
			point := extractRegionsOfInterestPoint(tfMap.(map[string]interface{}))
			polygon = append(polygon, point)
		}
	}
	return polygon
}

func extractSettings(tfList []interface{}) *types.StreamProcessorSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.StreamProcessorSettings{}

	if v, ok := tfMap["connected_home"].([]interface{}); ok && len(v) > 0 {
		result.ConnectedHome = extractConnectedHome(v)
	}
	if v, ok := tfMap["face_search"].([]interface{}); ok && len(v) > 0 {
		result.FaceSearch = extractFaceSearch(v)
	}
	return result
}

func extractConnectedHome(tfList []interface{}) *types.ConnectedHomeSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ConnectedHomeSettings{}

	if v, ok := tfMap["labels"]; ok && v.(*schema.Set).Len() > 0 {
		result.Labels = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := tfMap["min_confidence"].(float64); ok {
		result.MinConfidence = aws.Float32(float32(v))
	}
	return result
}

func extractConnectedHomeForUpdate(tfList []interface{}) *types.ConnectedHomeSettingsForUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ConnectedHomeSettingsForUpdate{}

	if v, ok := tfMap["labels"]; ok && v.(*schema.Set).Len() > 0 {
		result.Labels = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := tfMap["min_confidence"].(float64); ok {
		result.MinConfidence = aws.Float32(float32(v))
	}
	return result
}

func extractFaceSearch(tfList []interface{}) *types.FaceSearchSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.FaceSearchSettings{}

	if v, ok := tfMap["collection_id"].(string); ok && v != "" {
		result.CollectionId = aws.String(v)
	}
	if v, ok := tfMap["face_match_threshold"].(float64); ok {
		result.FaceMatchThreshold = aws.Float32(float32(v))
	}
	return result
}
