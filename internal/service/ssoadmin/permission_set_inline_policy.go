package ssoadmin

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePermissionSetInlinePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionSetInlinePolicyPut,
		Read:   resourcePermissionSetInlinePolicyRead,
		Update: resourcePermissionSetInlinePolicyPut,
		Delete: resourcePermissionSetInlinePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"inline_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourcePermissionSetInlinePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)

	policy, err := structure.NormalizeJsonString(d.Get("inline_policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("inline_policy").(string), err)
	}

	input := &ssoadmin.PutInlinePolicyToPermissionSetInput{
		InlinePolicy:     aws.String(policy),
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err = conn.PutInlinePolicyToPermissionSet(input)
	if err != nil {
		return fmt.Errorf("error putting Inline Policy for SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", permissionSetArn, instanceArn))

	// (Re)provision ALL accounts after making the above changes
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return resourcePermissionSetInlinePolicyRead(d, meta)
}

func resourcePermissionSetInlinePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	permissionSetArn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID: %w", err)
	}

	input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	output, err := conn.GetInlinePolicyForPermissionSet(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Inline Policy for SSO Permission Set (%s) not found, removing from state", permissionSetArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Inline Policy for SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	if output == nil {
		return fmt.Errorf("error reading Inline Policy for SSO Permission Set (%s): empty output", permissionSetArn)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("inline_policy").(string), aws.StringValue(output.InlinePolicy))

	if err != nil {
		return err
	}

	d.Set("inline_policy", policyToSet)

	d.Set("instance_arn", instanceArn)
	d.Set("permission_set_arn", permissionSetArn)

	return nil
}

func resourcePermissionSetInlinePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	permissionSetArn, instanceArn, err := ParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID: %w", err)
	}

	input := &ssoadmin.DeleteInlinePolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err = conn.DeleteInlinePolicyFromPermissionSet(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error detaching Inline Policy from SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	return nil
}
