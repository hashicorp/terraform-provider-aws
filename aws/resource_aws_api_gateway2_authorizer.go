package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsApiGateway2Authorizer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGateway2AuthorizerCreate,
		Read:   resourceAwsApiGateway2AuthorizerRead,
		Update: resourceAwsApiGateway2AuthorizerUpdate,
		Delete: resourceAwsApiGateway2AuthorizerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGateway2AuthorizerImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorizer_credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"authorizer_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.AuthorizerTypeRequest,
				}, false),
			},
			"authorizer_uri": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"identity_sources": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceAwsApiGateway2AuthorizerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateAuthorizerInput{
		ApiId:          aws.String(d.Get("api_id").(string)),
		AuthorizerType: aws.String(d.Get("authorizer_type").(string)),
		AuthorizerUri:  aws.String(d.Get("authorizer_uri").(string)),
		IdentitySource: expandStringSet(d.Get("identity_sources").(*schema.Set)),
		Name:           aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("authorizer_credentials_arn"); ok {
		req.AuthorizerCredentialsArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 authorizer: %s", req)
	resp, err := conn.CreateAuthorizer(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 authorizer: %s", err)
	}

	d.SetId(aws.StringValue(resp.AuthorizerId))

	return resourceAwsApiGateway2AuthorizerRead(d, meta)
}

func resourceAwsApiGateway2AuthorizerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 authorizer: %s", err)
	}

	d.Set("authorizer_credentials_arn", resp.AuthorizerCredentialsArn)
	d.Set("authorizer_type", resp.AuthorizerType)
	d.Set("authorizer_uri", resp.AuthorizerUri)
	if err := d.Set("identity_sources", flattenStringSet(resp.IdentitySource)); err != nil {
		return fmt.Errorf("error setting identity_sources: %s", err)
	}
	d.Set("name", resp.Name)

	return nil
}

func resourceAwsApiGateway2AuthorizerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	}
	if d.HasChange("authorizer_credentials_arn") {
		req.AuthorizerCredentialsArn = aws.String(d.Get("authorizer_credentials_arn").(string))
	}
	if d.HasChange("authorizer_type") {
		req.AuthorizerType = aws.String(d.Get("authorizer_type").(string))
	}
	if d.HasChange("authorizer_uri") {
		req.AuthorizerUri = aws.String(d.Get("authorizer_uri").(string))
	}
	if d.HasChange("identity_sources") {
		req.IdentitySource = expandStringSet(d.Get("identity_sources").(*schema.Set))
	}
	if d.HasChange("name") {
		req.Name = aws.String(d.Get("name").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 authorizer: %s", req)
	_, err := conn.UpdateAuthorizer(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 authorizer: %s", err)
	}

	return resourceAwsApiGateway2AuthorizerRead(d, meta)
}

func resourceAwsApiGateway2AuthorizerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 authorizer (%s)", d.Id())
	_, err := conn.DeleteAuthorizer(&apigatewayv2.DeleteAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 authorizer: %s", err)
	}

	return nil
}

func resourceAwsApiGateway2AuthorizerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/authorizer-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
