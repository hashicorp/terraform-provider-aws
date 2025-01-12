// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_describe_organization", name="Isolated Describe Organization")
func resourceDescribeOrganization() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: resourceDescribeOrganizationRead,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDescribeOrganizationImport,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("feature_set", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from ALL to CONSOLIDATED_BILLING for feature_set should force a new resource.
				return awstypes.OrganizationFeatureSet(old.(string)) == awstypes.OrganizationFeatureSetAll && awstypes.OrganizationFeatureSet(new.(string)) == awstypes.OrganizationFeatureSetConsolidatedBilling
			}),
		),

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
						names.AttrStatus: {
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
			"aws_service_access_principals": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enabled_policy_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.PolicyType](),
				},
			},
			"feature_set": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.OrganizationFeatureSetAll,
				ValidateDiagFunc: enum.Validate[awstypes.OrganizationFeatureSet](),
			},
			"master_account_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"non_master_accounts": {
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
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"roots": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
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
						"policy_types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStatus: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceDescribeOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	org, err := findOrganization(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Organization does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organization (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, org.Arn)
	d.Set("feature_set", org.FeatureSet)
	d.Set("master_account_arn", org.MasterAccountArn)
	d.Set("master_account_email", org.MasterAccountEmail)
	d.Set("master_account_id", org.MasterAccountId)

	return diags
}

func resourceDescribeOrganizationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	org, err := findOrganization(ctx, conn)

	if err != nil {
		return nil, err
	}

	// Check that any Org ID specified for import matches the current Org ID.
	if got, want := aws.ToString(org.Id), d.Id(); got != want {
		return nil, fmt.Errorf("current Organizations Organization ID (%s) does not match (%s)", got, want)
	}

	return []*schema.ResourceData{d}, nil
}
