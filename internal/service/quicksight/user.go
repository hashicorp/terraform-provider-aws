// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultUserNamespace = "default"
)

// @SDKResource("aws_quicksight_user", name="User")
func resourceUser() *schema.Resource {
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
					// TODO ValidateDiagFunc: enum.Validate[awstypes.IdentityType](),
					ValidateFunc: validation.StringInSlice(enum.Slice(
						awstypes.IdentityTypeIam,
						awstypes.IdentityTypeQuicksight,
					), false),
				},
				names.AttrNamespace: {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  defaultUserNamespace,
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
				"user_invitation_url": {
					Type:     schema.TypeString,
					Computed: true,
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
					// TODO ValidateDiagFunc: enum.Validate[awstypes.UserRole](),
					ValidateFunc: validation.StringInSlice(enum.Slice(
						awstypes.UserRoleReader,
						awstypes.UserRoleAuthor,
						awstypes.UserRoleAdmin,
						awstypes.UserRoleReaderPro,
						awstypes.UserRoleAuthorPro,
						awstypes.UserRoleAdminPro,
					), false),
				},
			}
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	email := d.Get(names.AttrEmail).(string)
	namespace := d.Get(names.AttrNamespace).(string)
	input := &quicksight.RegisterUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Email:        aws.String(email),
		IdentityType: awstypes.IdentityType(d.Get("identity_type").(string)),
		Namespace:    aws.String(namespace),
		UserRole:     awstypes.UserRole(d.Get("user_role").(string)),
	}

	if v, ok := d.GetOk("iam_arn"); ok {
		input.IamArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("session_name"); ok {
		input.SessionName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrUserName); ok {
		input.UserName = aws.String(v.(string))
	}

	output, err := conn.RegisterUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering QuickSight User (%s): %s", email, err)
	}

	d.SetId(userCreateResourceID(awsAccountID, namespace, aws.ToString(output.User.UserName)))

	if awstypes.IdentityType(d.Get("identity_type").(string)) == awstypes.IdentityTypeQuicksight {
		userInvitationUrl := aws.ToString(output.UserInvitationUrl)
		d.Set("user_invitation_url", userInvitationUrl)
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, userName, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	user, err := findUserByThreePartKey(ctx, conn, awsAccountID, namespace, userName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, user.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrEmail, user.Email)
	d.Set(names.AttrNamespace, namespace)
	d.Set("user_role", user.Role)
	d.Set(names.AttrUserName, user.UserName)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, userName, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &quicksight.UpdateUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Email:        aws.String(d.Get(names.AttrEmail).(string)),
		Namespace:    aws.String(namespace),
		Role:         awstypes.UserRole(d.Get("user_role").(string)),
		UserName:     aws.String(userName),
	}

	_, err = conn.UpdateUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating QuickSight User (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, userName, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteUser(ctx, &quicksight.DeleteUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight User (%s): %s", d.Id(), err)
	}

	return diags
}

const userResourceIDSeparator = "/"

func userCreateResourceID(awsAccountID, namespace, userName string) string {
	parts := []string{awsAccountID, namespace, userName}
	id := strings.Join(parts, userResourceIDSeparator)

	return id
}

func userParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, userResourceIDSeparator, 3)

	if len(parts) < 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sNAMESPACE%[2]sUSER_NAME", id, userResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findUserByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace, userName string) (*awstypes.User, error) {
	input := &quicksight.DescribeUserInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		UserName:     aws.String(userName),
	}

	return findUser(ctx, conn, input)
}

func findUser(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeUserInput) (*awstypes.User, error) {
	output, err := conn.DescribeUser(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.User == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.User, nil
}
