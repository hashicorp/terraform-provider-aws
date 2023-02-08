package inspector2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceEnabler() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnablerCreate,
		ReadWithoutTimeout:   resourceEnablerRead,
		UpdateWithoutTimeout: resourceEnablerCreate,
		DeleteWithoutTimeout: resourceEnablerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
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
	}
}

const (
	ResNameEnabler = "Enabler"
)

func resourceEnablerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Client()

	in := &inspector2.EnableInput{
		AccountIds:    flex.ExpandStringValueSet(d.Get("account_ids").(*schema.Set)),
		ResourceTypes: expandResourceScanTypes(flex.ExpandStringValueSet(d.Get("resource_types").(*schema.Set))),
		ClientToken:   aws.String(resource.UniqueId()),
	}

	id := EnablerID(in.AccountIds, flex.ExpandStringValueSet(d.Get("resource_types").(*schema.Set)))

	out, err := conn.Enable(ctx, in)
	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameEnabler, id, err)
	}

	if out == nil {
		return create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameEnabler, id, errors.New("empty output"))
	}

	d.SetId(id)

	if err := waitEnabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionWaitingForCreation, ResNameEnabler, d.Id(), err)
	}

	return resourceEnablerRead(ctx, d, meta)
}

func resourceEnablerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Client()

	s, err := FindAccountStatuses(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)
	}

	// probably a NotFound is not possible but including for linting/completeness
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Enabler (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var enabledAccounts []string
	for _, a := range s {
		if a.Status == string(types.StatusEnabled) {
			enabledAccounts = append(enabledAccounts, a.AccountID)
		}
	}

	if err := d.Set("account_ids", flex.FlattenStringValueSet(enabledAccounts)); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionReading, ResNameEnabler, d.Id(), err)
	}

	return nil
}

func resourceEnablerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Client()

	in := &inspector2.DisableInput{
		AccountIds:    flex.ExpandStringValueSet(d.Get("account_ids").(*schema.Set)),
		ResourceTypes: expandResourceScanTypes(flex.ExpandStringValueSet(d.Get("resource_types").(*schema.Set))),
	}

	_, err := conn.Disable(ctx, in)
	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionDeleting, ResNameEnabler, d.Id(), err)
	}

	if err := waitDisabled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionWaitingForDeletion, ResNameEnabler, d.Id(), err)
	}

	return nil
}

const (
	StatusDisabledEnabled = string(types.StatusDisabled + types.StatusEnabled)
	StatusInProgress      = "IN_PROGRESS"
)

