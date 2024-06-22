// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_workforce")
func ResourceWorkforce() *schema.Resource {
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

func resourceWorkforceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("workforce_name").(string)
	input := &sagemaker.CreateWorkforceInput{
		WorkforceName: aws.String(name),
	}

	if v, ok := d.GetOk("cognito_config"); ok {
		input.CognitoConfig = expandWorkforceCognitoConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("oidc_config"); ok {
		input.OidcConfig = expandWorkforceOIDCConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_ip_config"); ok {
		input.SourceIpConfig = expandWorkforceSourceIPConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("workforce_vpc_config"); ok {
		input.WorkforceVpcConfig = expandWorkforceVPCConfig(v.([]interface{}))
	}

	_, err := conn.CreateWorkforceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Workforce (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := WaitWorkforceActive(ctx, conn, name); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Workforce (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkforceRead(ctx, d, meta)...)
}

func resourceWorkforceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	workforce, err := FindWorkforceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Workforce (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Workforce (%s): %s", d.Id(), err)
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

func resourceWorkforceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.UpdateWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	}

	if d.HasChange("source_ip_config") {
		input.SourceIpConfig = expandWorkforceSourceIPConfig(d.Get("source_ip_config").([]interface{}))
	}

	if d.HasChange("oidc_config") {
		input.OidcConfig = expandWorkforceOIDCConfig(d.Get("oidc_config").([]interface{}))
	}

	if d.HasChange("workforce_vpc_config") {
		input.WorkforceVpcConfig = expandWorkforceVPCConfig(d.Get("workforce_vpc_config").([]interface{}))
	}

	_, err := conn.UpdateWorkforceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker Workforce (%s): %s", d.Id(), err)
	}

	if _, err := WaitWorkforceActive(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Workforce (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceWorkforceRead(ctx, d, meta)...)
}

func resourceWorkforceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	log.Printf("[DEBUG] Deleting SageMaker Workforce: %s", d.Id())
	_, err := conn.DeleteWorkforceWithContext(ctx, &sagemaker.DeleteWorkforceInput{
		WorkforceName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, "ValidationException", "No workforce") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Workforce (%s): %s", d.Id(), err)
	}

	if _, err := WaitWorkforceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Workforce (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandWorkforceSourceIPConfig(l []interface{}) *sagemaker.SourceIpConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SourceIpConfig{
		Cidrs: flex.ExpandStringSet(m["cidrs"].(*schema.Set)),
	}

	return config
}

func flattenWorkforceSourceIPConfig(config *sagemaker.SourceIpConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cidrs": flex.FlattenStringSet(config.Cidrs),
	}

	return []map[string]interface{}{m}
}

func expandWorkforceCognitoConfig(l []interface{}) *sagemaker.CognitoConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.CognitoConfig{
		ClientId: aws.String(m[names.AttrClientID].(string)),
		UserPool: aws.String(m["user_pool"].(string)),
	}

	return config
}

func flattenWorkforceCognitoConfig(config *sagemaker.CognitoConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrClientID: aws.StringValue(config.ClientId),
		"user_pool":        aws.StringValue(config.UserPool),
	}

	return []map[string]interface{}{m}
}

func expandWorkforceOIDCConfig(l []interface{}) *sagemaker.OidcConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OidcConfig{
		AuthorizationEndpoint: aws.String(m["authorization_endpoint"].(string)),
		ClientId:              aws.String(m[names.AttrClientID].(string)),
		ClientSecret:          aws.String(m[names.AttrClientSecret].(string)),
		Issuer:                aws.String(m[names.AttrIssuer].(string)),
		JwksUri:               aws.String(m["jwks_uri"].(string)),
		LogoutEndpoint:        aws.String(m["logout_endpoint"].(string)),
		TokenEndpoint:         aws.String(m["token_endpoint"].(string)),
		UserInfoEndpoint:      aws.String(m["user_info_endpoint"].(string)),
	}

	if v, ok := m["authentication_request_extra_params"].(map[string]interface{}); ok && v != nil {
		config.AuthenticationRequestExtraParams = flex.ExpandStringMap(v)
	}

	if v, ok := m[names.AttrScope].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}

	return config
}

func flattenWorkforceOIDCConfig(config *sagemaker.OidcConfigForResponse, clientSecret string) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"authentication_request_extra_params": aws.StringValueMap(config.AuthenticationRequestExtraParams),
		"authorization_endpoint":              aws.StringValue(config.AuthorizationEndpoint),
		names.AttrClientID:                    aws.StringValue(config.ClientId),
		names.AttrClientSecret:                clientSecret,
		names.AttrIssuer:                      aws.StringValue(config.Issuer),
		"jwks_uri":                            aws.StringValue(config.JwksUri),
		"logout_endpoint":                     aws.StringValue(config.LogoutEndpoint),
		names.AttrScope:                       aws.StringValue(config.Scope),
		"token_endpoint":                      aws.StringValue(config.TokenEndpoint),
		"user_info_endpoint":                  aws.StringValue(config.UserInfoEndpoint),
	}

	return []map[string]interface{}{m}
}

func expandWorkforceVPCConfig(l []interface{}) *sagemaker.WorkforceVpcConfigRequest {
	if len(l) == 0 || l[0] == nil {
		return &sagemaker.WorkforceVpcConfigRequest{}
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.WorkforceVpcConfigRequest{
		SecurityGroupIds: flex.ExpandStringSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
		Subnets:          flex.ExpandStringSet(m[names.AttrSubnets].(*schema.Set)),
		VpcId:            aws.String(m[names.AttrVPCID].(string)),
	}

	return config
}

func flattenWorkforceVPCConfig(config *sagemaker.WorkforceVpcConfigResponse) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrSecurityGroupIDs: flex.FlattenStringSet(config.SecurityGroupIds),
		names.AttrSubnets:          flex.FlattenStringSet(config.Subnets),
		names.AttrVPCEndpointID:    aws.StringValue(config.VpcEndpointId),
		names.AttrVPCID:            aws.StringValue(config.VpcId),
	}

	return []map[string]interface{}{m}
}
