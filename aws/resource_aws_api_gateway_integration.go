package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsApiGatewayIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayIntegrationCreate,
		Read:   resourceAwsApiGatewayIntegrationRead,
		Update: resourceAwsApiGatewayIntegrationUpdate,
		Delete: resourceAwsApiGatewayIntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID/HTTP-METHOD", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				httpMethod := idParts[2]
				d.Set("http_method", httpMethod)
				d.Set("resource_id", resourceID)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("agi-%s-%s-%s", restApiID, resourceID, httpMethod))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateHTTPMethod(),
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(apigateway.IntegrationType_Values(), false),
			},

			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      apigateway.ConnectionTypeInternet,
				ValidateFunc: validation.StringInSlice(apigateway.ConnectionType_Values(), false),
			},

			"connection_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"uri": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"credentials": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"integration_http_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateHTTPMethod(),
			},

			"request_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"request_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"content_handling": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateApiGatewayIntegrationContentHandling(),
			},

			"passthrough_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				//lintignore:AWSV001 // no enum in AWS SDK
				ValidateFunc: validation.StringInSlice([]string{
					"WHEN_NO_MATCH",
					"WHEN_NO_TEMPLATES",
					"NEVER",
				}, false),
			},

			"cache_key_parameters": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			"cache_namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"timeout_milliseconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(50, 29000),
				Default:      29000,
			},

			"tls_config": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
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
		},
	}
}

func resourceAwsApiGatewayIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	log.Print("[DEBUG] Creating API Gateway Integration")

	input := &apigateway.PutIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		Type:       aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("cache_key_parameters"); ok && v.(*schema.Set).Len() > 0 {
		input.CacheKeyParameters = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("cache_namespace"); ok {
		input.CacheNamespace = aws.String(v.(string))
	} else if input.CacheKeyParameters != nil {
		input.CacheNamespace = aws.String(d.Get("resource_id").(string))
	}

	if v, ok := d.GetOk("connection_id"); ok {
		input.ConnectionId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connection_type"); ok {
		input.ConnectionType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_handling"); ok {
		input.ContentHandling = aws.String(v.(string))
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
		input.RequestParameters = expandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("request_templates"); ok && len(v.(map[string]interface{})) > 0 {
		input.RequestTemplates = expandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("timeout_milliseconds"); ok {
		input.TimeoutInMillis = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tls_config"); ok && len(v.([]interface{})) > 0 {
		input.TlsConfig = expandApiGatewayTlsConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("uri"); ok {
		input.Uri = aws.String(v.(string))
	}

	_, err := conn.PutIntegration(input)

	if err != nil {
		return fmt.Errorf("Error creating API Gateway Integration: %s", err)
	}

	d.SetId(fmt.Sprintf("agi-%s-%s-%s", d.Get("rest_api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))

	return resourceAwsApiGatewayIntegrationRead(d, meta)
}

func resourceAwsApiGatewayIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	log.Printf("[DEBUG] Reading API Gateway Integration: %s", d.Id())
	integration, err := conn.GetIntegration(&apigateway.GetIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway Integration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Integration: %s", integration)

	if err := d.Set("cache_key_parameters", flattenStringList(integration.CacheKeyParameters)); err != nil {
		return fmt.Errorf("error setting cache_key_parameters: %s", err)
	}
	d.Set("cache_namespace", integration.CacheNamespace)
	d.Set("connection_id", integration.ConnectionId)
	d.Set("connection_type", apigateway.ConnectionTypeInternet)
	if integration.ConnectionType != nil {
		d.Set("connection_type", integration.ConnectionType)
	}
	d.Set("content_handling", integration.ContentHandling)
	d.Set("credentials", integration.Credentials)
	d.Set("integration_http_method", integration.HttpMethod)
	d.Set("passthrough_behavior", integration.PassthroughBehavior)

	if err := d.Set("request_parameters", aws.StringValueMap(integration.RequestParameters)); err != nil {
		return fmt.Errorf("error setting request_parameters: %s", err)
	}

	// We need to explicitly convert key = nil values into key = "", which aws.StringValueMap() removes
	requestTemplateMap := make(map[string]string)
	for key, valuePointer := range integration.RequestTemplates {
		requestTemplateMap[key] = aws.StringValue(valuePointer)
	}
	if err := d.Set("request_templates", requestTemplateMap); err != nil {
		return fmt.Errorf("error setting request_templates: %s", err)
	}

	d.Set("timeout_milliseconds", integration.TimeoutInMillis)
	d.Set("type", integration.Type)
	d.Set("uri", integration.Uri)

	if err := d.Set("tls_config", flattenApiGatewayTlsConfig(integration.TlsConfig)); err != nil {
		return fmt.Errorf("error setting tls_config: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	log.Printf("[DEBUG] Updating API Gateway Integration: %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)

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
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
					Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
				})
			}
		}

		for k, v := range ns {
			// Handle replaces
			if _, ok := os[k]; ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpReplace),
					Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
					Value: aws.String(v.(string)),
				})
			}

			// Handle additions
			if _, ok := os[k]; !ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpAdd),
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
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
					Path: aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
				})
			}
		}

		for k, v := range ns {
			// Handle replaces
			if _, ok := os[k]; ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpReplace),
					Path:  aws.String(fmt.Sprintf("/%s/%s", prefix, strings.Replace(k, "/", "~1", -1))),
					Value: aws.String(v.(string)),
				})
			}

			// Handle additions
			if _, ok := os[k]; !ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpAdd),
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
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpRemove),
				Path:  aws.String(fmt.Sprintf("/cacheKeyParameters/%s", v.(string))),
				Value: aws.String(""),
			})
		}

		additionList := ns.Difference(os)
		for _, v := range additionList.List() {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpAdd),
				Path:  aws.String(fmt.Sprintf("/cacheKeyParameters/%s", v.(string))),
				Value: aws.String(""),
			})
		}
	}

	if d.HasChange("cache_namespace") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/cacheNamespace"),
			Value: aws.String(d.Get("cache_namespace").(string)),
		})
	}

	// The documentation https://docs.aws.amazon.com/apigateway/api-reference/link-relation/integration-update/ says
	// that uri changes are only supported for non-mock types. Because the uri value is not used in mock
	// resources, it means that the uri can always be updated
	if d.HasChange("uri") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/uri"),
			Value: aws.String(d.Get("uri").(string)),
		})
	}

	if d.HasChange("content_handling") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/contentHandling"),
			Value: aws.String(d.Get("content_handling").(string)),
		})
	}

	if d.HasChange("connection_type") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/connectionType"),
			Value: aws.String(d.Get("connection_type").(string)),
		})
	}

	if d.HasChange("connection_id") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/connectionId"),
			Value: aws.String(d.Get("connection_id").(string)),
		})
	}

	if d.HasChange("timeout_milliseconds") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/timeoutInMillis"),
			Value: aws.String(strconv.Itoa(d.Get("timeout_milliseconds").(int))),
		})
	}

	if d.HasChange("tls_config") {
		if v, ok := d.GetOk("tls_config"); ok && len(v.([]interface{})) > 0 {
			m := v.([]interface{})[0].(map[string]interface{})

			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/tlsConfig/insecureSkipVerification"),
				Value: aws.String(strconv.FormatBool(m["insecure_skip_verification"].(bool))),
			})
		}
	}

	params := &apigateway.UpdateIntegrationInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		ResourceId:      aws.String(d.Get("resource_id").(string)),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: operations,
	}

	_, err := conn.UpdateIntegration(params)
	if err != nil {
		return fmt.Errorf("Error updating API Gateway Integration: %s", err)
	}

	d.SetId(fmt.Sprintf("agi-%s-%s-%s", d.Get("rest_api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))

	return resourceAwsApiGatewayIntegrationRead(d, meta)
}

func resourceAwsApiGatewayIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Deleting API Gateway Integration: %s", d.Id())

	_, err := conn.DeleteIntegration(&apigateway.DeleteIntegrationInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})

	if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway Integration (%s): %s", d.Id(), err)
	}

	return nil
}

func expandApiGatewayTlsConfig(vConfig []interface{}) *apigateway.TlsConfig {
	config := &apigateway.TlsConfig{}

	if len(vConfig) == 0 || vConfig[0] == nil {
		return config
	}
	mConfig := vConfig[0].(map[string]interface{})

	if insecureSkipVerification, ok := mConfig["insecure_skip_verification"].(bool); ok {
		config.InsecureSkipVerification = aws.Bool(insecureSkipVerification)
	}
	return config
}

func flattenApiGatewayTlsConfig(config *apigateway.TlsConfig) []interface{} {
	if config == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"insecure_skip_verification": aws.BoolValue(config.InsecureSkipVerification),
	}}
}
