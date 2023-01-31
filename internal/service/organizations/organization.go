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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

const policyTypeStatusDisabled = "DISABLED"

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOrganizationCreate,
		ReadWithoutTimeout:   resourceOrganizationRead,
		UpdateWithoutTimeout: resourceOrganizationUpdate,
		DeleteWithoutTimeout: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("feature_set", func(_ context.Context, old, new, meta interface{}) bool {
				// Only changes from ALL to CONSOLIDATED_BILLING for feature_set should force a new resource
				return old.(string) == organizations.OrganizationFeatureSetAll && new.(string) == organizations.OrganizationFeatureSetConsolidatedBilling
			}),
		),

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"aws_service_access_principals": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
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
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
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
		},
	}
}

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	createOpts := &organizations.CreateOrganizationInput{
		FeatureSet: aws.String(d.Get("feature_set").(string)),
	}
	log.Printf("[DEBUG] Creating Organization: %#v", createOpts)

	resp, err := conn.CreateOrganizationWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating organization: %s", err)
	}

	org := resp.Organization
	d.SetId(aws.StringValue(org.Id))

	awsServiceAccessPrincipals := d.Get("aws_service_access_principals").(*schema.Set).List()
	for _, principalRaw := range awsServiceAccessPrincipals {
		principal := principalRaw.(string)
		input := &organizations.EnableAWSServiceAccessInput{
			ServicePrincipal: aws.String(principal),
		}

		log.Printf("[DEBUG] Enabling AWS Service Access in Organization: %s", input)
		_, err := conn.EnableAWSServiceAccessWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling AWS Service Access (%s) in Organization: %s", principal, err)
		}
	}

	enabledPolicyTypes := d.Get("enabled_policy_types").(*schema.Set).List()

	if len(enabledPolicyTypes) > 0 {
		defaultRoot, err := getOrganizationDefaultRoot(ctx, conn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting AWS Organization (%s) default root: %s", d.Id(), err)
		}

		for _, v := range enabledPolicyTypes {
			enabledPolicyType := v.(string)
			input := &organizations.EnablePolicyTypeInput{
				PolicyType: aws.String(enabledPolicyType),
				RootId:     defaultRoot.Id,
			}

			if _, err := conn.EnablePolicyTypeWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling policy type (%s) in Organization (%s): %s", enabledPolicyType, d.Id(), err)
			}

			if err := waitForOrganizationDefaultRootPolicyTypeEnable(ctx, conn, enabledPolicyType); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for policy type (%s) enabling in Organization (%s): %s", enabledPolicyType, d.Id(), err)
			}
		}
	}

	return append(diags, resourceOrganizationRead(ctx, d, meta)...)
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	log.Printf("[INFO] Reading Organization: %s", d.Id())
	org, err := conn.DescribeOrganizationWithContext(ctx, &organizations.DescribeOrganizationInput{})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
		log.Printf("[WARN] Organization does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Organization: %s", err)
	}

	log.Printf("[INFO] Listing Accounts for Organization: %s", d.Id())
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

	log.Printf("[INFO] Listing Roots for Organization: %s", d.Id())
	var roots []*organizations.Root
	err = conn.ListRootsPagesWithContext(ctx, &organizations.ListRootsInput{}, func(page *organizations.ListRootsOutput, lastPage bool) bool {
		roots = append(roots, page.Roots...)
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing AWS Organization (%s) roots: %s", d.Id(), err)
	}

	if err := d.Set("accounts", flattenAccounts(accounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting accounts: %s", err)
	}

	d.Set("arn", org.Organization.Arn)
	d.Set("feature_set", org.Organization.FeatureSet)
	d.Set("master_account_arn", org.Organization.MasterAccountArn)
	d.Set("master_account_email", org.Organization.MasterAccountEmail)
	d.Set("master_account_id", org.Organization.MasterAccountId)

	if err := d.Set("non_master_accounts", flattenAccounts(nonMasterAccounts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting non_master_accounts: %s", err)
	}

	if err := d.Set("roots", FlattenRoots(roots)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting roots: %s", err)
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

	if err := d.Set("aws_service_access_principals", awsServiceAccessPrincipals); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting aws_service_access_principals: %s", err)
	}

	enabledPolicyTypes := make([]string, 0)

	for _, policyType := range roots[0].PolicyTypes {
		if aws.StringValue(policyType.Status) == organizations.PolicyTypeStatusEnabled {
			enabledPolicyTypes = append(enabledPolicyTypes, aws.StringValue(policyType.Type))
		}
	}

	if err := d.Set("enabled_policy_types", enabledPolicyTypes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting enabled_policy_types: %s", err)
	}

	return diags
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	if d.HasChange("aws_service_access_principals") {
		oldRaw, newRaw := d.GetChange("aws_service_access_principals")
		oldSet := oldRaw.(*schema.Set)
		newSet := newRaw.(*schema.Set)

		for _, disablePrincipalRaw := range oldSet.Difference(newSet).List() {
			principal := disablePrincipalRaw.(string)
			input := &organizations.DisableAWSServiceAccessInput{
				ServicePrincipal: aws.String(principal),
			}

			log.Printf("[DEBUG] Disabling AWS Service Access in Organization: %s", input)
			_, err := conn.DisableAWSServiceAccessWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling AWS Service Access (%s) in Organization: %s", principal, err)
			}
		}

		for _, enablePrincipalRaw := range newSet.Difference(oldSet).List() {
			principal := enablePrincipalRaw.(string)
			input := &organizations.EnableAWSServiceAccessInput{
				ServicePrincipal: aws.String(principal),
			}

			log.Printf("[DEBUG] Enabling AWS Service Access in Organization: %s", input)
			_, err := conn.EnableAWSServiceAccessWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling AWS Service Access (%s) in Organization: %s", principal, err)
			}
		}
	}

	if d.HasChange("enabled_policy_types") {
		defaultRootID := d.Get("roots.0.id").(string)
		o, n := d.GetChange("enabled_policy_types")
		oldSet := o.(*schema.Set)
		newSet := n.(*schema.Set)

		for _, v := range oldSet.Difference(newSet).List() {
			policyType := v.(string)
			input := &organizations.DisablePolicyTypeInput{
				PolicyType: aws.String(policyType),
				RootId:     aws.String(defaultRootID),
			}

			log.Printf("[DEBUG] Disabling Policy Type in Organization: %s", input)
			if _, err := conn.DisablePolicyTypeWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling policy type (%s) in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
			}

			if err := waitForOrganizationDefaultRootPolicyTypeDisable(ctx, conn, policyType); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for policy type (%s) disabling in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
			}
		}

		for _, v := range newSet.Difference(oldSet).List() {
			policyType := v.(string)
			input := &organizations.EnablePolicyTypeInput{
				PolicyType: aws.String(policyType),
				RootId:     aws.String(defaultRootID),
			}

			log.Printf("[DEBUG] Enabling Policy Type in Organization: %s", input)
			if _, err := conn.EnablePolicyTypeWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling policy type (%s) in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
			}

			if err := waitForOrganizationDefaultRootPolicyTypeEnable(ctx, conn, policyType); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for policy type (%s) enabling in Organization (%s) Root (%s): %s", policyType, d.Id(), defaultRootID, err)
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
	conn := meta.(*conns.AWSClient).OrganizationsConn()

	log.Printf("[INFO] Deleting Organization: %s", d.Id())

	_, err := conn.DeleteOrganizationWithContext(ctx, &organizations.DeleteOrganizationInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organization: %s", err)
	}

	return diags
}

func flattenAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(accounts) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"arn":    aws.StringValue(account.Arn),
			"email":  aws.StringValue(account.Email),
			"id":     aws.StringValue(account.Id),
			"name":   aws.StringValue(account.Name),
			"status": aws.StringValue(account.Status),
		})
	}
	return result
}

func FlattenRoots(roots []*organizations.Root) []map[string]interface{} {
	if len(roots) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, r := range roots {
		result = append(result, map[string]interface{}{
			"id":           aws.StringValue(r.Id),
			"name":         aws.StringValue(r.Name),
			"arn":          aws.StringValue(r.Arn),
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
			"status": aws.StringValue(s.Status),
			"type":   aws.StringValue(s.Type),
		})
	}
	return result
}

func getOrganizationDefaultRoot(ctx context.Context, conn *organizations.Organizations) (*organizations.Root, error) {
	var roots []*organizations.Root

	err := conn.ListRootsPagesWithContext(ctx, &organizations.ListRootsInput{}, func(page *organizations.ListRootsOutput, lastPage bool) bool {
		roots = append(roots, page.Roots...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("no roots found")
	}

	return roots[0], nil
}

func getOrganizationDefaultRootPolicyTypeRefreshFunc(ctx context.Context, conn *organizations.Organizations, policyType string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		defaultRoot, err := getOrganizationDefaultRoot(ctx, conn)

		if err != nil {
			return nil, "", fmt.Errorf("error getting default root: %s", err)
		}

		for _, pt := range defaultRoot.PolicyTypes {
			if aws.StringValue(pt.Type) == policyType {
				return pt, aws.StringValue(pt.Status), nil
			}
		}

		return &organizations.PolicyTypeSummary{}, policyTypeStatusDisabled, nil
	}
}

func waitForOrganizationDefaultRootPolicyTypeDisable(ctx context.Context, conn *organizations.Organizations, policyType string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			organizations.PolicyTypeStatusEnabled,
			organizations.PolicyTypeStatusPendingDisable,
		},
		Target:  []string{policyTypeStatusDisabled},
		Refresh: getOrganizationDefaultRootPolicyTypeRefreshFunc(ctx, conn, policyType),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForOrganizationDefaultRootPolicyTypeEnable(ctx context.Context, conn *organizations.Organizations, policyType string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			policyTypeStatusDisabled,
			organizations.PolicyTypeStatusPendingEnable,
		},
		Target:  []string{organizations.PolicyTypeStatusEnabled},
		Refresh: getOrganizationDefaultRootPolicyTypeRefreshFunc(ctx, conn, policyType),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
