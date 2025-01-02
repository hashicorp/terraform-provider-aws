// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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

// @SDKResource("aws_iot_ca_certificate", name="CA Certificate")
// @Tags(identifierAttribute="arn")
func resourceCACertificate() *schema.Resource {
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
			names.AttrARN: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.CertificateModeDefault,
				ValidateDiagFunc: enum.Validate[awstypes.CertificateMode](),
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
						names.AttrRoleARN: {
							Type:         schema.TypeString,
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
				if mode := diff.Get("certificate_mode").(string); mode == string(awstypes.CertificateModeDefault) {
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
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	input := &iot.RegisterCACertificateInput{
		AllowAutoRegistration: d.Get("allow_auto_registration").(bool),
		CaCertificate:         aws.String(d.Get("ca_certificate_pem").(string)),
		CertificateMode:       awstypes.CertificateMode(d.Get("certificate_mode").(string)),
		SetAsActive:           d.Get("active").(bool),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("registration_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RegistrationConfig = expandRegistrationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("verification_certificate_pem"); ok {
		input.VerificationCertificate = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.RegisterCACertificate(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering IoT CA Certificate: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*iot.RegisterCACertificateOutput).CertificateId))

	return append(diags, resourceCACertificateRead(ctx, d, meta)...)
}

func resourceCACertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findCACertificateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT CA Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT CA Certificate (%s): %s", d.Id(), err)
	}

	certificateDescription := output.CertificateDescription
	d.Set("active", string(certificateDescription.Status) == string(awstypes.CACertificateStatusActive))
	d.Set("allow_auto_registration", string(certificateDescription.AutoRegistrationStatus) == string(awstypes.AutoRegistrationStatusEnable))
	d.Set(names.AttrARN, certificateDescription.CertificateArn)
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
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &iot.UpdateCACertificateInput{
			CertificateId: aws.String(d.Id()),
		}

		if d.Get("active").(bool) {
			input.NewStatus = awstypes.CACertificateStatusActive
		} else {
			input.NewStatus = awstypes.CACertificateStatusInactive
		}

		if d.Get("allow_auto_registration").(bool) {
			input.NewAutoRegistrationStatus = awstypes.AutoRegistrationStatusEnable
		} else {
			input.NewAutoRegistrationStatus = awstypes.AutoRegistrationStatusDisable
		}

		if d.HasChange("registration_config") {
			if v, ok := d.GetOk("registration_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.RegistrationConfig = expandRegistrationConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := tfresource.RetryWhenIsA[*awstypes.InvalidRequestException](ctx, propagationTimeout, func() (interface{}, error) {
			return conn.UpdateCACertificate(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT CA Certificate (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCACertificateRead(ctx, d, meta)...)
}

func resourceCACertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.Get("active").(bool) {
		_, err := conn.UpdateCACertificate(ctx, &iot.UpdateCACertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     awstypes.CACertificateStatusInactive,
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling IoT CA Certificate (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting IoT CA Certificate: %s", d.Id())
	_, err := conn.DeleteCACertificate(ctx, &iot.DeleteCACertificateInput{
		CertificateId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT CA Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func findCACertificateByID(ctx context.Context, conn *iot.Client, id string) (*iot.DescribeCACertificateOutput, error) {
	input := &iot.DescribeCACertificateInput{
		CertificateId: aws.String(id),
	}

	output, err := conn.DescribeCACertificate(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func expandRegistrationConfig(tfMap map[string]interface{}) *awstypes.RegistrationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RegistrationConfig{}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
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

func flattenRegistrationConfig(apiObject *awstypes.RegistrationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.TemplateBody; v != nil {
		tfMap["template_body"] = aws.ToString(v)
	}

	if v := apiObject.TemplateName; v != nil {
		tfMap["template_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenCertificateValidity(apiObject *awstypes.CertificateValidity) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NotAfter; v != nil {
		tfMap["not_after"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := apiObject.NotBefore; v != nil {
		tfMap["not_before"] = aws.ToTime(v).Format(time.RFC3339)
	}

	return tfMap
}
