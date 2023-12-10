// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssoadmin_trusted_token_issuer")
// @Tags
func ResourceTrustedTokenIssuer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrustedTokenIssuerCreate,
		ReadWithoutTimeout:   resourceTrustedTokenIssuerRead,
		UpdateWithoutTimeout: resourceTrustedTokenIssuerUpdate,
		DeleteWithoutTimeout: resourceTrustedTokenIssuerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_token": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"trusted_token_issuer_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oidc_jwt_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"claim_attribute_path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"identity_store_attribute_path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"issuer_url": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true, // Not part of OidcJwtUpdateConfiguration struct, have to recreate at change
									},
									"jwks_retrieval_option": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.JwksRetrievalOption](),
									},
								},
							},
						},
					},
				},
			},
			"trusted_token_issuer_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.TrustedTokenIssuerType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTrustedTokenIssuerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	input := &ssoadmin.CreateTrustedTokenIssuerInput{
		InstanceArn:                     aws.String(d.Get("instance_arn").(string)),
		Name:                            aws.String(d.Get("name").(string)),
		TrustedTokenIssuerConfiguration: expandTrustedTokenIssuerConfiguration(d.Get("trusted_token_issuer_configuration").([]interface{})),
		TrustedTokenIssuerType:          types.TrustedTokenIssuerType(d.Get("trusted_token_issuer_type").(string)),
		Tags:                            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("client_token"); ok {
		input.ClientToken = aws.String(v.(string))
	}

	output, err := conn.CreateTrustedTokenIssuer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Trusted Token Issuer (%s): %s", d.Get("name").(string), err)
	}

	d.SetId(aws.ToString(output.TrustedTokenIssuerArn))

	return append(diags, resourceTrustedTokenIssuerRead(ctx, d, meta)...)
}

func resourceTrustedTokenIssuerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	output, err := FindTrustedTokenIssuerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Trusted Token Issuer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Trusted Token Issuer (%s): %s", d.Id(), err)
	}

	instanceARN, _ := TrustedTokenIssuerParseInstanceARN(meta.(*conns.AWSClient), d.Id())

	d.Set("name", output.Name)
	d.Set("arn", output.TrustedTokenIssuerArn)
	d.Set("instance_arn", instanceARN)
	d.Set("trusted_token_issuer_configuration", flattenTrustedTokenIssuerConfiguration(output.TrustedTokenIssuerConfiguration))
	d.Set("trusted_token_issuer_type", output.TrustedTokenIssuerType)

	// listTags requires both trusted token issuer and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	tags, err := listTags(ctx, conn, d.Id(), instanceARN)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Trusted Token Issuer (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourceTrustedTokenIssuerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ssoadmin.UpdateTrustedTokenIssuerInput{
			TrustedTokenIssuerArn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("trusted_token_issuer_configuration") {
			input.TrustedTokenIssuerConfiguration = expandTrustedTokenIssuerUpdateConfiguration(d.Get("trusted_token_issuer_configuration").([]interface{}))
		}

		_, err := conn.UpdateTrustedTokenIssuer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSO Trusted Token Issuer (%s): %s", d.Id(), err)
		}
	}

	// updateTags requires both trusted token issuer and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	if d.HasChange("tags_all") {
		oldTagsAll, newTagsAll := d.GetChange("tags_all")
		if err := updateTags(ctx, conn, d.Id(), d.Get("instance_arn").(string), oldTagsAll, newTagsAll); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SSO Trusted Token Issuer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTrustedTokenIssuerRead(ctx, d, meta)...)
}

func resourceTrustedTokenIssuerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	input := &ssoadmin.DeleteTrustedTokenIssuerInput{
		TrustedTokenIssuerArn: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting SSO Trusted Token Issuer: %s", d.Id())
	_, err := conn.DeleteTrustedTokenIssuer(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Trusted Token Issuer (%s): %s", d.Id(), err)
	}

	return diags
}

