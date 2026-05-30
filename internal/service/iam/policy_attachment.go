// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_policy_attachment", name="Policy Attachment")
func resourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		UpdateWithoutTimeout: resourcePolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"groups": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"groups", "roles", "users"},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"roles": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"groups", "roles", "users"},
			},
			"users": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				AtLeastOneOf: []string{"groups", "roles", "users"},
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyARN := d.Get("policy_arn").(string)
	var groups, roles, users []string
	if v, ok := d.GetOk("groups"); ok && v.(*schema.Set).Len() > 0 {
		groups = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("roles"); ok && v.(*schema.Set).Len() > 0 {
		roles = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("users"); ok && v.(*schema.Set).Len() > 0 {
		users = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	diags = sdkdiag.AppendFromErr(diags, attachPolicyToGroups(ctx, conn, groups, policyARN))
	diags = sdkdiag.AppendFromErr(diags, attachPolicyToRoles(ctx, conn, roles, policyARN))
	diags = sdkdiag.AppendFromErr(diags, attachPolicyToUsers(ctx, conn, users, policyARN))

	if diags.HasError() {
		return diags
	}

	d.SetId(d.Get(names.AttrName).(string))

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyARN := d.Get("policy_arn").(string)
	groups, roles, users, err := findEntitiesForPolicyByARN(ctx, conn, policyARN)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IAM Policy Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy Attachment (%s): %s", d.Id(), err)
	}

	d.Set("groups", groups)
	d.Set("roles", roles)
	d.Set("users", users)

	return diags
}

func resourcePolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if d.HasChange("groups") {
		diags = sdkdiag.AppendFromErr(diags, updateGroups(ctx, conn, d))
	}
	if d.HasChange("roles") {
		diags = sdkdiag.AppendFromErr(diags, updateRoles(ctx, conn, d))
	}
	if d.HasChange("users") {
		diags = sdkdiag.AppendFromErr(diags, updateUsers(ctx, conn, d))
	}

	if diags.HasError() {
		return diags
	}

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	policyARN := d.Get("policy_arn").(string)
	var groups, roles, users []string
	if v, ok := d.GetOk("groups"); ok && v.(*schema.Set).Len() > 0 {
		groups = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("roles"); ok && v.(*schema.Set).Len() > 0 {
		roles = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("users"); ok && v.(*schema.Set).Len() > 0 {
		users = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	diags = sdkdiag.AppendFromErr(diags, detachPolicyFromGroups(ctx, conn, groups, policyARN))
	diags = sdkdiag.AppendFromErr(diags, detachPolicyFromRoles(ctx, conn, roles, policyARN))
	diags = sdkdiag.AppendFromErr(diags, detachPolicyFromUsers(ctx, conn, users, policyARN))

	return diags
}

func attachPolicyToGroups(ctx context.Context, conn *iam.Client, groups []string, policyARN string) error {
	var errs []error

	for _, group := range groups {
		errs = append(errs, attachPolicyToGroup(ctx, conn, group, policyARN))
	}

	return errors.Join(errs...)
}

func attachPolicyToRoles(ctx context.Context, conn *iam.Client, roles []string, policyARN string) error {
	var errs []error

	for _, role := range roles {
		errs = append(errs, attachPolicyToRole(ctx, conn, role, policyARN))
	}

	return errors.Join(errs...)
}

func attachPolicyToUsers(ctx context.Context, conn *iam.Client, users []string, policyARN string) error {
	var errs []error

	for _, user := range users {
		errs = append(errs, attachPolicyToUser(ctx, conn, user, policyARN))
	}

	return errors.Join(errs...)
}

func updateGroups(ctx context.Context, conn *iam.Client, d *schema.ResourceData) error {
	policyARN := d.Get("policy_arn").(string)
	o, n := d.GetChange("groups")
	os, ns := o.(*schema.Set), n.(*schema.Set)
	add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

	if err := detachPolicyFromGroups(ctx, conn, del, policyARN); err != nil {
		return err
	}
	if err := attachPolicyToGroups(ctx, conn, add, policyARN); err != nil {
		return err
	}

	return nil
}

func updateRoles(ctx context.Context, conn *iam.Client, d *schema.ResourceData) error {
	policyARN := d.Get("policy_arn").(string)
	o, n := d.GetChange("roles")
	os, ns := o.(*schema.Set), n.(*schema.Set)
	add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

	if err := detachPolicyFromRoles(ctx, conn, del, policyARN); err != nil {
		return err
	}
	if err := attachPolicyToRoles(ctx, conn, add, policyARN); err != nil {
		return err
	}

	return nil
}

func updateUsers(ctx context.Context, conn *iam.Client, d *schema.ResourceData) error {
	policyARN := d.Get("policy_arn").(string)
	o, n := d.GetChange("users")
	os, ns := o.(*schema.Set), n.(*schema.Set)
	add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

	if err := detachPolicyFromUsers(ctx, conn, del, policyARN); err != nil {
		return err
	}
	if err := attachPolicyToUsers(ctx, conn, add, policyARN); err != nil {
		return err
	}

	return nil
}

func detachPolicyFromGroups(ctx context.Context, conn *iam.Client, groups []string, policyARN string) error {
	var errs []error

	for _, group := range groups {
		errs = append(errs, detachPolicyFromGroup(ctx, conn, group, policyARN))
	}

	return errors.Join(errs...)
}

func detachPolicyFromRoles(ctx context.Context, conn *iam.Client, roles []string, policyARN string) error {
	var errs []error

	for _, role := range roles {
		errs = append(errs, detachPolicyFromRole(ctx, conn, role, policyARN))
	}

	return errors.Join(errs...)
}

func detachPolicyFromUsers(ctx context.Context, conn *iam.Client, users []string, policyARN string) error {
	var errs []error

	for _, user := range users {
		errs = append(errs, detachPolicyFromUser(ctx, conn, user, policyARN))
	}

	return errors.Join(errs...)
}

func findEntitiesForPolicyByARN(ctx context.Context, conn *iam.Client, arn string) ([]string, []string, []string, error) {
	input := iam.ListEntitiesForPolicyInput{
		PolicyArn: aws.String(arn),
	}
	groups, roles, users, err := findEntitiesForPolicy(ctx, conn, &input)

	if err != nil {
		return nil, nil, nil, err
	}

	if len(groups) == 0 && len(roles) == 0 && len(users) == 0 {
		return nil, nil, nil, tfresource.NewEmptyResultError()
	}

	groupName := tfslices.ApplyToAll(groups, func(v awstypes.PolicyGroup) string { return aws.ToString(v.GroupName) })
	roleNames := tfslices.ApplyToAll(roles, func(v awstypes.PolicyRole) string { return aws.ToString(v.RoleName) })
	userNames := tfslices.ApplyToAll(users, func(v awstypes.PolicyUser) string { return aws.ToString(v.UserName) })

	return groupName, roleNames, userNames, nil
}

func findEntitiesForPolicy(ctx context.Context, conn *iam.Client, input *iam.ListEntitiesForPolicyInput) ([]awstypes.PolicyGroup, []awstypes.PolicyRole, []awstypes.PolicyUser, error) {
	var groups []awstypes.PolicyGroup
	var roles []awstypes.PolicyRole
	var users []awstypes.PolicyUser

	pages := iam.NewListEntitiesForPolicyPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, nil, nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, nil, nil, err
		}

		for _, v := range page.PolicyGroups {
			if p := &v; !inttypes.IsZero(p) {
				groups = append(groups, v)
			}
		}
		for _, v := range page.PolicyRoles {
			if p := &v; !inttypes.IsZero(p) {
				roles = append(roles, v)
			}
		}
		for _, v := range page.PolicyUsers {
			if p := &v; !inttypes.IsZero(p) {
				users = append(users, v)
			}
		}
	}

	return groups, roles, users, nil
}
