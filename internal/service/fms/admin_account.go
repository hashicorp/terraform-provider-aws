// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package fms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fms_admin_account", name="Admin Account")
// @Region(global=true)
func resourceAdminAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAdminAccountCreate,
		ReadWithoutTimeout:   resourceAdminAccountRead,
		UpdateWithoutTimeout: resourceAdminAccountUpdate,
		DeleteWithoutTimeout: resourceAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"admin_scope": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_scope": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"accounts": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"all_accounts_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"exclude_specified_accounts": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"organizational_unit_scope": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"organizational_units": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"all_organizational_units_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"exclude_specified_organizational_units": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"region_scope": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"regions": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"all_regions_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"policy_type_scope": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.StringInSlice([]string{
												"WAF",
												"WAFV2",
												"SHIELD_ADVANCED",
												"SECURITY_GROUPS_COMMON",
												"SECURITY_GROUPS_CONTENT_AUDIT",
												"SECURITY_GROUPS_USAGE_AUDIT",
												"NETWORK_FIREWALL",
												"DNS_FIREWALL",
												"THIRD_PARTY_FIREWALL",
												"IMPORT_NETWORK_FIREWALL",
												"NETWORK_ACL_COMMON",
											}, false),
										},
									},
									"all_policy_types_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
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

func resourceAdminAccountCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}

	// Ensure there is not an existing FMS Admin Account.
	output, err := findAdminAccount(ctx, conn)

	switch {
	case retry.NotFound(err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading FMS Admin Account (%s): %s", accountID, err)
	default:
		return sdkdiag.AppendErrorf(diags, "FMS Admin Account (%s) already associated: import this Terraform resource to manage", aws.ToString(output.AdminAccount))
	}

	input := &fms.PutAdminAccountInput{
		AdminAccount: aws.String(accountID),
	}

	if v, ok := d.GetOk("admin_scope"); ok && len(v.([]interface{})) > 0 {
		input.AdminScope = expandAdminScope(v.([]interface{}))
	}

	_, err = conn.PutAdminAccount(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FMS Admin Account (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	if _, err := waitAdminAccountCreated(ctx, conn, accountID, input.AdminScope, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAdminAccountRead(ctx, d, meta)...)
}

func resourceAdminAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	output, err := findAdminAccount(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] FMS Admin Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FMS Admin Account (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, output.AdminAccount)

	// Get admin scope
	scopeOutput, err := conn.GetAdminScope(ctx, &fms.GetAdminScopeInput{
		AdminAccount: output.AdminAccount,
	})

	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return sdkdiag.AppendErrorf(diags, "reading FMS Admin Account (%s) scope: %s", d.Id(), err)
	}

	if scopeOutput != nil && scopeOutput.AdminScope != nil {
		if err := d.Set("admin_scope", flattenAdminScope(scopeOutput.AdminScope)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting admin_scope: %s", err)
		}
	} else {
		d.Set("admin_scope", nil)
	}

	return diags
}

func resourceAdminAccountUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	if d.HasChange("admin_scope") {
		input := &fms.PutAdminAccountInput{
			AdminAccount: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("admin_scope"); ok && len(v.([]interface{})) > 0 {
			input.AdminScope = expandAdminScope(v.([]interface{}))
		}

		_, err := conn.PutAdminAccount(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FMS Admin Account (%s): %s", d.Id(), err)
		}

		if _, err := waitAdminAccountUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceAdminAccountRead(ctx, d, meta)...)
}

func resourceAdminAccountDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FMSClient(ctx)

	_, err := conn.DisassociateAdminAccount(ctx, &fms.DisassociateAdminAccountInput{})

	if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "No default admin could be found for account") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating FMS Admin Account (%s): %s", d.Id(), err)
	}

	if _, err := waitAdminAccountDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FMS Admin Account (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAdminAccount(ctx context.Context, conn *fms.Client) (*fms.GetAdminAccountOutput, error) {
	input := &fms.GetAdminAccountInput{}

	output, err := conn.GetAdminAccount(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if status := output.RoleStatus; status == awstypes.AccountRoleStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func statusPutAdminAccount(conn *fms.Client, accountID string, adminScope *awstypes.AdminScope) retry.StateRefreshFunc {
	// This is all wrapped in a StateRefreshFunc since PutAdminAccount returns
	// success even though it failed if called too quickly after creating an Organization.
	return func(ctx context.Context) (any, string, error) {
		input := &fms.PutAdminAccountInput{
			AdminAccount: aws.String(accountID),
			AdminScope:   adminScope,
		}

		_, err := conn.PutAdminAccount(ctx, input)

		if err != nil {
			return nil, "", err
		}

		output, err := conn.GetAdminAccount(ctx, &fms.GetAdminAccountInput{})

		// FMS returns an AccessDeniedException if no account is associated,
		// but does not define this in its error codes.
		if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "is not currently delegated by AWS FM") {
			return nil, "", nil
		}

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if aws.ToString(output.AdminAccount) != accountID {
			return nil, "", nil
		}

		return output, string(output.RoleStatus), err
	}
}

func statusAdminAccount(conn *fms.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAdminAccount(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.RoleStatus), nil
	}
}

func waitAdminAccountCreated(ctx context.Context, conn *fms.Client, accountID string, adminScope *awstypes.AdminScope, timeout time.Duration) (*fms.GetAdminAccountOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.AccountRoleStatusDeleted, // Recreating association can return this status.
			awstypes.AccountRoleStatusCreating,
		),
		Target:  enum.Slice(awstypes.AccountRoleStatusReady),
		Refresh: statusPutAdminAccount(conn, accountID, adminScope),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fms.GetAdminAccountOutput); ok {
		return output, err
	}

	return nil, err
}

func waitAdminAccountUpdated(ctx context.Context, conn *fms.Client, accountID string, timeout time.Duration) (*fms.GetAdminAccountOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.AccountRoleStatusCreating,
		),
		Target:  enum.Slice(awstypes.AccountRoleStatusReady),
		Refresh: statusAdminAccount(conn),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fms.GetAdminAccountOutput); ok {
		return output, err
	}

	return nil, err
}

func waitAdminAccountDeleted(ctx context.Context, conn *fms.Client, timeout time.Duration) (*fms.GetAdminAccountOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.AccountRoleStatusDeleting,
			awstypes.AccountRoleStatusPendingDeletion,
			awstypes.AccountRoleStatusReady,
		),
		Target:  []string{},
		Refresh: statusAdminAccount(conn),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fms.GetAdminAccountOutput); ok {
		return output, err
	}

	return nil, err
}

func expandAdminScope(tfList []interface{}) *awstypes.AdminScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AdminScope{}

	if v, ok := tfMap["account_scope"].([]interface{}); ok && len(v) > 0 {
		apiObject.AccountScope = expandAccountScope(v)
	}

	if v, ok := tfMap["organizational_unit_scope"].([]interface{}); ok && len(v) > 0 {
		apiObject.OrganizationalUnitScope = expandOrganizationalUnitScope(v)
	}

	if v, ok := tfMap["region_scope"].([]interface{}); ok && len(v) > 0 {
		apiObject.RegionScope = expandRegionScope(v)
	}

	if v, ok := tfMap["policy_type_scope"].([]interface{}); ok && len(v) > 0 {
		apiObject.PolicyTypeScope = expandPolicyTypeScope(v)
	}

	return apiObject
}

func expandAccountScope(tfList []interface{}) *awstypes.AccountScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AccountScope{}

	if v, ok := tfMap["accounts"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Accounts = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["all_accounts_enabled"].(bool); ok {
		apiObject.AllAccountsEnabled = v
	}

	if v, ok := tfMap["exclude_specified_accounts"].(bool); ok {
		apiObject.ExcludeSpecifiedAccounts = v
	}

	return apiObject
}

