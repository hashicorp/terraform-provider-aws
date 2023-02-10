package opsworks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
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
		ReadWithoutTimeout:   resourcePermissionRead,
		UpdateWithoutTimeout: resourceSetPermission,
		DeleteWithoutTimeout: schema.NoopContext,

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
				Required: true,
				ForceNew: true,
			},
			"user_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSetPermission(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn()

	iamUserARN := d.Get("user_arn").(string)
	stackID := d.Get("stack_id").(string)
	id := iamUserARN + stackID
	input := &opsworks.SetPermissionInput{
		AllowSudo:  aws.Bool(d.Get("allow_sudo").(bool)),
		AllowSsh:   aws.Bool(d.Get("allow_ssh").(bool)),
		IamUserArn: aws.String(iamUserARN),
		StackId:    aws.String(stackID),
	}

	if d.IsNewResource() {
		if v, ok := d.GetOk("level"); ok {
			input.Level = aws.String(v.(string))
		}
	} else if d.HasChange("level") {
		input.Level = aws.String(d.Get("level").(string))
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.SetPermissionWithContext(ctx, input)
	}, opsworks.ErrCodeResourceNotFoundException, "Unable to find user with ARN")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting OpsWorks Permission (%s): %s", id, err)
	}

	if d.IsNewResource() {
		d.SetId(id)
	}

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn()

	permission, err := FindPermissionByTwoPartKey(ctx, conn, d.Get("user_arn").(string), d.Get("stack_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks Permission %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Permission (%s): %s", d.Id(), err)
	}

	d.Set("allow_ssh", permission.AllowSsh)
	d.Set("allow_sudo", permission.AllowSudo)
	d.Set("level", permission.Level)
	d.Set("stack_id", permission.StackId)
	d.Set("user_arn", permission.IamUserArn)

	return diags
}

func FindPermissionByTwoPartKey(ctx context.Context, conn *opsworks.OpsWorks, iamUserARN, stackID string) (*opsworks.Permission, error) {
	input := &opsworks.DescribePermissionsInput{
		IamUserArn: aws.String(iamUserARN),
		StackId:    aws.String(stackID),
	}

	output, err := conn.DescribePermissionsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Permissions) == 0 || output.Permissions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Permissions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Permissions[0], nil
}
