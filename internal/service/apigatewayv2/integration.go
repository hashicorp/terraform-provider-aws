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

// @SDKResource("aws_apigatewayv2_integration", name="Integration")
func resourceIntegration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIntegrationCreate,
		ReadWithoutTimeout:   resourceIntegrationRead,
		UpdateWithoutTimeout: resourceIntegrationUpdate,
		DeleteWithoutTimeout: resourceIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceIntegrationImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrConnectionID: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"connection_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ConnectionTypeInternet,
				ValidateDiagFunc: enum.Validate[awstypes.ConnectionType](),
			},
			"content_handling_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ContentHandlingStrategy](),
			},
			"credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"integration_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validHTTPMethod(),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Default HTTP method for Lambda integration is POST.
					if v := d.Get("integration_type").(string); (v == string(awstypes.IntegrationTypeAws) || v == string(awstypes.IntegrationTypeAwsProxy)) && old == "POST" && new == "" {
						return true
					}

					return false
				},
			},
			"integration_response_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_subtype": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"integration_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IntegrationType](),
			},
			"integration_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"passthrough_behavior": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.PassthroughBehaviorWhenNoMatch,
				ValidateDiagFunc: enum.Validate[awstypes.PassthroughBehavior](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Not set for HTTP APIs.
					if old == "" && new == string(awstypes.PassthroughBehaviorWhenNoMatch) {
						return true
					}
					return false
				},
			},
			"payload_format_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.0",
				ValidateFunc: validation.StringInSlice([]string{
					"1.0",
					"2.0",
				}, false),
			},
			"request_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				// Length between [1-512].
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"request_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				// Length between [0-32768].
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"response_parameters": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mappings": {
							Type:     schema.TypeMap,
							Required: true,
							// Length between [1-512].
							Elem: &schema.Schema{Type: schema.TypeString},
						},
						names.AttrStatusCode: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"template_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"timeout_milliseconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"tls_config": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server_name_to_verify": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	input := &apigatewayv2.CreateIntegrationInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		IntegrationType: awstypes.IntegrationType(d.Get("integration_type").(string)),
	}

	if v, ok := d.GetOk(names.AttrConnectionID); ok {
		input.ConnectionId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connection_type"); ok {
		input.ConnectionType = awstypes.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("content_handling_strategy"); ok {
		input.ContentHandlingStrategy = awstypes.ContentHandlingStrategy(v.(string))
	}

	if v, ok := d.GetOk("credentials_arn"); ok {
		input.CredentialsArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("integration_method"); ok {
		input.IntegrationMethod = aws.String(v.(string))
	}

	if v, ok := d.GetOk("integration_subtype"); ok {
		input.IntegrationSubtype = aws.String(v.(string))
	}

	if v, ok := d.GetOk("integration_uri"); ok {
		input.IntegrationUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("passthrough_behavior"); ok {
		input.PassthroughBehavior = awstypes.PassthroughBehavior(v.(string))
	}

	if v, ok := d.GetOk("payload_format_version"); ok {
		input.PayloadFormatVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_parameters"); ok {
		input.RequestParameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("request_templates"); ok {
		input.RequestTemplates = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("response_parameters"); ok && v.(*schema.Set).Len() > 0 {
		input.ResponseParameters = expandIntegrationResponseParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_selection_expression"); ok {
		input.TemplateSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("timeout_milliseconds"); ok {
		input.TimeoutInMillis = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("tls_config"); ok {
		input.TlsConfig = expandTLSConfig(v.([]interface{}))
	}

	output, err := conn.CreateIntegration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Integration: %s", err)
	}

	d.SetId(aws.ToString(output.IntegrationId))

	return append(diags, resourceIntegrationRead(ctx, d, meta)...)
}

func resourceIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findIntegrationByTwoPartKey(ctx, conn, d.Get("api_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Integration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrConnectionID, output.ConnectionId)
	d.Set("connection_type", output.ConnectionType)
	d.Set("content_handling_strategy", output.ContentHandlingStrategy)
	d.Set("credentials_arn", output.CredentialsArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("integration_method", output.IntegrationMethod)
	d.Set("integration_response_selection_expression", output.IntegrationResponseSelectionExpression)
	d.Set("integration_subtype", output.IntegrationSubtype)
	d.Set("integration_type", output.IntegrationType)
	d.Set("integration_uri", output.IntegrationUri)
	d.Set("passthrough_behavior", output.PassthroughBehavior)
	d.Set("payload_format_version", output.PayloadFormatVersion)
	d.Set("request_parameters", output.RequestParameters)
	d.Set("request_templates", output.RequestTemplates)
	if err := d.Set("response_parameters", flattenIntegrationResponseParameters(output.ResponseParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting response_parameters: %s", err)
	}
	d.Set("template_selection_expression", output.TemplateSelectionExpression)
	d.Set("timeout_milliseconds", output.TimeoutInMillis)
	if err := d.Set("tls_config", flattenTLSConfig(output.TlsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tls_config: %s", err)
	}

	return diags
}

func resourceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	input := &apigatewayv2.UpdateIntegrationInput{
		ApiId:         aws.String(d.Get("api_id").(string)),
		IntegrationId: aws.String(d.Id()),
		// Always specify the integration type.
		IntegrationType: awstypes.IntegrationType(d.Get("integration_type").(string)),
	}

	if d.HasChange(names.AttrConnectionID) {
		input.ConnectionId = aws.String(d.Get(names.AttrConnectionID).(string))
	}

	if d.HasChange("connection_type") {
		input.ConnectionType = awstypes.ConnectionType(d.Get("connection_type").(string))
	}

	if d.HasChange("content_handling_strategy") {
		input.ContentHandlingStrategy = awstypes.ContentHandlingStrategy(d.Get("content_handling_strategy").(string))
	}

	if d.HasChange("credentials_arn") {
		input.CredentialsArn = aws.String(d.Get("credentials_arn").(string))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("integration_method") {
		input.IntegrationMethod = aws.String(d.Get("integration_method").(string))
	}

	// Always specify any integration subtype.
	if v, ok := d.GetOk("integration_subtype"); ok {
		input.IntegrationSubtype = aws.String(v.(string))
	}

	if d.HasChange("integration_uri") {
		input.IntegrationUri = aws.String(d.Get("integration_uri").(string))
	}

	if d.HasChange("passthrough_behavior") {
		input.PassthroughBehavior = awstypes.PassthroughBehavior(d.Get("passthrough_behavior").(string))
	}

	if d.HasChange("payload_format_version") {
		input.PayloadFormatVersion = aws.String(d.Get("payload_format_version").(string))
	}

	if d.HasChange("request_parameters") {
		o, n := d.GetChange("request_parameters")
		add, del, nop := flex.DiffStringValueMaps(o.(map[string]interface{}), n.(map[string]interface{}))

		// Parameters are removed by setting the associated value to "".
		for k := range del {
			del[k] = ""
		}
		variables := del
		for k, v := range add {
			variables[k] = v
		}
		// Also specify any request parameters that are unchanged as for AWS service integrations some parameters are always required:
		// https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-aws-services-reference.html
		for k, v := range nop {
			variables[k] = v
		}

		input.RequestParameters = variables
	}

	if d.HasChange("request_templates") {
		input.RequestTemplates = flex.ExpandStringValueMap(d.Get("request_templates").(map[string]interface{}))
	}

	if d.HasChange("response_parameters") {
		o, n := d.GetChange("response_parameters")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		del := os.Difference(ns).List()

		input.ResponseParameters = expandIntegrationResponseParameters(ns.List())

		// Parameters are removed by setting the associated value to {}.
		for _, tfMapRaw := range del {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			if v, ok := tfMap[names.AttrStatusCode].(string); ok && v != "" {
				if input.ResponseParameters == nil {
					input.ResponseParameters = map[string]map[string]string{}
				}
				input.ResponseParameters[v] = map[string]string{}
			}
		}
	}

	if d.HasChange("template_selection_expression") {
		input.TemplateSelectionExpression = aws.String(d.Get("template_selection_expression").(string))
	}

	if d.HasChange("timeout_milliseconds") {
		input.TimeoutInMillis = aws.Int32(int32(d.Get("timeout_milliseconds").(int)))
	}

	if d.HasChange("tls_config") {
		input.TlsConfig = expandTLSConfig(d.Get("tls_config").([]interface{}))
	}

	_, err := conn.UpdateIntegration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Integration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIntegrationRead(ctx, d, meta)...)
}

func resourceIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Integration: %s", d.Id())
	_, err := conn.DeleteIntegration(ctx, &apigatewayv2.DeleteIntegrationInput{
		ApiId:         aws.String(d.Get("api_id").(string)),
		IntegrationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Integration (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceIntegrationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/integration-id'", d.Id())
	}

	apiID := parts[0]
	integrationID := parts[1]

	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findIntegrationByTwoPartKey(ctx, conn, apiID, integrationID)

	if err != nil {
		return nil, err
	}

	if aws.ToBool(output.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 Integration (%s) was created via quick create", integrationID)
	}

	d.SetId(integrationID)
	d.Set("api_id", apiID)

	return []*schema.ResourceData{d}, nil
}

func findIntegrationByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, integrationID string) (*apigatewayv2.GetIntegrationOutput, error) {
	input := &apigatewayv2.GetIntegrationInput{
		ApiId:         aws.String(apiID),
		IntegrationId: aws.String(integrationID),
	}

	return findIntegration(ctx, conn, input)
}

func findIntegration(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetIntegrationInput) (*apigatewayv2.GetIntegrationOutput, error) {
	output, err := conn.GetIntegration(ctx, input)

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

func expandTLSConfig(vConfig []interface{}) *awstypes.TlsConfigInput {
	config := &awstypes.TlsConfigInput{}

	if len(vConfig) == 0 || vConfig[0] == nil {
		return config
	}
	mConfig := vConfig[0].(map[string]interface{})

	if vServerNameToVerify, ok := mConfig["server_name_to_verify"].(string); ok && vServerNameToVerify != "" {
		config.ServerNameToVerify = aws.String(vServerNameToVerify)
	}

	return config
}

func flattenTLSConfig(config *awstypes.TlsConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"server_name_to_verify": aws.ToString(config.ServerNameToVerify),
	}}
}

func expandIntegrationResponseParameters(tfList []interface{}) map[string]map[string]string {
	if len(tfList) == 0 {
		return nil
	}

	responseParameters := map[string]map[string]string{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if vStatusCode, ok := tfMap[names.AttrStatusCode].(string); ok && vStatusCode != "" {
			if v, ok := tfMap["mappings"].(map[string]interface{}); ok && len(v) > 0 {
				responseParameters[vStatusCode] = flex.ExpandStringValueMap(v)
			}
		}
	}

	return responseParameters
}

func flattenIntegrationResponseParameters(responseParameters map[string]map[string]string) []interface{} {
	if len(responseParameters) == 0 {
		return nil
	}

	var tfList []interface{}

	for statusCode, mappings := range responseParameters {
		if len(mappings) == 0 {
			continue
		}

		tfMap := map[string]interface{}{}

		tfMap[names.AttrStatusCode] = statusCode
		tfMap["mappings"] = mappings

		tfList = append(tfList, tfMap)
	}

	return tfList
}
