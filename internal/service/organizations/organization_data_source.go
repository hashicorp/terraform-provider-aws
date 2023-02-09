package organizations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrganizationRead,

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
						"status": {
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
			"non_master_accounts": {
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
						"status": {
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
			"roots": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy_types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
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
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	org, err := conn.DescribeOrganizationWithContext(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing organization: %s", err)
	}

	d.SetId(aws.StringValue(org.Organization.Id))
	d.Set("arn", org.Organization.Arn)
	d.Set("feature_set", org.Organization.FeatureSet)
	d.Set("master_account_arn", org.Organization.MasterAccountArn)
	d.Set("master_account_email", org.Organization.MasterAccountEmail)
	d.Set("master_account_id", org.Organization.MasterAccountId)

	if aws.StringValue(org.Organization.MasterAccountId) == meta.(*conns.AWSClient).AccountID {
		var accounts []*organizations.Account
		var nonMasterAccounts []*organizations.Account
		err = conn.ListAccountsPagesWithContext(ctx, &organizations.ListAccountsInput{}, func(page *organizations.ListAccountsOutput, lastPage bool) bool {
			for _, account := range page.Accounts {
				if aws.StringValue(account.Id) != aws.StringValue(org.Organization.MasterAccountId) {
					nonMasterAccounts = append(nonMasterAccounts, account)
				}

				accounts = append(accounts, account)
			}

			return !lastPage
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing AWS Organization (%s) accounts: %s", d.Id(), err)
		}

		var roots []*organizations.Root
		err = conn.ListRootsPagesWithContext(ctx, &organizations.ListRootsInput{}, func(page *organizations.ListRootsOutput, lastPage bool) bool {
			roots = append(roots, page.Roots...)
			return !lastPage
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing AWS Organization (%s) roots: %s", d.Id(), err)
		}

		awsServiceAccessPrincipals := make([]string, 0)
		// ConstraintViolationException: The request failed because the organization does not have all features enabled. Please enable all features in your organization and then retry.
		if aws.StringValue(org.Organization.FeatureSet) == organizations.OrganizationFeatureSetAll {
			err = conn.ListAWSServiceAccessForOrganizationPagesWithContext(ctx, &organizations.ListAWSServiceAccessForOrganizationInput{}, func(page *organizations.ListAWSServiceAccessForOrganizationOutput, lastPage bool) bool {
				for _, enabledServicePrincipal := range page.EnabledServicePrincipals {
					awsServiceAccessPrincipals = append(awsServiceAccessPrincipals, aws.StringValue(enabledServicePrincipal.ServicePrincipal))
				}
				return !lastPage
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing AWS Service Access for Organization (%s): %s", d.Id(), err)
			}
		}

		enabledPolicyTypes := make([]string, 0)
		for _, policyType := range roots[0].PolicyTypes {
			if aws.StringValue(policyType.Status) == organizations.PolicyTypeStatusEnabled {
				enabledPolicyTypes = append(enabledPolicyTypes, aws.StringValue(policyType.Type))
			}
		}

		if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
		}

		if err := d.Set("aws_service_access_principals", awsServiceAccessPrincipals); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting aws_service_access_principals: %s", err)
		}

		if err := d.Set("enabled_policy_types", enabledPolicyTypes); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting enabled_policy_types: %s", err)
		}

		if err := d.Set("non_master_accounts", flattenAccounts(nonMasterAccounts)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting non_master_accounts: %s", err)
		}

		if err := d.Set("roots", FlattenRoots(roots)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting roots: %s", err)
		}
	}
	return diags
}
