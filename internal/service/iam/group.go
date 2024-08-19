// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_group", name="Group")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`^[0-9A-Za-z=,.@\-_+]+$`),
					"must only contain alphanumeric characters, hyphens, underscores, commas, periods, @ symbols, plus and equals signs",
				),
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
			if d.HasChanges(names.AttrName, names.AttrPath) {
				return d.SetNewComputed(names.AttrARN)
			}

			return nil
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &iam.CreateGroupInput{
		GroupName: aws.String(name),
		Path:      aws.String(d.Get(names.AttrPath).(string)),
	}

	output, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Group.GroupName))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findGroupByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IAM Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	group, err := findGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, group.Arn)
	d.Set(names.AttrName, group.GroupName)
	d.Set(names.AttrPath, group.Path)
	d.Set("unique_id", group.GroupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	o, n := d.GetChange(names.AttrName)
	input := &iam.UpdateGroupInput{
		GroupName:    aws.String(o.(string)),
		NewGroupName: aws.String(n.(string)),
		NewPath:      aws.String(d.Get(names.AttrPath).(string)),
	}

	_, err := conn.UpdateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Group (%s): %s", d.Id(), err)
	}

	d.SetId(n.(string))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	log.Printf("[DEBUG] Deleting IAM Group: %s", d.Id())
	_, err := conn.DeleteGroup(ctx, &iam.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findGroupByName(ctx context.Context, conn *iam.Client, name string) (*awstypes.Group, error) {
	input := &iam.GetGroupInput{
		GroupName: aws.String(name),
	}

	output, err := conn.GetGroup(ctx, input)

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
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

func DeleteGroupPolicyAttachments(ctx context.Context, conn *iam.Client, groupName string) error {
	var attachedPolicies []awstypes.AttachedPolicy
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	pages := iam.NewListAttachedGroupPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing IAM Group (%s) policy attachments for deletion: %w", groupName, err)
		}

		attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)
	}

	for _, attachedPolicy := range attachedPolicies {
		input := &iam.DetachGroupPolicyInput{
			GroupName: aws.String(groupName),
			PolicyArn: attachedPolicy.PolicyArn,
		}

		_, err := conn.DetachGroupPolicy(ctx, input)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("detaching IAM Group (%s) policy (%s): %w", groupName, aws.ToString(attachedPolicy.PolicyArn), err)
		}
	}

	return nil
}

func DeleteGroupPolicies(ctx context.Context, conn *iam.Client, groupName string) error {
	var inlinePolicies []string
	input := &iam.ListGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	pages := iam.NewListGroupPoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing IAM Group (%s) inline policies for deletion: %w", groupName, err)
		}

		inlinePolicies = append(inlinePolicies, page.PolicyNames...)
	}

	for _, policyName := range inlinePolicies {
		input := &iam.DeleteGroupPolicyInput{
			GroupName:  aws.String(groupName),
			PolicyName: aws.String(policyName),
		}

		_, err := conn.DeleteGroupPolicy(ctx, input)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("deleting IAM Group (%s) inline policy (%s): %w", groupName, policyName, err)
		}
	}

	return nil
}
