package ivschat

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoggingConfigurationCreate,
		ReadWithoutTimeout:   resourceLoggingConfigurationRead,
		UpdateWithoutTimeout: resourceLoggingConfigurationUpdate,
		DeleteWithoutTimeout: resourceLoggingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ExactlyOneOf: []string{
								"destination_configuration.0.cloudwatch_logs",
								"destination_configuration.0.firehose",
								"destination_configuration.0.s3",
							},
							ConflictsWith: []string{
								"destination_configuration.0.firehose",
								"destination_configuration.0.s3",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\.\-_/#A-Za-z0-9]{1,512}$`), "must contain only lowercase alphanumeric characters, hyphen, dot, underscore, forward slash, or hash sign, and between 1 and 512 characters"),
									},
								},
							},
						},
						"firehose": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ExactlyOneOf: []string{
								"destination_configuration.0.cloudwatch_logs",
								"destination_configuration.0.firehose",
								"destination_configuration.0.s3",
							},
							ConflictsWith: []string{
								"destination_configuration.0.cloudwatch_logs",
								"destination_configuration.0.s3",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delivery_stream_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]{1,64}$`), "must contain only lowercase alphanumeric characters, hyphen, dot, or underscore, and between 1 and 64 characters"),
									},
								},
							},
						},
						"s3": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ExactlyOneOf: []string{
								"destination_configuration.0.cloudwatch_logs",
								"destination_configuration.0.firehose",
								"destination_configuration.0.s3",
							},
							ConflictsWith: []string{
								"destination_configuration.0.cloudwatch_logs",
								"destination_configuration.0.firehose",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9-.]{3,63}$`), "must contain only lowercase alphanumeric characters, hyphen, or dot, and between 3 and 63 characters"),
									},
								},
							},
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameLoggingConfiguration = "Logging Configuration"
)

func resourceLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient()

	in := &ivschat.CreateLoggingConfigurationInput{
		DestinationConfiguration: expandDestinationConfiguration(d.Get("destination_configuration").([]interface{})),
	}

	if v, ok := d.GetOk("name"); ok {
		in.Name = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateLoggingConfiguration(ctx, in)
	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionCreating, ResNameLoggingConfiguration, d.Get("name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.IVSChat, create.ErrActionCreating, ResNameLoggingConfiguration, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionWaitingForCreation, ResNameLoggingConfiguration, d.Id(), err)
	}

	return resourceLoggingConfigurationRead(ctx, d, meta)
}

func resourceLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient()

	out, err := findLoggingConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVSChat LoggingConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionReading, ResNameLoggingConfiguration, d.Id(), err)
	}

	d.Set("arn", out.Arn)

	if err := d.Set("destination_configuration", flattenDestinationConfiguration(out.DestinationConfiguration)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionSetting, ResNameLoggingConfiguration, d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("state", out.State)

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionReading, ResNameLoggingConfiguration, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionSetting, ResNameLoggingConfiguration, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionSetting, ResNameLoggingConfiguration, d.Id(), err)
	}

	return nil
}

func resourceLoggingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient()

	update := false

	in := &ivschat.UpdateLoggingConfigurationInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges("name") {
		in.Name = aws.String(d.Get("name").(string))
		update = true
	}

	if d.HasChanges("destination_configuration") {
		in.DestinationConfiguration = expandDestinationConfiguration(d.Get("destination_configuration").([]interface{}))
		update = true
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.IVS, create.ErrActionUpdating, ResNameLoggingConfiguration, d.Id(), err)
		}
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating IVSChat LoggingConfiguration (%s): %#v", d.Id(), in)
	out, err := conn.UpdateLoggingConfiguration(ctx, in)
	if err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionUpdating, ResNameLoggingConfiguration, d.Id(), err)
	}

	if _, err := waitLoggingConfigurationUpdated(ctx, conn, aws.ToString(out.Arn), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionWaitingForUpdate, ResNameLoggingConfiguration, d.Id(), err)
	}

	return resourceLoggingConfigurationRead(ctx, d, meta)
}

func resourceLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSChatClient()

	log.Printf("[INFO] Deleting IVSChat LoggingConfiguration %s", d.Id())

	_, err := conn.DeleteLoggingConfiguration(ctx, &ivschat.DeleteLoggingConfigurationInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.IVSChat, create.ErrActionDeleting, ResNameLoggingConfiguration, d.Id(), err)
	}

	if _, err := waitLoggingConfigurationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.IVSChat, create.ErrActionWaitingForDeletion, ResNameLoggingConfiguration, d.Id(), err)
	}

	return nil
}

func flattenDestinationConfiguration(apiObject types.DestinationConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *types.DestinationConfigurationMemberCloudWatchLogs:
		m["cloudwatch_logs"] = flattenCloudWatchDestinationConfiguration(v.Value)

	case *types.DestinationConfigurationMemberFirehose:
		m["firehose"] = flattenFirehoseDestinationConfiguration(v.Value)

	case *types.DestinationConfigurationMemberS3:
		m["s3"] = flattenS3DestinationConfiguration(v.Value)

	case *types.UnknownUnionMember:
		log.Println("unknown tag:", v.Tag)

	default:
		log.Println("union is nil or unknown type")
	}

	return []interface{}{m}
}

func flattenCloudWatchDestinationConfiguration(apiObject types.CloudWatchLogsDestinationConfiguration) []interface{} {
	m := map[string]interface{}{}

	if v := apiObject.LogGroupName; v != nil {
		m["log_group_name"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenFirehoseDestinationConfiguration(apiObject types.FirehoseDestinationConfiguration) []interface{} {
	m := map[string]interface{}{}

	if v := apiObject.DeliveryStreamName; v != nil {
		m["delivery_stream_name"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenS3DestinationConfiguration(apiObject types.S3DestinationConfiguration) []interface{} {
	m := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		m["bucket_name"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandDestinationConfiguration(vSettings []interface{}) types.DestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	tfMap := vSettings[0].(map[string]interface{})

	if v, ok := tfMap["cloudwatch_logs"].([]interface{}); ok && len(v) > 0 {
		return &types.DestinationConfigurationMemberCloudWatchLogs{
			Value: *expandCloudWatchLogsDestinationConfiguration(v),
		}
	} else if v, ok := tfMap["firehose"].([]interface{}); ok && len(v) > 0 {
		return &types.DestinationConfigurationMemberFirehose{
			Value: *expandFirehouseDestinationConfiguration(v),
		}
	} else if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		return &types.DestinationConfigurationMemberS3{
			Value: *expandS3DestinationConfiguration(v),
		}
	} else {
		return nil
	}
}

func expandCloudWatchLogsDestinationConfiguration(vSettings []interface{}) *types.CloudWatchLogsDestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	a := &types.CloudWatchLogsDestinationConfiguration{}

	tfMap := vSettings[0].(map[string]interface{})

	if v, ok := tfMap["log_group_name"].(string); ok && v != "" {
		a.LogGroupName = aws.String(v)
	}

	return a
}

func expandFirehouseDestinationConfiguration(vSettings []interface{}) *types.FirehoseDestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	a := &types.FirehoseDestinationConfiguration{}

	tfMap := vSettings[0].(map[string]interface{})

	if v, ok := tfMap["delivery_stream_name"].(string); ok && v != "" {
		a.DeliveryStreamName = aws.String(v)
	}

	return a
}

func expandS3DestinationConfiguration(vSettings []interface{}) *types.S3DestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	a := &types.S3DestinationConfiguration{}

	tfMap := vSettings[0].(map[string]interface{})

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	return a
}
