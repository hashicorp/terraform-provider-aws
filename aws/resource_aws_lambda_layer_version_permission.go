package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsLambdaLayerVersionPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaLayerVersionPermissionAdd,
		Read:   resourceAwsLambdaLayerVersionPermissionGet,
		Delete: resourceAwsLambdaLayerVersionPermissionRemove,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
			"policy": {
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

	d.SetId(*layerArn + ":" + strconv.FormatInt(*layerVersion, 10))
	d.Set("revision_id", result.RevisionId)

	return resourceAwsLambdaLayerVersionPermissionGet(d, meta)
}

func resourceAwsLambdaLayerVersionPermissionGet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerName, layerArn, version, err := resourceAwsLambdaLayerVersionPermissionParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parsing lambda layer ID: %s", err)
	}

	layerVersionPolicyOutput, err := conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(version),
	})

	if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Lambda Layer Version (%s) not found, removing it's permission from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lambda Layer version permission (%s): %s", d.Id(), err)
	}

	policyDoc := &IAMPolicyDoc{}

	if err := json.Unmarshal([]byte(aws.StringValue(layerVersionPolicyOutput.Policy)), policyDoc); err != nil {
		return err
	}

	principal := ""
	identifiers := policyDoc.Statements[0].Principals[0].Identifiers
	if reflect.TypeOf(identifiers).String() == "[]string" && identifiers.([]string)[0] == "*" {
		principal = "*"
	} else {
		policy_principal_arn, err := arn.Parse(policyDoc.Statements[0].Principals[0].Identifiers.(string))
		if err != nil {
			return fmt.Errorf("error reading principal arn from Lambda Layer version permission (%s): %s", d.Id(), err)
		}
		principal = policy_principal_arn.AccountID
	}

	if err := d.Set("layer_arn", layerArn); err != nil {
		return fmt.Errorf("Error setting lambda layer permission layer_arn: %s", err)
	}
	if err := d.Set("layer_version", version); err != nil {
		return fmt.Errorf("Error setting lambda layer permission layer_version: %s", err)
	}
	if err := d.Set("statement_id", policyDoc.Statements[0].Sid); err != nil {
		return fmt.Errorf("Error setting lambda layer permission statement_id: %s", err)
	}
	if err := d.Set("action", policyDoc.Statements[0].Actions); err != nil {
		return fmt.Errorf("Error setting lambda layer permission action: %s", err)
	}
	if err := d.Set("principal", principal); err != nil {
		return fmt.Errorf("Error setting lambda layer permission statement_id: %s", err)
	}
	if len(policyDoc.Statements[0].Conditions) > 0 {
		if err := d.Set("organization_id", policyDoc.Statements[0].Conditions[0].Values.([]string)[0]); err != nil {
			return fmt.Errorf("Error setting lambda layer permission organization_id: %s", err)
		}
	}
	if err := d.Set("policy", layerVersionPolicyOutput.Policy); err != nil {
		return fmt.Errorf("Error setting lambda layer permission policy: %s", err)
	}
	if err := d.Set("revision_id", layerVersionPolicyOutput.RevisionId); err != nil {
		return fmt.Errorf("Error setting lambda layer permission revision_id: %s", err)
	}

	return nil
}

func resourceAwsLambdaLayerVersionPermissionRemove(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	layerName, _, version, err := resourceAwsLambdaLayerVersionPermissionParseId(d.Id())
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

func resourceAwsLambdaLayerVersionPermissionParseId(id string) (layerName string, layerARN string, version int64, err error) {
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
	layerARN = strings.TrimSuffix(id, ":"+parts[2])
	version, err = strconv.ParseInt(parts[2], 10, 64)
	return
}
