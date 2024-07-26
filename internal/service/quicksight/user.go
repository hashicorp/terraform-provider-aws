// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	DefaultUserNamespace = "default"
)

// @SDKResource("aws_quicksight_user", name="User")
func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},

				names.AttrAWSAccountID: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
					ForceNew: true,
				},

				names.AttrEmail: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},

				"iam_arn": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},

				"identity_type": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.StringInSlice([]string{
						quicksight.IdentityTypeIam,
						quicksight.IdentityTypeQuicksight,
					}, false),
				},

				names.AttrNamespace: {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  DefaultUserNamespace,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 63),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
					),
				},

				"session_name": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},

				names.AttrUserName: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.NoZeroValues,
				},

				"user_role": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
					ValidateFunc: validation.StringInSlice([]string{
						quicksight.UserRoleReader,
						quicksight.UserRoleAuthor,
						quicksight.UserRoleAdmin,
					}, false),
				},
			}
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID

	namespace := d.Get(names.AttrNamespace).(string)

	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.RegisterUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Email:        aws.String(d.Get(names.AttrEmail).(string)),
		IdentityType: aws.String(d.Get("identity_type").(string)),
		Namespace:    aws.String(namespace),
		UserRole:     aws.String(d.Get("user_role").(string)),
	}

	if v, ok := d.GetOk("iam_arn"); ok {
		createOpts.IamArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_name"); ok {
		createOpts.SessionName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrUserName); ok {
		createOpts.UserName = aws.String(v.(string))
	}

	resp, err := conn.RegisterUserWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering QuickSight user: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", awsAccountID, namespace, aws.StringValue(resp.User.UserName)))

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID, namespace, userName, err := UserParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight User (%s): %s", d.Id(), err)
	}

	descOpts := &quicksight.DescribeUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	}

	resp, err := conn.DescribeUserWithContext(ctx, descOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] QuickSight User %s is not found", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.User.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrEmail, resp.User.Email)
	d.Set(names.AttrNamespace, namespace)
	d.Set("user_role", resp.User.Role)
	d.Set(names.AttrUserName, resp.User.UserName)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID, namespace, userName, err := UserParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating QuickSight User (%s): %s", d.Id(), err)
	}

	updateOpts := &quicksight.UpdateUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Email:        aws.String(d.Get(names.AttrEmail).(string)),
		Namespace:    aws.String(namespace),
		Role:         aws.String(d.Get("user_role").(string)),
		UserName:     aws.String(userName),
	}

	_, err = conn.UpdateUserWithContext(ctx, updateOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating QuickSight User (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID, namespace, userName, err := UserParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight User (%s): %s", d.Id(), err)
	}

	deleteOpts := &quicksight.DeleteUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	}

	if _, err := conn.DeleteUserWithContext(ctx, deleteOpts); err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight User (%s): %s", d.Id(), err)
	}

	return diags
}

func UserParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/USER_NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
