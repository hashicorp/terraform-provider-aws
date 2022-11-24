package opsworks

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceSetPermission,
		Update: resourceSetPermission,
		Delete: resourcePermissionDelete,
		Read:   resourcePermissionRead,

		Schema: map[string]*schema.Schema{
			"allow_ssh": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"allow_sudo": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"user_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"level": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"deny",
					"show",
					"deploy",
					"manage",
					"iam_only",
				}, false),
			},
			"stack_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	req := &opsworks.DescribePermissionsInput{
		IamUserArn: aws.String(d.Get("user_arn").(string)),
		StackId:    aws.String(d.Get("stack_id").(string)),
	}

	log.Printf("[DEBUG] Reading OpsWorks permissions for: %s on stack: %s", d.Get("user_arn"), d.Get("stack_id"))

	resp, err := client.DescribePermissions(req)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
			log.Printf("[INFO] OpsWorks Permissions (%s, %s) not found, removing from state", d.Get("user_arn"), d.Get("stack_id"))
			d.SetId("")
			return nil
		}
		return err
	}

	found := false
	id := ""
	for _, permission := range resp.Permissions {
		id = *permission.IamUserArn + *permission.StackId

		if d.Get("user_arn").(string)+d.Get("stack_id").(string) == id {
			found = true
			d.SetId(id)
			d.Set("allow_ssh", permission.AllowSsh)
			d.Set("allow_sudo", permission.AllowSudo)
			d.Set("user_arn", permission.IamUserArn)
			d.Set("stack_id", permission.StackId)
			d.Set("level", permission.Level)
		}
	}

	if !found {
		d.SetId("")
		log.Printf("[INFO] The correct permission could not be found for: %s on stack: %s", d.Get("user_arn"), d.Get("stack_id"))
	}

	return nil
}

func resourceSetPermission(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*conns.AWSClient).OpsWorksConn

	req := &opsworks.SetPermissionInput{
		AllowSudo:  aws.Bool(d.Get("allow_sudo").(bool)),
		AllowSsh:   aws.Bool(d.Get("allow_ssh").(bool)),
		IamUserArn: aws.String(d.Get("user_arn").(string)),
		StackId:    aws.String(d.Get("stack_id").(string)),
	}

	if d.HasChange("level") {
		req.Level = aws.String(d.Get("level").(string))
	}

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := client.SetPermission(req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, opsworks.ErrCodeResourceNotFoundException, "Unable to find user with ARN") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = client.SetPermission(req)
	}

	if err != nil {
		return err
	}

	return resourcePermissionRead(d, meta)
}
