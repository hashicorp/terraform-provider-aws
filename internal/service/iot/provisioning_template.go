package iot

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	provisioningHookPayloadVersion2020_04_01 = "2020-04-01"
)

func provisioningHookPayloadVersion_Values() []string {
	return []string{
		provisioningHookPayloadVersion2020_04_01,
	}
}

func ResourceProvisioningTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProvisioningTemplateCreate,
		ReadWithoutTimeout:   resourceProvisioningTemplateRead,
		UpdateWithoutTimeout: resourceProvisioningTemplateUpdate,
		DeleteWithoutTimeout: resourceProvisioningTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 36),
					validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters and/or the following: _-"),
				),
			},
			"pre_provisioning_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"payload_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      provisioningHookPayloadVersion2020_04_01,
							ValidateFunc: validation.StringInSlice(provisioningHookPayloadVersion_Values(), false),
						},
						"target_arn": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"template_body": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringIsJSON,
					validation.StringLenBetween(0, 10240),
				),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProvisioningTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &iot.CreateProvisioningTemplateInput{
		Enabled:      aws.Bool(d.Get("enabled").(bool)),
		TemplateName: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
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

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating IoT Provisioning Template: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContainsContext(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateProvisioningTemplateWithContext(ctx, input)
		},
		iot.ErrCodeInvalidRequestException, "The provisioning role cannot be assumed by AWS IoT")

	if err != nil {
		return diag.Errorf("error creating IoT Provisioning Template (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*iot.CreateProvisioningTemplateOutput).TemplateName))

	return resourceProvisioningTemplateRead(ctx, d, meta)
}

func resourceProvisioningTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindProvisioningTemplateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Provisioning Template %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading IoT Provisioning Template (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.TemplateArn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set("description", output.Description)
	d.Set("enabled", output.Enabled)
	d.Set("name", output.TemplateName)
	if output.PreProvisioningHook != nil {
		if err := d.Set("pre_provisioning_hook", []interface{}{flattenProvisioningHook(output.PreProvisioningHook)}); err != nil {
			return diag.Errorf("error setting pre_provisioning_hook: %s", err)
		}
	} else {
		d.Set("pre_provisioning_hook", nil)
	}
	d.Set("provisioning_role_arn", output.ProvisioningRoleArn)
	d.Set("template_body", output.TemplateBody)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("error listing tags for IoT Provisioning Template (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceProvisioningTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("template_body") {
		input := &iot.CreateProvisioningTemplateVersionInput{
			SetAsDefault: aws.Bool(true),
			TemplateBody: aws.String(d.Get("template_body").(string)),
			TemplateName: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Creating IoT Provisioning Template version: %s", input)
		_, err := conn.CreateProvisioningTemplateVersionWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error creating IoT Provisioning Template (%s) version: %s", d.Id(), err)
		}
	}

	if d.HasChanges("description", "enabled", "provisioning_role_arn") {
		input := &iot.UpdateProvisioningTemplateInput{
			Description:         aws.String(d.Get("description").(string)),
			Enabled:             aws.Bool(d.Get("enabled").(bool)),
			ProvisioningRoleArn: aws.String(d.Get("provisioning_role_arn").(string)),
			TemplateName:        aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating IoT Provisioning Template: %s", input)
		_, err := tfresource.RetryWhenAWSErrMessageContainsContext(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateProvisioningTemplateWithContext(ctx, input)
			},
			iot.ErrCodeInvalidRequestException, "The provisioning role cannot be assumed by AWS IoT")

		if err != nil {
			return diag.Errorf("error updating IoT Provisioning Template (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceProvisioningTemplateRead(ctx, d, meta)
}

func resourceProvisioningTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IoTConn

	log.Printf("[INFO] Deleting IoT Provisioning Template: %s", d.Id())
	_, err := conn.DeleteProvisioningTemplateWithContext(ctx, &iot.DeleteProvisioningTemplateInput{
		TemplateName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting IoT Provisioning Template (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenProvisioningHook(apiObject *iot.ProvisioningHook) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PayloadVersion; v != nil {
		tfMap["payload_version"] = aws.StringValue(v)
	}

	if v := apiObject.TargetArn; v != nil {
		tfMap["target_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func expandProvisioningHook(tfMap map[string]interface{}) *iot.ProvisioningHook {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.ProvisioningHook{}

	if v, ok := tfMap["payload_version"].(string); ok && v != "" {
		apiObject.PayloadVersion = aws.String(v)
	}

	if v, ok := tfMap["target_arn"].(string); ok && v != "" {
		apiObject.TargetArn = aws.String(v)
	}

	return apiObject
}

func FindProvisioningTemplateByName(ctx context.Context, conn *iot.IoT, name string) (*iot.DescribeProvisioningTemplateOutput, error) {
	input := &iot.DescribeProvisioningTemplateInput{
		TemplateName: aws.String(name),
	}

	output, err := conn.DescribeProvisioningTemplateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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
