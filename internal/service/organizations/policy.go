// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_policy", name="Policy")
// @Tags(identifierAttribute="id")
func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"skip_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      organizations.PolicyTypeServiceControlPolicy,
				ValidateFunc: validation.StringInSlice(organizations.PolicyType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	name := d.Get("name").(string)
	input := &organizations.CreatePolicyInput{
		Content:     aws.String(d.Get("content").(string)),
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
		Type:        aws.String(d.Get("type").(string)),
		Tags:        getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 4*time.Minute, func() (interface{}, error) {
		return conn.CreatePolicyWithContext(ctx, input)
	}, organizations.ErrCodeFinalizingOrganizationException)

	if err != nil {
		return diag.Errorf("creating Organizations Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*organizations.CreatePolicyOutput).Policy.PolicySummary.Id))

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	policy, err := findPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Policy %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Organizations Policy (%s): %s", d.Id(), err)
	}

	policySummary := policy.PolicySummary
	d.Set("arn", policySummary.Arn)
	d.Set("content", policy.Content)
	d.Set("description", policySummary.Description)
	d.Set("name", policySummary.Name)
	d.Set("type", policySummary.Type)

	if aws.BoolValue(policySummary.AwsManaged) {
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
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &organizations.UpdatePolicyInput{
			PolicyId: aws.String(d.Id()),
		}

		if d.HasChange("content") {
			input.Content = aws.String(d.Get("content").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		_, err := conn.UpdatePolicyWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Organizations policy (%s): %s", d.Id(), err)
		}
	}

	return resourcePolicyRead(ctx, d, meta)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if v, ok := d.GetOk("skip_destroy"); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy: %s", d.Id())
		return nil
	}

	log.Printf("[DEBUG] Deleting Organizations Policy: %s", d.Id())
	_, err := conn.DeletePolicyWithContext(ctx, &organizations.DeletePolicyInput{
		PolicyId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodePolicyNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Organizations policy (%s): %s", d.Id(), err)
	}

	return nil
}

func resourcePolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	policy, err := findPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	if aws.BoolValue(policy.PolicySummary.AwsManaged) {
		return nil, fmt.Errorf("AWS-managed Organizations policy (%s) cannot be imported. Use the policy ID directly in your configuration.", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}

func findPolicyByID(ctx context.Context, conn *organizations.Organizations, id string) (*organizations.Policy, error) {
	input := &organizations.DescribePolicyInput{
		PolicyId: aws.String(id),
	}

	output, err := conn.DescribePolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException, organizations.ErrCodePolicyNotFoundException) {
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
