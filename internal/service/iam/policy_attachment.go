// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_iam_policy_attachment")
func ResourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		UpdateWithoutTimeout: resourcePolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"users": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"policy_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	name := d.Get("name").(string)
	arn := d.Get("policy_arn").(string)
	users := flex.ExpandStringSet(d.Get("users").(*schema.Set))
	roles := flex.ExpandStringSet(d.Get("roles").(*schema.Set))
	groups := flex.ExpandStringSet(d.Get("groups").(*schema.Set))

	if len(users) == 0 && len(roles) == 0 && len(groups) == 0 {
		return sdkdiag.AppendErrorf(diags, "No Users, Roles, or Groups specified for IAM Policy Attachment %s", name)
	} else {
		var userErr, roleErr, groupErr error
		if users != nil {
			userErr = attachPolicyToUsers(ctx, conn, users, arn)
		}
		if roles != nil {
			roleErr = attachPolicyToRoles(ctx, conn, roles, arn)
		}
		if groups != nil {
			groupErr = attachPolicyToGroups(ctx, conn, groups, arn)
		}
		if userErr != nil || roleErr != nil || groupErr != nil {
			return composeErrors(fmt.Sprint("attaching policy with IAM Policy Attachment ", name, ":"), userErr, roleErr, groupErr)
		}
	}
	d.SetId(d.Get("name").(string))
	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	arn := d.Get("policy_arn").(string)
	name := d.Get("name").(string)

	_, err := conn.GetPolicyWithContext(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(arn),
	})

	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			log.Printf("[WARN] IAM Policy Attachment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy Attachment (%s): %s", d.Id(), err)
	}

	ul := make([]string, 0)
	rl := make([]string, 0)
	gl := make([]string, 0)

	args := iam.ListEntitiesForPolicyInput{
		PolicyArn: aws.String(arn),
	}
	err = conn.ListEntitiesForPolicyPagesWithContext(ctx, &args, func(page *iam.ListEntitiesForPolicyOutput, lastPage bool) bool {
		for _, u := range page.PolicyUsers {
			ul = append(ul, *u.UserName)
		}

		for _, r := range page.PolicyRoles {
			rl = append(rl, *r.RoleName)
		}

		for _, g := range page.PolicyGroups {
			gl = append(gl, *g.GroupName)
		}
		return true
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Policy Attachment (%s): %s", d.Id(), err)
	}

	userErr := d.Set("users", ul)
	roleErr := d.Set("roles", rl)
	groupErr := d.Set("groups", gl)

	if userErr != nil || roleErr != nil || groupErr != nil {
		return composeErrors(fmt.Sprint("setting user, role, or group list from IAM Policy Attachment ", name, ":"), userErr, roleErr, groupErr)
	}

	return diags
}

func resourcePolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	name := d.Get("name").(string)
	var userErr, roleErr, groupErr error

	if d.HasChange("users") {
		userErr = updateUsers(ctx, conn, d)
	}
	if d.HasChange("roles") {
		roleErr = updateRoles(ctx, conn, d)
	}
	if d.HasChange("groups") {
		groupErr = updateGroups(ctx, conn, d)
	}
	if userErr != nil || roleErr != nil || groupErr != nil {
		return composeErrors(fmt.Sprint("updating user, role, or group list from IAM Policy Attachment ", name, ":"), userErr, roleErr, groupErr)
	}
	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)
	name := d.Get("name").(string)
	arn := d.Get("policy_arn").(string)
	users := flex.ExpandStringSet(d.Get("users").(*schema.Set))
	roles := flex.ExpandStringSet(d.Get("roles").(*schema.Set))
	groups := flex.ExpandStringSet(d.Get("groups").(*schema.Set))

	var userErr, roleErr, groupErr error
	if len(users) != 0 {
		userErr = detachPolicyFromUsers(ctx, conn, users, arn)
	}
	if len(roles) != 0 {
		roleErr = detachPolicyFromRoles(ctx, conn, roles, arn)
	}
	if len(groups) != 0 {
		groupErr = detachPolicyFromGroups(ctx, conn, groups, arn)
	}
	if userErr != nil || roleErr != nil || groupErr != nil {
		return append(diags, composeErrors(fmt.Sprint("removing user, role, or group list from IAM Policy Detach ", name, ":"), userErr, roleErr, groupErr)...)
	}
	return diags
}

