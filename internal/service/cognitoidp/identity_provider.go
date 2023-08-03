// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_identity_provider")
func ResourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIdentityProviderCreate,
		ReadWithoutTimeout:   resourceIdentityProviderRead,
		UpdateWithoutTimeout: resourceIdentityProviderUpdate,
		DeleteWithoutTimeout: resourceIdentityProviderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"attribute_mapping": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"idp_identifiers": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 40),
						validation.StringMatch(regexp.MustCompile(`^[\w\s+=.@-]+$`), "see https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#API_CreateIdentityProvider_RequestSyntax"),
					),
				},
			},

			"provider_details": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"provider_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[^_][\p{L}\p{M}\p{S}\p{N}\p{P}][^_]+$`), "see https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#API_CreateIdentityProvider_RequestSyntax"),
				),
			},

			"provider_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(cognitoidentityprovider.IdentityProviderTypeType_Values(), false),
			},

			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIdentityProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)
	log.Print("[DEBUG] Creating Cognito Identity Provider")

	providerName := d.Get("provider_name").(string)
	userPoolID := d.Get("user_pool_id").(string)
	params := &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName: aws.String(providerName),
		ProviderType: aws.String(d.Get("provider_type").(string)),
		UserPoolId:   aws.String(userPoolID),
	}

	if v, ok := d.GetOk("attribute_mapping"); ok {
		params.AttributeMapping = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("provider_details"); ok {
		params.ProviderDetails = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("idp_identifiers"); ok {
		params.IdpIdentifiers = flex.ExpandStringList(v.([]interface{}))
	}

	_, err := conn.CreateIdentityProviderWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito Identity Provider: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", userPoolID, providerName))

	return append(diags, resourceIdentityProviderRead(ctx, d, meta)...)
}

func resourceIdentityProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)
	log.Printf("[DEBUG] Reading Cognito Identity Provider: %s", d.Id())

	userPoolID, providerName, err := DecodeIdentityProviderID(d.Id())
	if err != nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameIdentityProvider, d.Id(), err)
	}

	ret, err := conn.DescribeIdentityProviderWithContext(ctx, &cognitoidentityprovider.DescribeIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameIdentityProvider, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameIdentityProvider, d.Id(), err)
	}

	if !d.IsNewResource() && (ret == nil || ret.IdentityProvider == nil) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameIdentityProvider, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && (ret == nil || ret.IdentityProvider == nil) {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameIdentityProvider, d.Id(), errors.New("not found after creation"))
	}

	ip := ret.IdentityProvider
	d.Set("provider_name", ip.ProviderName)
	d.Set("provider_type", ip.ProviderType)
	d.Set("user_pool_id", ip.UserPoolId)

	if err := d.Set("attribute_mapping", aws.StringValueMap(ip.AttributeMapping)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attribute_mapping error: %s", err)
	}

	if err := d.Set("provider_details", aws.StringValueMap(ip.ProviderDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting provider_details error: %s", err)
	}

	if err := d.Set("idp_identifiers", flex.FlattenStringList(ip.IdpIdentifiers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting idp_identifiers error: %s", err)
	}

	return diags
}

func resourceIdentityProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)
	log.Print("[DEBUG] Updating Cognito Identity Provider")

	userPoolID, providerName, err := DecodeIdentityProviderID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	params := &cognitoidentityprovider.UpdateIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	}

	if d.HasChange("attribute_mapping") {
		params.AttributeMapping = flex.ExpandStringMap(d.Get("attribute_mapping").(map[string]interface{}))
	}

	if d.HasChange("provider_details") {
		params.ProviderDetails = flex.ExpandStringMap(d.Get("provider_details").(map[string]interface{}))
	}

	if d.HasChange("idp_identifiers") {
		params.IdpIdentifiers = flex.ExpandStringList(d.Get("idp_identifiers").([]interface{}))
	}

	_, err = conn.UpdateIdentityProviderWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIdentityProviderRead(ctx, d, meta)...)
}

func resourceIdentityProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)
	log.Printf("[DEBUG] Deleting Cognito Identity Provider: %s", d.Id())

	userPoolID, providerName, err := DecodeIdentityProviderID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteIdentityProviderWithContext(ctx, &cognitoidentityprovider.DeleteIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeIdentityProviderID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format UserPoolID:ProviderName, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
