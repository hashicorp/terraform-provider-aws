package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayV2Route() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2RouteCreate,
		Read:   resourceAwsApiGatewayV2RouteRead,
		Update: resourceAwsApiGatewayV2RouteUpdate,
		Delete: resourceAwsApiGatewayV2RouteDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				d.Set("request_validator_id", resourceID)
				d.Set("api_id", restApiID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorization_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authorizer_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"api_key_required": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"authorization_scopes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"model_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"request_parameters": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			// "request_models": {
			// 	Type:     schema.TypeSet,
			// 	Optional: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"name": {
			// 				Type:     schema.TypeString,
			// 				Required: true,
			// 			},
			// 			"required": {
			// 				Type:     schema.TypeBool,
			// 				Required: true,
			// 			},
			// 		},
			// 	},
			// },
			"target": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"operation_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"route_response_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsApiGatewayV2RouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	log.Printf("[DEBUG] Creating API Gateway V2 for API %s", d.Get("route_key").(string))

	var err error
	createRouteInput := &apigatewayv2.CreateRouteInput{
		ApiId:    aws.String(d.Get("api_id").(string)),
		RouteKey: aws.String(d.Get("route_key").(string)),
	}
	if v, ok := d.GetOk("api_key_required"); ok {
		createRouteInput.ApiKeyRequired = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("authorization_type"); ok {
		createRouteInput.AuthorizationType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("model_selection_expression"); ok {
		createRouteInput.ModelSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("operation_name"); ok {
		createRouteInput.OperationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("route_response_selection_expression"); ok {
		createRouteInput.RouteResponseSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target"); ok {
		createRouteInput.Target = aws.String(v.(string))
	}

	resource, err := conn.CreateRoute(createRouteInput)

	if err != nil {
		return fmt.Errorf("Error creating API Gateway V2: %s", err)
	}

	d.SetId(*resource.RouteId)

	return resourceAwsApiGatewayV2RouteRead(d, meta)
}

func resourceAwsApiGatewayV2RouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Reading API Gateway V2 %s", d.Id())
	resource, err := conn.GetApi(&apigatewayv2.GetApiInput{
		ApiId: aws.String(d.Id()),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] API Gateway V2 (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", resource.Name)
	d.Set("api_id", resource.ApiId)
	d.Set("description", resource.Description)
	d.Set("route_selection_expression", resource.RouteSelectionExpression)
	d.Set("protocol_type", resource.ProtocolType)
	d.Set("api_key_selection_expression", resource.ApiKeySelectionExpression)

	return nil
}

func resourceAwsApiGatewayV2RouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Updating API Gateway Resource %s", d.Id())
	updateApiConfig := &apigatewayv2.UpdateApiInput{
		ApiId:                    aws.String(d.Get("api_id").(string)),
		Description:              aws.String(d.Get("description").(string)),
		Name:                     aws.String(d.Get("name").(string)),
		RouteSelectionExpression: aws.String(d.Get("route_selection_expression").(string)),
	}

	if v, ok := d.GetOk("api_key_selection_expression"); ok {
		updateApiConfig.ApiKeySelectionExpression = aws.String(v.(string))
	}

	_, err := conn.UpdateApi(updateApiConfig)

	if err != nil {
		return err
	}

	return resourceAwsApiGatewayV2RouteRead(d, meta)
}

func resourceAwsApiGatewayV2RouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	log.Printf("[DEBUG] Deleting API Gateway V2: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] schema is %#v", d)
		_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
			ApiId: aws.String(d.Id()),
		})
		if err == nil {
			return nil
		}

		if apigatewayErr, ok := err.(awserr.Error); ok && apigatewayErr.Code() == "NotFoundException" {
			return nil
		}

		return resource.NonRetryableError(err)
	})
}
