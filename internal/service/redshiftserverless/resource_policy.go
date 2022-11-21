package redshiftserverless

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourcePolicyPut,
		Read:   resourceResourcePolicyRead,
		Update: resourceResourcePolicyPut,
		Delete: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceResourcePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	arn := d.Get("resource_arn").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	input := redshiftserverless.PutResourcePolicyInput{
		ResourceArn: aws.String(arn),
		Policy:      aws.String(policy),
	}

	out, err := conn.PutResourcePolicy(&input)

	if err != nil {
		return fmt.Errorf("error setting Redshift Serverless Resource Policy (%s): %w", arn, err)
	}

	d.SetId(aws.StringValue(out.ResourcePolicy.ResourceArn))

	return resourceResourcePolicyRead(d, meta)
}

func resourceResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	out, err := FindResourcePolicyByARN(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Serverless Resource Policy (%s): %w", d.Id(), err)
	}

	d.Set("resource_arn", out.ResourceArn)

	doc := resourcePolicyDoc{}
	log.Printf("policy is %s:", aws.StringValue(out.Policy))

	if err := json.Unmarshal([]byte(aws.StringValue(out.Policy)), &doc); err != nil {
		return fmt.Errorf("error unmarshaling policy: %w", err)
	}

	doc.Statement.Resources = nil

	policyDoc := tfiam.IAMPolicyDoc{}

	policyDoc.Id = doc.Id
	policyDoc.Version = doc.Version
	policyDoc.Statements = []*tfiam.IAMPolicyStatement{doc.Statement}

	formattedPolicy, err := json.Marshal(policyDoc)
	if err != nil {
		return fmt.Errorf("error marshling policy: %w", err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), string(formattedPolicy))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is an invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourceResourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	log.Printf("[DEBUG] Deleting Redshift Serverless Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicy(&redshiftserverless.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Redshift Serverless Resource Policy (%s): %w", d.Id(), err)
	}

	return nil
}

type resourcePolicyDoc struct {
	Version   string                    `json:",omitempty"`
	Id        string                    `json:",omitempty"`
	Statement *tfiam.IAMPolicyStatement `json:"Statement,omitempty"`
}
