package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsApiGateway2Api() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGateway2ApiCreate,
		Read:   resourceAwsApiGateway2ApiRead,
		Update: resourceAwsApiGateway2ApiUpdate,
		Delete: resourceAwsApiGateway2ApiDelete,
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
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.ProtocolTypeWebsocket,
				}, false),
			},
			"route_selection_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsApiGateway2ApiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateApiInput{
		Name:                     aws.String(d.Get("name").(string)),
		ProtocolType:             aws.String(d.Get("protocol_type").(string)),
		RouteSelectionExpression: aws.String(d.Get("route_selection_expression").(string)),
	}
	if v, ok := d.GetOk("api_key_selection_expression"); ok {
		req.ApiKeySelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
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

	return resourceAwsApiGateway2ApiRead(d, meta)
}

func resourceAwsApiGateway2ApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetApi(&apigatewayv2.GetApiInput{
		ApiId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API: %s", err)
	}

	d.Set("api_endpoint", resp.ApiEndpoint)
	d.Set("api_key_selection_expression", resp.ApiKeySelectionExpression)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("protocol_type", resp.ProtocolType)
	d.Set("route_selection_expression", resp.RouteSelectionExpression)
	d.Set("version", resp.Version)

	return nil
}

func resourceAwsApiGateway2ApiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateApiInput{
		ApiId: aws.String(d.Id()),
	}
	if d.HasChange("api_key_selection_expression") {
		req.ApiKeySelectionExpression = aws.String(d.Get("api_key_selection_expression").(string))
	}
	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
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
		return fmt.Errorf("error updating API Gateway v2 API: %s", err)
	}

	return resourceAwsApiGateway2ApiRead(d, meta)
}

func resourceAwsApiGateway2ApiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 API (%s)", d.Id())
	_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
		ApiId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 API: %s", err)
	}

	return nil
}
