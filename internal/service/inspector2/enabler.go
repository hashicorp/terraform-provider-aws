// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/mapstructure"
)

// @SDKResource("aws_inspector2_enabler")
func ResourceEnabler() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnablerCreate,
		ReadWithoutTimeout:   resourceEnablerRead,
		UpdateWithoutTimeout: resourceEnablerUpdate,
		DeleteWithoutTimeout: resourceEnablerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_ids": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			"resource_types": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.ResourceScanType](),
				},
			},
		},

		CustomizeDiff: customdiff.All(
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				accountIDs := getAccountIDs(d)
				if l := len(accountIDs); l > 1 {
					client := meta.(*conns.AWSClient)

					if slices.Contains(accountIDs, client.AccountID) {
						return fmt.Errorf(`"account_ids" can contain either the administrator account or one or more member accounts. Contains %v`, accountIDs)
					}
				}
				return nil
			},
		),
	}
}

type resourceGetter interface {
	Get(key string) any
}

func getAccountIDs(d resourceGetter) []string {
	return flex.ExpandStringValueSet(d.Get("account_ids").(*schema.Set))
}

const (
	ResNameEnabler = "Enabler"
)

func resourceEnablerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	accountIDs := getAccountIDs(d)

	typeEnable := flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set))

	in := &inspector2.EnableInput{
		AccountIds:    accountIDs,
		ResourceTypes: typeEnable,
		ClientToken:   aws.String(sdkid.UniqueId()),
	}

	id := enablerID(accountIDs, typeEnable)

	var out *inspector2.EnableOutput
	err := tfresource.Retry(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error
		out, err = conn.Enable(ctx, in)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if out == nil {
			return retry.RetryableError(tfresource.NewEmptyResultError(nil))
		}

		if len(out.FailedAccounts) == 0 {
			return nil
		}

		var errs []error
		for _, acct := range out.FailedAccounts {
			errs = append(errs, newFailedAccountError(acct))
		}
		err = errors.Join(errs...)

		if tfslices.All(out.FailedAccounts, func(acct types.FailedAccount) bool {
			switch acct.ErrorCode {
			case types.ErrorCodeAccessDenied, // Account membership not propagated
				types.ErrorCodeSsmThrottled,
				types.ErrorCodeEventbridgeThrottled,
				types.ErrorCodeEnableInProgress,
				types.ErrorCodeDisableInProgress,
				types.ErrorCodeSuspendInProgress:
				return true
			}
			return false
		}) {
			return retry.RetryableError(err)
		}

		return retry.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		out, err = conn.Enable(ctx, in)
	}
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionCreating, ResNameEnabler, id, err)
	}

	d.SetId(id)

	st, err := waitEnabled(ctx, conn, accountIDs, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForCreation, ResNameEnabler, d.Id(), err)
	}

	var disableAccountIDs []string
	for acctID, acctStatus := range st {
		resourceStatuses := acctStatus.ResourceStatuses
		for _, resourceType := range typeEnable {
			delete(resourceStatuses, resourceType)
		}
		if len(resourceStatuses) > 0 {
			disableAccountIDs = append(disableAccountIDs, acctID)
			in := &inspector2.DisableInput{
				AccountIds:    []string{acctID},
				ResourceTypes: tfmaps.Keys(resourceStatuses),
			}

			_, err := conn.Disable(ctx, in)
			if err != nil {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, err)
			}
		}
	}

	if len(disableAccountIDs) > 0 {
		if _, err := waitEnabled(ctx, conn, disableAccountIDs, d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForUpdate, ResNameEnabler, id, err)
		}
	}

	return append(diags, resourceEnablerRead(ctx, d, meta)...)
}

func resourceEnablerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	accountIDs, _, err := parseEnablerID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)
	}

	s, err := AccountStatuses(ctx, conn, accountIDs)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Enabler (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)
	}

	var enabledAccounts []string
	for k, a := range s {
		if a.Status == types.StatusEnabled {
			enabledAccounts = append(enabledAccounts, k)
		}
	}

	accountStatuses := tfmaps.Values(s)
	accountStatus := accountStatuses[0]
	var resourceTypes []types.ResourceScanType
	for k, a := range accountStatus.ResourceStatuses {
		if a == types.StatusEnabled {
			resourceTypes = append(resourceTypes, k)
		}
	}

	if err := d.Set("account_ids", flex.FlattenStringValueSet(enabledAccounts)); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)
	}
	if err := d.Set("resource_types", flex.FlattenStringValueSet(enum.Slice(resourceTypes...))); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)
	}

	return diags
}

func resourceEnablerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client(ctx)

	typeEnable := flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set))
	var typeDisable []types.ResourceScanType
	if d.HasChange("resource_types") {
		for _, v := range types.ResourceScanType("").Values() {
			if !slices.Contains(typeEnable, v) {
				typeDisable = append(typeDisable, v)
			}
		}
	}

	var acctEnable, acctRemove []string
	acctEnable = flex.ExpandStringValueSet(d.Get("account_ids").(*schema.Set))
	if d.HasChange("account_ids") {
		o, n := d.GetChange("account_ids")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		acctRemove = flex.ExpandStringValueSet(os.Difference(ns))
	}

	id := enablerID(getAccountIDs(d), flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set)))

	d.SetId(id)

	if len(acctEnable) > 0 {
		if len(typeEnable) > 0 {
			in := &inspector2.EnableInput{
				AccountIds:    acctEnable,
				ResourceTypes: typeEnable,
				ClientToken:   aws.String(sdkid.UniqueId()),
			}

			out, err := conn.Enable(ctx, in)
			if err != nil {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, err)
			}

			if out == nil {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, tfresource.NewEmptyResultError(nil))
			}

			if len(out.FailedAccounts) > 0 {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, errors.New("failed accounts"))
			}

			if _, err := waitEnabled(ctx, conn, acctEnable, d.Timeout(schema.TimeoutCreate)); err != nil {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForUpdate, ResNameEnabler, id, err)
			}
		}

		if len(typeDisable) > 0 {
			in := &inspector2.DisableInput{
				AccountIds:    acctEnable,
				ResourceTypes: typeDisable,
			}

			_, err := conn.Disable(ctx, in)
			if err != nil {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, err)
			}

			if _, err := waitEnabled(ctx, conn, acctEnable, d.Timeout(schema.TimeoutCreate)); err != nil {
				return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForUpdate, ResNameEnabler, id, err)
			}
		}
	}

	if len(acctRemove) > 0 {
		diags = append(diags, disableAccounts(ctx, conn, d, acctRemove)...)
	}

	return append(diags, resourceEnablerRead(ctx, d, meta)...)
}

func resourceEnablerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient)
	conn := client.Inspector2Client(ctx)

	accountIDs := getAccountIDs(d)
	admin := slices.Contains(accountIDs, client.AccountID)
	members := tfslices.Filter(accountIDs, func(s string) bool {
		return s != client.AccountID
	})
	if len(members) > 0 {
		// Catch legacy case mixing admin account and member accounts
		if admin {
			diags = append(diags, errs.NewWarningDiagnostic(
				"Inconsistent Amazon Inspector State",
				"The Organization Administrator Account cannot be deleted while there are associated member accounts. Disabling Inspector for the member accounts. ",
			))
		}

		diags = append(diags, disableAccounts(ctx, conn, d, members)...)
		if diags.HasError() {
			return diags
		}
	} else if admin {
		diags = append(diags, disableAccounts(ctx, conn, d, []string{client.AccountID})...)
	}

	return diags
}

func disableAccounts(ctx context.Context, conn *inspector2.Client, d *schema.ResourceData, accountIDs []string) diag.Diagnostics {
	var diags diag.Diagnostics
	in := &inspector2.DisableInput{
		AccountIds:    accountIDs,
		ResourceTypes: types.ResourceScanType("").Values(),
	}

	out, err := conn.Disable(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionDeleting, ResNameEnabler, d.Id(), err)
	}
	if out == nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionDeleting, ResNameEnabler, d.Id(), tfresource.NewEmptyResultError(nil))
	}

	var errs []error
	for _, acct := range out.FailedAccounts {
		if acct.ErrorCode != types.ErrorCodeAccessDenied {
			errs = append(errs, newFailedAccountError(acct))
		}
	}
	err = errors.Join(errs...)

	if err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionDeleting, ResNameEnabler, d.Id(), err)
	}

	if err := waitDisabled(ctx, conn, accountIDs, d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.Inspector2, create.ErrActionWaitingForDeletion, ResNameEnabler, d.Id(), err)
	}

	return diags
}

type failedAccountError struct {
	accountID string
	code      types.ErrorCode
	message   string
}

func newFailedAccountError(a types.FailedAccount) error {
	return &failedAccountError{
		accountID: aws.ToString(a.AccountId),
		code:      a.ErrorCode,
		message:   aws.ToString(a.ErrorMessage),
	}
}

func (e failedAccountError) Error() string {
	return fmt.Sprintf("account %s: %s: %s", e.accountID, e.code, e.message)
}

const (
	statusComplete   = "COMPLETE"
	statusInProgress = "IN_PROGRESS"
)

