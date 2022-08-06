package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegistryPolicyPut,
		Read:   resourceRegistryPolicyRead,
		Update: resourceRegistryPolicyPut,
		Delete: resourceRegistryPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRegistryPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	input := ecr.PutRegistryPolicyInput{
		PolicyText: aws.String(policy),
	}

	out, err := conn.PutRegistryPolicy(&input)
	if err != nil {
		return fmt.Errorf("Error creating ECR Registry Policy: %w", err)
	}

	regID := aws.StringValue(out.RegistryId)

	d.SetId(regID)

	return resourceRegistryPolicyRead(d, meta)
}

func resourceRegistryPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	log.Printf("[DEBUG] Reading registry policy %s", d.Id())
	out, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			log.Printf("[WARN] ECR Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("registry_id", out.RegistryId)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(out.PolicyText))

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

func resourceRegistryPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	_, err := conn.DeleteRegistryPolicy(&ecr.DeleteRegistryPolicyInput{})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}
