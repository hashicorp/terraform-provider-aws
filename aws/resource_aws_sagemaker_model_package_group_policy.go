package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfsagemaker "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsSagemakerModelPackageGroupPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerModelPackageGroupPolicyPut,
		Read:   resourceAwsSagemakerModelPackageGroupPolicyRead,
		Update: resourceAwsSagemakerModelPackageGroupPolicyPut,
		Delete: resourceAwsSagemakerModelPackageGroupPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"model_package_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsSagemakerModelPackageGroupPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.PutModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
		ResourcePolicy:        aws.String(d.Get("resource_policy").(string)),
	}

	_, err := conn.PutModelPackageGroupPolicy(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Model Package Group Policy %s: %w", name, err)
	}

	d.SetId(name)

	return resourceAwsSagemakerModelPackageGroupPolicyRead(d, meta)
}

func resourceAwsSagemakerModelPackageGroupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	mpg, err := finder.ModelPackageGroupPolicyByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker Model Package Group Policy (%s); removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Model Package Group Policy (%s): %w", d.Id(), err)
	}

	d.Set("model_package_group_name", d.Id())
	d.Set("resource_policy", mpg.ResourcePolicy)

	return nil
}

func resourceAwsSagemakerModelPackageGroupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroupPolicy(input); err != nil {
		if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Cannot find Model Package Group") ||
			tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Cannot find resource policy") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Model Package Group Policy (%s): %w", d.Id(), err)
	}

	return nil
}
