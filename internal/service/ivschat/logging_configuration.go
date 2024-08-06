// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivschat

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
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

// @SDKResource("aws_ivschat_logging_configuration", name="Logging Configuration")
// @Tags(identifierAttribute="id")
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
			names.AttrARN: {
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
						names.AttrCloudWatchLogs: {
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
									names.AttrLogGroupName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_./#-]{1,512}$`), "must contain only lowercase alphanumeric characters, hyphen, dot, underscore, forward slash, or hash sign, and between 1 and 512 characters"),
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
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]{1,64}$`), "must contain only lowercase alphanumeric characters, hyphen, dot, or underscore, and between 1 and 64 characters"),
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
									names.AttrBucketName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9a-z.-]{3,63}$`), "must contain only lowercase alphanumeric characters, hyphen, or dot, and between 3 and 63 characters"),
									},
								},
							},
						},
					},
				},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameLoggingConfiguration = "Logging Configuration"
)

func resourceLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	in := &ivschat.CreateLoggingConfigurationInput{
		DestinationConfiguration: expandDestinationConfiguration(d.Get("destination_configuration").([]interface{})),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		in.Name = aws.String(v.(string))
	}

	out, err := conn.CreateLoggingConfiguration(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionCreating, ResNameLoggingConfiguration, d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionCreating, ResNameLoggingConfiguration, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionWaitingForCreation, ResNameLoggingConfiguration, d.Id(), err)
	}

	return append(diags, resourceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	out, err := findLoggingConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVSChat LoggingConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionReading, ResNameLoggingConfiguration, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)

	if err := d.Set("destination_configuration", flattenDestinationConfiguration(out.DestinationConfiguration)); err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionSetting, ResNameLoggingConfiguration, d.Id(), err)
	}

	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrState, out.State)

	return diags
}

func resourceLoggingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	update := false

	in := &ivschat.UpdateLoggingConfigurationInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges(names.AttrName) {
		in.Name = aws.String(d.Get(names.AttrName).(string))
		update = true
	}

	if d.HasChanges("destination_configuration") {
		in.DestinationConfiguration = expandDestinationConfiguration(d.Get("destination_configuration").([]interface{}))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating IVSChat LoggingConfiguration (%s): %#v", d.Id(), in)
	out, err := conn.UpdateLoggingConfiguration(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionUpdating, ResNameLoggingConfiguration, d.Id(), err)
	}

	if _, err := waitLoggingConfigurationUpdated(ctx, conn, aws.ToString(out.Arn), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionWaitingForUpdate, ResNameLoggingConfiguration, d.Id(), err)
	}

	return append(diags, resourceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IVSChatClient(ctx)

	log.Printf("[INFO] Deleting IVSChat LoggingConfiguration %s", d.Id())

	_, err := conn.DeleteLoggingConfiguration(ctx, &ivschat.DeleteLoggingConfigurationInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionDeleting, ResNameLoggingConfiguration, d.Id(), err)
	}

	if _, err := waitLoggingConfigurationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.IVSChat, create.ErrActionWaitingForDeletion, ResNameLoggingConfiguration, d.Id(), err)
	}

	return diags
}

func flattenDestinationConfiguration(apiObject types.DestinationConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *types.DestinationConfigurationMemberCloudWatchLogs:
		m[names.AttrCloudWatchLogs] = flattenCloudWatchDestinationConfiguration(v.Value)

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
		m[names.AttrLogGroupName] = aws.ToString(v)
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
		m[names.AttrBucketName] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandDestinationConfiguration(vSettings []interface{}) types.DestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	tfMap := vSettings[0].(map[string]interface{})

	if v, ok := tfMap[names.AttrCloudWatchLogs].([]interface{}); ok && len(v) > 0 {
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

	if v, ok := tfMap[names.AttrLogGroupName].(string); ok && v != "" {
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

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	return a
}
