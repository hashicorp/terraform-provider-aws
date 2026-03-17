// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package organizations

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_policy", name="Policy")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @CustomImport
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/organizations/types;awstypes;awstypes.Policy")
// @Testing(serialize=true)
// @Testing(preIdentityVersion="6.4.0")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
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
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	outputRaw, err := tfresource.RetryWhenIsA[any, *awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreatePolicy(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*organizations.CreatePolicyOutput).Policy.PolicySummary.Id))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	policy, err := findPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Organizations Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
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
		return append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "AWS-managed Organizations policies cannot be imported",
			Detail:   fmt.Sprintf("This resource should be removed from your Terraform state using `terraform state rm` (https://www.terraform.io/docs/commands/state/rm.html) and references should use the ID (%s) directly.", d.Id()),
		})
	}

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy: %s", d.Id())
		return diags
	}

	log.Printf("[DEBUG] Deleting Organizations Policy: %s", d.Id())
	_, err := conn.DeletePolicy(ctx, &organizations.DeletePolicyInput{
		PolicyId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importer.Import(ctx, d, meta); err != nil {
		return nil, err
	}

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
	input := organizations.DescribePolicyInput{
		PolicyId: aws.String(id),
	}

	return findPolicy(ctx, conn, &input)
}

func findPolicy(ctx context.Context, conn *organizations.Client, input *organizations.DescribePolicyInput) (*awstypes.Policy, error) {
	output, err := conn.DescribePolicy(ctx, input)

	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) || errs.IsA[*awstypes.PolicyNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil || output.Policy.PolicySummary == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Policy, nil
}
