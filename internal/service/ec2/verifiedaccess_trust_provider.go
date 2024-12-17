// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_trust_provider", name="Verified Access Trust Provider")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVerifiedAccessTrustProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessTrustProviderCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessTrustProviderRead,
		UpdateWithoutTimeout: resourceVerifiedAccessTrustProviderUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessTrustProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
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
						names.AttrClientID: {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						names.AttrClientSecret: {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						names.AttrIssuer: {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						names.AttrScope: {
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

func resourceVerifiedAccessTrustProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVerifiedAccessTrustProviderInput{
		ClientToken:         aws.String(id.UniqueId()),
		PolicyReferenceName: aws.String(d.Get("policy_reference_name").(string)),
		TagSpecifications:   getTagSpecificationsIn(ctx, types.ResourceTypeVerifiedAccessTrustProvider),
		TrustProviderType:   types.TrustProviderType(d.Get("trust_provider_type").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("device_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeviceOptions = expandCreateVerifiedAccessTrustProviderDeviceOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("device_trust_provider_type"); ok {
		input.DeviceTrustProviderType = types.DeviceTrustProviderType(v.(string))
	}

	if v, ok := d.GetOk("oidc_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OidcOptions = expandCreateVerifiedAccessTrustProviderOIDCOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("user_trust_provider_type"); ok {
		input.UserTrustProviderType = types.UserTrustProviderType(v.(string))
	}

	output, err := conn.CreateVerifiedAccessTrustProvider(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Trust Provider: %s", err)
	}

	d.SetId(aws.ToString(output.VerifiedAccessTrustProvider.VerifiedAccessTrustProviderId))

	return append(diags, resourceVerifiedAccessTrustProviderRead(ctx, d, meta)...)
}

func resourceVerifiedAccessTrustProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := findVerifiedAccessTrustProviderByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Trust Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Trust Provider (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, output.Description)
	if v := output.DeviceOptions; v != nil {
		if err := d.Set("device_options", flattenDeviceOptions(v)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting device_options: %s", err)
		}
	} else {
		d.Set("device_options", nil)
	}
	d.Set("device_trust_provider_type", output.DeviceTrustProviderType)
	if v := output.OidcOptions; v != nil {
		if err := d.Set("oidc_options", flattenOIDCOptions(v, d.Get("oidc_options.0.client_secret").(string))); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting oidc_options: %s", err)
		}
	} else {
		d.Set("oidc_options", nil)
	}
	d.Set("policy_reference_name", output.PolicyReferenceName)
	d.Set("trust_provider_type", output.TrustProviderType)
	d.Set("user_trust_provider_type", output.UserTrustProviderType)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceVerifiedAccessTrustProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyVerifiedAccessTrustProviderInput{
			ClientToken:                   aws.String(id.UniqueId()),
			VerifiedAccessTrustProviderId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("oidc_options") {
			if v, ok := d.GetOk("oidc_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.OidcOptions = expandModifyVerifiedAccessTrustProviderOIDCOptions(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := conn.ModifyVerifiedAccessTrustProvider(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Trust Provider (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVerifiedAccessTrustProviderRead(ctx, d, meta)...)
}

func resourceVerifiedAccessTrustProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting Verified Access Trust Provider: %s", d.Id())
	_, err := conn.DeleteVerifiedAccessTrustProvider(ctx, &ec2.DeleteVerifiedAccessTrustProviderInput{
		ClientToken:                   aws.String(id.UniqueId()),
		VerifiedAccessTrustProviderId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessTrustProviderIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Trust Provider (%s): %s", d.Id(), err)
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
		names.AttrClientSecret: clientSecret,
	}

	if v := apiObject.AuthorizationEndpoint; v != nil {
		tfMap["authorization_endpoint"] = aws.ToString(v)
	}
	if v := apiObject.ClientId; v != nil {
		tfMap[names.AttrClientID] = aws.ToString(v)
	}
	if v := apiObject.Issuer; v != nil {
		tfMap[names.AttrIssuer] = aws.ToString(v)
	}
	if v := apiObject.Scope; v != nil {
		tfMap[names.AttrScope] = aws.ToString(v)
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
	if v, ok := tfMap[names.AttrClientID].(string); ok && v != "" {
		apiObject.ClientId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrClientSecret].(string); ok && v != "" {
		apiObject.ClientSecret = aws.String(v)
	}
	if v, ok := tfMap[names.AttrIssuer].(string); ok && v != "" {
		apiObject.Issuer = aws.String(v)
	}
	if v, ok := tfMap[names.AttrScope].(string); ok && v != "" {
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

	if v, ok := tfMap[names.AttrScope].(string); ok && v != "" {
		apiObject.Scope = aws.String(v)
	}

	return apiObject
}
