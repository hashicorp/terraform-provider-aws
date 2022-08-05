package sagemaker

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceModelPackageGroupPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceModelPackageGroupPolicyPut,
		Read:   resourceModelPackageGroupPolicyRead,
		Update: resourceModelPackageGroupPolicyPut,
		Delete: resourceModelPackageGroupPolicyDelete,
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
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceModelPackageGroupPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	policy, err := structure.NormalizeJsonString(d.Get("resource_policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("resource_policy").(string), err)
	}

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.PutModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
		ResourcePolicy:        aws.String(policy),
	}

	_, err = conn.PutModelPackageGroupPolicy(input)
	if err != nil {
		return fmt.Errorf("creating SageMaker Model Package Group Policy %s: %w", name, err)
	}

	d.SetId(name)

	return resourceModelPackageGroupPolicyRead(d, meta)
}

func resourceModelPackageGroupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	mpg, err := FindModelPackageGroupPolicyByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker Model Package Group Policy (%s); removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading SageMaker Model Package Group Policy (%s): %w", d.Id(), err)
	}

	d.Set("model_package_group_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get("resource_policy").(string), aws.StringValue(mpg.ResourcePolicy))

	if err != nil {
		return err
	}

	d.Set("resource_policy", policyToSet)

	return nil
}

func resourceModelPackageGroupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	input := &sagemaker.DeleteModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroupPolicy(input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find Model Package Group") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find resource policy") {
			return nil
		}
		return fmt.Errorf("deleting SageMaker Model Package Group Policy (%s): %w", d.Id(), err)
	}

	return nil
}
