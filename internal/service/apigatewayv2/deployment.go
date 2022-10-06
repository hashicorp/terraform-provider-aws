package apigatewayv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentCreate,
		Read:   resourceDeploymentRead,
		Update: resourceDeploymentUpdate,
		Delete: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"auto_deployed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	req := &apigatewayv2.CreateDeploymentInput{
		ApiId: aws.String(d.Get("api_id").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 deployment: %s", req)
	resp, err := conn.CreateDeployment(req)
	if err != nil {
		return fmt.Errorf("creating API Gateway v2 deployment: %s", err)
	}

	d.SetId(aws.StringValue(resp.DeploymentId))

	if _, err := WaitDeploymentDeployed(conn, d.Get("api_id").(string), d.Id()); err != nil {
		return fmt.Errorf("waiting for API Gateway v2 deployment (%s) creation: %s", d.Id(), err)
	}

	return resourceDeploymentRead(d, meta)
}

func resourceDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	outputRaw, _, err := StatusDeployment(conn, d.Get("api_id").(string), d.Id())()
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading API Gateway v2 deployment: %s", err)
	}

	output := outputRaw.(*apigatewayv2.GetDeploymentOutput)
	d.Set("auto_deployed", output.AutoDeployed)
	d.Set("description", output.Description)

	return nil
}

func resourceDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

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
		return fmt.Errorf("updating API Gateway v2 deployment: %s", err)
	}

	if _, err := WaitDeploymentDeployed(conn, d.Get("api_id").(string), d.Id()); err != nil {
		return fmt.Errorf("waiting for API Gateway v2 deployment (%s) update: %s", d.Id(), err)
	}

	return resourceDeploymentRead(d, meta)
}

func resourceDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	log.Printf("[DEBUG] Deleting API Gateway v2 deployment (%s)", d.Id())
	_, err := conn.DeleteDeployment(&apigatewayv2.DeleteDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting API Gateway v2 deployment: %s", err)
	}

	return nil
}

func resourceDeploymentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/deployment-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
