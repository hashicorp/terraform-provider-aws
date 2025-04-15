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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
				"invocation_connectivity_parameters": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"resource_parameters": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"resource_association_arn": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"resource_configuration_arn": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
					},
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

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &eventbridge.CreateConnectionInput{
		AuthorizationType: types.ConnectionAuthorizationType(d.Get("authorization_type").(string)),
		AuthParameters:    expandCreateConnectionAuthRequestParameters(d.Get("auth_parameters").([]any)),
		Name:              aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_connectivity_parameters"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.InvocationConnectivityParameters = expandConnectivityResourceParameters(v.([]any)[0].(map[string]any))
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

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	if output.InvocationConnectivityParameters != nil {
		if err := d.Set("invocation_connectivity_parameters", []any{flattenDescribeConnectionConnectivityParameters(output.InvocationConnectivityParameters)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting invocation_connectivity_parameters: %s", err)
		}
	} else {
		d.Set("invocation_connectivity_parameters", nil)
	}
	d.Set(names.AttrName, output.Name)
	d.Set("secret_arn", output.SecretArn)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.UpdateConnectionInput{
		Name: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("authorization_type"); ok {
		input.AuthorizationType = types.ConnectionAuthorizationType(v.(string))
	}

	if v, ok := d.GetOk("auth_parameters"); ok {
		input.AuthParameters = expandUpdateConnectionAuthRequestParameters(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_connectivity_parameters"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.InvocationConnectivityParameters = expandConnectivityResourceParameters(v.([]any)[0].(map[string]any))
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

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	return func() (any, string, error) {
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

func expandCreateConnectionAuthRequestParameters(tfList []any) *types.CreateConnectionAuthRequestParameters {
	apiObject := &types.CreateConnectionAuthRequestParameters{}

	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap["api_key"].([]any); ok && len(v) > 0 {
			apiObject.ApiKeyAuthParameters = expandCreateConnectionAPIKeyAuthRequestParameters(v)
		}
		if v, ok := tfMap["basic"].([]any); ok && len(v) > 0 {
			apiObject.BasicAuthParameters = expandCreateConnectionBasicAuthRequestParameters(v)
		}
		if v, ok := tfMap["oauth"].([]any); ok && len(v) > 0 {
			apiObject.OAuthParameters = expandCreateConnectionOAuthAuthRequestParameters(v)
		}
		if v, ok := tfMap["invocation_http_parameters"].([]any); ok && len(v) > 0 {
			apiObject.InvocationHttpParameters = expandConnectionHTTPParameters(v)
		}
	}

	return apiObject
}

func expandCreateConnectionAPIKeyAuthRequestParameters(tfList []any) *types.CreateConnectionApiKeyAuthRequestParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.CreateConnectionApiKeyAuthRequestParameters{}
	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.ApiKeyName = aws.String(v)
		}
		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.ApiKeyValue = aws.String(v)
		}
	}

	return apiObject
}

func expandCreateConnectionBasicAuthRequestParameters(tfList []any) *types.CreateConnectionBasicAuthRequestParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.CreateConnectionBasicAuthRequestParameters{}
	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
			apiObject.Username = aws.String(v)
		}
		if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
			apiObject.Password = aws.String(v)
		}
	}

	return apiObject
}

func expandCreateConnectionOAuthAuthRequestParameters(tfList []any) *types.CreateConnectionOAuthRequestParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.CreateConnectionOAuthRequestParameters{}
	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap["authorization_endpoint"].(string); ok && v != "" {
			apiObject.AuthorizationEndpoint = aws.String(v)
		}
		if v, ok := tfMap["http_method"].(string); ok && v != "" {
			apiObject.HttpMethod = types.ConnectionOAuthHttpMethod(v)
		}
		if v, ok := tfMap["oauth_http_parameters"].([]any); ok && len(v) > 0 {
			apiObject.OAuthHttpParameters = expandConnectionHTTPParameters(v)
		}
		if v, ok := tfMap["client_parameters"].([]any); ok && len(v) > 0 {
			apiObject.ClientParameters = expandCreateConnectionOAuthClientRequestParameters(v)
		}
	}

	return apiObject
}

