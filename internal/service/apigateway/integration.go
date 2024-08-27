// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
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

// @SDKResource("aws_api_gateway_integration", name="Integration")
func resourceIntegration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIntegrationCreate,
		ReadWithoutTimeout:   resourceIntegrationRead,
		UpdateWithoutTimeout: resourceIntegrationUpdate,
		DeleteWithoutTimeout: resourceIntegrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID/HTTP-METHOD", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				httpMethod := idParts[2]
				d.Set("http_method", httpMethod)
				d.Set(names.AttrResourceID, resourceID)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("agi-%s-%s-%s", restApiID, resourceID, httpMethod))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"cache_key_parameters": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"cache_namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"connection_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ConnectionTypeInternet,
				ValidateDiagFunc: enum.Validate[types.ConnectionType](),
			},
			"content_handling": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validIntegrationContentHandling(),
			},
			"credentials": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validHTTPMethod(),
			},
			"integration_http_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validHTTPMethod(),
			},
			"passthrough_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"WHEN_NO_MATCH",
					"WHEN_NO_TEMPLATES",
					"NEVER",
				}, false),
			},
			"request_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"request_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"timeout_milliseconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(50, 300000),
				Default:      29000,
			},
			"tls_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"insecure_skip_verification": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.IntegrationType](),
			},
			names.AttrURI: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.PutIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		Type:       types.IntegrationType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("cache_key_parameters"); ok && v.(*schema.Set).Len() > 0 {
		input.CacheKeyParameters = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("cache_namespace"); ok {
		input.CacheNamespace = aws.String(v.(string))
	} else if input.CacheKeyParameters != nil {
		input.CacheNamespace = aws.String(d.Get(names.AttrResourceID).(string))
	}

	if v, ok := d.GetOk(names.AttrConnectionID); ok {
		input.ConnectionId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connection_type"); ok {
		input.ConnectionType = types.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("content_handling"); ok {
		input.ContentHandling = types.ContentHandlingStrategy(v.(string))
	}

	if v, ok := d.GetOk("credentials"); ok {
		input.Credentials = aws.String(v.(string))
	}

	if v, ok := d.GetOk("integration_http_method"); ok {
		input.IntegrationHttpMethod = aws.String(v.(string))
	}

	if v, ok := d.GetOk("passthrough_behavior"); ok {
		input.PassthroughBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.RequestParameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("request_templates"); ok && len(v.(map[string]interface{})) > 0 {
		input.RequestTemplates = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("timeout_milliseconds"); ok {
		input.TimeoutInMillis = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("tls_config"); ok && len(v.([]interface{})) > 0 {
		input.TlsConfig = expandTLSConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrURI); ok {
		input.Uri = aws.String(v.(string))
	}

	_, err := conn.PutIntegration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Integration: %s", err)
	}

	d.SetId(fmt.Sprintf("agi-%s-%s-%s", d.Get("rest_api_id").(string), d.Get(names.AttrResourceID).(string), d.Get("http_method").(string)))

	return append(diags, resourceIntegrationRead(ctx, d, meta)...)
}

func resourceIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	integration, err := findIntegrationByThreePartKey(ctx, conn, d.Get("http_method").(string), d.Get(names.AttrResourceID).(string), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Integration (%s): %s", d.Id(), err)
	}

	d.Set("cache_key_parameters", integration.CacheKeyParameters)
	d.Set("cache_namespace", integration.CacheNamespace)
	d.Set(names.AttrConnectionID, integration.ConnectionId)
	d.Set("connection_type", types.ConnectionTypeInternet)
	if integration.ConnectionType != "" {
		d.Set("connection_type", integration.ConnectionType)
	}
	d.Set("content_handling", integration.ContentHandling)
	d.Set("credentials", integration.Credentials)
	d.Set("integration_http_method", integration.HttpMethod)
	d.Set("passthrough_behavior", integration.PassthroughBehavior)
	d.Set("request_parameters", integration.RequestParameters)
	// We need to explicitly convert key = nil values into key = "", which aws.ToStringMap() removes
	requestTemplates := make(map[string]string)
	for k, v := range integration.RequestTemplates {
		requestTemplates[k] = v
	}
	d.Set("request_templates", requestTemplates)
	d.Set("timeout_milliseconds", integration.TimeoutInMillis)
	d.Set(names.AttrType, integration.Type)
	d.Set(names.AttrURI, integration.Uri)

	if err := d.Set("tls_config", flattenTLSConfig(integration.TlsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tls_config: %s", err)
	}

	return diags
}

func resourceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)

	// https://docs.aws.amazon.com/apigateway/api-reference/link-relation/integration-update/#remarks
	// According to the above documentation, only a few parts are addable / removable.
	if d.HasChange("request_templates") {
		o, n := d.GetChange("request_templates")
		prefix := "requestTemplates"

		os := o.(map[string]interface{})
		ns := n.(map[string]interface{})

		// Handle Removal
		for k := range os {
			if _, ok := ns[k]; !ok {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
				})
			}
		}

		for k, v := range ns {
			// Handle replaces
			if _, ok := os[k]; ok {
				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
					Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
					Value: aws.String(v.(string)),
				})
			}

			// Handle additions
			if _, ok := os[k]; !ok {
				operations = append(operations, types.PatchOperation{
					Op:    types.OpAdd,
					Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
					Value: aws.String(v.(string)),
				})
			}
		}
	}

	if d.HasChange("request_parameters") {
		o, n := d.GetChange("request_parameters")
		prefix := "requestParameters"

		os := o.(map[string]interface{})
		ns := n.(map[string]interface{})

		// Handle Removal
		for k := range os {
			if _, ok := ns[k]; !ok {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
				})
			}
		}

		for k, v := range ns {
			// Handle replaces
			if _, ok := os[k]; ok {
				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
					Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
					Value: aws.String(v.(string)),
				})
			}

			// Handle additions
			if _, ok := os[k]; !ok {
				operations = append(operations, types.PatchOperation{
					Op:    types.OpAdd,
					Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
					Value: aws.String(v.(string)),
				})
			}
		}
	}

	if d.HasChange("cache_key_parameters") {
		o, n := d.GetChange("cache_key_parameters")

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpRemove,
				Path:  aws.String(fmt.Sprintf("/cacheKeyParameters/%s", v.(string))),
				Value: aws.String(""),
			})
		}

		additionList := ns.Difference(os)
		for _, v := range additionList.List() {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpAdd,
				Path:  aws.String(fmt.Sprintf("/cacheKeyParameters/%s", v.(string))),
				Value: aws.String(""),
			})
		}
	}

	if d.HasChange("cache_namespace") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/cacheNamespace"),
			Value: aws.String(d.Get("cache_namespace").(string)),
		})
	}

	// The documentation https://docs.aws.amazon.com/apigateway/api-reference/link-relation/integration-update/ says
	// that uri changes are only supported for non-mock types. Because the uri value is not used in mock
	// resources, it means that the uri can always be updated
	if d.HasChange(names.AttrURI) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/uri"),
			Value: aws.String(d.Get(names.AttrURI).(string)),
		})
	}

	if d.HasChange("content_handling") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/contentHandling"),
			Value: aws.String(d.Get("content_handling").(string)),
		})
	}

	if d.HasChange("connection_type") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/connectionType"),
			Value: aws.String(d.Get("connection_type").(string)),
		})
	}

	if d.HasChange(names.AttrConnectionID) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/connectionId"),
			Value: aws.String(d.Get(names.AttrConnectionID).(string)),
		})
	}

	if d.HasChange("timeout_milliseconds") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/timeoutInMillis"),
			Value: aws.String(strconv.Itoa(d.Get("timeout_milliseconds").(int))),
		})
	}

	if d.HasChange("tls_config") {
		if v, ok := d.GetOk("tls_config"); ok && len(v.([]interface{})) > 0 {
			m := v.([]interface{})[0].(map[string]interface{})

			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/tlsConfig/insecureSkipVerification"),
				Value: aws.String(strconv.FormatBool(m["insecure_skip_verification"].(bool))),
			})
		}
	}

	input := &apigateway.UpdateIntegrationInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		PatchOperations: operations,
		ResourceId:      aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
	}

	_, err := conn.UpdateIntegration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Integration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIntegrationRead(ctx, d, meta)...)
}

func resourceIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Integration: %s", d.Id())
	_, err := conn.DeleteIntegration(ctx, &apigateway.DeleteIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Integration (%s): %s", d.Id(), err)
	}

	return diags
}

func findIntegrationByThreePartKey(ctx context.Context, conn *apigateway.Client, httpMethod, resourceID, apiID string) (*apigateway.GetIntegrationOutput, error) {
	input := &apigateway.GetIntegrationInput{
		HttpMethod: aws.String(httpMethod),
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
	}

	output, err := conn.GetIntegration(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

func expandTLSConfig(vConfig []interface{}) *types.TlsConfig {
	config := &types.TlsConfig{}

	if len(vConfig) == 0 || vConfig[0] == nil {
		return config
	}
	mConfig := vConfig[0].(map[string]interface{})

	if insecureSkipVerification, ok := mConfig["insecure_skip_verification"].(bool); ok {
		config.InsecureSkipVerification = insecureSkipVerification
	}
	return config
}

func flattenTLSConfig(config *types.TlsConfig) []interface{} {
	if config == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"insecure_skip_verification": config.InsecureSkipVerification,
	}}
}
