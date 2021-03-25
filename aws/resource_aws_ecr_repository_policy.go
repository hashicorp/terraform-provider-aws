package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsEcrRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcrRepositoryPolicyPut,
		Read:   resourceAwsEcrRepositoryPolicyRead,
		Update: resourceAwsEcrRepositoryPolicyPut,
		Delete: resourceAwsEcrRepositoryPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEcrRepositoryPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	input := ecr.SetRepositoryPolicyInput{
		RepositoryName: aws.String(d.Get("repository").(string)),
		PolicyText:     aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Creating ECR repository policy: %#v", input)

	// Retry due to IAM eventual consistency
	var err error
	var out *ecr.SetRepositoryPolicyOutput
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		out, err = conn.SetRepositoryPolicy(&input)

		if isAWSErr(err, ecr.ErrCodeInvalidParameterException, "Invalid repository policy provided") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.SetRepositoryPolicy(&input)
	}
	if err != nil {
		return fmt.Errorf("error creating ECR Repository Policy: %w", err)
	}

	log.Printf("[DEBUG] ECR repository policy created: %s", aws.StringValue(out.RepositoryName))

	d.SetId(aws.StringValue(out.RepositoryName))

	return resourceAwsEcrRepositoryPolicyRead(d, meta)
}

func resourceAwsEcrRepositoryPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	log.Printf("[DEBUG] Reading repository policy %s", d.Id())
	out, err := conn.GetRepositoryPolicy(&ecr.GetRepositoryPolicyInput{
		RepositoryName: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") ||
			isAWSErr(err, ecr.ErrCodeRepositoryPolicyNotFoundException, "") {
			log.Printf("[WARN] ECR Repository Policy %s not found, removing", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Received repository policy %#v", out)

	d.Set("repository", out.RepositoryName)
	d.Set("registry_id", out.RegistryId)
	d.Set("policy", out.PolicyText)

	return nil
}

func resourceAwsEcrRepositoryPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	_, err := conn.DeleteRepositoryPolicy(&ecr.DeleteRepositoryPolicyInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
	})
	if err != nil {
		if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") ||
			isAWSErr(err, ecr.ErrCodeRepositoryPolicyNotFoundException, "") {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] repository policy %s deleted.", d.Id())

	return nil
}
