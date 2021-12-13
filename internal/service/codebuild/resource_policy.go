package codebuild

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				ValidateFunc:     verify.ValidIAMPolicyJSON,
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
	conn := meta.(*conns.AWSClient).CodeBuildConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("policy").(string), err)
	}

	input := &codebuild.PutResourcePolicyInput{
		Policy:      aws.String(policy),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	resp, err := conn.PutResourcePolicy(input)
	if err != nil {
		return fmt.Errorf("error creating CodeBuild Resource Policy: %w", err)
	}

	d.SetId(aws.StringValue(resp.ResourceArn))

	return resourceResourcePolicyRead(d, meta)
}

func resourceResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	output, err := FindResourcePolicyByARN(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error Listing CodeBuild Resource Policys: %w", err)
	}

	if output == nil {
		log.Printf("[WARN] CodeBuild Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(output.Policy))

	if err != nil {
		return err
	}

	d.Set("resource_arn", d.Id())
	d.Set("policy", policyToSet)

	return nil
}

func resourceResourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	deleteOpts := &codebuild.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteResourcePolicy(deleteOpts); err != nil {
		return fmt.Errorf("error deleting CodeBuild Resource Policy (%s): %w", d.Id(), err)
	}

	return nil
}
