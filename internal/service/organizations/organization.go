// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_organizations_organization", name="Organization")
func resourceOrganization() *schema.Resource {
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

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	input := &organizations.CreateOrganizationInput{
		FeatureSet: awstypes.OrganizationFeatureSet(d.Get("feature_set").(string)),
	}

	output, err := conn.CreateOrganization(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Organization: %s", err)
	}

	d.SetId(aws.ToString(output.Organization.Id))

	if v, ok := d.GetOk("aws_service_access_principals"); ok && v.(*schema.Set).Len() > 0 {
		for _, v := range flex.ExpandStringValueSet(v.(*schema.Set)) {
			if err := enableServicePrincipal(ctx, conn, v); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if v, ok := d.GetOk("enabled_policy_types"); ok && v.(*schema.Set).Len() > 0 {
		defaultRoot, err := findDefaultRoot(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Organizations Organization (%s) default root: %s", d.Id(), err)
		}

		defaultRootID := aws.ToString(defaultRoot.Id)

		for _, v := range flex.ExpandStringValueSet(v.(*schema.Set)) {
			if err := enablePolicyType(ctx, conn, awstypes.PolicyType(v), defaultRootID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceOrganizationRead(ctx, d, meta)...)
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	accounts, err := findAccounts(ctx, conn, &organizations.ListAccountsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organization (%s) accounts: %s", d.Id(), err)
	}

	managementAccountID := aws.ToString(org.MasterAccountId)
	var managementAccountName *string
	for _, v := range accounts {
		if aws.ToString(v.Id) == managementAccountID {
			managementAccountName = v.Name
		}
	}
	nonManagementAccounts := tfslices.Filter(accounts, func(v awstypes.Account) bool {
		return aws.ToString(v.Id) != managementAccountID
	})

	roots, err := findRoots(ctx, conn, &organizations.ListRootsInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Organization (%s) roots: %s", d.Id(), err)
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
	if org.FeatureSet == awstypes.OrganizationFeatureSetAll {
		awsServiceAccessPrincipals, err = findEnabledServicePrincipalNames(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Organizations Organization (%s) service principals: %s", d.Id(), err)
		}
	}

	d.Set("aws_service_access_principals", awsServiceAccessPrincipals)

	var enabledPolicyTypes []awstypes.PolicyType

	for _, v := range roots[0].PolicyTypes {
		if v.Status == awstypes.PolicyTypeStatusEnabled {
			enabledPolicyTypes = append(enabledPolicyTypes, v.Type)
		}
	}

	d.Set("enabled_policy_types", enabledPolicyTypes)

	return diags
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	if d.HasChange("aws_service_access_principals") {
		o, n := d.GetChange("aws_service_access_principals")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		for _, v := range del {
			if err := disableServicePrincipal(ctx, conn, v); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		for _, v := range add {
			if err := enableServicePrincipal(ctx, conn, v); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("enabled_policy_types") {
		defaultRootID := d.Get("roots.0.id").(string)
		o, n := d.GetChange("enabled_policy_types")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		for _, v := range del {
			if err := disablePolicyType(ctx, conn, awstypes.PolicyType(v), defaultRootID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		for _, v := range add {
			if err := enablePolicyType(ctx, conn, awstypes.PolicyType(v), defaultRootID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("feature_set") {
		input := &organizations.EnableAllFeaturesInput{}

		_, err := conn.EnableAllFeatures(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling all features in Organizations Organization (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceOrganizationRead(ctx, d, meta)...)
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsClient(ctx)

	log.Printf("[INFO] Deleting Organization: %s", d.Id())
	_, err := conn.DeleteOrganization(ctx, &organizations.DeleteOrganizationInput{})

	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Organization (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceOrganizationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func disableServicePrincipal(ctx context.Context, conn *organizations.Client, servicePrincipal string) error {
	input := &organizations.DisableAWSServiceAccessInput{
		ServicePrincipal: aws.String(servicePrincipal),
	}

	_, err := conn.DisableAWSServiceAccess(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling AWS Service Access (%s) in Organizations Organization: %w", servicePrincipal, err)
	}

	return nil
}

func enableServicePrincipal(ctx context.Context, conn *organizations.Client, servicePrincipal string) error {
	input := &organizations.EnableAWSServiceAccessInput{
		ServicePrincipal: aws.String(servicePrincipal),
	}

	_, err := conn.EnableAWSServiceAccess(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling AWS Service Access (%s) in Organizations Organization: %w", servicePrincipal, err)
	}

	return nil
}

func disablePolicyType(ctx context.Context, conn *organizations.Client, policyType awstypes.PolicyType, rootID string) error {
	input := &organizations.DisablePolicyTypeInput{
		PolicyType: policyType,
		RootId:     aws.String(rootID),
	}

	_, err := conn.DisablePolicyType(ctx, input)

	if err != nil {
		return fmt.Errorf("disabling policy type (%s) in Organizations Organization root (%s): %w", policyType, rootID, err)
	}

	if _, err := waitDefaultRootPolicyTypeDisabled(ctx, conn, policyType); err != nil {
		return fmt.Errorf("waiting for Organizations Organization policy (%s) disable: %w", policyType, err)
	}

	return nil
}

func enablePolicyType(ctx context.Context, conn *organizations.Client, policyType awstypes.PolicyType, rootID string) error {
	input := &organizations.EnablePolicyTypeInput{
		PolicyType: policyType,
		RootId:     aws.String(rootID),
	}

	_, err := conn.EnablePolicyType(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling policy type (%s) in Organizations Organization root (%s): %w", policyType, rootID, err)
	}

	if _, err := waitDefaultRootPolicyTypeEnabled(ctx, conn, policyType); err != nil {
		return fmt.Errorf("waiting for Organizations Organization policy (%s) enable: %w", policyType, err)
	}

	return nil
}

func findOrganization(ctx context.Context, conn *organizations.Client) (*awstypes.Organization, error) {
	input := &organizations.DescribeOrganizationInput{}

	output, err := conn.DescribeOrganization(ctx, input)

	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) {
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

func findAccounts(ctx context.Context, conn *organizations.Client, input *organizations.ListAccountsInput) ([]awstypes.Account, error) {
	var output []awstypes.Account

	pages := organizations.NewListAccountsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Accounts...)
	}

	return output, nil
}

func findEnabledServicePrincipalNames(ctx context.Context, conn *organizations.Client) ([]string, error) {
	input := &organizations.ListAWSServiceAccessForOrganizationInput{}

	output, err := findEnabledServicePrincipals(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfslices.ApplyToAll(output, func(v awstypes.EnabledServicePrincipal) string {
		return aws.ToString(v.ServicePrincipal)
	}), nil
}

func findEnabledServicePrincipals(ctx context.Context, conn *organizations.Client, input *organizations.ListAWSServiceAccessForOrganizationInput) ([]awstypes.EnabledServicePrincipal, error) {
	var output []awstypes.EnabledServicePrincipal

	pages := organizations.NewListAWSServiceAccessForOrganizationPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.EnabledServicePrincipals...)
	}

	return output, nil
}

func findRoots(ctx context.Context, conn *organizations.Client, input *organizations.ListRootsInput) ([]awstypes.Root, error) {
	var output []awstypes.Root

	pages := organizations.NewListRootsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Roots...)
	}

	return output, nil
}

func findDefaultRoot(ctx context.Context, conn *organizations.Client) (*awstypes.Root, error) {
	input := &organizations.ListRootsInput{}

	output, err := findRoots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertFirstValueResult(output)
}

func flattenAccounts(apiObjects []awstypes.Account) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrARN:    aws.ToString(apiObject.Arn),
			names.AttrEmail:  aws.ToString(apiObject.Email),
			names.AttrID:     aws.ToString(apiObject.Id),
			names.AttrName:   aws.ToString(apiObject.Name),
			names.AttrStatus: apiObject.Status,
		})
	}

	return tfList
}

func flattenRoots(apiObjects []awstypes.Root) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, r := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrID:   aws.ToString(r.Id),
			names.AttrName: aws.ToString(r.Name),
			names.AttrARN:  aws.ToString(r.Arn),
			"policy_types": flattenRootPolicyTypeSummaries(r.PolicyTypes),
		})
	}

	return tfList
}

func flattenRootPolicyTypeSummaries(apiObjects []awstypes.PolicyTypeSummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrStatus: apiObject.Status,
			names.AttrType:   apiObject.Type,
		})
	}

	return tfList
}

func statusDefaultRootPolicyType(ctx context.Context, conn *organizations.Client, policyType awstypes.PolicyType) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		defaultRoot, err := findDefaultRoot(ctx, conn)

		if err != nil {
			return nil, "", err
		}

		for _, v := range defaultRoot.PolicyTypes {
			if v.Type == policyType {
				return &v, string(v.Status), nil
			}
		}

		return &awstypes.PolicyTypeSummary{}, string(policyTypeStatusDisabled), nil
	}
}

const policyTypeStatusDisabled awstypes.PolicyTypeStatus = "DISABLED"

func waitDefaultRootPolicyTypeDisabled(ctx context.Context, conn *organizations.Client, policyType awstypes.PolicyType) (*awstypes.PolicyTypeSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PolicyTypeStatusEnabled, awstypes.PolicyTypeStatusPendingDisable),
		Target:  enum.Slice(policyTypeStatusDisabled),
		Refresh: statusDefaultRootPolicyType(ctx, conn, policyType),
		Timeout: 5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PolicyTypeSummary); ok {
		return output, err
	}

	return nil, err
}

func waitDefaultRootPolicyTypeEnabled(ctx context.Context, conn *organizations.Client, policyType awstypes.PolicyType) (*awstypes.PolicyTypeSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(policyTypeStatusDisabled, awstypes.PolicyTypeStatusPendingEnable),
		Target:  enum.Slice(awstypes.PolicyTypeStatusEnabled),
		Refresh: statusDefaultRootPolicyType(ctx, conn, policyType),
		Timeout: 5 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PolicyTypeSummary); ok {
		return output, err
	}

	return nil, err
}
