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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_organization")
func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationCreate,
		ReadWithoutTimeout:   resourceOrganizationRead,
		UpdateWithoutTimeout: resourceOrganizationUpdate,
		DeleteWithoutTimeout: resourceOrganizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceOrganizationImport,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("feature_set", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from ALL to CONSOLIDATED_BILLING for feature_set should force a new resource.
				return old.(string) == organizations.OrganizationFeatureSetAll && new.(string) == organizations.OrganizationFeatureSetConsolidatedBilling
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
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(organizations.PolicyType_Values(), false),
				},
			},
			"feature_set": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      organizations.OrganizationFeatureSetAll,
				ValidateFunc: validation.StringInSlice(organizations.OrganizationFeatureSet_Values(), true),
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

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	input := &organizations.CreateOrganizationInput{
		FeatureSet: aws.String(d.Get("feature_set").(string)),
	}

	output, err := conn.CreateOrganizationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organization: %s", err)
	}

	d.SetId(aws.StringValue(output.Organization.Id))

	if v, ok := d.GetOk("aws_service_access_principals"); ok && v.(*schema.Set).Len() > 0 {
		for _, principal := range flex.ExpandStringValueSet(v.(*schema.Set)) {
			input := &organizations.EnableAWSServiceAccessInput{
				ServicePrincipal: aws.String(principal),
			}

			_, err := conn.EnableAWSServiceAccessWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling AWS Service Access (%s) in Organization (%s): %s", principal, d.Id(), err)
			}
		}
	}

	if v, ok := d.GetOk("enabled_policy_types"); ok && v.(*schema.Set).Len() > 0 {
		defaultRoot, err := findDefaultRoot(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Organization (%s) default root: %s", d.Id(), err)
		}

		defaultRootID := aws.StringValue(defaultRoot.Id)

		for _, policyType := range flex.ExpandStringValueSet(v.(*schema.Set)) {
			input := &organizations.EnablePolicyTypeInput{
				PolicyType: aws.String(policyType),
				RootId:     aws.String(defaultRootID),
			}

			if _, err := conn.EnablePolicyTypeWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling policy type (%s) in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
			}

			if err := waitDefaultRootPolicyTypeEnabled(ctx, conn, policyType); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for policy type (%s) in Organization (%s) enable: %s", policyType, d.Id(), err)
			}
		}
	}

	return append(diags, resourceOrganizationRead(ctx, d, meta)...)
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	org, err := findOrganization(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organization does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organization: %s", err)
	}

	accounts, err := findAccounts(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organization (%s) accounts: %s", d.Id(), err)
	}

	managementAccountID := aws.StringValue(org.MasterAccountId)
	managementAccountName := ""
	for _, v := range accounts {
		if aws.StringValue(v.Id) == managementAccountID {
			managementAccountName = aws.StringValue(v.Name)
		}
	}
	nonManagementAccounts := tfslices.Filter(accounts, func(v *organizations.Account) bool {
		return aws.StringValue(v.Id) != managementAccountID
	})

	roots, err := findRoots(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organization (%s) roots: %s", d.Id(), err)
	}

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
	}
	d.Set(names.AttrARN, org.Arn)
	d.Set("feature_set", org.FeatureSet)
	d.Set("master_account_arn", org.MasterAccountArn)
	d.Set("master_account_email", org.MasterAccountEmail)
	d.Set("master_account_id", org.MasterAccountId)
	d.Set("master_account_name", managementAccountName)
	if err := d.Set("non_master_accounts", flattenAccounts(nonManagementAccounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting non_master_accounts: %s", err)
	}
	if err := d.Set("roots", flattenRoots(roots)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting roots: %s", err)
	}

	var awsServiceAccessPrincipals []string

	// ConstraintViolationException: The request failed because the organization does not have all features enabled. Please enable all features in your organization and then retry.
	if aws.StringValue(org.FeatureSet) == organizations.OrganizationFeatureSetAll {
		awsServiceAccessPrincipals, err = findEnabledServicePrincipalNames(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Organization (%s) service principals: %s", d.Id(), err)
		}
	}

	d.Set("aws_service_access_principals", awsServiceAccessPrincipals)

	var enabledPolicyTypes []string

	for _, policyType := range roots[0].PolicyTypes {
		if aws.StringValue(policyType.Status) == organizations.PolicyTypeStatusEnabled {
			enabledPolicyTypes = append(enabledPolicyTypes, aws.StringValue(policyType.Type))
		}
	}

	d.Set("enabled_policy_types", enabledPolicyTypes)

	return diags
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	if d.HasChange("aws_service_access_principals") {
		o, n := d.GetChange("aws_service_access_principals")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		for _, principal := range del {
			if err := disableServicePrincipal(ctx, conn, principal); err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling AWS Service Access (%s) in Organization (%s): %s", principal, d.Id(), err)
			}
		}

		for _, principal := range add {
			input := &organizations.EnableAWSServiceAccessInput{
				ServicePrincipal: aws.String(principal),
			}

			_, err := conn.EnableAWSServiceAccessWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling AWS Service Access (%s) in Organization (%s): %s", principal, d.Id(), err)
			}
		}
	}

	if d.HasChange("enabled_policy_types") {
		defaultRootID := d.Get("roots.0.id").(string)
		o, n := d.GetChange("enabled_policy_types")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		for _, policyType := range del {
			input := &organizations.DisablePolicyTypeInput{
				PolicyType: aws.String(policyType),
				RootId:     aws.String(defaultRootID),
			}

			if _, err := conn.DisablePolicyTypeWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling policy type (%s) in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
			}

			if err := waitDefaultRootPolicyTypeDisabled(ctx, conn, policyType); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for policy type (%s) in Organization (%s) disable: %s", policyType, d.Id(), err)
			}
		}

		for _, policyType := range add {
			input := &organizations.EnablePolicyTypeInput{
				PolicyType: aws.String(policyType),
				RootId:     aws.String(defaultRootID),
			}

			if _, err := conn.EnablePolicyTypeWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling policy type (%s) in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
			}

			if err := waitDefaultRootPolicyTypeEnabled(ctx, conn, policyType); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for policy type (%s) in Organization (%s) enable: %s", policyType, d.Id(), err)
			}
		}
	}

	if d.HasChange("feature_set") {
		if _, err := conn.EnableAllFeaturesWithContext(ctx, &organizations.EnableAllFeaturesInput{}); err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling all features in Organization (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceOrganizationRead(ctx, d, meta)...)
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	log.Printf("[INFO] Deleting Organization: %s", d.Id())
	_, err := conn.DeleteOrganizationWithContext(ctx, &organizations.DeleteOrganizationInput{})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organization (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceOrganizationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	org, err := findOrganization(ctx, conn)

	if err != nil {
		return nil, err
	}

	// Check that any Org ID specified for import matches the current Org ID.
	if got, want := aws.StringValue(org.Id), d.Id(); got != want {
		return nil, fmt.Errorf("current Organization ID (%s) does not match (%s)", got, want)
	}

	return []*schema.ResourceData{d}, nil
}

func findOrganization(ctx context.Context, conn *organizations.Organizations) (*organizations.Organization, error) {
	input := &organizations.DescribeOrganizationInput{}

	output, err := conn.DescribeOrganizationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Organization == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Organization, nil
}

func findAccounts(ctx context.Context, conn *organizations.Organizations) ([]*organizations.Account, error) {
	input := &organizations.ListAccountsInput{}
	var output []*organizations.Account

	err := conn.ListAccountsPagesWithContext(ctx, input, func(page *organizations.ListAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Accounts...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findEnabledServicePrincipalNames(ctx context.Context, conn *organizations.Organizations) ([]string, error) {
	output, err := findEnabledServicePrincipals(ctx, conn)

	if err != nil {
		return nil, err
	}

	return tfslices.ApplyToAll(output, func(v *organizations.EnabledServicePrincipal) string {
		return aws.StringValue(v.ServicePrincipal)
	}), nil
}

func findEnabledServicePrincipals(ctx context.Context, conn *organizations.Organizations) ([]*organizations.EnabledServicePrincipal, error) {
	input := &organizations.ListAWSServiceAccessForOrganizationInput{}
	var output []*organizations.EnabledServicePrincipal

	err := conn.ListAWSServiceAccessForOrganizationPagesWithContext(ctx, input, func(page *organizations.ListAWSServiceAccessForOrganizationOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.EnabledServicePrincipals...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findRoots(ctx context.Context, conn *organizations.Organizations) ([]*organizations.Root, error) {
	input := &organizations.ListRootsInput{}
	var output []*organizations.Root

	err := conn.ListRootsPagesWithContext(ctx, input, func(page *organizations.ListRootsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Roots...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findDefaultRoot(ctx context.Context, conn *organizations.Organizations) (*organizations.Root, error) {
	output, err := findRoots(ctx, conn)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output[0], nil
}

func flattenAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(accounts) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			names.AttrARN:    aws.StringValue(account.Arn),
			names.AttrEmail:  aws.StringValue(account.Email),
			names.AttrID:     aws.StringValue(account.Id),
			names.AttrName:   aws.StringValue(account.Name),
			names.AttrStatus: aws.StringValue(account.Status),
		})
	}
	return result
}

func flattenRoots(roots []*organizations.Root) []map[string]interface{} {
	if len(roots) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, r := range roots {
		result = append(result, map[string]interface{}{
			names.AttrID:   aws.StringValue(r.Id),
			names.AttrName: aws.StringValue(r.Name),
			names.AttrARN:  aws.StringValue(r.Arn),
			"policy_types": flattenRootPolicyTypeSummaries(r.PolicyTypes),
		})
	}
	return result
}

func flattenRootPolicyTypeSummaries(summaries []*organizations.PolicyTypeSummary) []map[string]interface{} {
	if len(summaries) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, s := range summaries {
		result = append(result, map[string]interface{}{
			names.AttrStatus: aws.StringValue(s.Status),
			names.AttrType:   aws.StringValue(s.Type),
		})
	}
	return result
}

func statusDefaultRootPolicyType(ctx context.Context, conn *organizations.Organizations, policyType string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		defaultRoot, err := findDefaultRoot(ctx, conn)

		if err != nil {
			return nil, "", err
		}

		for _, v := range defaultRoot.PolicyTypes {
			if aws.StringValue(v.Type) == policyType {
				return v, aws.StringValue(v.Status), nil
			}
		}

		return &organizations.PolicyTypeSummary{}, policyTypeStatusDisabled, nil
	}
}

const policyTypeStatusDisabled = "DISABLED"

func waitDefaultRootPolicyTypeDisabled(ctx context.Context, conn *organizations.Organizations, policyType string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{organizations.PolicyTypeStatusEnabled, organizations.PolicyTypeStatusPendingDisable},
		Target:  []string{policyTypeStatusDisabled},
		Refresh: statusDefaultRootPolicyType(ctx, conn, policyType),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDefaultRootPolicyTypeEnabled(ctx context.Context, conn *organizations.Organizations, policyType string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{policyTypeStatusDisabled, organizations.PolicyTypeStatusPendingEnable},
		Target:  []string{organizations.PolicyTypeStatusEnabled},
		Refresh: statusDefaultRootPolicyType(ctx, conn, policyType),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func disableServicePrincipal(ctx context.Context, conn *organizations.Organizations, servicePrincipal string) error {
	input := &organizations.DisableAWSServiceAccessInput{
		ServicePrincipal: aws.String(servicePrincipal),
	}

	_, err := conn.DisableAWSServiceAccessWithContext(ctx, input)

	return err
}
