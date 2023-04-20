package inspector2

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/go-multierror"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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
						return fmt.Errorf(`"account_ids" can contain either the administrator account or one or more member accounts. Contains the administrator account and %d other accounts`, l-1)
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
	conn := meta.(*conns.AWSClient).Inspector2Client()

	accountIDs := getAccountIDs(d)

	in := &inspector2.EnableInput{
		AccountIds:    accountIDs,
		ResourceTypes: flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set)),
		ClientToken:   aws.String(sdkid.UniqueId()),
	}

	id := enablerID(accountIDs, flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set)))

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

		var errs *multierror.Error
		for _, acct := range out.FailedAccounts {
			errs = multierror.Append(errs, newFailedAccountError(acct))
		}
		err = failedAccountsError(*errs)

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
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameEnabler, id, err)...)
	}

	d.SetId(id)

	if err := waitEnabled(ctx, conn, accountIDs, d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionWaitingForCreation, ResNameEnabler, d.Id(), err)...)
	}

	return append(diags, resourceEnablerRead(ctx, d, meta)...)
}

func resourceEnablerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	accountIDs, _, err := parseEnablerID(d.Id())
	if err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)...)
	}

	s, err := AccountStatuses(ctx, conn, accountIDs)
	if err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)...)
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Enabler (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var enabledAccounts []string
	for k, a := range s {
		if a.Status == types.StatusEnabled {
			enabledAccounts = append(enabledAccounts, k)
		}
	}

	accountStatuses := maps.Values(s)
	x := accountStatuses[0]
	var resourceTypes []types.ResourceScanType
	for k, a := range x.ResourceStatuses {
		if a == types.StatusEnabled {
			resourceTypes = append(resourceTypes, k)
		}
	}

	if err := d.Set("account_ids", flex.FlattenStringValueSet(enabledAccounts)); err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)...)
	}
	if err := d.Set("resource_types", flex.FlattenStringValueSet(enum.Slice(resourceTypes...))); err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)...)
	}

	return diags
}

func resourceEnablerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Inspector2Client()

	var enable, disable []types.ResourceScanType
	if d.HasChange("resource_types") {
		enable = flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set))
		for _, v := range types.ResourceScanType("").Values() {
			if !slices.Contains(enable, v) {
				disable = append(disable, v)
			}
		}
	}

	accountIDs := getAccountIDs(d)

	id := enablerID(accountIDs, flex.ExpandStringyValueSet[types.ResourceScanType](d.Get("resource_types").(*schema.Set)))

	if len(enable) > 0 {
		in := &inspector2.EnableInput{
			AccountIds:    accountIDs,
			ResourceTypes: enable,
			ClientToken:   aws.String(sdkid.UniqueId()),
		}

		_, err := conn.Enable(ctx, in)
		if err != nil {
			return append(diags, create.DiagError(names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, err)...)
		}

		if err := waitEnabled(ctx, conn, accountIDs, d.Timeout(schema.TimeoutCreate)); err != nil {
			return append(diags, create.DiagError(names.Inspector2, create.ErrActionWaitingForUpdate, ResNameEnabler, d.Id(), err)...)
		}
	}

	if len(disable) > 0 {
		in := &inspector2.DisableInput{
			AccountIds:    accountIDs,
			ResourceTypes: disable,
		}
		_, err := conn.Disable(ctx, in)
		if err != nil {
			return append(diags, create.DiagError(names.Inspector2, create.ErrActionUpdating, ResNameEnabler, id, err)...)
		}
	}

	d.SetId(id)

	if err := waitEnabled(ctx, conn, accountIDs, d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionWaitingForUpdate, ResNameEnabler, d.Id(), err)...)
	}

	return append(diags, resourceEnablerRead(ctx, d, meta)...)
}

func resourceEnablerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient)
	conn := client.Inspector2Client()

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

		diags = append(diags, disableAccount(ctx, conn, d, members)...)
		if diags.HasError() {
			return diags
		}
	} else if admin {
		diags = append(diags, disableAccount(ctx, conn, d, []string{client.AccountID})...)
	}

	return diags
}

func disableAccount(ctx context.Context, conn *inspector2.Client, d *schema.ResourceData, accountIDs []string) diag.Diagnostics {
	var diags diag.Diagnostics
	in := &inspector2.DisableInput{
		AccountIds:    accountIDs,
		ResourceTypes: types.ResourceScanType("").Values(),
	}

	var out *inspector2.DisableOutput
	err := tfresource.Retry(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error
		out, err = conn.Disable(ctx, in)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		if out == nil {
			return retry.RetryableError(tfresource.NewEmptyResultError(nil))
		}

		if len(out.FailedAccounts) == 0 {
			return nil
		}

		var errs *multierror.Error
		for _, acct := range out.FailedAccounts {
			errs = multierror.Append(errs, newFailedAccountError(acct))
		}
		err = failedAccountsError(*errs)

		return retry.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		out, err = conn.Disable(ctx, in)
	}
	if err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionDeleting, ResNameEnabler, d.Id(), err)...)
	}

	if err := waitDisabled(ctx, conn, accountIDs, d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.Inspector2, create.ErrActionWaitingForDeletion, ResNameEnabler, d.Id(), err)...)
	}

	return diags
}

type failedAccountsError multierror.Error

func (e failedAccountsError) Error() string {
	m := multierror.Error(e)
	return m.Error()
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
	StatusComplete   = "COMPLETE"
	StatusInProgress = "IN_PROGRESS"
)

func waitEnabled(ctx context.Context, conn *inspector2.Client, accountIDs []string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusInProgress},
		Target:  []string{StatusComplete},
		Refresh: statusEnablerAccountAndResourceTypes(ctx, conn, accountIDs),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitDisabled(ctx context.Context, conn *inspector2.Client, accountIDs []string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{StatusInProgress},
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

		if tfslices.All(maps.Values(st), accountStatusEquals(types.StatusDisabled)) {
			return nil, "", nil
		}

		if tfslices.Any(maps.Values(st), func(v AccountResourceStatus) bool {
			if slices.Contains(pendingStates, v.Status) {
				return true
			}
			if tfslices.Any(maps.Values(v.ResourceStatuses), func(v types.Status) bool {
				return slices.Contains(pendingStates, v)
			}) {
				return true
			}
			if v.Status == types.StatusEnabled && tfslices.All(maps.Values(v.ResourceStatuses), tfslices.FilterEquals(types.StatusDisabled)) {
				return true
			}
			return false
		}) {
			return st, StatusInProgress, nil
		}

		return st, StatusComplete, nil
	}
}

// statusEnablerAccount checks only the status of Inspector for the account as a whole
func statusEnablerAccount(ctx context.Context, conn *inspector2.Client, accountIDs []string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		st, err := AccountStatuses(ctx, conn, accountIDs)
		if err != nil {
			return nil, "", err
		}

		if tfslices.All(maps.Values(st), accountStatusEquals(types.StatusDisabled)) {
			return nil, "", nil
		}
		return st, StatusInProgress, nil
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
			err = multierror.Append(err, e)
			continue
		}
		for k, v := range m {
			status.ResourceStatuses[types.ResourceScanType(strings.ToUpper(k))] = v.Status
		}
		results[aws.ToString(a.AccountId)] = status
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
