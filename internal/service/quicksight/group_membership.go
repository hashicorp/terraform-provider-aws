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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_group_membership", name="Group Membership")
func resourceGroupMembership() *schema.Resource {
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

func resourceGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	namespace := d.Get(names.AttrNamespace).(string)
	groupName := d.Get(names.AttrGroupName).(string)
	memberName := d.Get("member_name").(string)
	id := groupMembershipCreateResourceID(awsAccountID, namespace, groupName, memberName)
	input := &quicksight.CreateGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		MemberName:   aws.String(memberName),
		Namespace:    aws.String(namespace),
	}

	_, err := conn.CreateGroupMembership(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Group Membership (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceGroupMembershipRead(ctx, d, meta)...)
}

func resourceGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, groupName, memberName, err := groupMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	member, err := findGroupMembershipByFourPartKey(ctx, conn, awsAccountID, namespace, groupName, memberName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Group Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Group Membership (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set(names.AttrGroupName, groupName)
	d.Set("member_name", member.MemberName)
	d.Set(names.AttrNamespace, namespace)

	return diags
}

func resourceGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, namespace, groupName, memberName, err := groupMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteGroupMembership(ctx, &quicksight.DeleteGroupMembershipInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		MemberName:   aws.String(memberName),
		Namespace:    aws.String(namespace),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Group Membership (%s): %s", d.Id(), err)
	}

	return diags
}

const groupMembershipResourceIDSeparator = "/"

func groupMembershipCreateResourceID(awsAccountID, namespace, groupName, memberName string) string {
	parts := []string{awsAccountID, namespace, groupName, memberName}
	id := strings.Join(parts, groupMembershipResourceIDSeparator)

	return id
}

func groupMembershipParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, groupMembershipResourceIDSeparator, 4)

	if len(parts) < 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sNAMESPACE%[2]sGROUP_NAME%[2]sUSER_NAME", id, groupMembershipResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func findGroupMembershipByFourPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, namespace, groupName, memberName string) (*awstypes.GroupMember, error) {
	input := &quicksight.ListGroupMembershipsInput{
		AwsAccountId: aws.String(awsAccountID),
		GroupName:    aws.String(groupName),
		Namespace:    aws.String(namespace),
	}

	return findGroupMembership(ctx, conn, input, func(v *awstypes.GroupMember) bool {
		return aws.ToString(v.MemberName) == memberName
	})
}

func findGroupMembership(ctx context.Context, conn *quicksight.Client, input *quicksight.ListGroupMembershipsInput, filter tfslices.Predicate[*awstypes.GroupMember]) (*awstypes.GroupMember, error) {
	output, err := findGroupMemberships(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGroupMemberships(ctx context.Context, conn *quicksight.Client, input *quicksight.ListGroupMembershipsInput, filter tfslices.Predicate[*awstypes.GroupMember]) ([]awstypes.GroupMember, error) {
	var output []awstypes.GroupMember

	pages := quicksight.NewListGroupMembershipsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.GroupMemberList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
