// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_device_fleet", name="Device Fleet")
// @Tags(identifierAttribute="arn")
func ResourceDeviceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeviceFleetCreate,
		ReadWithoutTimeout:   resourceDeviceFleetRead,
		UpdateWithoutTimeout: resourceDeviceFleetUpdate,
		DeleteWithoutTimeout: resourceDeviceFleetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 800),
			},
			"device_fleet_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"enable_iot_role_alias": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"iot_role_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"s3_output_location": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeviceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("device_fleet_name").(string)
	input := &sagemaker.CreateDeviceFleetInput{
		DeviceFleetName:    aws.String(name),
		OutputConfig:       expandFeatureDeviceFleetOutputConfig(d.Get("output_config").([]interface{})),
		EnableIotRoleAlias: aws.Bool(d.Get("enable_iot_role_alias").(bool)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.CreateDeviceFleetWithContext(ctx, input)
	}, "ValidationException")
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Device Fleet %s: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDeviceFleetRead(ctx, d, meta)...)
}

func resourceDeviceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deviceFleet, err := FindDeviceFleetByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker Device Fleet (%s); removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Device Fleet (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(deviceFleet.DeviceFleetArn)
	d.Set("device_fleet_name", deviceFleet.DeviceFleetName)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrRoleARN, deviceFleet.RoleArn)
	d.Set(names.AttrDescription, deviceFleet.Description)

	iotAlias := aws.StringValue(deviceFleet.IotRoleAlias)
	d.Set("iot_role_alias", iotAlias)
	d.Set("enable_iot_role_alias", len(iotAlias) > 0)

	if err := d.Set("output_config", flattenFeatureDeviceFleetOutputConfig(deviceFleet.OutputConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_config for SageMaker Device Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDeviceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateDeviceFleetInput{
			DeviceFleetName:    aws.String(d.Id()),
			EnableIotRoleAlias: aws.Bool(d.Get("enable_iot_role_alias").(bool)),
			OutputConfig:       expandFeatureDeviceFleetOutputConfig(d.Get("output_config").([]interface{})),
			RoleArn:            aws.String(d.Get(names.AttrRoleARN).(string)),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		log.Printf("[DEBUG] sagemaker DeviceFleet update config: %s", input.String())
		_, err := conn.UpdateDeviceFleetWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Device Fleet: %s", err)
		}
	}

	return append(diags, resourceDeviceFleetRead(ctx, d, meta)...)
}

func resourceDeviceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteDeviceFleetInput{
		DeviceFleetName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteDeviceFleetWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "DeviceFleet with name") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Device Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func expandFeatureDeviceFleetOutputConfig(l []interface{}) *sagemaker.EdgeOutputConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.EdgeOutputConfig{
		S3OutputLocation: aws.String(m["s3_output_location"].(string)),
	}

	if v, ok := m[names.AttrKMSKeyID].(string); ok && v != "" {
		config.KmsKeyId = aws.String(m[names.AttrKMSKeyID].(string))
	}

	return config
}

func flattenFeatureDeviceFleetOutputConfig(config *sagemaker.EdgeOutputConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"s3_output_location": aws.StringValue(config.S3OutputLocation),
	}

	if config.KmsKeyId != nil {
		m[names.AttrKMSKeyID] = aws.StringValue(config.KmsKeyId)
	}

	return []map[string]interface{}{m}
}
