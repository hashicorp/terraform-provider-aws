// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_authorizer", name="Authorizer")
func resourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthorizerCreate,
		ReadWithoutTimeout:   resourceAuthorizerRead,
		UpdateWithoutTimeout: resourceAuthorizerUpdate,
		DeleteWithoutTimeout: resourceAuthorizerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAuthorizerImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(30 * time.Minute),
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
						names.AttrIssuer: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrName: {
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

	apiID := d.Get("api_id").(string)
	outputGA, err := findAPIByID(ctx, conn, apiID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", apiID, err)
	}

	authorizerType := awstypes.AuthorizerType(d.Get("authorizer_type").(string))
	name := d.Get(names.AttrName).(string)
	protocolType := outputGA.ProtocolType
	input := &apigatewayv2.CreateAuthorizerInput{
		ApiId:          aws.String(apiID),
		AuthorizerType: authorizerType,
		IdentitySource: flex.ExpandStringValueSet(d.Get("identity_sources").(*schema.Set)),
		Name:           aws.String(name),
	}

	if v, ok := d.GetOk("authorizer_credentials_arn"); ok {
		input.AuthorizerCredentialsArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("authorizer_payload_format_version"); ok {
		input.AuthorizerPayloadFormatVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("authorizer_result_ttl_in_seconds"); ok {
		input.AuthorizerResultTtlInSeconds = aws.Int32(int32(v.(int)))
	} else if protocolType == awstypes.ProtocolTypeHttp && authorizerType == awstypes.AuthorizerTypeRequest && len(input.IdentitySource) > 0 {
		// Default in the AWS Console is 300 seconds.
		// Explicitly set on creation so that we can correctly detect changes to the 0 value.
		// This value should only be set when IdentitySources have been defined
		input.AuthorizerResultTtlInSeconds = aws.Int32(300)
	}

	if v, ok := d.GetOk("authorizer_uri"); ok {
		input.AuthorizerUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_simple_responses"); ok {
		input.EnableSimpleResponses = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("jwt_configuration"); ok {
		input.JwtConfiguration = expandJWTConfiguration(v.([]interface{}))
	}

	outputCA, err := conn.CreateAuthorizer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Authorizer (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputCA.AuthorizerId))

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findAuthorizerByTwoPartKey(ctx, conn, d.Get("api_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 Authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Authorizer (%s): %s", d.Id(), err)
	}

	d.Set("authorizer_credentials_arn", output.AuthorizerCredentialsArn)
	d.Set("authorizer_payload_format_version", output.AuthorizerPayloadFormatVersion)
	d.Set("authorizer_result_ttl_in_seconds", output.AuthorizerResultTtlInSeconds)
	d.Set("authorizer_type", output.AuthorizerType)
	d.Set("authorizer_uri", output.AuthorizerUri)
	d.Set("enable_simple_responses", output.EnableSimpleResponses)
	d.Set("identity_sources", output.IdentitySource)
	if err := d.Set("jwt_configuration", flattenJWTConfiguration(output.JwtConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting jwt_configuration: %s", err)
	}
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	input := &apigatewayv2.UpdateAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	}

	if d.HasChange("authorizer_credentials_arn") {
		input.AuthorizerCredentialsArn = aws.String(d.Get("authorizer_credentials_arn").(string))
	}

	if d.HasChange("authorizer_payload_format_version") {
		input.AuthorizerPayloadFormatVersion = aws.String(d.Get("authorizer_payload_format_version").(string))
	}

	if d.HasChange("authorizer_result_ttl_in_seconds") {
		input.AuthorizerResultTtlInSeconds = aws.Int32(int32(d.Get("authorizer_result_ttl_in_seconds").(int)))
	}

	if d.HasChange("authorizer_type") {
		input.AuthorizerType = awstypes.AuthorizerType(d.Get("authorizer_type").(string))
	}

	if d.HasChange("authorizer_uri") {
		input.AuthorizerUri = aws.String(d.Get("authorizer_uri").(string))
	}

	if d.HasChange("enable_simple_responses") {
		input.EnableSimpleResponses = aws.Bool(d.Get("enable_simple_responses").(bool))
	}

	if d.HasChange("identity_sources") {
		input.IdentitySource = flex.ExpandStringValueSet(d.Get("identity_sources").(*schema.Set))
	}

	if d.HasChange(names.AttrName) {
		input.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange("jwt_configuration") {
		input.JwtConfiguration = expandJWTConfiguration(d.Get("jwt_configuration").([]interface{}))
	}

	_, err := conn.UpdateAuthorizer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Authorizer (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Authorizer: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteAuthorizer(ctx, &apigatewayv2.DeleteAuthorizerInput{
			ApiId:        aws.String(d.Get("api_id").(string)),
			AuthorizerId: aws.String(d.Id()),
		})
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Authorizer (%s): %s", d.Id(), err)
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

func findAuthorizerByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, authorizerID string) (*apigatewayv2.GetAuthorizerOutput, error) {
	input := &apigatewayv2.GetAuthorizerInput{
		ApiId:        aws.String(apiID),
		AuthorizerId: aws.String(authorizerID),
	}

	return findAuthorizer(ctx, conn, input)
}

func findAuthorizer(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetAuthorizerInput) (*apigatewayv2.GetAuthorizerOutput, error) {
	output, err := conn.GetAuthorizer(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func expandJWTConfiguration(vConfiguration []interface{}) *awstypes.JWTConfiguration {
	configuration := &awstypes.JWTConfiguration{}

	if len(vConfiguration) == 0 || vConfiguration[0] == nil {
		return configuration
	}
	mConfiguration := vConfiguration[0].(map[string]interface{})

	if vAudience, ok := mConfiguration["audience"].(*schema.Set); ok && vAudience.Len() > 0 {
		configuration.Audience = flex.ExpandStringValueSet(vAudience)
	}
	if vIssuer, ok := mConfiguration[names.AttrIssuer].(string); ok && vIssuer != "" {
		configuration.Issuer = aws.String(vIssuer)
	}

	return configuration
}

func flattenJWTConfiguration(configuration *awstypes.JWTConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"audience":       flex.FlattenStringValueSet(configuration.Audience),
		names.AttrIssuer: aws.ToString(configuration.Issuer),
	}}
}
