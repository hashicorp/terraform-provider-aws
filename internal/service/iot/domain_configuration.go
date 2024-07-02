// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_domain_configuration", name="Domain Configuration")
// @Tags(identifierAttribute="arn")
func resourceDomainConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainConfigurationCreate,
		ReadWithoutTimeout:   resourceDomainConfigurationRead,
		UpdateWithoutTimeout: resourceDomainConfigurationUpdate,
		DeleteWithoutTimeout: resourceDomainConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorizer_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_authorizer_override": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"default_authorizer_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"domain_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_certificate_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"service_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.ServiceTypeData,
				ValidateDiagFunc: enum.Validate[awstypes.ServiceType](),
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DomainConfigurationStatusEnabled,
				ValidateDiagFunc: enum.Validate[awstypes.DomainConfigurationStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tls_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"validation_certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceDomainConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iot.CreateDomainConfigurationInput{
		DomainConfigurationName: aws.String(name),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("authorizer_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AuthorizerConfig = expandAuthorizerConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDomainName); ok {
		input.DomainName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_certificate_arns"); ok && v.(*schema.Set).Len() > 0 {
		input.ServerCertificateArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("service_type"); ok {
		input.ServiceType = awstypes.ServiceType(v.(string))
	}

	if v, ok := d.GetOk("tls_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TlsConfig = expandTlsConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("validation_certificate_arn"); ok {
		input.ValidationCertificateArn = aws.String(v.(string))
	}

	output, err := conn.CreateDomainConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Domain Configuration (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DomainConfigurationName))

	return append(diags, resourceDomainConfigurationRead(ctx, d, meta)...)
}

func resourceDomainConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findDomainConfigurationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Domain Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Domain Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.DomainConfigurationArn)
	if output.AuthorizerConfig != nil {
		if err := d.Set("authorizer_config", []interface{}{flattenAuthorizerConfig(output.AuthorizerConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authorizer_config: %s", err)
		}
	} else {
		d.Set("authorizer_config", nil)
	}
	d.Set(names.AttrDomainName, output.DomainName)
	d.Set("domain_type", output.DomainType)
	d.Set(names.AttrName, output.DomainConfigurationName)
	d.Set("server_certificate_arns", tfslices.ApplyToAll(output.ServerCertificates, func(v awstypes.ServerCertificateSummary) string {
		return aws.ToString(v.ServerCertificateArn)
	}))
	d.Set("service_type", output.ServiceType)
	d.Set(names.AttrStatus, output.DomainConfigurationStatus)
	if output.TlsConfig != nil {
		if err := d.Set("tls_config", []interface{}{flattenTlsConfig(output.TlsConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tls_config: %s", err)
		}
	} else {
		d.Set("tls_config", nil)
	}
	d.Set("validation_certificate_arn", d.Get("validation_certificate_arn"))

	return diags
}

func resourceDomainConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &iot.UpdateDomainConfigurationInput{
			DomainConfigurationName: aws.String(d.Id()),
		}

		if d.HasChange("authorizer_config") {
			if v, ok := d.GetOk("authorizer_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.AuthorizerConfig = expandAuthorizerConfig(v.([]interface{})[0].(map[string]interface{}))
			} else {
				input.RemoveAuthorizerConfig = true
			}
		}

		if d.HasChange(names.AttrStatus) {
			input.DomainConfigurationStatus = awstypes.DomainConfigurationStatus(d.Get(names.AttrStatus).(string))
		}

		if d.HasChange("tls_config") {
			if v, ok := d.GetOk("tls_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TlsConfig = expandTlsConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := conn.UpdateDomainConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Domain Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainConfigurationRead(ctx, d, meta)...)
}

func resourceDomainConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.Get(names.AttrStatus).(string) == string(awstypes.DomainConfigurationStatusEnabled) {
		_, err := conn.UpdateDomainConfiguration(ctx, &iot.UpdateDomainConfigurationInput{
			DomainConfigurationName:   aws.String(d.Id()),
			DomainConfigurationStatus: awstypes.DomainConfigurationStatusDisabled,
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling IoT Domain Configuration (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting IoT Domain Configuration: %s", d.Id())
	_, err := conn.DeleteDomainConfiguration(ctx, &iot.DeleteDomainConfigurationInput{
		DomainConfigurationName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Domain Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainConfigurationByName(ctx context.Context, conn *iot.Client, name string) (*iot.DescribeDomainConfigurationOutput, error) {
	input := &iot.DescribeDomainConfigurationInput{
		DomainConfigurationName: aws.String(name),
	}

	output, err := conn.DescribeDomainConfiguration(ctx, input)

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

func expandAuthorizerConfig(tfMap map[string]interface{}) *awstypes.AuthorizerConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AuthorizerConfig{}

	if v, ok := tfMap["allow_authorizer_override"].(bool); ok {
		apiObject.AllowAuthorizerOverride = aws.Bool(v)
	}

	if v, ok := tfMap["default_authorizer_name"].(string); ok && v != "" {
		apiObject.DefaultAuthorizerName = aws.String(v)
	}

	return apiObject
}

func expandTlsConfig(tfMap map[string]interface{}) *awstypes.TlsConfig { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TlsConfig{}

	if v, ok := tfMap["security_policy"].(string); ok && v != "" {
		apiObject.SecurityPolicy = aws.String(v)
	}

	return apiObject
}

func flattenAuthorizerConfig(apiObject *awstypes.AuthorizerConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowAuthorizerOverride; v != nil {
		tfMap["allow_authorizer_override"] = aws.ToBool(v)
	}

	if v := apiObject.DefaultAuthorizerName; v != nil {
		tfMap["default_authorizer_name"] = aws.ToString(v)
	}

	return tfMap
}

func flattenTlsConfig(apiObject *awstypes.TlsConfig) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityPolicy; v != nil {
		tfMap["security_policy"] = aws.ToString(v)
	}

	return tfMap
}
