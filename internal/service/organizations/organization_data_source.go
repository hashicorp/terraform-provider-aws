// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_organizations_organization")
func DataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationRead,

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
						names.AttrStatus: {
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
			"aws_service_access_principals": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enabled_policy_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"feature_set": {
				Type:     schema.TypeString,
				Computed: true,
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
						names.AttrStatus: {
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
			"roots": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrARN: {
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

func dataSourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	org, err := findOrganization(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organization: %s", err)
	}

	d.SetId(aws.StringValue(org.Id))
	d.Set(names.AttrARN, org.Arn)
	d.Set("feature_set", org.FeatureSet)
	d.Set("master_account_arn", org.MasterAccountArn)
	d.Set("master_account_email", org.MasterAccountEmail)
	managementAccountID := aws.StringValue(org.MasterAccountId)
	d.Set("master_account_id", managementAccountID)

	isManagementAccount := managementAccountID == meta.(*conns.AWSClient).AccountID
	isDelegatedAdministrator := true
	accounts, err := findAccounts(ctx, conn)

	managementAccountName := ""
	for _, v := range accounts {
		if aws.StringValue(v.Id) == managementAccountID {
			managementAccountName = aws.StringValue(v.Name)
		}
	}
	d.Set("master_account_name", managementAccountName)

	if err != nil {
		if isManagementAccount || !tfawserr.ErrCodeEquals(err, organizations.ErrCodeAccessDeniedException) {
			return sdkdiag.AppendErrorf(diags, "reading Organization (%s) accounts: %s", d.Id(), err)
		}

		isDelegatedAdministrator = false
	}

	if isManagementAccount || isDelegatedAdministrator {
		nonManagementAccounts := tfslices.Filter(accounts, func(v *organizations.Account) bool {
			return aws.StringValue(v.Id) != managementAccountID
		})

		roots, err := findRoots(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Organization (%s) roots: %s", d.Id(), err)
		}

		var awsServiceAccessPrincipals []string

		// ConstraintViolationException: The request failed because the organization does not have all features enabled. Please enable all features in your organization and then retry.
		if aws.StringValue(org.FeatureSet) == organizations.OrganizationFeatureSetAll {
			awsServiceAccessPrincipals, err = findEnabledServicePrincipalNames(ctx, conn)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Organization (%s) service principals: %s", d.Id(), err)
			}
		}

		var enabledPolicyTypes []string

		for _, policyType := range roots[0].PolicyTypes {
			if aws.StringValue(policyType.Status) == organizations.PolicyTypeStatusEnabled {
				enabledPolicyTypes = append(enabledPolicyTypes, aws.StringValue(policyType.Type))
			}
		}

		if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
		}
		d.Set("aws_service_access_principals", awsServiceAccessPrincipals)
		d.Set("enabled_policy_types", enabledPolicyTypes)
		if err := d.Set("non_master_accounts", flattenAccounts(nonManagementAccounts)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting non_master_accounts: %s", err)
		}
		if err := d.Set("roots", flattenRoots(roots)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting roots: %s", err)
		}
	}

	return diags
}