func waitEnabled(ctx context.Context, conn *inspector2.Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   append(enum.Slice(types.StatusEnabling, types.StatusDisabled), StatusDisabledEnabled, StatusInProgress),
		Target:                    enum.Slice(types.StatusEnabled),
		Refresh:                   statusEnable(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     10 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitDisabled(ctx context.Context, conn *inspector2.Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: append(enum.Slice(types.StatusDisabling, types.StatusEnabled), StatusDisabledEnabled, StatusInProgress),
		Target:  enum.Slice(types.StatusDisabled),
		Refresh: statusEnable(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusEnable(ctx context.Context, conn *inspector2.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		st, err := FindAccountStatuses(ctx, conn, id)

		if errs.Contains(err, string(types.ErrorCodeAlreadyEnabled)) {
			return 42, string(types.StatusEnabled), nil
		}

		if errs.Contains(err, string(types.ErrorCodeEnableInProgress)) {
			return nil, StatusInProgress, nil
		}

		if errs.Contains(err, string(types.ErrorCodeDisableInProgress)) {
			return nil, StatusInProgress, nil
		}

		if errs.Contains(err, string(types.ErrorCodeInternalError)) {
			// Can mean DISABLE was called right after ENABLE when status hasn't changed yet.
			// Otherwise, it would return a types.ErrorCodeResourceScanNotDisabled
			return nil, StatusInProgress, nil
		}

		if err != nil {
			return nil, "", err
		}

		hasEnabled := false
		hasDisabled := false

		for _, s := range st {
			if s.Status != string(types.StatusEnabled) && s.Status != string(types.StatusDisabled) {
				return 42, s.Status, nil
			}

			if !hasEnabled && s.Status == string(types.StatusEnabled) {
				hasEnabled = true
			}

			if !hasDisabled && s.Status == string(types.StatusDisabled) {
				hasDisabled = true
			}
		}

		// can only have status of enabled and/or disabled or would have returned in for

		if hasDisabled && hasEnabled {
			return 42, StatusDisabledEnabled, nil
		}

		if hasDisabled {
			return 42, string(types.StatusDisabled), nil
		}

		return 42, string(types.StatusEnabled), nil
	}
}

type AccountStatus struct {
	AccountID string
	Status    string
}

func FindAccountStatuses(ctx context.Context, conn *inspector2.Client, id string) ([]AccountStatus, error) {
	accountIDs, resourceTypes, err := parseEnablerID(id)
	if err != nil {
		return nil, fmt.Errorf("parse error (%s): %s", id, err)
	}

	ec2 := false
	ecr := false

	for _, v := range resourceTypes {
		if v == string(types.ResourceScanTypeEc2) {
			ec2 = true
			continue
		}
		if v == string(types.ResourceScanTypeEcr) {
			ecr = true
		}
	}

	// there's no describe/list but calling disable without a resource type returns an error
	// and information about state
	in := &inspector2.DisableInput{}

	if len(accountIDs) > 0 {
		in.AccountIds = accountIDs
	}

	out, err := conn.Disable(ctx, in)

	if err != nil {
		return nil, fmt.Errorf("calling Disable: %s", err)
	}

	if out == nil {
		return nil, fmt.Errorf("empty response calling Disable")
	}

	var errs *multierror.Error
	var s []AccountStatus
	for _, a := range out.Accounts {
		// in one situation, info is returned in Accounts (rather than FailedAccounts below): when
		// everything is disabled
		if a.ResourceStatus == nil {
			continue
		}

		s = append(s, AccountStatus{
			AccountID: aws.ToString(a.AccountId),
			Status:    compositeStatus(ec2, ecr, string(a.ResourceStatus.Ec2), string(a.ResourceStatus.Ecr)),
		})
	}

	for _, a := range out.FailedAccounts {
		if a.ErrorCode != types.ErrorCodeResourceScanNotDisabled {
			// because calling disable without a resource type, information is returned in failed accounts
			// with ErrorCode of types.ErrorCodeResourceScanNotDisabled
			errs = multierror.Append(errs, fmt.Errorf("(%s) %s: %s (%s)", aws.ToString(a.AccountId), a.ErrorCode, aws.ToString(a.ErrorMessage), a.Status))
			continue
		}

		if a.ResourceStatus == nil {
			continue
		}

		s = append(s, AccountStatus{
			AccountID: aws.ToString(a.AccountId),
			Status:    compositeStatus(ec2, ecr, string(a.ResourceStatus.Ec2), string(a.ResourceStatus.Ecr)),
		})
	}

	return s, errs.ErrorOrNil()
}

// compositeStatus returns the status of ec2 and/or ecr scans depending on which are set by resource.
// If you configure both, compositeStatus returns the most troubling status (e.g., *ing rather than *ed).
func compositeStatus(ec2, ecr bool, ec2Status, ecrStatus string) string {
	if ec2 && !ecr {
		return ec2Status
	}

	if !ec2 && ecr {
		return ecrStatus
	}

	// both are configured

	if ec2Status == ecrStatus {
		return ec2Status
	}

	// ING suffix beats anything (i.e., ENABLING, DISABLING, SUSPENDING)
	if strings.HasSuffix(ec2Status, "ING") {
		return ec2Status
	}

	if strings.HasSuffix(ecrStatus, "ING") {
		return ecrStatus
	}

	// not the same & neither is *ING

	if !strings.HasPrefix(ec2Status, "SUS") && !strings.HasPrefix(ecrStatus, "SUS") {
		return StatusDisabledEnabled
	}

	return string(types.StatusSuspended)
}

func expandResourceScanTypes(s []string) []types.ResourceScanType {
	vs := make([]types.ResourceScanType, 0, len(s))
	for _, v := range s {
		if v != "" {
			vs = append(vs, types.ResourceScanType(v))
		}
	}
	return vs
}

func EnablerID(accountIDs []string, types []string) string {
	return fmt.Sprintf("%s-%s", strings.Join(accountIDs, ":"), strings.Join(types, ":"))
}

func parseEnablerID(id string) ([]string, []string, error) {
	parts := strings.Split(id, "-")

	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("unexpected ID format (%s), expected <account-ids (':' separated)>-<types (':' separated)>", id)
	}

	return strings.Split(parts[0], ":"), strings.Split(parts[1], ":"), nil
}
