package ssoadmin

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePermissionSetInlinePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoAdminPermissionSetInlinePolicyPut,
		Read:   resourcePermissionSetInlinePolicyRead,
		Update: resourceAwsSsoAdminPermissionSetInlinePolicyPut,
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

func resourceAwsSsoAdminPermissionSetInlinePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)

	input := &ssoadmin.PutInlinePolicyToPermissionSetInput{
		InlinePolicy:     aws.String(d.Get("inline_policy").(string)),
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err := conn.PutInlinePolicyToPermissionSet(input)
	if err != nil {
		return fmt.Errorf("error putting Inline Policy for SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", permissionSetArn, instanceArn))

	// (Re)provision ALL accounts after making the above changes
	if err := provisionSsoAdminPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return resourcePermissionSetInlinePolicyRead(d, meta)
}

func resourcePermissionSetInlinePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	permissionSetArn, instanceArn, err := parseSsoAdminResourceID(d.Id())
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

	d.Set("inline_policy", output.InlinePolicy)
	d.Set("instance_arn", instanceArn)
	d.Set("permission_set_arn", permissionSetArn)

	return nil
}

func resourcePermissionSetInlinePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	permissionSetArn, instanceArn, err := parseSsoAdminResourceID(d.Id())
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