func composeErrors(desc string, uErr error, rErr error, gErr error) diag.Diagnostics {
	errMsg := fmt.Sprint(desc)
	errs := []error{uErr, rErr, gErr}
	for _, e := range errs {
		if e != nil {
			errMsg = errMsg + "\n– " + e.Error()
		}
	}
	return diag.Errorf(errMsg)
}

func attachPolicyToUsers(ctx context.Context, conn *iam.IAM, users []*string, arn string) error {
	for _, u := range users {
		_, err := conn.AttachUserPolicyWithContext(ctx, &iam.AttachUserPolicyInput{
			UserName:  u,
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func attachPolicyToRoles(ctx context.Context, conn *iam.IAM, roles []*string, arn string) error {
	for _, r := range roles {
		_, err := conn.AttachRolePolicyWithContext(ctx, &iam.AttachRolePolicyInput{
			RoleName:  r,
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func attachPolicyToGroups(ctx context.Context, conn *iam.IAM, groups []*string, arn string) error {
	for _, g := range groups {
		_, err := conn.AttachGroupPolicyWithContext(ctx, &iam.AttachGroupPolicyInput{
			GroupName: g,
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func updateUsers(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
	arn := d.Get("policy_arn").(string)
	o, n := d.GetChange("users")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	remove := flex.ExpandStringSet(os.Difference(ns))
	add := flex.ExpandStringSet(ns.Difference(os))

	if rErr := detachPolicyFromUsers(ctx, conn, remove, arn); rErr != nil {
		return rErr
	}
	if aErr := attachPolicyToUsers(ctx, conn, add, arn); aErr != nil {
		return aErr
	}
	return nil
}
func updateRoles(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
	arn := d.Get("policy_arn").(string)
	o, n := d.GetChange("roles")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	remove := flex.ExpandStringSet(os.Difference(ns))
	add := flex.ExpandStringSet(ns.Difference(os))

	if rErr := detachPolicyFromRoles(ctx, conn, remove, arn); rErr != nil {
		return rErr
	}
	if aErr := attachPolicyToRoles(ctx, conn, add, arn); aErr != nil {
		return aErr
	}
	return nil
}
func updateGroups(ctx context.Context, conn *iam.IAM, d *schema.ResourceData) error {
	arn := d.Get("policy_arn").(string)
	o, n := d.GetChange("groups")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	remove := flex.ExpandStringSet(os.Difference(ns))
	add := flex.ExpandStringSet(ns.Difference(os))

	if rErr := detachPolicyFromGroups(ctx, conn, remove, arn); rErr != nil {
		return rErr
	}
	if aErr := attachPolicyToGroups(ctx, conn, add, arn); aErr != nil {
		return aErr
	}
	return nil
}
func detachPolicyFromUsers(ctx context.Context, conn *iam.IAM, users []*string, arn string) error {
	for _, u := range users {
		_, err := conn.DetachUserPolicyWithContext(ctx, &iam.DetachUserPolicyInput{
			UserName:  u,
			PolicyArn: aws.String(arn),
		})
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
func detachPolicyFromRoles(ctx context.Context, conn *iam.IAM, roles []*string, arn string) error {
	for _, r := range roles {
		_, err := conn.DetachRolePolicyWithContext(ctx, &iam.DetachRolePolicyInput{
			RoleName:  r,
			PolicyArn: aws.String(arn),
		})
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
func detachPolicyFromGroups(ctx context.Context, conn *iam.IAM, groups []*string, arn string) error {
	for _, g := range groups {
		_, err := conn.DetachGroupPolicyWithContext(ctx, &iam.DetachGroupPolicyInput{
			GroupName: g,
			PolicyArn: aws.String(arn),
		})
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}
