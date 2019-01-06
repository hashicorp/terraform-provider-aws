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

func resourceAwsApiGatewayV2() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2Create,
		Read:   resourceAwsApiGatewayV2Read,
		Update: resourceAwsApiGatewayV2Update,
		Delete: resourceAwsApiGatewayV2Delete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				d.Set("request_validator_id", resourceID)
				d.Set("rest_api_id", restApiID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"route_selection_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsApiGatewayV2Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	log.Printf("[DEBUG] Creating API Gateway V2 for API %s", d.Get("name").(string))

	var err error
	resource, err := conn.CreateApi(&apigatewayv2.CreateApiInput{
		Name:                     aws.String(d.Get("name").(string)),
		ProtocolType:             aws.String(d.Get("protocol_type").(string)),
		RouteSelectionExpression: aws.String(d.Get("route_selection_expression").(string)),
		Description:              aws.String(d.Get("description").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error creating API Gateway V2: %s", err)
	}

	d.SetId(*resource.ApiId)

	return resourceAwsApiGatewayV2Read(d, meta)
}

func resourceAwsApiGatewayV2Read(d *schema.ResourceData, meta interface{}) error {
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
	d.Set("description", resource.Description)
	d.Set("route_selection_expression", resource.RouteSelectionExpression)
	d.Set("protocol_type", resource.ProtocolType)

	return nil
}

func resourceAwsApiGatewayV2Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Updating API Gateway Resource %s", d.Id())
	_, err := conn.UpdateApi(&apigatewayv2.UpdateApiInput{
		ApiId:                    aws.String(d.Id()),
		Description:              aws.String(d.Get("description").(string)),
		Name:                     aws.String(d.Get("name").(string)),
		RouteSelectionExpression: aws.String(d.Get("route_selection_expression").(string)),
	})

	if err != nil {
		return err
	}

	return resourceAwsApiGatewayV2Read(d, meta)
}

func resourceAwsApiGatewayV2Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	log.Printf("[DEBUG] Deleting API Gateway V2: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] schema is %#v", d)
		_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
			ApiId: aws.String(d.Get("api_id").(string)),
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
