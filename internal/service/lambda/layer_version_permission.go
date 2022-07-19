package lambda

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLayerVersionPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceLayerVersionPermissionCreate,
		Read:   resourceLayerVersionPermissionRead,
		Delete: resourceLayerVersionPermissionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"layer_name": {
				Type: schema.TypeString,
				ValidateFunc: validation.Any(
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), ""),
					verify.ValidARN,
				),
				Required: true,
				ForceNew: true,
			},
			"version_number": {
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

func resourceLayerVersionPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	layerName := d.Get("layer_name").(string)
	versionNumber := d.Get("version_number").(int)

	params := &lambda.AddLayerVersionPermissionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(int64(versionNumber)),
		Action:        aws.String(d.Get("action").(string)),
		Principal:     aws.String(d.Get("principal").(string)),
		StatementId:   aws.String(d.Get("statement_id").(string)),
	}

	if v, ok := d.GetOk("organization_id"); ok {
		params.OrganizationId = aws.String(v.(string))
	}

	_, err := conn.AddLayerVersionPermission(params)
	if err != nil {
		return fmt.Errorf("error adding Lambda Layer Version Permission (layer: %s, version: %d): %w", layerName, versionNumber, err)
	}

	d.SetId(fmt.Sprintf("%s,%d", layerName, versionNumber))

	return resourceLayerVersionPermissionRead(d, meta)
}

func resourceLayerVersionPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	layerName, versionNumber, err := ResourceLayerVersionPermissionParseId(d.Id())
	if err != nil {
		return err
	}

	input := &lambda.GetLayerVersionPolicyInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	layerVersionPolicyOutput, err := conn.GetLayerVersionPolicy(input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Lambda Layer Version Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lambda Layer Version Permission (%s): %w", d.Id(), err)
	}

	policyDoc := &IAMPolicyDoc{}

	if err := json.Unmarshal([]byte(aws.StringValue(layerVersionPolicyOutput.Policy)), policyDoc); err != nil {
		return err
	}

	d.Set("layer_name", layerName)
	d.Set("version_number", versionNumber)
	d.Set("policy", layerVersionPolicyOutput.Policy)
	d.Set("revision_id", layerVersionPolicyOutput.RevisionId)

	if policyDoc != nil && len(policyDoc.Statements) > 0 {
		d.Set("statement_id", policyDoc.Statements[0].Sid)

		if actions := policyDoc.Statements[0].Actions; actions != nil {
			var action string
			t := reflect.TypeOf(actions)
			if t.String() == "[]string" && len(actions.([]string)) > 0 {
				action = actions.([]string)[0]
			} else if t.String() == "string" {
				action = actions.(string)
			}

			d.Set("action", action)
		}

		if len(policyDoc.Statements[0].Conditions) > 0 && policyDoc.Statements[0].Conditions[0].Values != nil {
			var organizationId string
			values := policyDoc.Statements[0].Conditions[0].Values
			t := reflect.TypeOf(values)
			if t.String() == "[]string" && len(values.([]string)) > 0 {
				organizationId = values.([]string)[0]
			} else if t.String() == "string" {
				organizationId = values.(string)
			}

			d.Set("organization_id", organizationId)
		}

		if len(policyDoc.Statements[0].Principals) > 0 && policyDoc.Statements[0].Principals[0].Identifiers != nil {
			var principal string
			identifiers := policyDoc.Statements[0].Principals[0].Identifiers
			t := reflect.TypeOf(identifiers)
			if t.String() == "[]string" && len(identifiers.([]string)) > 0 && identifiers.([]string)[0] == "*" {
				principal = "*"
			} else if t.String() == "string" {
				policyPrincipalArn, err := arn.Parse(identifiers.(string))
				if err != nil {
					return fmt.Errorf("error reading Principal ARN from Lambda Layer Version Permission (%s): %w", d.Id(), err)
				}
				principal = policyPrincipalArn.AccountID
			}

			d.Set("principal", principal)
		}
	}

	return nil
}

func resourceLayerVersionPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	layerName, versionNumber, err := ResourceLayerVersionPermissionParseId(d.Id())
	if err != nil {
		return err
	}

	input := &lambda.RemoveLayerVersionPermissionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
		StatementId:   aws.String(d.Get("statement_id").(string)),
	}

	_, err = conn.RemoveLayerVersionPermission(input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Lambda Layer Version Permission (%s): %w", d.Id(), err)
	}

	return nil
}

func ResourceLayerVersionPermissionParseId(id string) (string, int64, error) {
	parts := strings.Split(id, ",")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("unexpected format of ID (%s), expected LAYER_NAME,VERSION_NUMBER or LAYER_ARN,VERSION_NUMBER", id)
	}

	layerName := parts[0]
	versionNum, err := strconv.ParseInt(parts[1], 10, 64)

	if err != nil {
		return "", 0, err
	}

	return layerName, versionNum, nil
}
