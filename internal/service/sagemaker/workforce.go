// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_workforce", name="Workforce")
func resourceWorkforce() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkforceCreate,
		ReadWithoutTimeout:   resourceWorkforceRead,
		UpdateWithoutTimeout: resourceWorkforceUpdate,
		DeleteWithoutTimeout: resourceWorkforceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cognito_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"oidc_config", "cognito_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrClientID: {
							Type:     schema.TypeString,
							Required: true,
						},
						"user_pool": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"oidc_config": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"oidc_config", "cognito_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_request_extra_params": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"authorization_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							),
						},
						names.AttrClientID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						names.AttrClientSecret: {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						names.AttrIssuer: {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"jwks_uri": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"logout_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						names.AttrScope: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"token_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							)},
						"user_info_endpoint": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 500),
								validation.IsURLWithHTTPS,
							),
						},
					},
				},
			},
			"source_ip_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidrs": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
						},
					},
				},
			},
			"subdomain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workforce_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([0-9A-Za-z-])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"workforce_vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 16,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCEndpointID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceWorkforceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("workforce_name").(string)
	input := &sagemaker.CreateWorkforceInput{
		WorkforceName: aws.String(name),
	}

	if v, ok := d.GetOk("cognito_config"); ok {
		input.CognitoConfig = expandWorkforceCognitoConfig(v.([]any))
	}

	if v, ok := d.GetOk("oidc_config"); ok {
		input.OidcConfig = expandWorkforceOIDCConfig(v.([]any))
	}

	if v, ok := d.GetOk("source_ip_config"); ok {
		input.SourceIpConfig = expandWorkforceSourceIPConfig(v.([]any))
	}

	if v, ok := d.GetOk("workforce_vpc_config"); ok {
		input.WorkforceVpcConfig = expandWorkforceVPCConfig(v.([]any))
	}

	_, err := conn.CreateWorkforce(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Workforce (%s): %s", name, err)
	}

	d.SetId(name)

	if err := waitWorkforceActive(ctx, conn, name); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Workforce (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkforceRead(ctx, d, meta)...)
}

func resourceWorkforceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	workforce, err := findWorkforceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Workforce (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Workforce (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, workforce.WorkforceArn)
	d.Set("subdomain", workforce.SubDomain)
	d.Set("workforce_name", workforce.WorkforceName)

	if err := d.Set("cognito_config", flattenWorkforceCognitoConfig(workforce.CognitoConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_config : %s", err)
	}

	if workforce.OidcConfig != nil {
		if err := d.Set("oidc_config", flattenWorkforceOIDCConfig(workforce.OidcConfig, d.Get("oidc_config.0.client_secret").(string))); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting oidc_config: %s", err)
		}
	}

	if err := d.Set("source_ip_config", flattenWorkforceSourceIPConfig(workforce.SourceIpConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source_ip_config: %s", err)
	}

	if err := d.Set("workforce_vpc_config", flattenWorkforceVPCConfig(workforce.WorkforceVpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting workforce_vpc_config: %s", err)
	}

	return diags
}

func resourceWorkforceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.UpdateWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	}

	if d.HasChange("source_ip_config") {
		input.SourceIpConfig = expandWorkforceSourceIPConfig(d.Get("source_ip_config").([]any))
	}

	if d.HasChange("oidc_config") {
		input.OidcConfig = expandWorkforceOIDCConfig(d.Get("oidc_config").([]any))
	}

	if d.HasChange("workforce_vpc_config") {
		input.WorkforceVpcConfig = expandWorkforceVPCConfig(d.Get("workforce_vpc_config").([]any))
	}

	_, err := conn.UpdateWorkforce(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Workforce (%s): %s", d.Id(), err)
	}

	if err := waitWorkforceActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Workforce (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceWorkforceRead(ctx, d, meta)...)
}

func resourceWorkforceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI Workforce: %s", d.Id())
	_, err := conn.DeleteWorkforce(ctx, &sagemaker.DeleteWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No workforce") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Workforce (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkforceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Workforce (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findWorkforceByName(ctx context.Context, conn *sagemaker.Client, name string) (*awstypes.Workforce, error) {
	input := &sagemaker.DescribeWorkforceInput{
		WorkforceName: aws.String(name),
	}

	output, err := conn.DescribeWorkforce(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No workforce") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workforce == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workforce, nil
}

func expandWorkforceSourceIPConfig(l []any) *awstypes.SourceIpConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.SourceIpConfig{
		Cidrs: flex.ExpandStringValueSet(m["cidrs"].(*schema.Set)),
	}

	return config
}

func flattenWorkforceSourceIPConfig(config *awstypes.SourceIpConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"cidrs": flex.FlattenStringValueSet(config.Cidrs),
	}

	return []map[string]any{m}
}

func expandWorkforceCognitoConfig(l []any) *awstypes.CognitoConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.CognitoConfig{
		ClientId: aws.String(m[names.AttrClientID].(string)),
		UserPool: aws.String(m["user_pool"].(string)),
	}

	return config
}

func flattenWorkforceCognitoConfig(config *awstypes.CognitoConfig) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrClientID: aws.ToString(config.ClientId),
		"user_pool":        aws.ToString(config.UserPool),
	}

	return []map[string]any{m}
}

func expandWorkforceOIDCConfig(l []any) *awstypes.OidcConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.OidcConfig{
		AuthorizationEndpoint: aws.String(m["authorization_endpoint"].(string)),
		ClientId:              aws.String(m[names.AttrClientID].(string)),
		ClientSecret:          aws.String(m[names.AttrClientSecret].(string)),
		Issuer:                aws.String(m[names.AttrIssuer].(string)),
		JwksUri:               aws.String(m["jwks_uri"].(string)),
		LogoutEndpoint:        aws.String(m["logout_endpoint"].(string)),
		TokenEndpoint:         aws.String(m["token_endpoint"].(string)),
		UserInfoEndpoint:      aws.String(m["user_info_endpoint"].(string)),
	}

	if v, ok := m["authentication_request_extra_params"].(map[string]any); ok && v != nil {
		config.AuthenticationRequestExtraParams = flex.ExpandStringValueMap(v)
	}

	if v, ok := m[names.AttrScope].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}

	return config
}

func flattenWorkforceOIDCConfig(config *awstypes.OidcConfigForResponse, clientSecret string) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"authentication_request_extra_params": aws.StringMap(config.AuthenticationRequestExtraParams),
		"authorization_endpoint":              aws.ToString(config.AuthorizationEndpoint),
		names.AttrClientID:                    aws.ToString(config.ClientId),
		names.AttrClientSecret:                clientSecret,
		names.AttrIssuer:                      aws.ToString(config.Issuer),
		"jwks_uri":                            aws.ToString(config.JwksUri),
		"logout_endpoint":                     aws.ToString(config.LogoutEndpoint),
		names.AttrScope:                       aws.ToString(config.Scope),
		"token_endpoint":                      aws.ToString(config.TokenEndpoint),
		"user_info_endpoint":                  aws.ToString(config.UserInfoEndpoint),
	}

	return []map[string]any{m}
}

func expandWorkforceVPCConfig(l []any) *awstypes.WorkforceVpcConfigRequest {
	if len(l) == 0 || l[0] == nil {
		return &awstypes.WorkforceVpcConfigRequest{}
	}

	m := l[0].(map[string]any)

	config := &awstypes.WorkforceVpcConfigRequest{
		SecurityGroupIds: flex.ExpandStringValueSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
		Subnets:          flex.ExpandStringValueSet(m[names.AttrSubnets].(*schema.Set)),
		VpcId:            aws.String(m[names.AttrVPCID].(string)),
	}

	return config
}

func flattenWorkforceVPCConfig(config *awstypes.WorkforceVpcConfigResponse) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrSecurityGroupIDs: flex.FlattenStringValueSet(config.SecurityGroupIds),
		names.AttrSubnets:          flex.FlattenStringValueSet(config.Subnets),
		names.AttrVPCEndpointID:    aws.ToString(config.VpcEndpointId),
		names.AttrVPCID:            aws.ToString(config.VpcId),
	}

	return []map[string]any{m}
}
