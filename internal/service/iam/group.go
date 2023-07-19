// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iam_group")
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9A-Za-z=,.@\-_+]+$`),
					"must only contain alphanumeric characters, hyphens, underscores, commas, periods, @ symbols, plus and equals signs",
				),
			},
			"path": {
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
			if d.HasChanges("name", "path") {
				return d.SetNewComputed("arn")
			}

			return nil
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	name := d.Get("name").(string)
	input := &iam.CreateGroupInput{
		GroupName: aws.String(name),
		Path:      aws.String(d.Get("path").(string)),
	}

	output, err := conn.CreateGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Group.GroupName))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindGroupByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IAM Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	group, err := FindGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", group.Arn)
	d.Set("name", group.GroupName)
	d.Set("path", group.Path)
	d.Set("unique_id", group.GroupId)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	o, n := d.GetChange("name")
	input := &iam.UpdateGroupInput{
		GroupName:    aws.String(o.(string)),
		NewGroupName: aws.String(n.(string)),
		NewPath:      aws.String(d.Get("path").(string)),
	}

	_, err := conn.UpdateGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IAM Group (%s): %s", d.Id(), err)
	}

	d.SetId(n.(string))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	log.Printf("[DEBUG] Deleting IAM Group: %s", d.Id())
	_, err := conn.DeleteGroupWithContext(ctx, &iam.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindGroupByName(ctx context.Context, conn *iam.IAM, name string) (*iam.Group, error) {
	input := &iam.GetGroupInput{
		GroupName: aws.String(name),
	}

	output, err := conn.GetGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
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

func DeleteGroupPolicyAttachments(ctx context.Context, conn *iam.IAM, groupName string) error {
	var attachedPolicies []*iam.AttachedPolicy
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	err := conn.ListAttachedGroupPoliciesPagesWithContext(ctx, input, func(page *iam.ListAttachedGroupPoliciesOutput, lastPage bool) bool {
		attachedPolicies = append(attachedPolicies, page.AttachedPolicies...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing IAM Group (%s) policy attachments for deletion: %w", groupName, err)
	}

	for _, attachedPolicy := range attachedPolicies {
		input := &iam.DetachGroupPolicyInput{
			GroupName: aws.String(groupName),
			PolicyArn: attachedPolicy.PolicyArn,
		}

		_, err := conn.DetachGroupPolicyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("detaching IAM Group (%s) policy (%s): %w", groupName, aws.StringValue(attachedPolicy.PolicyArn), err)
		}
	}

	return nil
}

func DeleteGroupPolicies(ctx context.Context, conn *iam.IAM, groupName string) error {
	var inlinePolicies []*string
	input := &iam.ListGroupPoliciesInput{
		GroupName: aws.String(groupName),
	}

	err := conn.ListGroupPoliciesPagesWithContext(ctx, input, func(page *iam.ListGroupPoliciesOutput, lastPage bool) bool {
		inlinePolicies = append(inlinePolicies, page.PolicyNames...)
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing IAM Group (%s) inline policies for deletion: %w", groupName, err)
	}

	for _, policyName := range inlinePolicies {
		input := &iam.DeleteGroupPolicyInput{
			GroupName:  aws.String(groupName),
			PolicyName: policyName,
		}

		_, err := conn.DeleteGroupPolicyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("deleting IAM Group (%s) inline policy (%s): %w", groupName, aws.StringValue(policyName), err)
		}
	}

	return nil
}
