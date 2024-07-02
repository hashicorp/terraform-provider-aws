// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_policy", name="Policy")
// @Tags(identifierAttribute="id")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsJSON,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.PolicyTypeServiceControlPolicy,
				ValidateDiagFunc: enum.Validate[awstypes.PolicyType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &organizations.CreatePolicyInput{
		Content:     aws.String(d.Get(names.AttrContent).(string)),
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		Name:        aws.String(name),
		Type:        awstypes.PolicyType(d.Get(names.AttrType).(string)),
		Tags:        getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout, func() (interface{}, error) {
		return conn.CreatePolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*organizations.CreatePolicyOutput).Policy.PolicySummary.Id))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policy, err := findPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policy (%s): %s", d.Id(), err)
	}

	policySummary := policy.PolicySummary
	d.Set(names.AttrARN, policySummary.Arn)
	d.Set(names.AttrContent, policy.Content)
	d.Set(names.AttrDescription, policySummary.Description)
	d.Set(names.AttrName, policySummary.Name)
	d.Set(names.AttrType, policySummary.Type)

	if policySummary.AwsManaged {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "AWS-managed Organizations policies cannot be imported",
				Detail:   fmt.Sprintf("This resource should be removed from your Terraform state using `terraform state rm` (https://www.terraform.io/docs/commands/state/rm.html) and references should use the ID (%s) directly.", d.Id()),
			},
		}
	}

	return nil
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &organizations.UpdatePolicyInput{
			PolicyId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrContent) {
			input.Content = aws.String(d.Get(names.AttrContent).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdatePolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Organizations Policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy: %s", d.Id())
		return nil
	}

	log.Printf("[DEBUG] Deleting Organizations Policy: %s", d.Id())
	_, err := conn.DeletePolicy(ctx, &organizations.DeletePolicyInput{
		PolicyId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policy, err := findPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	if policy.PolicySummary.AwsManaged {
		return nil, fmt.Errorf("AWS-managed Organizations policy (%s) cannot be imported. Use the policy ID directly in your configuration", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}

func findPolicyByID(ctx context.Context, conn *organizations.Client, id string) (*awstypes.Policy, error) {
	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(id),
	}

	output, err := conn.DescribePolicy(ctx, input)

	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil || output.Policy.PolicySummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}
