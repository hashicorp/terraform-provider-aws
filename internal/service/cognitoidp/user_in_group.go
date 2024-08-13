// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_user_in_group", name="Group User")
func resourceUserInGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserInGroupCreate,
		ReadWithoutTimeout:   resourceUserInGroupRead,
		DeleteWithoutTimeout: resourceUserInGroupDelete,
		Schema: map[string]*schema.Schema{
			names.AttrGroupName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			names.AttrUserPoolID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
			names.AttrUsername: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceUserInGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.AdminAddUserToGroupInput{}

	if v, ok := d.GetOk(names.AttrGroupName); ok {
		input.GroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrUserPoolID); ok {
		input.UserPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrUsername); ok {
		input.Username = aws.String(v.(string))
	}

	_, err := conn.AdminAddUserToGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito Group User: %s", err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.UniqueId())

	return append(diags, resourceUserInGroupRead(ctx, d, meta)...)
}

func resourceUserInGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	err := findGroupUserByThreePartKey(ctx, conn, d.Get(names.AttrGroupName).(string), d.Get(names.AttrUserPoolID).(string), d.Get(names.AttrUsername).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito Group User %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito Group User (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceUserInGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	log.Printf("[DEBUG] Deleting Cognito Group User: %s", d.Id())
	_, err := conn.AdminRemoveUserFromGroup(ctx, &cognitoidentityprovider.AdminRemoveUserFromGroupInput{
		GroupName:  aws.String(d.Get(names.AttrGroupName).(string)),
		Username:   aws.String(d.Get(names.AttrUsername).(string)),
		UserPoolId: aws.String(d.Get(names.AttrUserPoolID).(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Group User (%s): %s", d.Id(), err)
	}

	return diags
}

func findGroupUserByThreePartKey(ctx context.Context, conn *cognitoidentityprovider.Client, groupName, userPoolID, username string) error {
	input := &cognitoidentityprovider.AdminListGroupsForUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(userPoolID),
	}

	pages := cognitoidentityprovider.NewAdminListGroupsForUserPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.UserNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return err
		}

		for _, v := range page.Groups {
			if aws.ToString(v.GroupName) == groupName {
				return nil
			}
		}
	}

	return &retry.NotFoundError{}
}