func FindTrustedTokenIssuerByARN(ctx context.Context, conn *ssoadmin.Client, trustedTokenIssuerARN string) (*ssoadmin.DescribeTrustedTokenIssuerOutput, error) {
	input := &ssoadmin.DescribeTrustedTokenIssuerInput{
		TrustedTokenIssuerArn: aws.String(trustedTokenIssuerARN),
	}

	output, err := conn.DescribeTrustedTokenIssuer(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

// Instance ARN is not returned by DescribeTrustedTokenIssuer but is needed for schema consistency when importing and tagging.
// Instance ARN can be extracted from the Trusted Token Issuer ARN.
func TrustedTokenIssuerParseInstanceARN(conn *conns.AWSClient, id string) (string, error) {
	parts := strings.Split(id, "/")

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return fmt.Sprintf("arn:%s:sso:::instance/%s", conn.Partition, parts[1]), nil
	}

	return "", fmt.Errorf("unable to construct Instance ARN from Trusted Token Issuer ARN: %s", id)
}

func expandTrustedTokenIssuerConfiguration(tfMap []interface{}) types.TrustedTokenIssuerConfiguration {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfList["oidc_jwt_configuration"]; ok {
		return &types.TrustedTokenIssuerConfigurationMemberOidcJwtConfiguration{
			Value: expandOIDCJWTConfiguration(v.([]interface{})),
		}
	}

	return nil
}

func expandOIDCJWTConfiguration(tfMap []interface{}) types.OidcJwtConfiguration {
	apiObject := types.OidcJwtConfiguration{}

	if len(tfMap) == 0 {
		return apiObject
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return apiObject
	}

	if v, ok := tfList["claim_attribute_path"]; ok {
		apiObject.ClaimAttributePath = aws.String(v.(string))
	}

	if v, ok := tfList["identity_store_attribute_path"]; ok {
		apiObject.IdentityStoreAttributePath = aws.String(v.(string))
	}

	if v, ok := tfList["issuer_url"]; ok {
		apiObject.IssuerUrl = aws.String(v.(string))
	}

	if v, ok := tfList["jwks_retrieval_option"]; ok {
		apiObject.JwksRetrievalOption = types.JwksRetrievalOption(v.(string))
	}

	return apiObject
}

func expandTrustedTokenIssuerUpdateConfiguration(tfMap []interface{}) types.TrustedTokenIssuerUpdateConfiguration {
	if len(tfMap) == 0 {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfList["oidc_jwt_configuration"]; ok {
		return &types.TrustedTokenIssuerUpdateConfigurationMemberOidcJwtConfiguration{
			Value: expandOIDCJWTUpdateConfiguration(v.([]interface{})),
		}
	}

	return nil
}

func expandOIDCJWTUpdateConfiguration(tfMap []interface{}) types.OidcJwtUpdateConfiguration {
	apiObject := types.OidcJwtUpdateConfiguration{}

	if len(tfMap) == 0 {
		return apiObject
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return apiObject
	}

	if v, ok := tfList["claim_attribute_path"]; ok {
		apiObject.ClaimAttributePath = aws.String(v.(string))
	}

	if v, ok := tfList["identity_store_attribute_path"]; ok {
		apiObject.IdentityStoreAttributePath = aws.String(v.(string))
	}

	if v, ok := tfList["jwks_retrieval_option"]; ok {
		apiObject.JwksRetrievalOption = types.JwksRetrievalOption(v.(string))
	}

	return apiObject
}

func flattenTrustedTokenIssuerConfiguration(apiObject types.TrustedTokenIssuerConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *types.TrustedTokenIssuerConfigurationMemberOidcJwtConfiguration:
		tfMap["oidc_jwt_configuration"] = flattenOIDCJWTConfiguration(v.Value)
	default:
		log.Println("union is nil or unknown type")
	}

	return []interface{}{tfMap}
}

func flattenOIDCJWTConfiguration(apiObject types.OidcJwtConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.ClaimAttributePath; v != nil {
		tfMap["claim_attribute_path"] = aws.ToString(v)
	}

	if v := apiObject.IdentityStoreAttributePath; v != nil {
		tfMap["identity_store_attribute_path"] = aws.ToString(v)
	}

	if v := apiObject.IssuerUrl; v != nil {
		tfMap["issuer_url"] = aws.ToString(v)
	}

	tfMap["jwks_retrieval_option"] = string(apiObject.JwksRetrievalOption)

	return []interface{}{tfMap}
}
