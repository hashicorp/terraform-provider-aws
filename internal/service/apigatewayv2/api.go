package apigatewayv2

import (
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAPI() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPICreate,
		Read:   resourceAPIRead,
		Update: resourceAPIUpdate,
		Delete: resourceAPIDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
							Set:      hashStringCaseInsensitive,
						},
						"allow_methods": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"allow_origins": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
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
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"disable_execute_api_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"fail_on_warnings": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"body": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONOrYAMLDiffs,
				ValidateFunc:     verify.ValidStringIsJSONOrYAML,
			},
			"protocol_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(apigatewayv2.ProtocolType_Values(), false),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceImportOpenAPI(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig

	if body, ok := d.GetOk("body"); ok {
		revertReq := &apigatewayv2.UpdateApiInput{
			ApiId:       aws.String(d.Id()),
			Name:        aws.String(d.Get("name").(string)),
			Description: aws.String(d.Get("description").(string)),
			Version:     aws.String(d.Get("version").(string)),
		}

		log.Printf("[DEBUG] Updating API Gateway from OpenAPI spec %s", d.Id())
		importReq := &apigatewayv2.ReimportApiInput{
			ApiId: aws.String(d.Id()),
			Body:  aws.String(body.(string)),
		}

		if value, ok := d.GetOk("fail_on_warnings"); ok {
			importReq.FailOnWarnings = aws.Bool(value.(bool))
		}

		_, err := conn.ReimportApi(importReq)

		if err != nil {
			return fmt.Errorf("error importing API Gateway v2 API (%s) OpenAPI specification: %s", d.Id(), err)
		}

		tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

		corsConfiguration := d.Get("cors_configuration")

		if err := resourceAPIRead(d, meta); err != nil {
			return err
		}

		if !reflect.DeepEqual(corsConfiguration, d.Get("cors_configuration")) {
			if len(corsConfiguration.([]interface{})) == 0 {
				log.Printf("[DEBUG] Deleting CORS configuration for API Gateway v2 API (%s)", d.Id())
				_, err := conn.DeleteCorsConfiguration(&apigatewayv2.DeleteCorsConfigurationInput{
					ApiId: aws.String(d.Id()),
				})
				if err != nil {
					return fmt.Errorf("error deleting CORS configuration for API Gateway v2 API (%s): %s", d.Id(), err)
				}
			} else {
				revertReq.CorsConfiguration = expandCORSConfiguration(corsConfiguration.([]interface{}))
			}
		}

		if err := UpdateTags(conn, d.Get("arn").(string), d.Get("tags_all"), tags); err != nil {
			return fmt.Errorf("error updating API Gateway v2 API (%s) tags: %s", d.Id(), err)
		}

		log.Printf("[DEBUG] Reverting API Gateway v2 API: %s", revertReq)
		_, err = conn.UpdateApi(revertReq)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 API (%s): %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAPICreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	protocolType := d.Get("protocol_type").(string)
	req := &apigatewayv2.CreateApiInput{
		Name:         aws.String(d.Get("name").(string)),
		ProtocolType: aws.String(protocolType),
		Tags:         Tags(tags.IgnoreAWS()),
	}
	if v, ok := d.GetOk("api_key_selection_expression"); ok {
		req.ApiKeySelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("cors_configuration"); ok {
		req.CorsConfiguration = expandCORSConfiguration(v.([]interface{}))
	}
	if v, ok := d.GetOk("credentials_arn"); ok {
		req.CredentialsArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("disable_execute_api_endpoint"); ok {
		req.DisableExecuteApiEndpoint = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("route_key"); ok {
		req.RouteKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_selection_expression"); ok {
		req.RouteSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target"); ok {
		req.Target = aws.String(v.(string))
	}
	if v, ok := d.GetOk("version"); ok {
		req.Version = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 API: %s", req)
	resp, err := conn.CreateApi(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 API: %s", err)
	}

	d.SetId(aws.StringValue(resp.ApiId))

	err = resourceImportOpenAPI(d, meta)
	if err != nil {
		return err
	}

	return resourceAPIRead(d, meta)
}

func resourceAPIRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.GetApi(&apigatewayv2.GetApiInput{
		ApiId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API (%s): %s", d.Id(), err)
	}

	d.Set("api_endpoint", resp.ApiEndpoint)
	d.Set("api_key_selection_expression", resp.ApiKeySelectionExpression)
	apiArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/apis/%s", d.Id()),
	}.String()
	d.Set("arn", apiArn)
	if err := d.Set("cors_configuration", flattenCORSConfiguration(resp.CorsConfiguration)); err != nil {
		return fmt.Errorf("error setting cors_configuration: %s", err)
	}
	d.Set("description", resp.Description)
	d.Set("disable_execute_api_endpoint", resp.DisableExecuteApiEndpoint)
	executionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionArn)
	d.Set("name", resp.Name)
	d.Set("protocol_type", resp.ProtocolType)
	d.Set("route_selection_expression", resp.RouteSelectionExpression)

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}
	d.Set("version", resp.Version)

	return nil
}

func resourceAPIUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	deleteCorsConfiguration := false
	if d.HasChange("cors_configuration") {
		v := d.Get("cors_configuration")
		if len(v.([]interface{})) == 0 {
			deleteCorsConfiguration = true

			log.Printf("[DEBUG] Deleting CORS configuration for API Gateway v2 API (%s)", d.Id())
			_, err := conn.DeleteCorsConfiguration(&apigatewayv2.DeleteCorsConfigurationInput{
				ApiId: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("error deleting CORS configuration for API Gateway v2 API (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("api_key_selection_expression", "description", "disable_execute_api_endpoint", "name", "route_selection_expression", "version") ||
		(d.HasChange("cors_configuration") && !deleteCorsConfiguration) {
		req := &apigatewayv2.UpdateApiInput{
			ApiId: aws.String(d.Id()),
		}

		if d.HasChange("api_key_selection_expression") {
			req.ApiKeySelectionExpression = aws.String(d.Get("api_key_selection_expression").(string))
		}
		if d.HasChange("cors_configuration") {
			req.CorsConfiguration = expandCORSConfiguration(d.Get("cors_configuration").([]interface{}))
		}
		if d.HasChange("description") {
			req.Description = aws.String(d.Get("description").(string))
		}
		if d.HasChange("disable_execute_api_endpoint") {
			req.DisableExecuteApiEndpoint = aws.Bool(d.Get("disable_execute_api_endpoint").(bool))
		}
		if d.HasChange("name") {
			req.Name = aws.String(d.Get("name").(string))
		}
		if d.HasChange("route_selection_expression") {
			req.RouteSelectionExpression = aws.String(d.Get("route_selection_expression").(string))
		}
		if d.HasChange("version") {
			req.Version = aws.String(d.Get("version").(string))
		}

		log.Printf("[DEBUG] Updating API Gateway v2 API: %s", req)
		_, err := conn.UpdateApi(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 API (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 API (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("body") {
		err := resourceImportOpenAPI(d, meta)
		if err != nil {
			return err
		}
	}

	return resourceAPIRead(d, meta)
}

func resourceAPIDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	log.Printf("[DEBUG] Deleting API Gateway v2 API (%s)", d.Id())
	_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
		ApiId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 API (%s): %s", d.Id(), err)
	}

	return nil
}

func expandCORSConfiguration(vConfiguration []interface{}) *apigatewayv2.Cors {
	configuration := &apigatewayv2.Cors{}

	if len(vConfiguration) == 0 || vConfiguration[0] == nil {
		return configuration
	}
	mConfiguration := vConfiguration[0].(map[string]interface{})

	if vAllowCredentials, ok := mConfiguration["allow_credentials"].(bool); ok {
		configuration.AllowCredentials = aws.Bool(vAllowCredentials)
	}
	if vAllowHeaders, ok := mConfiguration["allow_headers"].(*schema.Set); ok {
		configuration.AllowHeaders = flex.ExpandStringSet(vAllowHeaders)
	}
	if vAllowMethods, ok := mConfiguration["allow_methods"].(*schema.Set); ok {
		configuration.AllowMethods = flex.ExpandStringSet(vAllowMethods)
	}
	if vAllowOrigins, ok := mConfiguration["allow_origins"].(*schema.Set); ok {
		configuration.AllowOrigins = flex.ExpandStringSet(vAllowOrigins)
	}
	if vExposeHeaders, ok := mConfiguration["expose_headers"].(*schema.Set); ok {
		configuration.ExposeHeaders = flex.ExpandStringSet(vExposeHeaders)
	}
	if vMaxAge, ok := mConfiguration["max_age"].(int); ok {
		configuration.MaxAge = aws.Int64(int64(vMaxAge))
	}

	return configuration
}

func flattenCORSConfiguration(configuration *apigatewayv2.Cors) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"allow_credentials": aws.BoolValue(configuration.AllowCredentials),
		"allow_headers":     flattenCaseInsensitiveStringSet(configuration.AllowHeaders),
		"allow_methods":     flattenCaseInsensitiveStringSet(configuration.AllowMethods),
		"allow_origins":     flattenCaseInsensitiveStringSet(configuration.AllowOrigins),
		"expose_headers":    flattenCaseInsensitiveStringSet(configuration.ExposeHeaders),
		"max_age":           int(aws.Int64Value(configuration.MaxAge)),
	}}
}
