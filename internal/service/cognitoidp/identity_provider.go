// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_identity_provider", name="Identity Provider")
func resourceIdentityProvider() *schema.Resource {
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
						validation.StringMatch(regexache.MustCompile(`^[\w\s+=.@-]+$`), "see https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#API_CreateIdentityProvider_RequestSyntax"),
					),
				},
			},
			"provider_details": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexache.MustCompile(`^[^_][\p{L}\p{M}\p{S}\p{N}\p{P}][^_]+$`), "see https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateIdentityProvider.html#API_CreateIdentityProvider_RequestSyntax"),
				),
			},
			"provider_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IdentityProviderTypeType](),
			},
			names.AttrUserPoolID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIdentityProviderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	providerName := d.Get(names.AttrProviderName).(string)
	userPoolID := d.Get(names.AttrUserPoolID).(string)
	id := identityProviderCreateResourceID(userPoolID, providerName)
	input := &cognitoidentityprovider.CreateIdentityProviderInput{
		ProviderName: aws.String(providerName),
		ProviderType: awstypes.IdentityProviderTypeType(d.Get("provider_type").(string)),
		UserPoolId:   aws.String(userPoolID),
	}

	if v, ok := d.GetOk("attribute_mapping"); ok && len(v.(map[string]interface{})) > 0 {
		input.AttributeMapping = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("idp_identifiers"); ok && len(v.([]interface{})) > 0 {
		input.IdpIdentifiers = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("provider_details"); ok && len(v.(map[string]interface{})) > 0 {
		input.ProviderDetails = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	_, err := conn.CreateIdentityProvider(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito Identity Provider (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceIdentityProviderRead(ctx, d, meta)...)
}

func resourceIdentityProviderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, providerName, err := identityProviderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	idp, err := findIdentityProviderByTwoPartKey(ctx, conn, userPoolID, providerName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito Identity Provider %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	d.Set("attribute_mapping", idp.AttributeMapping)
	d.Set("idp_identifiers", idp.IdpIdentifiers)
	d.Set("provider_details", idp.ProviderDetails)
	d.Set(names.AttrProviderName, idp.ProviderName)
	d.Set("provider_type", idp.ProviderType)
	d.Set(names.AttrUserPoolID, idp.UserPoolId)

	return diags
}

func resourceIdentityProviderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, providerName, err := identityProviderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &cognitoidentityprovider.UpdateIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	}

	if d.HasChange("attribute_mapping") {
		input.AttributeMapping = flex.ExpandStringValueMap(d.Get("attribute_mapping").(map[string]interface{}))
	}

	if d.HasChange("idp_identifiers") {
		input.IdpIdentifiers = flex.ExpandStringValueList(d.Get("idp_identifiers").([]interface{}))
	}

	if d.HasChange("provider_details") {
		v := flex.ExpandStringValueMap(d.Get("provider_details").(map[string]interface{}))
		delete(v, "ActiveEncryptionCertificate")
		input.ProviderDetails = v
	}

	_, err = conn.UpdateIdentityProvider(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIdentityProviderRead(ctx, d, meta)...)
}

func resourceIdentityProviderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, providerName, err := identityProviderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Cognito Identity Provider: %s", d.Id())
	_, err = conn.DeleteIdentityProvider(ctx, &cognitoidentityprovider.DeleteIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Identity Provider (%s): %s", d.Id(), err)
	}

	return diags
}

const identityProviderResourceIDSeparator = ":"

func identityProviderCreateResourceID(userPoolID, providerName string) string {
	parts := []string{userPoolID, providerName}
	id := strings.Join(parts, identityProviderResourceIDSeparator)

	return id
}

func identityProviderParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, identityProviderResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected UserPoolID%[2]sProviderName", id, identityProviderResourceIDSeparator)
}

func findIdentityProviderByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, providerName string) (*awstypes.IdentityProviderType, error) {
	input := &cognitoidentityprovider.DescribeIdentityProviderInput{
		ProviderName: aws.String(providerName),
		UserPoolId:   aws.String(userPoolID),
	}

	output, err := conn.DescribeIdentityProvider(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IdentityProvider == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.IdentityProvider, nil
}
