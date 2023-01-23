package opsworks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSetPermission,
		UpdateWithoutTimeout: resourceSetPermission,
		DeleteWithoutTimeout: resourcePermissionDelete,
		ReadWithoutTimeout:   resourcePermissionRead,

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

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn()

	req := &opsworks.DescribePermissionsInput{
		IamUserArn: aws.String(d.Get("user_arn").(string)),
		StackId:    aws.String(d.Get("stack_id").(string)),
	}

	log.Printf("[DEBUG] Reading OpsWorks prermissions for: %s on stack: %s", d.Get("user_arn"), d.Get("stack_id"))

	resp, err := client.DescribePermissionsWithContext(ctx, req)
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() == "ResourceNotFoundException" {
				log.Printf("[INFO] Permission not found")
				d.SetId("")
				return diags
			}
		}
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Permissions (%s): %s", d.Id(), err)
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

	return diags
}

func resourceSetPermission(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).OpsWorksConn()

	req := &opsworks.SetPermissionInput{
		AllowSudo:  aws.Bool(d.Get("allow_sudo").(bool)),
		AllowSsh:   aws.Bool(d.Get("allow_ssh").(bool)),
		IamUserArn: aws.String(d.Get("user_arn").(string)),
		StackId:    aws.String(d.Get("stack_id").(string)),
	}

	if d.HasChange("level") {
		req.Level = aws.String(d.Get("level").(string))
	}

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		_, err := client.SetPermissionWithContext(ctx, req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, opsworks.ErrCodeResourceNotFoundException, "Unable to find user with ARN") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = client.SetPermissionWithContext(ctx, req)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting OpsWorks Permissions (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}
