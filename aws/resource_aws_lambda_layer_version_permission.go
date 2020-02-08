package aws

import (
	// "errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	// "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsLambdaLayerVersionPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaLayerVersionPermissionAdd,
		Read:   resourceAwsLambdaLayerVersionPermissionGet,
		Delete: resourceAwsLambdaLayerVersionPermissionRemove,

		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
		// },

		Schema: map[string]*schema.Schema{
			"layer_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"layer_version": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"statement_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLambdaLayerVersionPermissionAdd(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerArn := aws.String(d.Get("layer_arn").(string))
	layerVersion := aws.Int64(int64(d.Get("layer_version").(int)))
	statementId := aws.String(d.Get("statement_id").(string))
	principal := aws.String(d.Get("principal").(string))
	organizationId, hasOrganizationId := d.GetOk("organization_id")
	action := aws.String(d.Get("action").(string))

	params := &lambda.AddLayerVersionPermissionInput{
		LayerName:     layerArn,
		VersionNumber: layerVersion,
		Action:        action,
		Principal:     principal,
		StatementId:   statementId,
	}

	if hasOrganizationId {
		params.OrganizationId = aws.String(organizationId.(string))
	}

	log.Printf("[DEBUG] Adding Lambda layer permissions: %s", params)
	result, err := conn.AddLayerVersionPermission(params)
	if err != nil {
		return fmt.Errorf("Error adding lambda layer permissions: %s", err)
	}

	// log.Printf(aws.StringValue(result.Statement))

	d.SetId(*layerArn + ":" + strconv.FormatInt(*layerVersion, 10))
	d.Set("revision_id", result.RevisionId)

	return resourceAwsLambdaLayerVersionPermissionGet(d, meta)
}

func resourceAwsLambdaLayerVersionPermissionGet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerName, version, err := resourceAwsLambdaLayerVersionPermissionParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing lambda layer ID: %s", err)
	}

	layerVersionPolicyOutput, err := conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(version),
	})

	log.Printf("[DEBUG] OUTPUT: %s", layerVersionPolicyOutput)

	if err != nil {
		return fmt.Errorf("error reading Lambda Layer version permission (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsLambdaLayerVersionPermissionRemove(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerName, version, err := resourceAwsLambdaLayerVersionPermissionParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing lambda layer ID: %s", err)
	}

	_, err = conn.RemoveLayerVersionPermission(&lambda.RemoveLayerVersionPermissionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(version),
		StatementId:   aws.String(d.Get("statement_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("error deleting Lambda Layer Version permission (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Lambda layer permission %q deleted", d.Get("statement_id").(string))
	return nil
}

func resourceAwsLambdaLayerVersionPermissionParseId(id string) (layerName string, version int64, err error) {
	arn, err := arn.Parse(id)
	if err != nil {
		return
	}
	parts := strings.Split(arn.Resource, ":")
	if len(parts) != 3 || parts[0] != "layer" {
		err = fmt.Errorf("lambda_layer ID must be a valid Layer ARN")
		return
	}

	layerName = parts[1]
	version, err = strconv.ParseInt(parts[2], 10, 64)
	return
}