func expandOrganizationalUnitScope(tfList []interface{}) *awstypes.OrganizationalUnitScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.OrganizationalUnitScope{}

	if v, ok := tfMap["organizational_units"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.OrganizationalUnits = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["all_organizational_units_enabled"].(bool); ok {
		apiObject.AllOrganizationalUnitsEnabled = v
	}

	if v, ok := tfMap["exclude_specified_organizational_units"].(bool); ok {
		apiObject.ExcludeSpecifiedOrganizationalUnits = v
	}

	return apiObject
}

func expandRegionScope(tfList []interface{}) *awstypes.RegionScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.RegionScope{}

	if v, ok := tfMap["regions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Regions = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["all_regions_enabled"].(bool); ok {
		apiObject.AllRegionsEnabled = v
	}

	return apiObject
}

func expandPolicyTypeScope(tfList []interface{}) *awstypes.PolicyTypeScope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.PolicyTypeScope{}

	if v, ok := tfMap["policy_types"].(*schema.Set); ok && v.Len() > 0 {
		policyTypes := make([]awstypes.SecurityServiceType, 0, v.Len())
		for _, item := range v.List() {
			policyTypes = append(policyTypes, awstypes.SecurityServiceType(item.(string)))
		}
		apiObject.PolicyTypes = policyTypes
	}

	if v, ok := tfMap["all_policy_types_enabled"].(bool); ok {
		apiObject.AllPolicyTypesEnabled = v
	}

	return apiObject
}

func flattenAdminScope(apiObject *awstypes.AdminScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccountScope; v != nil {
		tfMap["account_scope"] = flattenAccountScope(v)
	}

	if v := apiObject.OrganizationalUnitScope; v != nil {
		tfMap["organizational_unit_scope"] = flattenOrganizationalUnitScope(v)
	}

	if v := apiObject.RegionScope; v != nil {
		tfMap["region_scope"] = flattenRegionScope(v)
	}

	if v := apiObject.PolicyTypeScope; v != nil {
		tfMap["policy_type_scope"] = flattenPolicyTypeScope(v)
	}

	return []interface{}{tfMap}
}

func flattenAccountScope(apiObject *awstypes.AccountScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Accounts; v != nil {
		tfMap["accounts"] = flex.FlattenStringValueSet(v)
	}

	tfMap["all_accounts_enabled"] = apiObject.AllAccountsEnabled
	tfMap["exclude_specified_accounts"] = apiObject.ExcludeSpecifiedAccounts

	return []interface{}{tfMap}
}

func flattenOrganizationalUnitScope(apiObject *awstypes.OrganizationalUnitScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OrganizationalUnits; v != nil {
		tfMap["organizational_units"] = flex.FlattenStringValueSet(v)
	}

	tfMap["all_organizational_units_enabled"] = apiObject.AllOrganizationalUnitsEnabled
	tfMap["exclude_specified_organizational_units"] = apiObject.ExcludeSpecifiedOrganizationalUnits

	return []interface{}{tfMap}
}

func flattenRegionScope(apiObject *awstypes.RegionScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Regions; v != nil {
		tfMap["regions"] = flex.FlattenStringValueSet(v)
	}

	tfMap["all_regions_enabled"] = apiObject.AllRegionsEnabled

	return []interface{}{tfMap}
}

func flattenPolicyTypeScope(apiObject *awstypes.PolicyTypeScope) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PolicyTypes; v != nil {
		policyTypes := make([]string, 0, len(v))
		for _, pt := range v {
			policyTypes = append(policyTypes, string(pt))
		}
		tfMap["policy_types"] = policyTypes
	}

	tfMap["all_policy_types_enabled"] = apiObject.AllPolicyTypesEnabled

	return []interface{}{tfMap}
}
