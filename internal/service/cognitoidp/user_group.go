// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_user_group", name="User Group")
func resourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserGroupCreate,
		ReadWithoutTimeout:   resourceUserGroupRead,
		UpdateWithoutTimeout: resourceUserGroupUpdate,
		DeleteWithoutTimeout: resourceUserGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateGroup.html
		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserGroupName,
			},
			"precedence": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrUserPoolID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validUserPoolID,
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cognitoidentityprovider.CreateGroupInput{
		GroupName:  aws.String(name),
		UserPoolId: aws.String(d.Get(names.AttrUserPoolID).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("precedence"); ok {
		input.Precedence = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	output, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito User Group (%s): %s", name, err)
	}

	d.SetId(userGroupCreateResourceID(aws.ToString(output.Group.UserPoolId), aws.ToString(output.Group.GroupName)))

	return append(diags, resourceUserGroupRead(ctx, d, meta)...)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, groupName, err := userGroupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	group, err := findGroupByTwoPartKey(ctx, conn, userPoolID, groupName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito User Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrName, groupName)
	d.Set("precedence", group.Precedence)
	d.Set(names.AttrRoleARN, group.RoleArn)
	d.Set(names.AttrUserPoolID, userPoolID)

	return diags
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, groupName, err := userGroupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &cognitoidentityprovider.UpdateGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolID),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange("precedence") {
		input.Precedence = aws.Int32(int32(d.Get("precedence").(int)))
	}

	if d.HasChange(names.AttrRoleARN) {
		input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
	}

	_, err = conn.UpdateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito User Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUserGroupRead(ctx, d, meta)...)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, groupName, err := userGroupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Cognito User Group: %s", d.Id())
	_, err = conn.DeleteGroup(ctx, &cognitoidentityprovider.DeleteGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito User Group (%s): %s", d.Id(), err)
	}

	return diags
}

const userGroupResourceIDSeparator = "/"

func userGroupCreateResourceID(userPoolID, groupName string) string {
	parts := []string{userPoolID, groupName}
	id := strings.Join(parts, userGroupResourceIDSeparator)

	return id
}

func userGroupParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, userGroupResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected UserPoolID%[2]sGroupName", id, userGroupResourceIDSeparator)
}

func findGroupByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, groupName string) (*awstypes.GroupType, error) {
	input := &cognitoidentityprovider.GetGroupInput{
		GroupName:  aws.String(groupName),
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.GetGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Group == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Group, nil
}