func expandCreateConnectionOAuthClientRequestParameters(tfList []any) *types.CreateConnectionOAuthClientRequestParameters {
	apiObject := &types.CreateConnectionOAuthClientRequestParameters{}

	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrClientID].(string); ok && v != "" {
			apiObject.ClientID = aws.String(v)
		}
		if v, ok := tfMap[names.AttrClientSecret].(string); ok && v != "" {
			apiObject.ClientSecret = aws.String(v)
		}
	}

	return apiObject
}

func expandConnectionHTTPParameters(tfList []any) *types.ConnectionHttpParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.ConnectionHttpParameters{}
	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap["body"].([]any); ok && len(v) > 0 {
			apiObject.BodyParameters = expandConnectionHTTPParametersBody(v)
		}
		if v, ok := tfMap[names.AttrHeader].([]any); ok && len(v) > 0 {
			apiObject.HeaderParameters = expandConnectionHTTPParametersHeader(v)
		}
		if v, ok := tfMap["query_string"].([]any); ok && len(v) > 0 {
			apiObject.QueryStringParameters = expandConnectionHTTPParametersQueryString(v)
		}
	}

	return apiObject
}

func expandConnectionHTTPParametersBody(tfList []any) []types.ConnectionBodyParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ConnectionBodyParameter
	for _, item := range tfList {
		if item == nil {
			continue
		}

		apiObject := types.ConnectionBodyParameter{}
		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}
		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}
		if v, ok := tfMap["is_value_secret"].(bool); ok {
			apiObject.IsValueSecret = v
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandConnectionHTTPParametersHeader(tfList []any) []types.ConnectionHeaderParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ConnectionHeaderParameter
	for _, item := range tfList {
		if item == nil {
			continue
		}

		apiObject := types.ConnectionHeaderParameter{}
		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}
		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}
		if v, ok := tfMap["is_value_secret"].(bool); ok {
			apiObject.IsValueSecret = v
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandConnectionHTTPParametersQueryString(tfList []any) []types.ConnectionQueryStringParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ConnectionQueryStringParameter
	for _, item := range tfList {
		if item == nil {
			continue
		}

		apiObject := types.ConnectionQueryStringParameter{}
		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}
		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}
		if v, ok := tfMap["is_value_secret"].(bool); ok {
			apiObject.IsValueSecret = v
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenConnectionAuthParameters(apiObject *types.ConnectionAuthResponseParameters, d *schema.ResourceData) []map[string]any {
	tfMap := make(map[string]any)

	if apiObject.ApiKeyAuthParameters != nil {
		tfMap["api_key"] = flattenConnectionAPIKeyAuthParameters(apiObject.ApiKeyAuthParameters, d)
	}

	if apiObject.BasicAuthParameters != nil {
		tfMap["basic"] = flattenConnectionBasicAuthParameters(apiObject.BasicAuthParameters, d)
	}

	if apiObject.OAuthParameters != nil {
		tfMap["oauth"] = flattenConnectionOAuthParameters(apiObject.OAuthParameters, d)
	}

	if apiObject.InvocationHttpParameters != nil {
		tfMap["invocation_http_parameters"] = flattenConnectionHTTPParameters(apiObject.InvocationHttpParameters, d, "auth_parameters.0.invocation_http_parameters")
	}

	return []map[string]any{tfMap}
}

func flattenConnectionAPIKeyAuthParameters(apiObject *types.ConnectionApiKeyAuthResponseParameters, d *schema.ResourceData) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	if apiObject.ApiKeyName != nil {
		tfMap[names.AttrKey] = aws.ToString(apiObject.ApiKeyName)
	}

	if v, ok := d.GetOk("auth_parameters.0.api_key.0.value"); ok {
		tfMap[names.AttrValue] = v.(string)
	}

	return []map[string]any{tfMap}
}

func flattenConnectionBasicAuthParameters(apiObject *types.ConnectionBasicAuthResponseParameters, d *schema.ResourceData) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	if apiObject.Username != nil {
		tfMap[names.AttrUsername] = aws.ToString(apiObject.Username)
	}

	if v, ok := d.GetOk("auth_parameters.0.basic.0.password"); ok {
		tfMap[names.AttrPassword] = v.(string)
	}

	return []map[string]any{tfMap}
}

