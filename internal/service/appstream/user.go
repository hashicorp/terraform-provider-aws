// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_user", name="User")
func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthenticationType](),
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"first_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"last_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"send_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrUserName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType := d.Get(names.AttrUserName).(string), awstypes.AuthenticationType(d.Get("authentication_type").(string))
	id := userCreateResourceID(userName, authType)
	input := appstream.CreateUserInput{
		AuthenticationType: authType,
		UserName:           aws.String(userName),
	}

	if v, ok := d.GetOk("first_name"); ok {
		input.FirstName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("last_name"); ok {
		input.LastName = aws.String(v.(string))
	}

	if !d.Get("send_email_notification").(bool) {
		input.MessageAction = awstypes.MessageActionSuppress
	}

	_, err := conn.CreateUser(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream User (%s): %s", id, err)
	}

	d.SetId(id)

	const (
		timeout = 4 * time.Minute
	)
	_, err = tfresource.RetryWhenNotFound(ctx, timeout, func() (any, error) {
		return findUserByTwoPartKey(ctx, conn, userName, authType)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream User (%s) create: %s", id, err)
	}

	if !d.Get(names.AttrEnabled).(bool) {
		if err := disableUser(ctx, conn, userName, authType); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	user, err := findUserByTwoPartKey(ctx, conn, userName, authType)

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] AppStream User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream User (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, user.Arn)
	d.Set("authentication_type", user.AuthenticationType)
	d.Set(names.AttrCreatedTime, aws.ToTime(user.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrEnabled, user.Enabled)
	d.Set("first_name", user.FirstName)
	d.Set("last_name", user.LastName)
	d.Set(names.AttrUserName, user.UserName)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChange(names.AttrEnabled) {
		if d.Get(names.AttrEnabled).(bool) {
			if err := enableUser(ctx, conn, userName, authType); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := disableUser(ctx, conn, userName, authType); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	userName, authType, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting AppStream User: %s", d.Id())
	input := appstream.DeleteUserInput{
		AuthenticationType: authType,
		UserName:           aws.String(userName),
	}
	_, err = conn.DeleteUser(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream User (%s): %s", d.Id(), err)
	}

	return diags
}

const userResourceIDSeparator = "/"

func userCreateResourceID(userName string, authType awstypes.AuthenticationType) string {
	parts := []string{userName, string(authType)} // nosemgrep:ci.typed-enum-conversion
	id := strings.Join(parts, userResourceIDSeparator)

	return id
}

func userParseResourceID(id string) (string, awstypes.AuthenticationType, error) {
	parts := strings.SplitN(id, userResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected UserName%[2]sAuthenticationType", id, userResourceIDSeparator)
	}

	return parts[0], awstypes.AuthenticationType(parts[1]), nil
}

func enableUser(ctx context.Context, conn *appstream.Client, userName string, authType awstypes.AuthenticationType) error {
	input := appstream.EnableUserInput{
		AuthenticationType: authType,
		UserName:           aws.String(userName),
	}

	_, err := conn.EnableUser(ctx, &input)

	if err != nil {
		return fmt.Errorf("enabling AppStream User (%s/%s): %w", userName, authType, err)
	}

	return nil
}

func disableUser(ctx context.Context, conn *appstream.Client, userName string, authType awstypes.AuthenticationType) error {
	input := appstream.DisableUserInput{
		AuthenticationType: authType,
		UserName:           aws.String(userName),
	}

	_, err := conn.DisableUser(ctx, &input)

	if err != nil {
		return fmt.Errorf("disabling AppStream User (%s/%s): %w", userName, authType, err)
	}

	return nil
}

func findUserByTwoPartKey(ctx context.Context, conn *appstream.Client, userName string, authType awstypes.AuthenticationType) (*awstypes.User, error) {
	input := appstream.DescribeUsersInput{
		AuthenticationType: authType,
	}

	return findUser(ctx, conn, &input, func(v *awstypes.User) bool {
		return aws.ToString(v.UserName) == userName
	})
}

func findUser(ctx context.Context, conn *appstream.Client, input *appstream.DescribeUsersInput, filter tfslices.Predicate[*awstypes.User]) (*awstypes.User, error) {
	output, err := findUsers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findUsers(ctx context.Context, conn *appstream.Client, input *appstream.DescribeUsersInput, filter tfslices.Predicate[*awstypes.User]) ([]awstypes.User, error) {
	var output []awstypes.User

	err := describeUsersPages(ctx, conn, input, func(page *appstream.DescribeUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Users {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
