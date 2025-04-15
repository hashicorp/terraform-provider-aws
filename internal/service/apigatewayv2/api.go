// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"reflect"

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
	"github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/internal/yaml"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_api", name="API")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigatewayv2;apigatewayv2.GetApiOutput")
func resourceAPI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPICreate,
		ReadWithoutTimeout:   resourceAPIRead,
		UpdateWithoutTimeout: resourceAPIUpdate,
		DeleteWithoutTimeout: resourceAPIDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$request.header.x-api-key",
				ValidateFunc: validation.StringInSlice([]string{
					"$context.authorizer.usageIdentifierKey",
					"$request.header.x-api-key",
				}, false),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"body": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONOrYAMLDiffs,
				ValidateFunc:     verify.ValidStringIsJSONOrYAML,
			},
			"cors_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"allow_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      sdkv2.StringCaseInsensitiveSetFunc,
						},
						"allow_methods": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      sdkv2.StringCaseInsensitiveSetFunc,
						},
						"allow_origins": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      sdkv2.StringCaseInsensitiveSetFunc,
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      sdkv2.StringCaseInsensitiveSetFunc,
						},
						"max_age": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"disable_execute_api_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fail_on_warnings": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"protocol_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProtocolType](),
			},
			"route_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"route_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$request.method $request.path",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTarget: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAPICreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigatewayv2.CreateApiInput{
		Name:         aws.String(name),
		ProtocolType: awstypes.ProtocolType(d.Get("protocol_type").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("api_key_selection_expression"); ok {
		input.ApiKeySelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cors_configuration"); ok {
		input.CorsConfiguration = expandCORSConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("credentials_arn"); ok {
		input.CredentialsArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disable_execute_api_endpoint"); ok {
		input.DisableExecuteApiEndpoint = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("route_key"); ok {
		input.RouteKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("route_selection_expression"); ok {
		input.RouteSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrTarget); ok {
		input.Target = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.Version = aws.String(v.(string))
	}

	output, err := conn.CreateApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 API (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ApiId))

	err = reimportOpenAPIDefinition(ctx, d, meta)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceAPIRead(ctx, d, meta)...)
}

func resourceAPIRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findAPIByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", d.Id(), err)
	}

	d.Set("api_endpoint", output.ApiEndpoint)
	d.Set("api_key_selection_expression", output.ApiKeySelectionExpression)
	d.Set(names.AttrARN, apiARN(ctx, meta.(*conns.AWSClient), d.Id()))
	if err := d.Set("cors_configuration", flattenCORSConfiguration(output.CorsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cors_configuration: %s", err)
	}
	d.Set(names.AttrDescription, output.Description)
	d.Set("disable_execute_api_endpoint", output.DisableExecuteApiEndpoint)
	d.Set("execution_arn", apiInvokeARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set(names.AttrName, output.Name)
	d.Set("protocol_type", output.ProtocolType)
	d.Set("route_selection_expression", output.RouteSelectionExpression)
	d.Set(names.AttrVersion, output.Version)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceAPIUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	corsConfigurationDeleted := false
	if d.HasChange("cors_configuration") {
		if v := d.Get("cors_configuration"); len(v.([]any)) == 0 {
			corsConfigurationDeleted = true
			input := &apigatewayv2.DeleteCorsConfigurationInput{
				ApiId: aws.String(d.Id()),
			}

			_, err := conn.DeleteCorsConfiguration(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 API (%s) CORS configuration: %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("api_key_selection_expression", names.AttrDescription, "disable_execute_api_endpoint", names.AttrName, "route_selection_expression", names.AttrVersion) ||
		(d.HasChange("cors_configuration") && !corsConfigurationDeleted) {
		input := &apigatewayv2.UpdateApiInput{
			ApiId: aws.String(d.Id()),
		}

		if d.HasChange("api_key_selection_expression") {
			input.ApiKeySelectionExpression = aws.String(d.Get("api_key_selection_expression").(string))
		}

		if d.HasChange("cors_configuration") {
			input.CorsConfiguration = expandCORSConfiguration(d.Get("cors_configuration").([]any))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("disable_execute_api_endpoint") {
			input.DisableExecuteApiEndpoint = aws.Bool(d.Get("disable_execute_api_endpoint").(bool))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("route_selection_expression") {
			input.RouteSelectionExpression = aws.String(d.Get("route_selection_expression").(string))
		}

		if d.HasChange(names.AttrVersion) {
			input.Version = aws.String(d.Get(names.AttrVersion).(string))
		}

		_, err := conn.UpdateApi(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 API (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("body") {
		err := reimportOpenAPIDefinition(ctx, d, meta)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceAPIRead(ctx, d, meta)...)
}

func resourceAPIDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 API: %s", d.Id())
	input := apigatewayv2.DeleteApiInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteApi(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 API (%s): %s", d.Id(), err)
	}

	return diags
}

func reimportOpenAPIDefinition(ctx context.Context, d *schema.ResourceData, meta any) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	if body, ok := d.GetOk("body"); ok {
		body := body.(string)
		configuredCORSConfiguration := d.Get("cors_configuration")
		configuredDescription := d.Get(names.AttrDescription).(string)
		configuredName := d.Get(names.AttrName).(string)
		configuredVersion := d.Get(names.AttrVersion).(string)

		inputRA := &apigatewayv2.ReimportApiInput{
			ApiId: aws.String(d.Id()),
			Body:  aws.String(body),
		}

		if value, ok := d.GetOk("fail_on_warnings"); ok {
			inputRA.FailOnWarnings = aws.Bool(value.(bool))
		}

		_, err := conn.ReimportApi(ctx, inputRA)

		if err != nil {
			return fmt.Errorf("reimporting API Gateway v2 API (%s) OpenAPI definition: %w", d.Id(), err)
		}

		if diags := resourceAPIRead(ctx, d, meta); diags.HasError() {
			return sdkdiag.DiagnosticsError(diags)
		}

		m, err := decodeOpenAPIDefinition(body)

		if err != nil {
			return fmt.Errorf("decoding API Gateway v2 API (%s) OpenAPI definition: %w", d.Id(), err)
		}

		// Don't ovewrite the configured values if they are not present in the OpenAPI definition.
		description := configuredDescription
		name := configuredName
		version := configuredVersion
		if m, ok := m["info"].(map[string]any); ok {
			if _, ok := m[names.AttrDescription]; ok {
				description = d.Get(names.AttrDescription).(string)
			}
			if _, ok := m["title"]; ok {
				name = d.Get(names.AttrName).(string)
			}
			if _, ok := m[names.AttrVersion]; ok {
				version = d.Get(names.AttrVersion).(string)
			}
		}

		inputUA := &apigatewayv2.UpdateApiInput{
			ApiId:       aws.String(d.Id()),
			Description: aws.String(description),
			Name:        aws.String(name),
			Version:     aws.String(version),
		}

		if !reflect.DeepEqual(configuredCORSConfiguration, d.Get("cors_configuration")) {
			if len(configuredCORSConfiguration.([]any)) == 0 {
				input := &apigatewayv2.DeleteCorsConfigurationInput{
					ApiId: aws.String(d.Id()),
				}

				_, err := conn.DeleteCorsConfiguration(ctx, input)

				if err != nil {
					return fmt.Errorf("deleting API Gateway v2 API (%s) CORS configuration: %w", d.Id(), err)
				}
			} else {
				inputUA.CorsConfiguration = expandCORSConfiguration(configuredCORSConfiguration.([]any))
			}
		}

		_, err = conn.UpdateApi(ctx, inputUA)

		if err != nil {
			return fmt.Errorf("updating API Gateway v2 API (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func findAPIByID(ctx context.Context, conn *apigatewayv2.Client, id string) (*apigatewayv2.GetApiOutput, error) {
	input := &apigatewayv2.GetApiInput{
		ApiId: aws.String(id),
	}

	return findAPI(ctx, conn, input)
}

func findAPI(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetApiInput) (*apigatewayv2.GetApiOutput, error) {
	output, err := conn.GetApi(ctx, input)
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

func expandCORSConfiguration(vConfiguration []any) *awstypes.Cors {
	configuration := &awstypes.Cors{}

	if len(vConfiguration) == 0 || vConfiguration[0] == nil {
		return configuration
	}
	mConfiguration := vConfiguration[0].(map[string]any)

	if vAllowCredentials, ok := mConfiguration["allow_credentials"].(bool); ok {
		configuration.AllowCredentials = aws.Bool(vAllowCredentials)
	}
	if vAllowHeaders, ok := mConfiguration["allow_headers"].(*schema.Set); ok {
		configuration.AllowHeaders = flex.ExpandStringValueSet(vAllowHeaders)
	}
	if vAllowMethods, ok := mConfiguration["allow_methods"].(*schema.Set); ok {
		configuration.AllowMethods = flex.ExpandStringValueSet(vAllowMethods)
	}
	if vAllowOrigins, ok := mConfiguration["allow_origins"].(*schema.Set); ok {
		configuration.AllowOrigins = flex.ExpandStringValueSet(vAllowOrigins)
	}
	if vExposeHeaders, ok := mConfiguration["expose_headers"].(*schema.Set); ok {
		configuration.ExposeHeaders = flex.ExpandStringValueSet(vExposeHeaders)
	}
	if vMaxAge, ok := mConfiguration["max_age"].(int); ok {
		configuration.MaxAge = aws.Int32(int32(vMaxAge))
	}

	return configuration
}

func flattenCORSConfiguration(configuration *awstypes.Cors) []any {
	if configuration == nil {
		return []any{}
	}

	return []any{map[string]any{
		"allow_credentials": aws.ToBool(configuration.AllowCredentials),
		"allow_headers":     flex.FlattenStringValueSetCaseInsensitive(configuration.AllowHeaders),
		"allow_methods":     flex.FlattenStringValueSetCaseInsensitive(configuration.AllowMethods),
		"allow_origins":     flex.FlattenStringValueSetCaseInsensitive(configuration.AllowOrigins),
		"expose_headers":    flex.FlattenStringValueSetCaseInsensitive(configuration.ExposeHeaders),
		"max_age":           aws.ToInt32(configuration.MaxAge),
	}}
}

func apiARN(ctx context.Context, c *conns.AWSClient, apiID string) string {
	return c.RegionalARNNoAccount(ctx, "apigateway", "/apis/"+apiID)
}

func apiInvokeARN(ctx context.Context, c *conns.AWSClient, apiID string) string {
	return c.RegionalARN(ctx, "execute-api", apiID)
}

func decodeOpenAPIDefinition(v string) (map[string]any, error) {
	var output map[string]any

	err := json.DecodeFromString(v, &output)

	if err != nil {
		err = yaml.DecodeFromString(v, &output)
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
