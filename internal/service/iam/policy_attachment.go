// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"name": {
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

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

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

	d.SetId(d.Get("name").(string))

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	policyARN := d.Get("policy_arn").(string)
	groups, roles, users, err := findEntitiesForPolicyByARN(ctx, conn, policyARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourcePolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

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

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

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

func attachPolicyToGroups(ctx context.Context, conn *iam.IAM, groups []string, policyARN string) error {
	var errs []error

	for _, group := range groups {
		errs = append(errs, attachPolicyToGroup(ctx, conn, group, policyARN))
	}

	return errors.Join(errs...)
}

func attachPolicyToRoles(ctx context.Context, conn *iam.IAM, roles []string, policyARN string) error {
	var errs []error

	for _, role := range roles {
		errs = append(errs, attachPolicyToRole(ctx, conn, role, policyARN))
	}

	return errors.Join(errs...)
}

func attachPolicyToUsers(ctx context.Context, conn *iam.IAM, users []string, policyARN string) error {
	var errs []error

	for _, user := range users {
		errs = append(errs, attachPolicyToUser(ctx, conn, user, policyARN))
	}

	return errors.Join(errs...)
}

func updateGroups(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
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

func updateRoles(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
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

func updateUsers(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
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

func detachPolicyFromGroups(ctx context.Context, conn *iam.IAM, groups []string, policyARN string) error {
	var errs []error

	for _, group := range groups {
		errs = append(errs, detachPolicyFromGroup(ctx, conn, group, policyARN))
	}

	return errors.Join(errs...)
}

func detachPolicyFromRoles(ctx context.Context, conn *iam.IAM, roles []string, policyARN string) error {
	var errs []error

	for _, role := range roles {
		errs = append(errs, detachPolicyFromRole(ctx, conn, role, policyARN))
	}

	return errors.Join(errs...)
}

func detachPolicyFromUsers(ctx context.Context, conn *iam.IAM, users []string, policyARN string) error {
	var errs []error

	for _, user := range users {
		errs = append(errs, detachPolicyFromUser(ctx, conn, user, policyARN))
	}

	return errors.Join(errs...)
}

func findEntitiesForPolicyByARN(ctx context.Context, conn *iam.IAM, arn string) ([]string, []string, []string, error) {
	input := &iam.ListEntitiesForPolicyInput{
		PolicyArn: aws.String(arn),
	}
	groups, roles, users, err := findEntitiesForPolicy(ctx, conn, input)

	if err != nil {
		return nil, nil, nil, err
	}

	if len(groups) == 0 && len(roles) == 0 && len(users) == 0 {
		return nil, nil, nil, tfresource.NewEmptyResultError(input)
	}

	groupName := tfslices.ApplyToAll(groups, func(v *iam.PolicyGroup) string { return aws.StringValue(v.GroupName) })
	roleNames := tfslices.ApplyToAll(roles, func(v *iam.PolicyRole) string { return aws.StringValue(v.RoleName) })
	userNames := tfslices.ApplyToAll(users, func(v *iam.PolicyUser) string { return aws.StringValue(v.UserName) })

	return groupName, roleNames, userNames, nil
}

func findEntitiesForPolicy(ctx context.Context, conn *iam.IAM, input *iam.ListEntitiesForPolicyInput) ([]*iam.PolicyGroup, []*iam.PolicyRole, []*iam.PolicyUser, error) {
	var groups []*iam.PolicyGroup
	var roles []*iam.PolicyRole
	var users []*iam.PolicyUser

	err := conn.ListEntitiesForPolicyPagesWithContext(ctx, input, func(page *iam.ListEntitiesForPolicyOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PolicyGroups {
			if v != nil {
				groups = append(groups, v)
			}
		}
		for _, v := range page.PolicyRoles {
			if v != nil {
				roles = append(roles, v)
			}
		}
		for _, v := range page.PolicyUsers {
			if v != nil {
				users = append(users, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil, nil, nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, nil, err
	}

	return groups, roles, users, nil
}
