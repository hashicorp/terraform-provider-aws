package acmpca

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyPut,
		Read:   resourcePolicyRead,
		Update: resourcePolicyPut,
		Delete: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("resource_arn", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

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
		},
	}
}

func resourcePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	resourceArn := d.Get("resource_arn").(string)
	input := &acmpca.PutPolicyInput{
		ResourceArn: aws.String(resourceArn),
		Policy:      aws.String(d.Get("policy").(string)),
	}

	_, err := conn.PutPolicy(input)

	if err != nil {
		return fmt.Errorf("error putting policy on %s: %w", d.Get("resource_arn").(string), err)
	}

	d.SetId(resourceArn)

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	getPolicyInput := &acmpca.GetPolicyInput{
		ResourceArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACM PCA Policy: %s", getPolicyInput)

	policyOutput, err := conn.GetPolicy(getPolicyInput)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] ACM PCA Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ACM PCA Policy (%s): %w", d.Id(), err)
	}

	if policyOutput == nil {
		return fmt.Errorf("error reading ACM PCA Policy (%s): empty response", d.Id())
	}

	d.Set("policy", policyOutput.Policy)

	return nil
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	input := &acmpca.DeletePolicyInput{
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}
	_, err := conn.DeletePolicy(input)

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestAlreadyProcessedException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestInProgressException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidRequestException, "Self-signed policy can not be revoked") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting ACM PCA Policy (%s): %w", d.Id(), err)
	}

	return nil
}