func waitEnabled(ctx context.Context, conn *inspector2.Client, accountIDs []string, timeout time.Duration) (map[string]AccountResourceStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusInProgress},
		Target:  []string{statusComplete},
		Refresh: statusEnablerAccountAndResourceTypes(ctx, conn, accountIDs),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(map[string]AccountResourceStatus); ok {
		return output, err
	}

	return nil, err
}

func waitDisabled(ctx context.Context, conn *inspector2.Client, accountIDs []string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusInProgress},
		Target:  []string{},
		Refresh: statusEnablerAccount(ctx, conn, accountIDs),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

var (
	// terminalStates = []types.Status{
	// 	types.StatusEnabled,
	// 	types.StatusDisabled,
	// 	types.StatusSuspended,
	// }
	pendingStates = []types.Status{
		types.StatusEnabling,
		types.StatusDisabling,
		types.StatusSuspending,
	}
)

// statusEnablerAccountAndResourceTypes checks the status of Inspector for the account and resource types
func statusEnablerAccountAndResourceTypes(ctx context.Context, conn *inspector2.Client, accountIDs []string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		st, err := AccountStatuses(ctx, conn, accountIDs)
		if err != nil {
			return nil, "", err
		}

		if tfslices.All(tfmaps.Values(st), accountStatusEquals(types.StatusDisabled)) {
			return nil, "", nil
		}

		if tfslices.Any(tfmaps.Values(st), func(v AccountResourceStatus) bool {
			if slices.Contains(pendingStates, v.Status) {
				return true
			}
			if tfslices.Any(tfmaps.Values(v.ResourceStatuses), func(v types.Status) bool {
				return slices.Contains(pendingStates, v)
			}) {
				return true
			}
			if v.Status == types.StatusEnabled && tfslices.All(tfmaps.Values(v.ResourceStatuses), tfslices.PredicateEquals(types.StatusDisabled)) {
				return true
			}
			return false
		}) {
			return st, statusInProgress, nil
		}

		return st, statusComplete, nil
	}
}

// statusEnablerAccount checks only the status of Inspector for the account as a whole.
// It is only used for deletion, so the non-error states are in-progress or not-found
func statusEnablerAccount(ctx context.Context, conn *inspector2.Client, accountIDs []string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		st, err := AccountStatuses(ctx, conn, accountIDs)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		if tfslices.All(tfmaps.Values(st), accountStatusEquals(types.StatusDisabled)) {
			return nil, "", nil
		}

		return st, statusInProgress, nil
	}
}

type AccountResourceStatus struct {
	Status           types.Status
	ResourceStatuses map[types.ResourceScanType]types.Status
}

func accountStatusEquals(s types.Status) func(AccountResourceStatus) bool {
	return func(v AccountResourceStatus) bool {
		return v.Status == s
	}
}

func AccountStatuses(ctx context.Context, conn *inspector2.Client, accountIDs []string) (map[string]AccountResourceStatus, error) {
	in := &inspector2.BatchGetAccountStatusInput{
		AccountIds: accountIDs,
	}
	out, err := conn.BatchGetAccountStatus(ctx, in)
	if err != nil {
		return nil, err
	}

	var errs []error
	results := make(map[string]AccountResourceStatus, len(out.Accounts))
	for _, a := range out.Accounts {
		if a.AccountId == nil || a.State == nil {
			continue
		}
		status := AccountResourceStatus{
			Status:           a.State.Status,
			ResourceStatuses: make(map[types.ResourceScanType]types.Status, len(enum.Values[types.ResourceScanType]())),
		}
		var m map[string]*types.State
		e := mapstructure.Decode(a.ResourceState, &m)
		if e != nil {
			errs = append(errs, e)
			continue
		}
		for k, v := range m {
			if k == "LambdaCode" {
				k = "LAMBDA_CODE"
			}
			status.ResourceStatuses[types.ResourceScanType(strings.ToUpper(k))] = v.Status
		}
		results[aws.ToString(a.AccountId)] = status
	}
	err = errors.Join(errs...)

	if err != nil {
		return results, err
	}

	if len(results) == 0 {
		return results, &retry.NotFoundError{}
	}

	return results, err
}

func enablerID(accountIDs []string, types []types.ResourceScanType) string {
	sort.Strings(accountIDs)
	t := enum.Slice(types...)
	sort.Strings(t)
	return fmt.Sprintf("%s-%s", strings.Join(accountIDs, ":"), strings.Join(t, ":"))
}

func parseEnablerID(id string) ([]string, []string, error) {
	parts := strings.Split(id, "-")

	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("unexpected ID format (%s), expected <account-ids (':' separated)>-<types (':' separated)>", id)
	}

	return strings.Split(parts[0], ":"), strings.Split(parts[1], ":"), nil
}
