package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsEcrRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcrRegistryPolicyPut,
		Read:   resourceAwsEcrRegistryPolicyRead,
		Update: resourceAwsEcrRegistryPolicyPut,
		Delete: resourceAwsEcrRegistryPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEcrRegistryPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	input := ecr.PutRegistryPolicyInput{
		PolicyText: aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Creating ECR resository policy: %s", input)

	// Retry due to IAM eventual consistency
	var err error
	var out *ecr.PutRegistryPolicyOutput
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		out, err = conn.PutRegistryPolicy(&input)

		if tfawserr.ErrMessageContains(err, ecr.ErrCodeInvalidParameterException, "Invalid registry policy provided") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.PutRegistryPolicy(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating ECR Registry Policy: %w", err)
	}

	registryPolicy := *out
	regID := aws.StringValue(registryPolicy.RegistryId)

	log.Printf("[DEBUG] ECR registry policy created: %s", regID)
	d.SetId(regID)

	return resourceAwsEcrRegistryPolicyRead(d, meta)
}

func resourceAwsEcrRegistryPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

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

	log.Printf("[DEBUG] Received registry policy %s", out)

	registryPolicy := out
	d.Set("registry_id", registryPolicy.RegistryId)
	d.Set("policy", registryPolicy.PolicyText)

	return nil
}

func resourceAwsEcrRegistryPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	_, err := conn.DeleteRegistryPolicy(&ecr.DeleteRegistryPolicyInput{})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] registry policy %s deleted.", d.Id())

	return nil
}
