// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_connection", name="Connection")
func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			connectionHttpParameters := func(parent string) *schema.Resource {
				element := func() *schema.Resource {
					return &schema.Resource{
						Schema: map[string]*schema.Schema{
							"is_value_secret": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							names.AttrKey: {
								Type:     schema.TypeString,
								Optional: true,
							},
							names.AttrValue: {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
							},
						},
					}
				}
				atLeastOneOf := []string{
					fmt.Sprintf("%s.0.body", parent),
					fmt.Sprintf("%s.0.header", parent),
					fmt.Sprintf("%s.0.query_string", parent),
				}

				return &schema.Resource{
					Schema: map[string]*schema.Schema{
						"body": {
							Type:         schema.TypeList,
							Optional:     true,
							Elem:         element(),
							AtLeastOneOf: atLeastOneOf,
						},
						names.AttrHeader: {
							Type:         schema.TypeList,
							Optional:     true,
							Elem:         element(),
							AtLeastOneOf: atLeastOneOf,
						},
						"query_string": {
							Type:         schema.TypeList,
							Optional:     true,
							Elem:         element(),
							AtLeastOneOf: atLeastOneOf,
						},
					},
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"auth_parameters": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"api_key": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								ExactlyOneOf: []string{
									"auth_parameters.0.api_key",
									"auth_parameters.0.basic",
									"auth_parameters.0.oauth",
								},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrKey: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 512),
											),
										},
										names.AttrValue: {
											Type:      schema.TypeString,
											Required:  true,
											Sensitive: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 512),
											),
										},
									},
								},
							},
							"basic": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								ExactlyOneOf: []string{
									"auth_parameters.0.api_key",
									"auth_parameters.0.basic",
									"auth_parameters.0.oauth",
								},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrPassword: {
											Type:      schema.TypeString,
											Required:  true,
											Sensitive: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 512),
											),
										},
										names.AttrUsername: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 512),
											),
										},
									},
								},
							},
							"invocation_http_parameters": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem:     connectionHttpParameters("auth_parameters.0.invocation_http_parameters"),
							},
							"oauth": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								ExactlyOneOf: []string{
									"auth_parameters.0.api_key",
									"auth_parameters.0.basic",
									"auth_parameters.0.oauth",
								},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"authorization_endpoint": {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 2048),
											),
										},
										"client_parameters": {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrClientID: {
														Type:     schema.TypeString,
														Required: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 512),
														),
													},
													names.AttrClientSecret: {
														Type:      schema.TypeString,
														Required:  true,
														Sensitive: true,
														ValidateFunc: validation.All(
															validation.StringLenBetween(1, 512),
														),
													},
												},
											},
										},
										"http_method": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[types.ConnectionOAuthHttpMethod](),
										},
										"oauth_http_parameters": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem:     connectionHttpParameters("auth_parameters.0.oauth.0.oauth_http_parameters"),
										},
									},
								},
							},
						},
					},
				},
				"authorization_type": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.ConnectionAuthorizationType](),
				},
				names.AttrDescription: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 512),
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 64),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), ""),
					),
				},
				"secret_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &eventbridge.CreateConnectionInput{
		AuthorizationType: types.ConnectionAuthorizationType(d.Get("authorization_type").(string)),
		AuthParameters:    expandCreateConnectionAuthRequestParameters(d.Get("auth_parameters").([]interface{})),
		Name:              aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Connection (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitConnectionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Connection (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	output, err := findConnectionByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ConnectionArn)
	if output.AuthParameters != nil {
		if err := d.Set("auth_parameters", flattenConnectionAuthParameters(output.AuthParameters, d)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auth_parameters error: %s", err)
		}
	}
	d.Set("authorization_type", output.AuthorizationType)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrName, output.Name)
	d.Set("secret_arn", output.SecretArn)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.UpdateConnectionInput{
		Name: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("authorization_type"); ok {
		input.AuthorizationType = types.ConnectionAuthorizationType(v.(string))
	}

	if v, ok := d.GetOk("auth_parameters"); ok {
		input.AuthParameters = expandUpdateConnectionAuthRequestParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectionUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Connection (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	log.Printf("[INFO] Deleting EventBridge Connection: %s", d.Id())
	_, err := conn.DeleteConnection(ctx, &eventbridge.DeleteConnectionInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConnectionByName(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeConnectionOutput, error) {
	input := &eventbridge.DescribeConnectionInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeConnection(ctx, input)

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

func statusConnectionState(ctx context.Context, conn *eventbridge.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConnectionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConnectionState), nil
	}
}

func waitConnectionCreated(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeConnectionOutput, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConnectionStateCreating, types.ConnectionStateAuthorizing),
		Target:  enum.Slice(types.ConnectionStateAuthorized, types.ConnectionStateDeauthorized),
		Refresh: statusConnectionState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitConnectionUpdated(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeConnectionOutput, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConnectionStateUpdating, types.ConnectionStateAuthorizing, types.ConnectionStateDeauthorizing),
		Target:  enum.Slice(types.ConnectionStateAuthorized, types.ConnectionStateDeauthorized),
		Refresh: statusConnectionState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeConnectionOutput, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusConnectionState(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func expandCreateConnectionAuthRequestParameters(config []interface{}) *types.CreateConnectionAuthRequestParameters {
	authParameters := &types.CreateConnectionAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["api_key"]; ok {
			authParameters.ApiKeyAuthParameters = expandCreateConnectionAPIKeyAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["basic"]; ok {
			authParameters.BasicAuthParameters = expandCreateConnectionBasicAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["oauth"]; ok {
			authParameters.OAuthParameters = expandCreateConnectionOAuthAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["invocation_http_parameters"]; ok {
			authParameters.InvocationHttpParameters = expandConnectionHTTPParameters(val.([]interface{}))
		}
	}

	return authParameters
}

func expandCreateConnectionAPIKeyAuthRequestParameters(config []interface{}) *types.CreateConnectionApiKeyAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	apiKeyAuthParameters := &types.CreateConnectionApiKeyAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param[names.AttrKey].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyName = aws.String(val)
		}
		if val, ok := param[names.AttrValue].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyValue = aws.String(val)
		}
	}
	return apiKeyAuthParameters
}

func expandCreateConnectionBasicAuthRequestParameters(config []interface{}) *types.CreateConnectionBasicAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	basicAuthParameters := &types.CreateConnectionBasicAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param[names.AttrUsername].(string); ok && val != "" {
			basicAuthParameters.Username = aws.String(val)
		}
		if val, ok := param[names.AttrPassword].(string); ok && val != "" {
			basicAuthParameters.Password = aws.String(val)
		}
	}
	return basicAuthParameters
}

