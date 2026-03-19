// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package logs

import (
	"context"
	"fmt"
	"iter"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_resource_policy", name="Resource Policy")
func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				if arn.IsARN(d.Id()) {
					d.Set(names.AttrResourceARN, d.Id())
					d.Set("policy_scope", awstypes.PolicyScopeResource)
				} else {
					d.Set("policy_name", d.Id())
					d.Set("policy_scope", awstypes.PolicyScopeAccount)
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"policy_document": sdkv2.IAMPolicyDocumentSchemaRequired(),
			"policy_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ConflictsWith: []string{
					names.AttrResourceARN,
				},
				ExactlyOneOf: []string{"policy_name", names.AttrResourceARN},
			},
			"policy_scope": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				ConflictsWith: []string{
					"policy_name",
				},
				ExactlyOneOf: []string{"policy_name", names.AttrResourceARN},
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get("policy_name").(string)
	input := &cloudwatchlogs.PutResourcePolicyInput{
		PolicyDocument: aws.String(policy),
	}

	if v, ok := d.GetOk("policy_name"); ok {
		input.PolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrResourceARN); ok {
		input.ResourceArn = aws.String(v.(string))
	}

	if !d.IsNewResource() {
		if v, ok := d.GetOk("revision_id"); ok {
			input.ExpectedRevisionId = aws.String(v.(string))
		}
	}
	_, err = conn.PutResourcePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Logs Resource Policy (%s): %s", name, err)
	}

	if d.IsNewResource() {
		// For account-scoped policies, use the policy name as the ID.
		// For resource-scoped policies, use the resource ARN as the ID.
		if input.PolicyName != nil {
			d.SetId(name)
			d.Set("policy_scope", awstypes.PolicyScopeAccount)
		} else if input.ResourceArn != nil {
			d.SetId(aws.ToString(input.ResourceArn))
			d.Set("policy_scope", awstypes.PolicyScopeResource)
		}
	}

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	var resourcePolicy *awstypes.ResourcePolicy
	var err error
	if v, ok := d.GetOk("policy_scope"); ok && v.(string) == string(awstypes.PolicyScopeResource) {
		resourcePolicy, err = findResourcePolicyByResourceARN(ctx, conn, d.Id())
	} else {
		resourcePolicy, err = findResourcePolicyByName(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Resource Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.ToString(resourcePolicy.PolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("policy_document", policyToSet)
	d.Set("policy_scope", resourcePolicy.PolicyScope)
	d.Set(names.AttrResourceARN, resourcePolicy.ResourceArn)
	d.Set("revision_id", resourcePolicy.RevisionId)

	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	input := cloudwatchlogs.DeleteResourcePolicyInput{}
	if v, ok := d.GetOk("policy_scope"); ok && v.(string) == string(awstypes.PolicyScopeResource) {
		log.Printf("[DEBUG] Deleting CloudWatch Logs Resource Policy by ARN: %s", d.Id())
		revisionID := d.Get("revision_id").(string)
		if revisionID == "" {
			return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Resource Policy (%s): missing required revision_id", d.Id())
		}
		input.ResourceArn = aws.String(d.Id())
		input.ExpectedRevisionId = aws.String(revisionID)
	} else {
		log.Printf("[DEBUG] Deleting CloudWatch Logs Resource Policy by Name: %s", d.Id())
		input.PolicyName = aws.String(d.Id())
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicyByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.ResourcePolicy, error) {
	input := cloudwatchlogs.DescribeResourcePoliciesInput{}
	output, err := findResourcePolicy(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.ResourcePolicy) bool {
		return aws.ToString(v.PolicyName) == name
	}), tfslices.WithReturnFirstMatch[awstypes.ResourcePolicy]())

	if err != nil {
		return nil, err
	}

	if output.PolicyDocument == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, err
}

func findResourcePolicyByResourceARN(ctx context.Context, conn *cloudwatchlogs.Client, arn string) (*awstypes.ResourcePolicy, error) {
	input := cloudwatchlogs.DescribeResourcePoliciesInput{
		ResourceArn: aws.String(arn),
		PolicyScope: awstypes.PolicyScopeResource,
	}
	output, err := findResourcePolicy(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.ResourcePolicy) bool {
		return aws.ToString(v.ResourceArn) == arn
	}), tfslices.WithReturnFirstMatch[awstypes.ResourcePolicy]())

	if err != nil {
		return nil, err
	}

	if output.PolicyDocument == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, err
}

func findResourcePolicy(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeResourcePoliciesInput, optFns ...tfslices.FinderOptionsFunc[awstypes.ResourcePolicy]) (*awstypes.ResourcePolicy, error) {
	output, err := findResourcePolicies(ctx, conn, input, optFns...)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourcePolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeResourcePoliciesInput, optFns ...tfslices.FinderOptionsFunc[awstypes.ResourcePolicy]) ([]awstypes.ResourcePolicy, error) {
	return tfslices.CollectWithErrorAndConcat(listResourcePolicies(ctx, conn, input), optFns...)
}

func listResourcePolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeResourcePoliciesInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[[]awstypes.ResourcePolicy, error] {
	return func(yield func([]awstypes.ResourcePolicy, error) bool) {
		err := describeResourcePoliciesPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeResourcePoliciesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			if !yield(page.ResourcePolicies, nil) {
				return false
			}

			return !lastPage
		}, optFns...)

		if err != nil {
			yield(nil, fmt.Errorf("listing CloudWatch Logs Resource Policies: %w", err))
			return
		}
	}
}
