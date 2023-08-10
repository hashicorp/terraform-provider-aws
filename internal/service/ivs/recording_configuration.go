// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

// @SDKResource("aws_ivs_recording_configuration", name="Recording Configuration")
// @Tags(identifierAttribute="id")
func ResourceRecordingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecordingConfigurationCreate,
		ReadWithoutTimeout:   resourceRecordingConfigurationRead,
		DeleteWithoutTimeout: resourceRecordingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
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
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]{0,128}$`), "must contain only alphanumeric characters, hyphen, or underscore, and at most 128 characters"),
			},
			"recording_reconnect_window_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 300),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchemaForceNew(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"thumbnail_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recording_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(ivs.RecordingMode_Values(), false),
						},
						"target_interval_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(5, 60),
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameRecordingConfiguration = "Recording Configuration"
)

func resourceRecordingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	in := &ivs.CreateRecordingConfigurationInput{
		DestinationConfiguration: expandDestinationConfiguration(d.Get("destination_configuration").([]interface{})),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("name"); ok {
		in.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("recording_reconnect_window_seconds"); ok {
		in.RecordingReconnectWindowSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("thumbnail_configuration"); ok {
		in.ThumbnailConfiguration = expandThumbnailConfiguration(v.([]interface{}))

		if aws.StringValue(in.ThumbnailConfiguration.RecordingMode) == ivs.RecordingModeDisabled && in.ThumbnailConfiguration.TargetIntervalSeconds != nil {
			return diag.Errorf("thumbnail configuration target interval cannot be set if recording_mode is \"DISABLED\"")
		}
	}

	out, err := conn.CreateRecordingConfigurationWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionCreating, ResNameRecordingConfiguration, d.Get("name").(string), err)
	}

	if out == nil || out.RecordingConfiguration == nil {
		return create.DiagError(names.IVS, create.ErrActionCreating, ResNameRecordingConfiguration, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.RecordingConfiguration.Arn))

	if _, err := waitRecordingConfigurationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.IVS, create.ErrActionWaitingForCreation, ResNameRecordingConfiguration, d.Id(), err)
	}

	return resourceRecordingConfigurationRead(ctx, d, meta)
}

func resourceRecordingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	out, err := FindRecordingConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IVS RecordingConfiguration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionReading, ResNameRecordingConfiguration, d.Id(), err)
	}

	d.Set("arn", out.Arn)

	if err := d.Set("destination_configuration", flattenDestinationConfiguration(out.DestinationConfiguration)); err != nil {
		return create.DiagError(names.IVS, create.ErrActionSetting, ResNameRecordingConfiguration, d.Id(), err)
	}

	d.Set("name", out.Name)
	d.Set("recording_reconnect_window_seconds", out.RecordingReconnectWindowSeconds)
	d.Set("state", out.State)

	if err := d.Set("thumbnail_configuration", flattenThumbnailConfiguration(out.ThumbnailConfiguration)); err != nil {
		return create.DiagError(names.IVS, create.ErrActionSetting, ResNameRecordingConfiguration, d.Id(), err)
	}

	return nil
}

func resourceRecordingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IVSConn(ctx)

	log.Printf("[INFO] Deleting IVS RecordingConfiguration %s", d.Id())

	_, err := conn.DeleteRecordingConfigurationWithContext(ctx, &ivs.DeleteRecordingConfigurationInput{
		Arn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ivs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.IVS, create.ErrActionDeleting, ResNameRecordingConfiguration, d.Id(), err)
	}

	if _, err := waitRecordingConfigurationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.IVS, create.ErrActionWaitingForDeletion, ResNameRecordingConfiguration, d.Id(), err)
	}

	return nil
}

func flattenDestinationConfiguration(apiObject *ivs.DestinationConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := apiObject.S3; v != nil {
		m["s3"] = flattenS3DestinationConfiguration(v)
	}

	return []interface{}{m}
}

func flattenS3DestinationConfiguration(apiObject *ivs.S3DestinationConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		m["bucket_name"] = aws.StringValue(v)
	}

	return []interface{}{m}
}

func flattenThumbnailConfiguration(apiObject *ivs.ThumbnailConfiguration) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if v := apiObject.RecordingMode; v != nil {
		m["recording_mode"] = aws.StringValue(v)
	}

	if v := apiObject.TargetIntervalSeconds; v != nil {
		m["target_interval_seconds"] = aws.Int64Value(v)
	}

	return []interface{}{m}
}

func expandDestinationConfiguration(vSettings []interface{}) *ivs.DestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}
	tfMap := vSettings[0].(map[string]interface{})
	a := &ivs.DestinationConfiguration{}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		a.S3 = expandS3DestinationConfiguration(v)
	}

	return a
}

func expandS3DestinationConfiguration(vSettings []interface{}) *ivs.S3DestinationConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}

	tfMap := vSettings[0].(map[string]interface{})
	a := &ivs.S3DestinationConfiguration{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	return a
}

func expandThumbnailConfiguration(vSettings []interface{}) *ivs.ThumbnailConfiguration {
	if len(vSettings) == 0 || vSettings[0] == nil {
		return nil
	}
	a := &ivs.ThumbnailConfiguration{}
	tfMap := vSettings[0].(map[string]interface{})

	if v, ok := tfMap["recording_mode"].(string); ok && v != "" {
		a.RecordingMode = aws.String(v)
	}

	if v, ok := tfMap["target_interval_seconds"].(int); ok {
		a.TargetIntervalSeconds = aws.Int64(int64(v))
	}

	return a
}
