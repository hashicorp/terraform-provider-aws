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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyPut,
		Read:   resourcePolicyRead,
		Update: resourcePolicyPut,
		Delete: resourcePolicyDelete,

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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("policy").(string), err)
	}

	resourceARN := d.Get("resource_arn").(string)
	input := &acmpca.PutPolicyInput{
		Policy:      aws.String(policy),
		ResourceArn: aws.String(resourceARN),
	}

	log.Printf("[DEBUG] Putting ACM PCA Policy: %s", input)
	_, err = conn.PutPolicy(input)

	if err != nil {
		return fmt.Errorf("putting ACM PCA Policy (%s): %w", resourceARN, err)
	}

	d.SetId(resourceARN)

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	policy, err := FindPolicyByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading ACM PCA Policy (%s): %w", d.Id(), err)
	}

	d.Set("policy", policy)
	d.Set("resource_arn", d.Id())

	return nil
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMPCAConn

	log.Printf("[DEBUG] Deleting ACM PCA Policy: %s", d.Id())
	_, err := conn.DeletePolicy(&acmpca.DeletePolicyInput{
		ResourceArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestAlreadyProcessedException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestInProgressException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidRequestException, "Self-signed policy can not be revoked") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting ACM PCA Policy (%s): %w", d.Id(), err)
	}

	return nil
}
