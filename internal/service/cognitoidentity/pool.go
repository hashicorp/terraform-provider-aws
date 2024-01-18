// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_identity_pool", name="Pool")
// @Tags(identifierAttribute="arn")
func ResourcePool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePoolCreate,
		ReadWithoutTimeout:   resourcePoolRead,
		UpdateWithoutTimeout: resourcePoolUpdate,
		DeleteWithoutTimeout: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"identity_pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentityPoolName,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cognito_identity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validIdentityProvidersClientID,
						},
						"provider_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validIdentityProvidersProviderName,
						},
						"server_side_token_check": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"developer_provider_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // Forcing a new resource since it cannot be edited afterwards
				ValidateFunc: validProviderDeveloperName,
			},

			"allow_unauthenticated_identities": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"allow_classic_flow": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"openid_connect_provider_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},

			"saml_provider_arns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},

			"supported_login_providers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validSupportedLoginProviders,
				},
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn(ctx)

	input := &cognitoidentity.CreateIdentityPoolInput{
		IdentityPoolName:               aws.String(d.Get("identity_pool_name").(string)),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated_identities").(bool)),
		AllowClassicFlow:               aws.Bool(d.Get("allow_classic_flow").(bool)),
		IdentityPoolTags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("developer_provider_name"); ok {
		input.DeveloperProviderName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_login_providers"); ok {
		input.SupportedLoginProviders = expandSupportedLoginProviders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("cognito_identity_providers"); ok {
		input.CognitoIdentityProviders = expandIdentityProviders(v.(*schema.Set))
	}

	if v, ok := d.GetOk("saml_provider_arns"); ok {
		input.SamlProviderARNs = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_provider_arns"); ok {
		input.OpenIdConnectProviderARNs = flex.ExpandStringSet(v.(*schema.Set))
	}

	entity, err := conn.CreateIdentityPoolWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito Identity Pool: %s", err)
	}

	d.SetId(aws.StringValue(entity.IdentityPoolId))

	return append(diags, resourcePoolRead(ctx, d, meta)...)
}

func resourcePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn(ctx)

	ip, err := conn.DescribeIdentityPoolWithContext(ctx, &cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentity.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CognitoIdentity, create.ErrActionReading, ResNamePool, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CognitoIdentity, create.ErrActionReading, ResNamePool, d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "cognito-identity",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identitypool/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("identity_pool_name", ip.IdentityPoolName)
	d.Set("allow_unauthenticated_identities", ip.AllowUnauthenticatedIdentities)
	d.Set("allow_classic_flow", ip.AllowClassicFlow)
	d.Set("developer_provider_name", ip.DeveloperProviderName)

	setTagsOut(ctx, ip.IdentityPoolTags)

	if err := d.Set("cognito_identity_providers", flattenIdentityProviders(ip.CognitoIdentityProviders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_identity_providers error: %s", err)
	}

	if err := d.Set("openid_connect_provider_arns", flex.FlattenStringList(ip.OpenIdConnectProviderARNs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting openid_connect_provider_arns error: %s", err)
	}

	if err := d.Set("saml_provider_arns", flex.FlattenStringList(ip.SamlProviderARNs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_provider_arns error: %s", err)
	}

	if err := d.Set("supported_login_providers", aws.StringValueMap(ip.SupportedLoginProviders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting supported_login_providers error: %s", err)
	}

	return diags
}

func resourcePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn(ctx)
	log.Print("[DEBUG] Updating Cognito Identity Pool")

	if d.HasChangesExcept("tags_all", "tags") {
		params := &cognitoidentity.IdentityPool{
			IdentityPoolId:                 aws.String(d.Id()),
			AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated_identities").(bool)),
			AllowClassicFlow:               aws.Bool(d.Get("allow_classic_flow").(bool)),
			IdentityPoolName:               aws.String(d.Get("identity_pool_name").(string)),
			CognitoIdentityProviders:       expandIdentityProviders(d.Get("cognito_identity_providers").(*schema.Set)),
			SupportedLoginProviders:        expandSupportedLoginProviders(d.Get("supported_login_providers").(map[string]interface{})),
			OpenIdConnectProviderARNs:      flex.ExpandStringSet(d.Get("openid_connect_provider_arns").(*schema.Set)),
			SamlProviderARNs:               flex.ExpandStringList(d.Get("saml_provider_arns").([]interface{})),
		}

		_, err := conn.UpdateIdentityPoolWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cognito Identity Pool (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePoolRead(ctx, d, meta)...)
}

func resourcePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIdentityConn(ctx)
	log.Printf("[DEBUG] Deleting Cognito Identity Pool: %s", d.Id())

	_, err := conn.DeleteIdentityPoolWithContext(ctx, &cognitoidentity.DeleteIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito identity pool (%s): %s", d.Id(), err)
	}
	return diags
}
