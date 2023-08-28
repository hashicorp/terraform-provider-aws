// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_trust_provider", name="Verified Access Trust Provider")
// @Tags(identifierAttribute="id")
func ResourceVerifiedaccessTrustProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedaccessTrustProviderCreate,
		ReadWithoutTimeout:   resourceVerifiedaccessTrustProviderRead,
		UpdateWithoutTimeout: resourceVerifiedaccessTrustProviderUpdate,
		DeleteWithoutTimeout: resourceVerifiedaccessTrustProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"device_options": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tenant_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"device_trust_provider_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.DeviceTrustProviderType](),
			},
			"oidc_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_endpoint": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"client_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"client_secret": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"issuer": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"scope": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"token_endpoint": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"user_info_endpoint": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
					},
				},
			},
			"policy_reference_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), ""),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"trust_provider_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.TrustProviderType](),
			},
			"user_trust_provider_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.UserTrustProviderType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameVerifiedAccessTrustProvider = "Verified Access Trust Provider"
)

func resourceVerifiedaccessTrustProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	in := &ec2.CreateVerifiedAccessTrustProviderInput{
		PolicyReferenceName: aws.String(d.Get("policy_reference_name").(string)),
		TrustProviderType:   types.TrustProviderType(d.Get("trust_provider_type").(string)),
		TagSpecifications:   getTagSpecificationsInV2(ctx, types.ResourceTypeVerifiedAccessTrustProvider),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("device_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.DeviceOptions = expandCreateVerifiedAccessTrustProviderDeviceOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("device_trust_provider_type"); ok {
		in.DeviceTrustProviderType = types.DeviceTrustProviderType(v.(string))
	}

	if v, ok := d.GetOk("user_trust_provider_type"); ok {
		in.UserTrustProviderType = types.UserTrustProviderType(v.(string))
	}

	if v, ok := d.GetOk("oidc_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.OidcOptions = expandCreateVerifiedAccessTrustProviderOIDCOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	out, err := conn.CreateVerifiedAccessTrustProvider(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessTrustProvider, "", err)...)
	}

	if out == nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessTrustProvider, "", errors.New("empty output"))...)
	}

	d.SetId(aws.ToString(out.VerifiedAccessTrustProvider.VerifiedAccessTrustProviderId))

	return append(diags, resourceVerifiedaccessTrustProviderRead(ctx, d, meta)...)
}

func resourceVerifiedaccessTrustProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := FindVerifiedaccessTrustProviderByID(ctx, conn, d.Id())

	// ouxd := oux.DeviceOptions
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedaccessTrustProvider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessTrustProvider, d.Id(), err)...)
	}

	d.Set("description", output.VerifiedAccessTrustProviders[0].Description)

	if v := output.VerifiedAccessTrustProviders[0].DeviceOptions; v != nil {
		if err := d.Set("device_options", flattenDeviceOptions(v)); err != nil {
			return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessTrustProvider, d.Id(), err)
		}
	}

	d.Set("device_trust_provider_type", output.VerifiedAccessTrustProviders[0].DeviceTrustProviderType)

	if v := output.VerifiedAccessTrustProviders[0].OidcOptions; v != nil {
		if err := d.Set("oidc_options", flattenOIDCOptions(v, d.Get("oidc_options.0.client_secret").(string))); err != nil {
			return create.DiagError(names.EC2, create.ErrActionSetting, ResNameVerifiedAccessTrustProvider, d.Id(), err)
		}
	}

	d.Set("policy_reference_name", output.VerifiedAccessTrustProviders[0].PolicyReferenceName)
	d.Set("trust_provider_type", output.VerifiedAccessTrustProviders[0].TrustProviderType)
	d.Set("user_trust_provider_type", output.VerifiedAccessTrustProviders[0].UserTrustProviderType)
	setTagsOutV2(ctx, output.VerifiedAccessTrustProviders[0].Tags)
	return diags
}

func resourceVerifiedaccessTrustProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	update := false

	in := &ec2.ModifyVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: aws.String(d.Id()),
	}

	if d.HasChanges("description") {
		if v, ok := d.GetOk("description"); ok {
			in.Description = aws.String(v.(string))
			update = true
		}
	}

	if d.HasChanges("oidc_options") {
		if v, ok := d.GetOk("oidc_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			in.OidcOptions = expandModifyVerifiedAccessTrustProviderOIDCOptions(v.([]interface{})[0].(map[string]interface{}))
			update = true
		}
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating EC2 VerifiedaccessTrustProvider (%s): %#v", d.Id(), in)
	_, err := conn.ModifyVerifiedAccessTrustProvider(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessTrustProvider, d.Id(), err)...)
	}

	return append(diags, resourceVerifiedaccessTrustProviderRead(ctx, d, meta)...)
}

func resourceVerifiedaccessTrustProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 VerifiedaccessTrustProvider %s", d.Id())

	_, err := conn.DeleteVerifiedAccessTrustProvider(ctx, &ec2.DeleteVerifiedAccessTrustProviderInput{
		VerifiedAccessTrustProviderId: aws.String(d.Id()),
	})

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessTrustProvider, d.Id(), err)...)
	}

	return diags
}

func flattenDeviceOptions(apiObject *types.DeviceOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TenantId; v != nil {
		tfMap["tenant_id"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenOIDCOptions(apiObject *types.OidcOptions, clientSecret string) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"client_secret": clientSecret,
	}

	if v := apiObject.AuthorizationEndpoint; v != nil {
		tfMap["authorization_endpoint"] = aws.ToString(v)
	}
	if v := apiObject.ClientId; v != nil {
		tfMap["client_id"] = aws.ToString(v)
	}
	if v := apiObject.Issuer; v != nil {
		tfMap["issuer"] = aws.ToString(v)
	}
	if v := apiObject.Scope; v != nil {
		tfMap["scope"] = aws.ToString(v)
	}
	if v := apiObject.TokenEndpoint; v != nil {
		tfMap["token_endpoint"] = aws.ToString(v)
	}
	if v := apiObject.UserInfoEndpoint; v != nil {
		tfMap["user_info_endpoint"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandCreateVerifiedAccessTrustProviderDeviceOptions(tfMap map[string]interface{}) *types.CreateVerifiedAccessTrustProviderDeviceOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateVerifiedAccessTrustProviderDeviceOptions{}

	if v, ok := tfMap["tenant_id"].(string); ok && v != "" {
		apiObject.TenantId = aws.String(v)
	}

	return apiObject
}

func expandCreateVerifiedAccessTrustProviderOIDCOptions(tfMap map[string]interface{}) *types.CreateVerifiedAccessTrustProviderOidcOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CreateVerifiedAccessTrustProviderOidcOptions{}

	if v, ok := tfMap["authorization_endpoint"].(string); ok && v != "" {
		apiObject.AuthorizationEndpoint = aws.String(v)
	}
	if v, ok := tfMap["client_id"].(string); ok && v != "" {
		apiObject.ClientId = aws.String(v)
	}
	if v, ok := tfMap["client_secret"].(string); ok && v != "" {
		apiObject.ClientSecret = aws.String(v)
	}
	if v, ok := tfMap["issuer"].(string); ok && v != "" {
		apiObject.Issuer = aws.String(v)
	}
	if v, ok := tfMap["scope"].(string); ok && v != "" {
		apiObject.Scope = aws.String(v)
	}
	if v, ok := tfMap["token_endpoint"].(string); ok && v != "" {
		apiObject.TokenEndpoint = aws.String(v)
	}
	if v, ok := tfMap["user_info_endpoint"].(string); ok && v != "" {
		apiObject.UserInfoEndpoint = aws.String(v)
	}

	return apiObject
}

func expandModifyVerifiedAccessTrustProviderOIDCOptions(tfMap map[string]interface{}) *types.ModifyVerifiedAccessTrustProviderOidcOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ModifyVerifiedAccessTrustProviderOidcOptions{}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		apiObject.Scope = aws.String(v)
	}

	return apiObject
}
