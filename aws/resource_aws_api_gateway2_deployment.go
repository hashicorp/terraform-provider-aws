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

func resourceAwsApiGateway2Deployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGateway2DeploymentCreate,
		Read:   resourceAwsApiGateway2DeploymentRead,
		Update: resourceAwsApiGateway2DeploymentUpdate,
		Delete: resourceAwsApiGateway2DeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGateway2DeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
		},
	}
}

func resourceAwsApiGateway2DeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateDeploymentInput{
		ApiId: aws.String(d.Get("api_id").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 deployment: %s", req)
	resp, err := conn.CreateDeployment(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 deployment: %s", err)
	}

	d.SetId(aws.StringValue(resp.DeploymentId))

	return resourceAwsApiGateway2DeploymentRead(d, meta)
}

func resourceAwsApiGateway2DeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetDeployment(&apigatewayv2.GetDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 deployment: %s", err)
	}

	d.Set("description", resp.Description)

	return nil
}

func resourceAwsApiGateway2DeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	}
	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 deployment: %s", req)
	_, err := conn.UpdateDeployment(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 deployment: %s", err)
	}

	return resourceAwsApiGateway2DeploymentRead(d, meta)
}

func resourceAwsApiGateway2DeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 deployment (%s)", d.Id())
	_, err := conn.DeleteDeployment(&apigatewayv2.DeleteDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 deployment: %s", err)
	}

	return nil
}

func resourceAwsApiGateway2DeploymentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/deployment-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
