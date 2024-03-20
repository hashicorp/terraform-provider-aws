// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_apigatewayv2_authorizer")
func ResourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthorizerCreate,
		ReadWithoutTimeout:   resourceAuthorizerRead,
		UpdateWithoutTimeout: resourceAuthorizerUpdate,
		DeleteWithoutTimeout: resourceAuthorizerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAuthorizerImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorizer_credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"authorizer_payload_format_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1.0", "2.0"}, false),
			},
			"authorizer_result_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 3600),
			},
			"authorizer_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthorizerType](),
			},
			"authorizer_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"enable_simple_responses": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"identity_sources": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"jwt_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audience": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"issuer": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceAuthorizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiId := d.Get("api_id").(string)
	authorizerType := d.Get("authorizer_type").(string)

	apiOutput, err := FindAPIByID(ctx, conn, apiId)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", apiId, err)
	}

	protocolType := apiOutput.ProtocolType

	req := &apigatewayv2.CreateAuthorizerInput{
		ApiId:          aws.String(apiId),
		AuthorizerType: awstypes.AuthorizerType(authorizerType),
		IdentitySource: flex.ExpandStringValueSet(d.Get("identity_sources").(*schema.Set)),
		Name:           aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("authorizer_credentials_arn"); ok {
		req.AuthorizerCredentialsArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("authorizer_payload_format_version"); ok {
		req.AuthorizerPayloadFormatVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOkExists("authorizer_result_ttl_in_seconds"); ok {
		req.AuthorizerResultTtlInSeconds = aws.Int32(int32(v.(int)))
	} else if protocolType == awstypes.ProtocolTypeHttp && authorizerType == string(awstypes.AuthorizerTypeRequest) && len(req.IdentitySource) > 0 {
		// Default in the AWS Console is 300 seconds.
		// Explicitly set on creation so that we can correctly detect changes to the 0 value.
		// This value should only be set when IdentitySources have been defined
		req.AuthorizerResultTtlInSeconds = aws.Int32(300)
	}
	if v, ok := d.GetOk("authorizer_uri"); ok {
		req.AuthorizerUri = aws.String(v.(string))
	}
	if v, ok := d.GetOk("enable_simple_responses"); ok {
		req.EnableSimpleResponses = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("jwt_configuration"); ok {
		req.JwtConfiguration = expandJWTConfiguration(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 authorizer: %+v", req)
	resp, err := conn.CreateAuthorizer(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 authorizer: %s", err)
	}

	d.SetId(aws.ToString(resp.AuthorizerId))

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	resp, err := conn.GetAuthorizer(ctx, &apigatewayv2.GetAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 authorizer: %s", err)
	}

	d.Set("authorizer_credentials_arn", resp.AuthorizerCredentialsArn)
	d.Set("authorizer_payload_format_version", resp.AuthorizerPayloadFormatVersion)
	d.Set("authorizer_result_ttl_in_seconds", resp.AuthorizerResultTtlInSeconds)
	d.Set("authorizer_type", resp.AuthorizerType)
	d.Set("authorizer_uri", resp.AuthorizerUri)
	d.Set("enable_simple_responses", resp.EnableSimpleResponses)
	if err := d.Set("identity_sources", flex.FlattenStringValueSet(resp.IdentitySource)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting identity_sources: %s", err)
	}
	if err := d.Set("jwt_configuration", flattenJWTConfiguration(resp.JwtConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting jwt_configuration: %s", err)
	}
	d.Set("name", resp.Name)

	return diags
}

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.UpdateAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	}
	if d.HasChange("authorizer_credentials_arn") {
		req.AuthorizerCredentialsArn = aws.String(d.Get("authorizer_credentials_arn").(string))
	}
	if d.HasChange("authorizer_payload_format_version") {
		req.AuthorizerPayloadFormatVersion = aws.String(d.Get("authorizer_payload_format_version").(string))
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		req.AuthorizerResultTtlInSeconds = aws.Int32(int32(d.Get("authorizer_result_ttl_in_seconds").(int)))
	}
	if d.HasChange("authorizer_type") {
		req.AuthorizerType = awstypes.AuthorizerType(d.Get("authorizer_type").(string))
	}
	if d.HasChange("authorizer_uri") {
		req.AuthorizerUri = aws.String(d.Get("authorizer_uri").(string))
	}
	if d.HasChange("enable_simple_responses") {
		req.EnableSimpleResponses = aws.Bool(d.Get("enable_simple_responses").(bool))
	}
	if d.HasChange("identity_sources") {
		req.IdentitySource = flex.ExpandStringValueSet(d.Get("identity_sources").(*schema.Set))
	}
	if d.HasChange("name") {
		req.Name = aws.String(d.Get("name").(string))
	}
	if d.HasChange("jwt_configuration") {
		req.JwtConfiguration = expandJWTConfiguration(d.Get("jwt_configuration").([]interface{}))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 authorizer: %+v", req)
	_, err := conn.UpdateAuthorizer(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 authorizer: %s", err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 authorizer (%s)", d.Id())
	_, err := conn.DeleteAuthorizer(ctx, &apigatewayv2.DeleteAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 authorizer: %s", err)
	}

	return diags
}

func resourceAuthorizerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/authorizer-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}

func expandJWTConfiguration(vConfiguration []interface{}) *awstypes.JWTConfiguration {
	configuration := &awstypes.JWTConfiguration{}

	if len(vConfiguration) == 0 || vConfiguration[0] == nil {
		return configuration
	}
	mConfiguration := vConfiguration[0].(map[string]interface{})

	if vAudience, ok := mConfiguration["audience"].(*schema.Set); ok && vAudience.Len() > 0 {
		configuration.Audience = flex.ExpandStringValueSet(vAudience)
	}
	if vIssuer, ok := mConfiguration["issuer"].(string); ok && vIssuer != "" {
		configuration.Issuer = aws.String(vIssuer)
	}

	return configuration
}

func flattenJWTConfiguration(configuration *awstypes.JWTConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"audience": flex.FlattenStringValueSet(configuration.Audience),
		"issuer":   aws.ToString(configuration.Issuer),
	}}
}