func expandCreateConnectionOAuthAuthRequestParameters(config []interface{}) *types.CreateConnectionOAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	oAuthParameters := &types.CreateConnectionOAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["authorization_endpoint"].(string); ok && val != "" {
			oAuthParameters.AuthorizationEndpoint = aws.String(val)
		}
		if val, ok := param["http_method"].(string); ok && val != "" {
			oAuthParameters.HttpMethod = types.ConnectionOAuthHttpMethod(val)
		}
		if val, ok := param["oauth_http_parameters"]; ok {
			oAuthParameters.OAuthHttpParameters = expandConnectionHTTPParameters(val.([]interface{}))
		}
		if val, ok := param["client_parameters"]; ok {
			oAuthParameters.ClientParameters = expandCreateConnectionOAuthClientRequestParameters(val.([]interface{}))
		}
	}
	return oAuthParameters
}

func expandCreateConnectionOAuthClientRequestParameters(config []interface{}) *types.CreateConnectionOAuthClientRequestParameters {
	oAuthClientRequestParameters := &types.CreateConnectionOAuthClientRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param[names.AttrClientID].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientID = aws.String(val)
		}
		if val, ok := param[names.AttrClientSecret].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientSecret = aws.String(val)
		}
	}
	return oAuthClientRequestParameters
}

func expandConnectionHTTPParameters(config []interface{}) *types.ConnectionHttpParameters {
	if len(config) == 0 {
		return nil
	}
	httpParameters := &types.ConnectionHttpParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["body"]; ok {
			httpParameters.BodyParameters = expandConnectionHTTPParametersBody(val.([]interface{}))
		}
		if val, ok := param[names.AttrHeader]; ok {
			httpParameters.HeaderParameters = expandConnectionHTTPParametersHeader(val.([]interface{}))
		}
		if val, ok := param["query_string"]; ok {
			httpParameters.QueryStringParameters = expandConnectionHTTPParametersQueryString(val.([]interface{}))
		}
	}
	return httpParameters
}

func expandConnectionHTTPParametersBody(config []interface{}) []types.ConnectionBodyParameter {
	if len(config) == 0 {
		return nil
	}
	var parameters []types.ConnectionBodyParameter
	for _, c := range config {
		parameter := types.ConnectionBodyParameter{}

		input := c.(map[string]interface{})
		if val, ok := input[names.AttrKey].(string); ok && val != "" {
			parameter.Key = aws.String(val)
		}
		if val, ok := input[names.AttrValue].(string); ok && val != "" {
			parameter.Value = aws.String(val)
		}
		if val, ok := input["is_value_secret"].(bool); ok {
			parameter.IsValueSecret = val
		}
		parameters = append(parameters, parameter)
	}
	return parameters
}

