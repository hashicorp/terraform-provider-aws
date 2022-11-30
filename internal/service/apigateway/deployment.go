package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentCreate,
		Read:   resourceDeploymentRead,
		Update: resourceDeploymentUpdate,
		Delete: resourceDeploymentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stage_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"stage_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invoke_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	// Create the gateway
	log.Printf("[DEBUG] Creating API Gateway Deployment")

	deployment, err := conn.CreateDeployment(&apigateway.CreateDeploymentInput{
		RestApiId:        aws.String(d.Get("rest_api_id").(string)),
		StageName:        aws.String(d.Get("stage_name").(string)),
		Description:      aws.String(d.Get("description").(string)),
		StageDescription: aws.String(d.Get("stage_description").(string)),
		Variables:        flex.ExpandStringMap(d.Get("variables").(map[string]interface{})),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Deployment: %w", err)
	}

	d.SetId(aws.StringValue(deployment.Id))
	log.Printf("[DEBUG] API Gateway Deployment ID: %s", d.Id())

	return resourceDeploymentRead(d, meta)
}

func resourceDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Reading API Gateway Deployment %s", d.Id())
	restApiId := d.Get("rest_api_id").(string)
	out, err := conn.GetDeployment(&apigateway.GetDeploymentInput{
		RestApiId:    aws.String(restApiId),
		DeploymentId: aws.String(d.Id()),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Deployment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading API Gateway Deployment (%s): %w", d.Id(), err)
	}
	log.Printf("[DEBUG] Received API Gateway Deployment: %s", out)
	d.Set("description", out.Description)

	stageName := d.Get("stage_name").(string)

	d.Set("invoke_url", buildInvokeURL(meta.(*conns.AWSClient), restApiId, stageName))

	executionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("%s/%s", restApiId, stageName),
	}.String()
	d.Set("execution_arn", executionArn)

	if err := d.Set("created_date", out.CreatedDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_date: %s", err)
	}

	return nil
}

func resourceDeploymentUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	return operations
}

func resourceDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Updating API Gateway API Key: %s", d.Id())

	_, err := conn.UpdateDeployment(&apigateway.UpdateDeploymentInput{
		DeploymentId:    aws.String(d.Id()),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: resourceDeploymentUpdateOperations(d),
	})
	if err != nil {
		return err
	}

	return resourceDeploymentRead(d, meta)
}

func resourceDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Deleting API Gateway Deployment: %s", d.Id())

	// If the stage has been updated to point at a different deployment, then
	// the stage should not be removed when this deployment is deleted.
	shouldDeleteStage := false

	// API Gateway allows an empty state name (e.g. ""), but the AWS Go SDK
	// now has validation for the parameter, so we must check first.
	// InvalidParameter: 1 validation error(s) found.
	//  - minimum field size of 1, GetStageInput.StageName.
	stageName := d.Get("stage_name").(string)
	restApiId := d.Get("rest_api_id").(string)
	if stageName != "" {
		stage, err := FindStageByName(conn, restApiId, stageName)

		if err != nil && !tfresource.NotFound(err) {
			return fmt.Errorf("error getting referenced stage: %w", err)
		}

		if stage != nil && aws.StringValue(stage.DeploymentId) == d.Id() {
			shouldDeleteStage = true
		}
	}

	if shouldDeleteStage {
		if _, err := conn.DeleteStage(&apigateway.DeleteStageInput{
			StageName: aws.String(stageName),
			RestApiId: aws.String(restApiId),
		}); err == nil {
			return nil
		}
	}

	_, err := conn.DeleteDeployment(&apigateway.DeleteDeploymentInput{
		DeploymentId: aws.String(d.Id()),
		RestApiId:    aws.String(restApiId),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway Deployment (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceDeploymentImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("Unexpected format of ID (%s), use: 'REST-API-ID/DEPLOYMENT-ID'", d.Id())
	}

	restApiID := idParts[0]
	deploymentID := idParts[1]

	d.SetId(deploymentID)
	d.Set("rest_api_id", restApiID)

	return []*schema.ResourceData{d}, nil
}