func flattenConnectionOAuthParameters(apiObject *types.ConnectionOAuthResponseParameters, d *schema.ResourceData) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	if apiObject.AuthorizationEndpoint != nil {
		tfMap["authorization_endpoint"] = aws.ToString(apiObject.AuthorizationEndpoint)
	}
	tfMap["http_method"] = apiObject.HttpMethod
	tfMap["oauth_http_parameters"] = flattenConnectionHTTPParameters(apiObject.OAuthHttpParameters, d, "auth_parameters.0.oauth.0.oauth_http_parameters")
	tfMap["client_parameters"] = flattenConnectionOAuthClientResponseParameters(apiObject.ClientParameters, d)

	return []map[string]any{tfMap}
}

func flattenConnectionOAuthClientResponseParameters(apiObject *types.ConnectionOAuthClientResponseParameters, d *schema.ResourceData) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)
	if apiObject.ClientID != nil {
		tfMap[names.AttrClientID] = aws.ToString(apiObject.ClientID)
	}

	if v, ok := d.GetOk("auth_parameters.0.oauth.0.client_parameters.0.client_secret"); ok {
		tfMap[names.AttrClientSecret] = v.(string)
	}

	return []map[string]any{tfMap}
}

func flattenConnectionHTTPParameters(apiObject *types.ConnectionHttpParameters, d *schema.ResourceData, path string) []map[string]any {
	if apiObject == nil {
		return nil
	}

	var bodyParameters []map[string]any
	for i, param := range apiObject.BodyParameters {
		tfMap := make(map[string]any)
		tfMap["is_value_secret"] = param.IsValueSecret
		tfMap[names.AttrKey] = aws.ToString(param.Key)

		if param.Value != nil {
			tfMap[names.AttrValue] = aws.ToString(param.Value)
		} else if v, ok := d.GetOk(fmt.Sprintf("%s.0.body.%d.value", path, i)); ok {
			tfMap[names.AttrValue] = v.(string)
		}

		bodyParameters = append(bodyParameters, tfMap)
	}

	var headerParameters []map[string]any
	for i, param := range apiObject.HeaderParameters {
		tfMap := make(map[string]any)
		tfMap["is_value_secret"] = param.IsValueSecret
		tfMap[names.AttrKey] = aws.ToString(param.Key)

		if param.Value != nil {
			tfMap[names.AttrValue] = aws.ToString(param.Value)
		} else if v, ok := d.GetOk(fmt.Sprintf("%s.0.header.%d.value", path, i)); ok {
			tfMap[names.AttrValue] = v.(string)
		}
		headerParameters = append(headerParameters, tfMap)
	}

	var queryStringParameters []map[string]any
	for i, param := range apiObject.QueryStringParameters {
		tfMap := make(map[string]any)
		tfMap["is_value_secret"] = param.IsValueSecret
		tfMap[names.AttrKey] = aws.ToString(param.Key)

		if param.Value != nil {
			tfMap[names.AttrValue] = aws.ToString(param.Value)
		} else if v, ok := d.GetOk(fmt.Sprintf("%s.0.query_string.%d.value", path, i)); ok {
			tfMap[names.AttrValue] = v.(string)
		}
		queryStringParameters = append(queryStringParameters, tfMap)
	}

	parameters := make(map[string]any)
	parameters["body"] = bodyParameters
	parameters[names.AttrHeader] = headerParameters
	parameters["query_string"] = queryStringParameters

	return []map[string]any{parameters}
}

func expandUpdateConnectionAuthRequestParameters(tfList []any) *types.UpdateConnectionAuthRequestParameters {
	apiObject := &types.UpdateConnectionAuthRequestParameters{}

	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap["api_key"].([]any); ok && len(v) > 0 {
			apiObject.ApiKeyAuthParameters = expandUpdateConnectionAPIKeyAuthRequestParameters(v)
		}
		if v, ok := tfMap["basic"].([]any); ok && len(v) > 0 {
			apiObject.BasicAuthParameters = expandUpdateConnectionBasicAuthRequestParameters(v)
		}
		if v, ok := tfMap["oauth"].([]any); ok && len(v) > 0 {
			apiObject.OAuthParameters = expandUpdateConnectionOAuthAuthRequestParameters(v)
		}
		if v, ok := tfMap["invocation_http_parameters"].([]any); ok && len(v) > 0 {
			apiObject.InvocationHttpParameters = expandConnectionHTTPParameters(v)
		}
	}

	return apiObject
}

