// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_policy_attachment", name="Policy Attachment")
func resourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		UpdateWithoutTimeout: resourcePolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policyID := d.Get("policy_id").(string)
	targetID := d.Get("target_id").(string)
	id := policyAttachmentCreateResourceID(targetID, policyID)
	input := &organizations.AttachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout, func() (interface{}, error) {
		return conn.AttachPolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Policy Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	targetID, policyID, err := policyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findPolicyAttachmentByTwoPartKey(ctx, conn, targetID, policyID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Policy Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policy Attachment (%s): %s", d.Id(), err)
	}

	d.Set("policy_id", policyID)
	d.Set("target_id", targetID)

	return diags
}

func resourcePolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Update is just a pass-through to allow skip_destroy to be updated in-place.

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy Attachment: %s", d.Id())
		return nil
	}

	targetID, policyID, err := policyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Organizations Policy Attachment: %s", d.Id())
	_, err = conn.DetachPolicy(ctx, &organizations.DetachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	})

	if errs.IsA[*awstypes.PolicyNotAttachedException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) || errs.IsA[*awstypes.TargetNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Policy Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

const policyAttachmentResourceIDSeparator = ":"

func policyAttachmentCreateResourceID(targetID, policyID string) string {
	parts := []string{targetID, policyID}
	id := strings.Join(parts, policyAttachmentResourceIDSeparator)

	return id
}

func policyAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, policyAttachmentResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TARGETID%[2]sPOLICYID", id, policyAttachmentResourceIDSeparator)
}

func findPolicyAttachmentByTwoPartKey(ctx context.Context, conn *organizations.Client, targetID, policyID string) (*awstypes.PolicyTargetSummary, error) {
	input := &organizations.ListTargetsForPolicyInput{
		PolicyId: aws.String(policyID),
	}

	return findPolicyTarget(ctx, conn, input, func(v *awstypes.PolicyTargetSummary) bool {
		return aws.ToString(v.TargetId) == targetID
	})
}

func findPolicyTarget(ctx context.Context, conn *organizations.Client, input *organizations.ListTargetsForPolicyInput, filter tfslices.Predicate[*awstypes.PolicyTargetSummary]) (*awstypes.PolicyTargetSummary, error) {
	output, err := findPolicyTargets(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPolicyTargets(ctx context.Context, conn *organizations.Client, input *organizations.ListTargetsForPolicyInput, filter tfslices.Predicate[*awstypes.PolicyTargetSummary]) ([]awstypes.PolicyTargetSummary, error) {
	var output []awstypes.PolicyTargetSummary

	pages := organizations.NewListTargetsForPolicyPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) || errs.IsA[*awstypes.TargetNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Targets {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
