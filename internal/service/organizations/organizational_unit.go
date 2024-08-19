// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_organizational_unit", name="Organizational Unit")
// @Tags(identifierAttribute="id")
func resourceOrganizationalUnit() *schema.Resource {
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
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEmail: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"parent_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[0-9a-z]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateOrganizationalUnit.html#organizations-CreateOrganizationalUnit-request-ParentId"),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOrganizationalUnitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String(name),
		ParentId: aws.String(d.Get("parent_id").(string)),
		Tags:     getTagsIn(ctx),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.FinalizingOrganizationException](ctx, organizationFinalizationTimeout, func() (interface{}, error) {
		return conn.CreateOrganizationalUnit(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Organizational Unit (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*organizations.CreateOrganizationalUnitOutput).OrganizationalUnit.Id))

	return append(diags, resourceOrganizationalUnitRead(ctx, d, meta)...)
}

func resourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	ou, err := findOrganizationalUnitByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organizational Unit (%s): %s", d.Id(), err)
	}

	parentAccountID, err := findParentAccountID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Organizational Unit (%s) parents: %s", d.Id(), err)
	}

	accounts, err := findAccountsForParentByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Organizations Accounts for parent (%s): %s", d.Id(), err)
	}

	if err := d.Set("accounts", flattenOrganizationalUnitAccounts(accounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
	}
	d.Set(names.AttrARN, ou.Arn)
	d.Set(names.AttrName, ou.Name)
	d.Set("parent_id", parentAccountID)

	return diags
}

func resourceOrganizationalUnitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if d.HasChange(names.AttrName) {
		input := &organizations.UpdateOrganizationalUnitInput{
			Name:                 aws.String(d.Get(names.AttrName).(string)),
			OrganizationalUnitId: aws.String(d.Id()),
		}

		_, err := conn.UpdateOrganizationalUnit(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Organizations Organizational Unit (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceOrganizationalUnitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	log.Printf("[DEBUG] Deleting Organizations Organizational Unit: %s", d.Id())
	_, err := conn.DeleteOrganizationalUnit(ctx, &organizations.DeleteOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.OrganizationalUnitNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Organizational Unit (%s): %s", d.Id(), err)
	}

	return diags
}

func findOrganizationalUnitByID(ctx context.Context, conn *organizations.Client, id string) (*awstypes.OrganizationalUnit, error) {
	input := &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(id),
	}

	return findOrganizationalUnit(ctx, conn, input)
}

func findOrganizationalUnit(ctx context.Context, conn *organizations.Client, input *organizations.DescribeOrganizationalUnitInput) (*awstypes.OrganizationalUnit, error) {
	output, err := conn.DescribeOrganizationalUnit(ctx, input)

	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) || errs.IsA[*awstypes.OrganizationalUnitNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OrganizationalUnit == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.OrganizationalUnit, nil
}

func flattenOrganizationalUnitAccounts(apiObjects []awstypes.Account) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrARN:   aws.ToString(apiObject.Arn),
			names.AttrEmail: aws.ToString(apiObject.Email),
			names.AttrID:    aws.ToString(apiObject.Id),
			names.AttrName:  aws.ToString(apiObject.Name),
		})
	}

	return tfList
}