func expandUpdateConnectionAPIKeyAuthRequestParameters(tfList []any) *types.UpdateConnectionApiKeyAuthRequestParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.UpdateConnectionApiKeyAuthRequestParameters{}
	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.ApiKeyName = aws.String(v)
		}
		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.ApiKeyValue = aws.String(v)
		}
	}

	return apiObject
}

func expandUpdateConnectionBasicAuthRequestParameters(tfList []any) *types.UpdateConnectionBasicAuthRequestParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.UpdateConnectionBasicAuthRequestParameters{}
	for _, c := range tfList {
		if c == nil {
			continue
		}

		tfMap := c.(map[string]any)
		if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
			apiObject.Username = aws.String(v)
		}
		if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
			apiObject.Password = aws.String(v)
		}
	}

	return apiObject
}

func expandUpdateConnectionOAuthAuthRequestParameters(tfList []any) *types.UpdateConnectionOAuthRequestParameters {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &types.UpdateConnectionOAuthRequestParameters{}
	for _, c := range tfList {
		if c == nil {
			continue
		}

		tfMap := c.(map[string]any)
		if v, ok := tfMap["authorization_endpoint"].(string); ok && v != "" {
			apiObject.AuthorizationEndpoint = aws.String(v)
		}
		if v, ok := tfMap["http_method"].(string); ok && v != "" {
			apiObject.HttpMethod = types.ConnectionOAuthHttpMethod(v)
		}
		if v, ok := tfMap["oauth_http_parameters"].([]any); ok && len(v) > 0 {
			apiObject.OAuthHttpParameters = expandConnectionHTTPParameters(v)
		}
		if v, ok := tfMap["client_parameters"].([]any); ok && len(v) > 0 {
			apiObject.ClientParameters = expandUpdateConnectionOAuthClientRequestParameters(v)
		}
	}

	return apiObject
}

func expandUpdateConnectionOAuthClientRequestParameters(tfList []any) *types.UpdateConnectionOAuthClientRequestParameters {
	apiObject := &types.UpdateConnectionOAuthClientRequestParameters{}

	for _, item := range tfList {
		if item == nil {
			continue
		}

		tfMap := item.(map[string]any)
		if v, ok := tfMap[names.AttrClientID].(string); ok && v != "" {
			apiObject.ClientID = aws.String(v)
		}
		if v, ok := tfMap[names.AttrClientSecret].(string); ok && v != "" {
			apiObject.ClientSecret = aws.String(v)
		}
	}

	return apiObject
}

func expandConnectivityResourceParameters(tfMap map[string]any) *types.ConnectivityResourceParameters {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ConnectivityResourceParameters{}

	if v, ok := tfMap["resource_parameters"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ResourceParameters = expandConnectivityResourceConfigurationARN(v[0].(map[string]any))
	}

	return apiObject
}

func expandConnectivityResourceConfigurationARN(tfMap map[string]any) *types.ConnectivityResourceConfigurationArn {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ConnectivityResourceConfigurationArn{}

	if v, ok := tfMap["resource_configuration_arn"].(string); ok && v != "" {
		apiObject.ResourceConfigurationArn = aws.String(v)
	}

	return apiObject
}

func flattenDescribeConnectionConnectivityParameters(apiObject *types.DescribeConnectionConnectivityParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ResourceParameters; v != nil {
		tfMap["resource_parameters"] = []any{flattenDescribeConnectionResourceParameters(v)}
	}

	return tfMap
}

func flattenDescribeConnectionResourceParameters(apiObject *types.DescribeConnectionResourceParameters) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ResourceAssociationArn; v != nil {
		tfMap["resource_association_arn"] = aws.ToString(v)
	}

	if v := apiObject.ResourceConfigurationArn; v != nil {
		tfMap["resource_configuration_arn"] = aws.ToString(v)
	}

	return tfMap
}
