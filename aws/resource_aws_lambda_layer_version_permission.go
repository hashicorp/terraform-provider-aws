package aws

import (
	// "errors"
	"fmt"
	"log"
	"strconv"
	// "strings"

	"github.com/aws/aws-sdk-go/aws"
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
				Optional: true, // add default value lambda:GetLayerVersion ??
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
	action := aws.String("lambda:GetLayerVersion")
	if d.Get("action") != nil {
		action = aws.String(d.Get("action").(string))
	}

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

	// if v, ok := d.GetOk("compatible_runtimes"); ok && v.(*schema.Set).Len() > 0 {
	// 	params.CompatibleRuntimes = expandStringList(v.(*schema.Set).List())
	// }

	log.Printf("[DEBUG] Adding Lambda layer permissions: %s", params)
	result, err := conn.AddLayerVersionPermission(params)
	if err != nil {
		return fmt.Errorf("Error adding lambda layer permissions: %s", err)
	}

	log.Printf(aws.StringValue(result.Statement))

	d.SetId(*layerArn + ":" + strconv.FormatInt(*layerVersion, 10))

	return resourceAwsLambdaLayerVersionPermissionGet(d, meta)
}

func resourceAwsLambdaLayerVersionPermissionGet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerName, version, err := resourceAwsLambdaLayerVersionParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing lambda layer ID: %s", err)
	}

	layerVersionPolicyOutput, err := conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(version),
	})

	log.Printf("[DEBUG] OUTPUT: %s", layerVersionPolicyOutput)
	// 	if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
	// 		log.Printf("[WARN] Lambda Layer Version (%s) not found, removing from state", d.Id())
	// 		d.SetId("")
	// 		return nil
	// 	}

	// 	if err != nil {
	// 		return fmt.Errorf("error reading Lambda Layer version (%s): %s", d.Id(), err)
	// 	}

	// 	if err := d.Set("layer_name", layerName); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer name: %s", err)
	// 	}
	// 	if err := d.Set("version", strconv.FormatInt(version, 10)); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer version: %s", err)
	// 	}
	// 	if err := d.Set("arn", layerVersion.LayerVersionArn); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer version arn: %s", err)
	// 	}
	// 	if err := d.Set("layer_arn", layerVersion.LayerArn); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer arn: %s", err)
	// 	}
	// 	if err := d.Set("description", layerVersion.Description); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer description: %s", err)
	// 	}
	// 	if err := d.Set("license_info", layerVersion.LicenseInfo); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer license info: %s", err)
	// 	}
	// 	if err := d.Set("created_date", layerVersion.CreatedDate); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer created date: %s", err)
	// 	}
	// 	if err := d.Set("source_code_hash", layerVersion.Content.CodeSha256); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer source code hash: %s", err)
	// 	}
	// 	if err := d.Set("source_code_size", layerVersion.Content.CodeSize); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer source code size: %s", err)
	// 	}
	// 	if err := d.Set("compatible_runtimes", flattenStringList(layerVersion.CompatibleRuntimes)); err != nil {
	// 		return fmt.Errorf("Error setting lambda layer compatible runtimes: %s", err)
	// 	}

	return nil
}

func resourceAwsLambdaLayerVersionPermissionRemove(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerName, version, err := resourceAwsLambdaLayerVersionParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing lambda layer ID: %s", err)
	}

	_, err = conn.RemoveLayerVersionPermission(&lambda.RemoveLayerVersionPermissionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(int64(version)),
		StatementId:   aws.String(d.Get("statement_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("error deleting Lambda Layer Version permission (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Lambda layer permission %q deleted", d.Get("statement_id").(string))
	return nil
}
