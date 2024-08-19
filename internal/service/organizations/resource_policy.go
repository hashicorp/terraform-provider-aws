// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_resource_policy", name="Resource Policy")
// @Tags(identifierAttribute="id")
func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyCreate,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyUpdate,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceResourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrContent).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &organizations.PutResourcePolicyInput{
		Content: aws.String(policy),
		Tags:    getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout, func() (interface{}, error) {
		return conn.PutResourcePolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Resource Policy: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*organizations.PutResourcePolicyOutput).ResourcePolicy.ResourcePolicySummary.Id))

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policy, err := findResourcePolicy(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Resource Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Resource Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, policy.ResourcePolicySummary.Arn)
	if policyToSet, err := verify.PolicyToSet(d.Get(names.AttrContent).(string), aws.ToString(policy.Content)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set(names.AttrContent, policyToSet)
	}

	return diags
}

func resourceResourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		policy, err := structure.NormalizeJsonString(d.Get(names.AttrContent).(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &organizations.PutResourcePolicyInput{
			Content: aws.String(policy),
		}

		_, err = conn.PutResourcePolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Organizations Resource Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	log.Printf("[DEBUG] Deleting Organizations Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicy(ctx, &organizations.DeleteResourcePolicyInput{})

	if errs.IsA[*awstypes.ResourcePolicyNotFoundException](err) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicy(ctx context.Context, conn *organizations.Client) (*awstypes.ResourcePolicy, error) {
	input := &organizations.DescribeResourcePolicyInput{}

	output, err := conn.DescribeResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) || errs.IsA[*awstypes.ResourcePolicyNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourcePolicy == nil || output.ResourcePolicy.ResourcePolicySummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourcePolicy, nil
}