func expandConnectionHTTPParametersHeader(config []interface{}) []types.ConnectionHeaderParameter {
	if len(config) == 0 {
		return nil
	}
	var parameters []types.ConnectionHeaderParameter
	for _, c := range config {
		parameter := types.ConnectionHeaderParameter{}

		input := c.(map[string]interface{})
		if val, ok := input[names.AttrKey].(string); ok && val != "" {
			parameter.Key = aws.String(val)
		}
		if val, ok := input[names.AttrValue].(string); ok && val != "" {
			parameter.Value = aws.String(val)
		}
		if val, ok := input["is_value_secret"].(bool); ok {
			parameter.IsValueSecret = val
		}
		parameters = append(parameters, parameter)
	}
	return parameters
}

func expandConnectionHTTPParametersQueryString(config []interface{}) []types.ConnectionQueryStringParameter {
	if len(config) == 0 {
		return nil
	}
	var parameters []types.ConnectionQueryStringParameter
	for _, c := range config {
		parameter := types.ConnectionQueryStringParameter{}

		input := c.(map[string]interface{})
		if val, ok := input[names.AttrKey].(string); ok && val != "" {
			parameter.Key = aws.String(val)
		}
		if val, ok := input[names.AttrValue].(string); ok && val != "" {
			parameter.Value = aws.String(val)
		}
		if val, ok := input["is_value_secret"].(bool); ok {
			parameter.IsValueSecret = val
		}
		parameters = append(parameters, parameter)
	}
	return parameters
}

