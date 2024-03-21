// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_cognito_user_in_group", name="Group User")
func resourceUserInGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserInGroupCreate,
		ReadWithoutTimeout:   resourceUserInGroupRead,
		DeleteWithoutTimeout: resourceUserInGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			"user_pool_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
			"username": {
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
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	groupName := d.Get("group_name").(string)
	userPoolId := d.Get("user_pool_id").(string)
	username := d.Get("username").(string)

	input := &cognitoidentityprovider.AdminAddUserToGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(username),
	}

	_, err := conn.AdminAddUserToGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding user to group: %s", err)
	}

	d.SetId(
		fmt.Sprintf("%s/%s/%s", userPoolId, groupName, username),
	)

	return append(diags, resourceUserInGroupRead(ctx, d, meta)...)
}

func resourceUserInGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	id := strings.Split(d.Id(), "/")
	userPoolId := id[0]
	groupName := id[1]
	username := id[2]

	found, err := FindCognitoUserInGroup(ctx, conn, groupName, userPoolId, username)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Group Membership: %s", err)
	}

	if !found {
		d.SetId("")
	}
	d.Set("group_name", groupName)
	d.Set("user_pool_id", userPoolId)
	d.Set("username", username)

	return diags
}

func resourceUserInGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn(ctx)

	groupName := d.Get("group_name").(string)
	userPoolID := d.Get("user_pool_id").(string)
	username := d.Get("username").(string)

	input := &cognitoidentityprovider.AdminRemoveUserFromGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
	}

	_, err := conn.AdminRemoveUserFromGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing user from group: %s", err)
	}

	return diags
}
