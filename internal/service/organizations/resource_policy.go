// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_resource_policy", name="Resource Policy")
// @Tags(identifierAttribute="id")
func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyCreate,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyUpdate,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
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
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("content").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &organizations.PutResourcePolicyInput{
		Content: aws.String(policy),
		Tags:    getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 4*time.Minute, func() (interface{}, error) {
		return conn.PutResourcePolicyWithContext(ctx, input)
	}, organizations.ErrCodeFinalizingOrganizationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Resource Policy: %s", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*organizations.PutResourcePolicyOutput).ResourcePolicy.ResourcePolicySummary.Id))

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	policy, err := findResourcePolicy(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Resource Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Resource Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", policy.ResourcePolicySummary.Arn)
	if policyToSet, err := verify.PolicyToSet(d.Get("content").(string), aws.StringValue(policy.Content)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("content", policyToSet)
	}

	return diags
}

func resourceResourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		policy, err := structure.NormalizeJsonString(d.Get("content").(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &organizations.PutResourcePolicyInput{
			Content: aws.String(policy),
		}

		_, err = conn.PutResourcePolicyWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Organizations Resource Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	log.Printf("[DEBUG] Deleting Organizations Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicyWithContext(ctx, &organizations.DeleteResourcePolicyInput{})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeResourcePolicyNotFoundException) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Resource Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourcePolicy(ctx context.Context, conn *organizations.Organizations) (*organizations.ResourcePolicy, error) {
	input := &organizations.DescribeResourcePolicyInput{}

	output, err := conn.DescribeResourcePolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException, organizations.ErrCodeResourcePolicyNotFoundException) {
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
