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
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ecr/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceLifecyclePolicyCreate,
		Read:   resourceLifecyclePolicyRead,
		Delete: resourceLifecyclePolicyDelete,

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
				ForceNew:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := &ecr.PutLifecyclePolicyInput{
		RepositoryName:      aws.String(d.Get("repository").(string)),
		LifecyclePolicyText: aws.String(d.Get("policy").(string)),
	}

	resp, err := conn.PutLifecyclePolicy(input)
	if err != nil {
		return err
	}
	d.SetId(aws.StringValue(resp.RepositoryName))
	d.Set("registry_id", resp.RegistryId)
	return resourceLifecyclePolicyRead(d, meta)
}

func resourceLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := &ecr.GetLifecyclePolicyInput{
		RepositoryName: aws.String(d.Id()),
	}

	var resp *ecr.GetLifecyclePolicyOutput

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.GetLifecyclePolicy(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeLifecyclePolicyNotFoundException) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.GetLifecyclePolicy(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeLifecyclePolicyNotFoundException) {
		log.Printf("[WARN] ECR Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
		log.Printf("[WARN] ECR Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECR Lifecycle Policy (%s): %w", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("error reading ECR Lifecycle Policy (%s): empty response", d.Id())
	}

	d.Set("repository", resp.RepositoryName)
	d.Set("registry_id", resp.RegistryId)
	d.Set("policy", resp.LifecyclePolicyText)

	return nil
}

func resourceLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := &ecr.DeleteLifecyclePolicyInput{
		RepositoryName: aws.String(d.Id()),
	}

	_, err := conn.DeleteLifecyclePolicy(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, ecr.ErrCodeRepositoryNotFoundException, "") {
			return nil
		}
		if tfawserr.ErrMessageContains(err, ecr.ErrCodeLifecyclePolicyNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}
