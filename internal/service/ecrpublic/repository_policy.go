package ecrpublic

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceRepositoryPolicyPut,
		Read:   resourceRepositoryPolicyRead,
		Update: resourceRepositoryPolicyPut,
		Delete: resourceRepositoryPolicyDelete,
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
			"repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	policyPutTimeout = 2 * time.Minute
)

func resourceRepositoryPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRPublicConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	repositoryName := d.Get("repository_name").(string)
	input := &ecrpublic.SetRepositoryPolicyInput{
		PolicyText:     aws.String(policy),
		RepositoryName: aws.String(repositoryName),
	}

	log.Printf("[DEBUG] Setting ECR Public Repository Policy: %s", input)
	outputRaw, err := tfresource.RetryWhen(policyPutTimeout,
		func() (interface{}, error) {
			return conn.SetRepositoryPolicy(input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, ecrpublic.ErrCodeInvalidParameterException, "Invalid repository policy provided") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("error setting ECR Public Repository (%s) Policy: %w", repositoryName, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.StringValue(outputRaw.(*ecrpublic.SetRepositoryPolicyOutput).RepositoryName))
	}

	return resourceRepositoryPolicyRead(d, meta)
}

func resourceRepositoryPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRPublicConn

	output, err := FindRepositoryPolicyByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Public Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECR Public Repository Policy (%s): %w", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(output.PolicyText))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is an invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)
	d.Set("registry_id", output.RegistryId)
	d.Set("repository_name", output.RepositoryName)

	return nil
}

func resourceRepositoryPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRPublicConn

	_, err := conn.DeleteRepositoryPolicy(&ecrpublic.DeleteRepositoryPolicyInput{
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		RepositoryName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ecrpublic.ErrCodeRepositoryNotFoundException, ecrpublic.ErrCodeRepositoryPolicyNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ECR Public Repository Policy (%s): %w", d.Id(), err)
	}

	return nil
}

func FindRepositoryPolicyByName(conn *ecrpublic.ECRPublic, name string) (*ecrpublic.GetRepositoryPolicyOutput, error) {
	input := &ecrpublic.GetRepositoryPolicyInput{
		RepositoryName: aws.String(name),
	}

	output, err := conn.GetRepositoryPolicy(input)

	if tfawserr.ErrCodeEquals(err, ecrpublic.ErrCodeRepositoryNotFoundException, ecrpublic.ErrCodeRepositoryPolicyNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