func flattenConnectionAuthParameters(authParameters *types.ConnectionAuthResponseParameters, d *schema.ResourceData) []map[string]interface{} {
	config := make(map[string]interface{})

	if authParameters.ApiKeyAuthParameters != nil {
		config["api_key"] = flattenConnectionAPIKeyAuthParameters(authParameters.ApiKeyAuthParameters, d)
	}

	if authParameters.BasicAuthParameters != nil {
		config["basic"] = flattenConnectionBasicAuthParameters(authParameters.BasicAuthParameters, d)
	}

	if authParameters.OAuthParameters != nil {
		config["oauth"] = flattenConnectionOAuthParameters(authParameters.OAuthParameters, d)
	}

	if authParameters.InvocationHttpParameters != nil {
		config["invocation_http_parameters"] = flattenConnectionHTTPParameters(authParameters.InvocationHttpParameters, d, "auth_parameters.0.invocation_http_parameters")
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenConnectionAPIKeyAuthParameters(apiKeyAuthParameters *types.ConnectionApiKeyAuthResponseParameters, d *schema.ResourceData) []map[string]interface{} {
	if apiKeyAuthParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if apiKeyAuthParameters.ApiKeyName != nil {
		config[names.AttrKey] = aws.ToString(apiKeyAuthParameters.ApiKeyName)
	}

	if v, ok := d.GetOk("auth_parameters.0.api_key.0.value"); ok {
		config[names.AttrValue] = v.(string)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenConnectionBasicAuthParameters(basicAuthParameters *types.ConnectionBasicAuthResponseParameters, d *schema.ResourceData) []map[string]interface{} {
	if basicAuthParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if basicAuthParameters.Username != nil {
		config[names.AttrUsername] = aws.ToString(basicAuthParameters.Username)
	}

	if v, ok := d.GetOk("auth_parameters.0.basic.0.password"); ok {
		config[names.AttrPassword] = v.(string)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenConnectionOAuthParameters(oAuthParameters *types.ConnectionOAuthResponseParameters, d *schema.ResourceData) []map[string]interface{} {
	if oAuthParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if oAuthParameters.AuthorizationEndpoint != nil {
		config["authorization_endpoint"] = aws.ToString(oAuthParameters.AuthorizationEndpoint)
	}
	config["http_method"] = oAuthParameters.HttpMethod
	config["oauth_http_parameters"] = flattenConnectionHTTPParameters(oAuthParameters.OAuthHttpParameters, d, "auth_parameters.0.oauth.0.oauth_http_parameters")
	config["client_parameters"] = flattenConnectionOAuthClientResponseParameters(oAuthParameters.ClientParameters, d)

	result := []map[string]interface{}{config}
	return result
}

func flattenConnectionOAuthClientResponseParameters(oAuthClientRequestParameters *types.ConnectionOAuthClientResponseParameters, d *schema.ResourceData) []map[string]interface{} {
	if oAuthClientRequestParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if oAuthClientRequestParameters.ClientID != nil {
		config[names.AttrClientID] = aws.ToString(oAuthClientRequestParameters.ClientID)
	}

	if v, ok := d.GetOk("auth_parameters.0.oauth.0.client_parameters.0.client_secret"); ok {
		config[names.AttrClientSecret] = v.(string)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenConnectionHTTPParameters(httpParameters *types.ConnectionHttpParameters, d *schema.ResourceData, path string) []map[string]interface{} {
	if httpParameters == nil {
		return nil
	}

	var bodyParameters []map[string]interface{}
	for i, param := range httpParameters.BodyParameters {
		config := make(map[string]interface{})
		config["is_value_secret"] = param.IsValueSecret
		config[names.AttrKey] = aws.ToString(param.Key)

		if param.Value != nil {
			config[names.AttrValue] = aws.ToString(param.Value)
		} else if v, ok := d.GetOk(fmt.Sprintf("%s.0.body.%d.value", path, i)); ok {
			config[names.AttrValue] = v.(string)
		}
		bodyParameters = append(bodyParameters, config)
	}

	var headerParameters []map[string]interface{}
	for i, param := range httpParameters.HeaderParameters {
		config := make(map[string]interface{})
		config["is_value_secret"] = param.IsValueSecret
		config[names.AttrKey] = aws.ToString(param.Key)

		if param.Value != nil {
			config[names.AttrValue] = aws.ToString(param.Value)
		} else if v, ok := d.GetOk(fmt.Sprintf("%s.0.header.%d.value", path, i)); ok {
			config[names.AttrValue] = v.(string)
		}
		headerParameters = append(headerParameters, config)
	}

	var queryStringParameters []map[string]interface{}
	for i, param := range httpParameters.QueryStringParameters {
		config := make(map[string]interface{})
		config["is_value_secret"] = param.IsValueSecret
		config[names.AttrKey] = aws.ToString(param.Key)

		if param.Value != nil {
			config[names.AttrValue] = aws.ToString(param.Value)
		} else if v, ok := d.GetOk(fmt.Sprintf("%s.0.query_string.%d.value", path, i)); ok {
			config[names.AttrValue] = v.(string)
		}
		queryStringParameters = append(queryStringParameters, config)
	}

	parameters := make(map[string]interface{})
	parameters["body"] = bodyParameters
	parameters[names.AttrHeader] = headerParameters
	parameters["query_string"] = queryStringParameters

	result := []map[string]interface{}{parameters}
	return result
}

func expandUpdateConnectionAuthRequestParameters(config []interface{}) *types.UpdateConnectionAuthRequestParameters {
	authParameters := &types.UpdateConnectionAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["api_key"]; ok {
			authParameters.ApiKeyAuthParameters = expandUpdateConnectionAPIKeyAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["basic"]; ok {
			authParameters.BasicAuthParameters = expandUpdateConnectionBasicAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["oauth"]; ok {
			authParameters.OAuthParameters = expandUpdateConnectionOAuthAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["invocation_http_parameters"]; ok {
			authParameters.InvocationHttpParameters = expandConnectionHTTPParameters(val.([]interface{}))
		}
	}

	return authParameters
}

func expandUpdateConnectionAPIKeyAuthRequestParameters(config []interface{}) *types.UpdateConnectionApiKeyAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	apiKeyAuthParameters := &types.UpdateConnectionApiKeyAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param[names.AttrKey].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyName = aws.String(val)
		}
		if val, ok := param[names.AttrValue].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyValue = aws.String(val)
		}
	}
	return apiKeyAuthParameters
}

func expandUpdateConnectionBasicAuthRequestParameters(config []interface{}) *types.UpdateConnectionBasicAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	basicAuthParameters := &types.UpdateConnectionBasicAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param[names.AttrUsername].(string); ok && val != "" {
			basicAuthParameters.Username = aws.String(val)
		}
		if val, ok := param[names.AttrPassword].(string); ok && val != "" {
			basicAuthParameters.Password = aws.String(val)
		}
	}
	return basicAuthParameters
}

func expandUpdateConnectionOAuthAuthRequestParameters(config []interface{}) *types.UpdateConnectionOAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	oAuthParameters := &types.UpdateConnectionOAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["authorization_endpoint"].(string); ok && val != "" {
			oAuthParameters.AuthorizationEndpoint = aws.String(val)
		}
		if val, ok := param["http_method"].(string); ok && val != "" {
			oAuthParameters.HttpMethod = types.ConnectionOAuthHttpMethod(val)
		}
		if val, ok := param["oauth_http_parameters"]; ok {
			oAuthParameters.OAuthHttpParameters = expandConnectionHTTPParameters(val.([]interface{}))
		}
		if val, ok := param["client_parameters"]; ok {
			oAuthParameters.ClientParameters = expandUpdateConnectionOAuthClientRequestParameters(val.([]interface{}))
		}
	}
	return oAuthParameters
}

func expandUpdateConnectionOAuthClientRequestParameters(config []interface{}) *types.UpdateConnectionOAuthClientRequestParameters {
	oAuthClientRequestParameters := &types.UpdateConnectionOAuthClientRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param[names.AttrClientID].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientID = aws.String(val)
		}
		if val, ok := param[names.AttrClientSecret].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientSecret = aws.String(val)
		}
	}
	return oAuthClientRequestParameters
}
