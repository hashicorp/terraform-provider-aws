// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_organizational_unit", name="Organizational Unit")
// @Tags(identifierAttribute="id")
func ResourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationalUnitCreate,
		ReadWithoutTimeout:   resourceOrganizationalUnitRead,
		UpdateWithoutTimeout: resourceOrganizationalUnitUpdate,
		DeleteWithoutTimeout: resourceOrganizationalUnitDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"parent_id": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateOrganizationalUnit.html#organizations-CreateOrganizationalUnit-request-ParentId"),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOrganizationalUnitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	// Create the organizational unit
	createOpts := &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String(d.Get("name").(string)),
		ParentId: aws.String(d.Get("parent_id").(string)),
		Tags:     getTagsIn(ctx),
	}

	var err error
	var resp *organizations.CreateOrganizationalUnitOutput
	err = retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		resp, err = conn.CreateOrganizationalUnitWithContext(ctx, createOpts)

		if tfawserr.ErrCodeEquals(err, organizations.ErrCodeFinalizingOrganizationException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreateOrganizationalUnitWithContext(ctx, createOpts)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Organizational Unit: %s", err)
	}

	// Store the ID
	d.SetId(aws.StringValue(resp.OrganizationalUnit.Id))

	return append(diags, resourceOrganizationalUnitRead(ctx, d, meta)...)
}

func resourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	describeOpts := &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeOrganizationalUnitWithContext(ctx, describeOpts)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organizational Unit (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organizational Unit (%s): empty response", d.Id())
	}

	ou := resp.OrganizationalUnit
	if ou == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading Organizations Organizational Unit (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	parentId, err := resourceOrganizationalUnitGetParentID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organizational Unit (%s) parents: %s", d.Id(), err)
	}

	var accounts []*organizations.Account
	input := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(d.Id()),
	}

	err = conn.ListAccountsForParentPagesWithContext(ctx, input, func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
		accounts = append(accounts, page.Accounts...)

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organizational Unit (%s) accounts: %s", d.Id(), err)
	}

	if err := d.Set("accounts", flattenOrganizationalUnitAccounts(accounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
	}

	d.Set("arn", ou.Arn)
	d.Set("name", ou.Name)
	d.Set("parent_id", parentId)

	return diags
}

func resourceOrganizationalUnitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if d.HasChange("name") {
		updateOpts := &organizations.UpdateOrganizationalUnitInput{
			Name:                 aws.String(d.Get("name").(string)),
			OrganizationalUnitId: aws.String(d.Id()),
		}

		_, err := conn.UpdateOrganizationalUnitWithContext(ctx, updateOpts)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Organizations Organizational Unit (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceOrganizationalUnitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	input := &organizations.DeleteOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationalUnitWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Organizational Unit (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceOrganizationalUnitGetParentID(ctx context.Context, conn *organizations.Organizations, childId string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	}
	var parents []*organizations.Parent

	err := conn.ListParentsPagesWithContext(ctx, input, func(page *organizations.ListParentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		parents = append(parents, page.Parents...)

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if len(parents) == 0 {
		return "", nil
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	parent := parents[0]
	return aws.StringValue(parent.Id), nil
}

func flattenOrganizationalUnitAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(accounts) == 0 {
		return nil
	}

	var result []map[string]interface{}

	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"arn":   aws.StringValue(account.Arn),
			"email": aws.StringValue(account.Email),
			"id":    aws.StringValue(account.Id),
			"name":  aws.StringValue(account.Name),
		})
	}

	return result
}
