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

// @SDKResource("aws_quicksight_group_membership", name="Group Membership")
func ResourceGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupMembershipCreate,
		ReadWithoutTimeout:   resourceGroupMembershipRead,
		DeleteWithoutTimeout: resourceGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
				names.AttrGroupName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"member_name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrNamespace: {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
					Default:  "default",
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 63),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "must contain only alphanumeric characters, hyphens, underscores, and periods"),
					),
				},
			}
		},
	}
}

func resourceGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
	namespace := d.Get(names.AttrNamespace).(string)
	groupName := d.Get(names.AttrGroupName).(string)
	memberName := d.Get("member_name").(string)

	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}

	createOpts := &quicksight.CreateGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		MemberName:   aws.String(memberName),
		Namespace:    aws.String(namespace),
	}

	resp, err := conn.CreateGroupMembershipWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding QuickSight user (%s) to group (%s): %s", memberName, groupName, err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s/%s", awsAccountID, namespace, groupName, aws.StringValue(resp.GroupMember.MemberName)))

	return append(diags, resourceGroupMembershipRead(ctx, d, meta)...)
}

func resourceGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID, namespace, groupName, userName, err := GroupMembershipParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "%s", err)
	}

	listInput := &quicksight.ListGroupMembershipsInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		GroupName:    aws.String(groupName),
	}

	found, err := FindGroupMembership(ctx, conn, listInput, userName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing QuickSight Group Memberships (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && !found {
		log.Printf("[WARN] QuickSight User-group membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrNamespace, namespace)
	d.Set("member_name", userName)
	d.Set(names.AttrGroupName, groupName)

	return diags
}

func resourceGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountID, namespace, groupName, userName, err := GroupMembershipParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "%s", err)
	}

	deleteOpts := &quicksight.DeleteGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		Namespace:    aws.String(namespace),
		MemberName:   aws.String(userName),
		GroupName:    aws.String(groupName),
	}

	if _, err := conn.DeleteGroupMembershipWithContext(ctx, deleteOpts); err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight User-group membership %s: %s", d.Id(), err)
	}

	return diags
}

func GroupMembershipParseID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, "/", 4)
	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/NAMESPACE/GROUP_NAME/USER_NAME", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}
