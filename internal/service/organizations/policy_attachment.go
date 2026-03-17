// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_policy_attachment", name="Policy Attachment")
// @IdentityAttribute("policy_id")
// @IdentityAttribute("target_id")
// @ImportIDHandler("policyAttachmentImportID")
// @Testing(serialize=true)
// @Testing(preIdentityVersion="6.4.0")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
func resourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		UpdateWithoutTimeout: resourcePolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

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

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policyID := d.Get("policy_id").(string)
	targetID := d.Get("target_id").(string)
	id := policyAttachmentCreateResourceID(targetID, policyID)
	input := organizations.AttachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	}

	_, err := tfresource.RetryWhenIsA[any, *awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout, func(ctx context.Context) (any, error) {
		return conn.AttachPolicy(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Policy Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	targetID := d.Get("target_id").(string)
	policyID := d.Get("policy_id").(string)

	_, err := findPolicyAttachmentByTwoPartKey(ctx, conn, targetID, policyID)

	if !d.IsNewResource() && retry.NotFound(err) {
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

func resourcePolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Update is just a pass-through to allow skip_destroy to be updated in-place.

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy Attachment: %s", d.Id())
		return diags
	}

	targetID := d.Get("target_id").(string)
	policyID := d.Get("policy_id").(string)

	_, err := conn.DetachPolicy(ctx, &organizations.DetachPolicyInput{
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

func findPolicyAttachmentByTwoPartKey(ctx context.Context, conn *organizations.Client, targetID, policyID string) (*awstypes.PolicyTargetSummary, error) {
	input := organizations.ListTargetsForPolicyInput{
		PolicyId: aws.String(policyID),
	}

	return findPolicyTarget(ctx, conn, &input, func(v *awstypes.PolicyTargetSummary) bool {
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
				LastError: err,
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

type policyAttachmentImportID struct{}

func (policyAttachmentImportID) Create(d *schema.ResourceData) string {
	return policyAttachmentCreateResourceID(d.Get("target_id").(string), d.Get("policy_id").(string))
}

func (policyAttachmentImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, policyAttachmentResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		result := map[string]any{
			"target_id": parts[0],
			"policy_id": parts[1],
		}
		return id, result, nil
	}

	return "", nil, fmt.Errorf("unexpected format for ID (%[1]s), expected TARGETID%[2]sPOLICYID", id, policyAttachmentResourceIDSeparator)
}
