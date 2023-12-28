// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_ca_certificate", name="CA Certificate")
// @Tags(identifierAttribute="arn")
func ResourceCACertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCACertificateCreate,
		ReadWithoutTimeout:   resourceCACertificateRead,
		UpdateWithoutTimeout: resourceCACertificateUpdate,
		DeleteWithoutTimeout: resourceCACertificateDelete,

		Schema: map[string]*schema.Schema{
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"allow_auto_registration": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_certificate_pem": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"certificate_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      iot.CertificateModeDefault,
				ValidateFunc: validation.StringInSlice(iot.CertificateMode_Values(), false),
			},
			"customer_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"generation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registration_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:         schema.TypeBool,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"template_body": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 10240),
						},
						"template_name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 36),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain only alphanumeric characters, underscores, and hyphens"),
							),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"validity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"not_after": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"not_before": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"verification_certificate_pem": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
		},

		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if mode := diff.Get("certificate_mode").(string); mode == iot.CertificateModeDefault {
					if v := diff.GetRawConfig().GetAttr("verification_certificate_pem"); v.IsKnown() {
						if v.IsNull() || v.AsString() == "" {
							return fmt.Errorf(`"verification_certificate_pem" is required when certificate_mode is %q`, mode)
						}
					}
				}

				return nil
			},
			verify.SetTagsDiff,
		),
	}
}

func resourceCACertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	input := &iot.RegisterCACertificateInput{
		AllowAutoRegistration: aws.Bool(d.Get("allow_auto_registration").(bool)),
		CaCertificate:         aws.String(d.Get("ca_certificate_pem").(string)),
		CertificateMode:       aws.String(d.Get("certificate_mode").(string)),
		SetAsActive:           aws.Bool(d.Get("active").(bool)),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("registration_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RegistrationConfig = expandRegistrationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("verification_certificate_pem"); ok {
		input.VerificationCertificate = aws.String(v.(string))
	}

	output, err := conn.RegisterCACertificateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering IoT CA Certificate: %s", err)
	}

	d.SetId(aws.StringValue(output.CertificateId))

	return append(diags, resourceCACertificateRead(ctx, d, meta)...)
}

func resourceCACertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	output, err := FindCACertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT CA Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT CA Certificate (%s): %s", d.Id(), err)
	}

	certificateDescription := output.CertificateDescription
	d.Set("active", aws.StringValue(certificateDescription.Status) == iot.CACertificateStatusActive)
	d.Set("allow_auto_registration", aws.StringValue(certificateDescription.AutoRegistrationStatus) == iot.AutoRegistrationStatusEnable)
	d.Set("arn", certificateDescription.CertificateArn)
	d.Set("ca_certificate_pem", certificateDescription.CertificatePem)
	d.Set("certificate_mode", certificateDescription.CertificateMode)
	d.Set("customer_version", certificateDescription.CustomerVersion)
	d.Set("generation_id", certificateDescription.GenerationId)
	if output.RegistrationConfig != nil {
		if err := d.Set("registration_config", []interface{}{flattenRegistrationConfig(output.RegistrationConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting registration_config: %s", err)
		}
	} else {
		d.Set("registration_config", nil)
	}
	if certificateDescription.Validity != nil {
		if err := d.Set("validity", []interface{}{flattenCertificateValidity(certificateDescription.Validity)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting validity: %s", err)
		}
	} else {
		d.Set("validity", nil)
	}

	return diags
}

func resourceCACertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &iot.UpdateCACertificateInput{
			CertificateId: aws.String(d.Id()),
		}

		if d.Get("active").(bool) {
			input.NewStatus = aws.String(iot.CACertificateStatusActive)
		} else {
			input.NewStatus = aws.String(iot.CACertificateStatusInactive)
		}

		if d.Get("allow_auto_registration").(bool) {
			input.NewAutoRegistrationStatus = aws.String(iot.AutoRegistrationStatusEnable)
		} else {
			input.NewAutoRegistrationStatus = aws.String(iot.AutoRegistrationStatusDisable)
		}

		if d.HasChange("registration_config") {
			if v, ok := d.GetOk("registration_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.RegistrationConfig = expandRegistrationConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := conn.UpdateCACertificateWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT CA Certificate (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCACertificateRead(ctx, d, meta)...)
}

func resourceCACertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	if d.Get("active").(bool) {
		log.Printf("[DEBUG] Disabling IoT CA Certificate: %s", d.Id())
		_, err := conn.UpdateCACertificateWithContext(ctx, &iot.UpdateCACertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     aws.String(iot.CACertificateStatusInactive),
		})

		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling IoT CA Certificate (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting IoT CA Certificate: %s", d.Id())
	_, err := conn.DeleteCACertificateWithContext(ctx, &iot.DeleteCACertificateInput{
		CertificateId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT CA Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func FindCACertificateByID(ctx context.Context, conn *iot.IoT, id string) (*iot.DescribeCACertificateOutput, error) {
	input := &iot.DescribeCACertificateInput{
		CertificateId: aws.String(id),
	}

	output, err := conn.DescribeCACertificateWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CertificateDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandRegistrationConfig(tfMap map[string]interface{}) *iot.RegistrationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &iot.RegistrationConfig{}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["template_body"].(string); ok && v != "" {
		apiObject.TemplateBody = aws.String(v)
	}

	if v, ok := tfMap["template_name"].(string); ok && v != "" {
		apiObject.TemplateName = aws.String(v)
	}

	return apiObject
}

func flattenRegistrationConfig(apiObject *iot.RegistrationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TemplateBody; v != nil {
		tfMap["template_body"] = aws.StringValue(v)
	}

	if v := apiObject.TemplateName; v != nil {
		tfMap["template_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenCertificateValidity(apiObject *iot.CertificateValidity) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NotAfter; v != nil {
		tfMap["not_after"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.NotBefore; v != nil {
		tfMap["not_before"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	return tfMap
}
