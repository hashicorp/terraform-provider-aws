// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iam_user_group_membership", name="User Group Membership")
func resourceUserGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserGroupMembershipCreate,
		ReadWithoutTimeout:   resourceUserGroupMembershipRead,
		UpdateWithoutTimeout: resourceUserGroupMembershipUpdate,
		DeleteWithoutTimeout: resourceUserGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserGroupMembershipImport,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"groups": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceUserGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	user := d.Get("user").(string)
	groupList := flex.ExpandStringValueSet(d.Get("groups").(*schema.Set))

	if err := addUserToGroups(ctx, conn, user, groupList); err != nil {
		return sdkdiag.AppendErrorf(diags, "assigning IAM User Group Membership (%s): %s", user, err)
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.UniqueId())

	return append(diags, resourceUserGroupMembershipRead(ctx, d, meta)...)
}

func resourceUserGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	user := d.Get("user").(string)
	groups := d.Get("groups").(*schema.Set)

	input := &iam.ListGroupsForUserInput{
		UserName: aws.String(user),
	}

	var gl []string

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		err := listGroupsForUserPages(ctx, conn, input, func(page *iam.ListGroupsForUserOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, group := range page.Groups {
				if groups.Contains(aws.ToString(group.GroupName)) {
					gl = append(gl, aws.ToString(group.GroupName))
				}
			}

			return !lastPage
		})

		if d.IsNewResource() && errs.IsA[*awstypes.NoSuchEntityException](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		err = listGroupsForUserPages(ctx, conn, input, func(page *iam.ListGroupsForUserOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, group := range page.Groups {
				if groups.Contains(aws.ToString(group.GroupName)) {
					gl = append(gl, aws.ToString(group.GroupName))
				}
			}

			return !lastPage
		})
	}

	var nse *awstypes.NoSuchEntityException
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, nse.ErrorCode()) {
		log.Printf("[WARN] IAM User Group Membership (%s) not found, removing from state", user)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM User Group Membership (%s): %s", user, err)
	}

	if err := d.Set("groups", gl); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting group list from IAM (%s), error: %s", user, err)
	}

	return diags
}

func resourceUserGroupMembershipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChange("groups") {
		user := d.Get("user").(string)

		o, n := d.GetChange("groups")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := flex.ExpandStringValueSet(os.Difference(ns))
		add := flex.ExpandStringValueSet(ns.Difference(os))

		if err := removeUserFromGroups(ctx, conn, user, remove); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User Group Membership (%s): %s", user, err)
		}

		if err := addUserToGroups(ctx, conn, user, add); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IAM User Group Membership (%s): %s", user, err)
		}
	}

	return append(diags, resourceUserGroupMembershipRead(ctx, d, meta)...)
}

func resourceUserGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)
	user := d.Get("user").(string)
	groups := flex.ExpandStringValueSet(d.Get("groups").(*schema.Set))

	if err := removeUserFromGroups(ctx, conn, user, groups); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM User Group Membership (%s): %s", user, err)
	}
	return diags
}

func removeUserFromGroups(ctx context.Context, conn *iam.Client, user string, groups []string) error {
	for _, group := range groups {
		if err := removeUserFromGroup(ctx, conn, user, group); err != nil {
			return err
		}
	}
	return nil
}

func addUserToGroups(ctx context.Context, conn *iam.Client, user string, groups []string) error {
	for _, group := range groups {
		if err := addUserToGroup(ctx, conn, user, group); err != nil {
			return err
		}
	}
	return nil
}

func addUserToGroup(ctx context.Context, conn *iam.Client, user, group string) error {
	_, err := conn.AddUserToGroup(ctx, &iam.AddUserToGroupInput{
		UserName:  aws.String(user),
		GroupName: aws.String(group),
	})
	if err != nil {
		return fmt.Errorf("adding User (%s) to Group (%s): %w", user, group, err)
	}
	return nil
}

func removeUserFromGroup(ctx context.Context, conn *iam.Client, user, group string) error {
	_, err := conn.RemoveUserFromGroup(ctx, &iam.RemoveUserFromGroupInput{
		UserName:  aws.String(user),
		GroupName: aws.String(group),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing IAM User (%s) from group (%s): %w", user, group, err)
	}

	return nil
}

func resourceUserGroupMembershipImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <user-name>/<group-name1>/...", d.Id())
	}

	userName := idParts[0]
	groupList := idParts[1:]

	d.Set("user", userName)
	d.Set("groups", groupList)

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(id.UniqueId())

	return []*schema.ResourceData{d}, nil
}
