// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type provisioningHookPayloadVersion string

const (
	provisioningHookPayloadVersion2020_04_01 provisioningHookPayloadVersion = "2020-04-01"
)

func (provisioningHookPayloadVersion) Values() []provisioningHookPayloadVersion {
	return []provisioningHookPayloadVersion{
		provisioningHookPayloadVersion2020_04_01,
	}
}

// @SDKResource("aws_iot_provisioning_template", name="Provisioning Template")
// @Tags(identifierAttribute="arn")
func resourceProvisioningTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProvisioningTemplateCreate,
		ReadWithoutTimeout:   resourceProvisioningTemplateRead,
		UpdateWithoutTimeout: resourceProvisioningTemplateUpdate,
		DeleteWithoutTimeout: resourceProvisioningTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 36),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters and/or the following: _-"),
				),
			},
			"pre_provisioning_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"payload_version": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          provisioningHookPayloadVersion2020_04_01,
							ValidateDiagFunc: enum.Validate[provisioningHookPayloadVersion](),
						},
						names.AttrTargetARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"provisioning_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"template_body": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(0, 10240),
				),
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TemplateType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProvisioningTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iot.CreateProvisioningTemplateInput{
		Enabled:      aws.Bool(d.Get(names.AttrEnabled).(bool)),
		Tags:         getTagsIn(ctx),
		TemplateName: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pre_provisioning_hook"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.PreProvisioningHook = expandProvisioningHook(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("provisioning_role_arn"); ok {
		input.ProvisioningRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.Get(names.AttrType).(awstypes.TemplateType); ok && v != "" {
		input.Type = v
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateProvisioningTemplate(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Provisioning Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*iot.CreateProvisioningTemplateOutput).TemplateName))

	return append(diags, resourceProvisioningTemplateRead(ctx, d, meta)...)
}

func resourceProvisioningTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findProvisioningTemplateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Provisioning Template %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Provisioning Template (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.TemplateArn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrEnabled, output.Enabled)
	d.Set(names.AttrName, output.TemplateName)
	if output.PreProvisioningHook != nil {
		if err := d.Set("pre_provisioning_hook", []interface{}{flattenProvisioningHook(output.PreProvisioningHook)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting pre_provisioning_hook: %s", err)
		}
	} else {
		d.Set("pre_provisioning_hook", nil)
	}
	d.Set("provisioning_role_arn", output.ProvisioningRoleArn)
	d.Set("template_body", output.TemplateBody)
	d.Set(names.AttrType, output.Type)

	return diags
}

func resourceProvisioningTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChange("template_body") {
		input := &iot.CreateProvisioningTemplateVersionInput{
			SetAsDefault: true,
			TemplateBody: aws.String(d.Get("template_body").(string)),
			TemplateName: aws.String(d.Id()),
		}

		_, err := conn.CreateProvisioningTemplateVersion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating IoT Provisioning Template (%s) version: %s", d.Id(), err)
		}
	}

	if d.HasChanges(names.AttrDescription, names.AttrEnabled, "provisioning_role_arn", "pre_provisioning_hook") {
		input := &iot.UpdateProvisioningTemplateInput{
			Description:         aws.String(d.Get(names.AttrDescription).(string)),
			Enabled:             aws.Bool(d.Get(names.AttrEnabled).(bool)),
			ProvisioningRoleArn: aws.String(d.Get("provisioning_role_arn").(string)),
			TemplateName:        aws.String(d.Id()),
		}

		if v, ok := d.GetOk("pre_provisioning_hook"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.PreProvisioningHook = expandProvisioningHook(v.([]interface{})[0].(map[string]interface{}))
		}

		_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateProvisioningTemplate(ctx, input)
			})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Provisioning Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProvisioningTemplateRead(ctx, d, meta)...)
}

func resourceProvisioningTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[INFO] Deleting IoT Provisioning Template: %s", d.Id())
	_, err := conn.DeleteProvisioningTemplate(ctx, &iot.DeleteProvisioningTemplateInput{
		TemplateName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Provisioning Template (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenProvisioningHook(apiObject *awstypes.ProvisioningHook) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PayloadVersion; v != nil {
		tfMap["payload_version"] = aws.ToString(v)
	}

	if v := apiObject.TargetArn; v != nil {
		tfMap[names.AttrTargetARN] = aws.ToString(v)
	}

	return tfMap
}

func expandProvisioningHook(tfMap map[string]interface{}) *awstypes.ProvisioningHook {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ProvisioningHook{}

	if v, ok := tfMap["payload_version"].(string); ok && v != "" {
		apiObject.PayloadVersion = aws.String(v)
	}

	if v, ok := tfMap[names.AttrTargetARN].(string); ok && v != "" {
		apiObject.TargetArn = aws.String(v)
	}

	return apiObject
}

func findProvisioningTemplateByName(ctx context.Context, conn *iot.Client, name string) (*iot.DescribeProvisioningTemplateOutput, error) {
	input := &iot.DescribeProvisioningTemplateInput{
		TemplateName: aws.String(name),
	}

	output, err := conn.DescribeProvisioningTemplate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
