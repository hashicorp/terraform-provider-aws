package waiter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// EksAddonCreated waits for a EKS Addon to return status "ACTIVE" or "CREATE_FAILED"
func EksAddonCreated(ctx context.Context, conn *eks.EKS, clusterName, addonName string, timeout time.Duration) (*eks.Addon, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.AddonStatusCreating},
		Target: []string{
			eks.AddonStatusActive,
			eks.AddonStatusCreateFailed,
		},
		Refresh: EksAddonCreatedStatus(ctx, conn, addonName, clusterName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if addon, ok := outputRaw.(*eks.Addon); ok {
		// If "CREATE_FAILED" status was returned, gather add-on health issues and return error
		if aws.StringValue(addon.Status) == eks.AddonStatusCreateFailed {
			var detailedErrors []string
			for i, addonIssue := range addon.Health.Issues {
				detailedErrors = append(detailedErrors, fmt.Sprintf("Error %d: Code: %s / Message: %s",
					i+1, aws.StringValue(addonIssue.Code), aws.StringValue(addonIssue.Message)))
			}

			return addon, fmt.Errorf("creation not successful (%s): Errors:\n%s",
				aws.StringValue(addon.Status), strings.Join(detailedErrors, "\n"))
		}

		return addon, err
	}

	return nil, err
}

// EksAddonUpdated waits for a EKS Addon update to return "Successful"
func EksAddonUpdated(ctx context.Context, conn *eks.EKS, clusterName, addonName, updateID string, timeout time.Duration) (*eks.Update, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target: []string{
			eks.UpdateStatusCancelled,
			eks.UpdateStatusFailed,
			eks.UpdateStatusSuccessful,
		},
		Refresh: EksAddonUpdateStatus(ctx, conn, clusterName, addonName, updateID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	update, ok := outputRaw.(*eks.Update)
	if !ok {
		return nil, err
	}

	if aws.StringValue(update.Status) == eks.UpdateStatusSuccessful {
		return nil, nil
	}

	var detailedErrors []string
	for i, updateError := range update.Errors {
		detailedErrors = append(detailedErrors, fmt.Sprintf("Error %d: Code: %s / Message: %s",
			i+1, aws.StringValue(updateError.ErrorCode), aws.StringValue(updateError.ErrorMessage)))
	}

	return update, fmt.Errorf("EKS Addon (%s:%s) update (%s) not successful (%s): Errors:\n%s",
		clusterName, addonName, updateID, aws.StringValue(update.Status), strings.Join(detailedErrors, "\n"))
}

// EksAddonDeleted waits for a EKS Addon to return "Deleted"
func EksAddonDeleted(ctx context.Context, conn *eks.EKS, clusterName, addonName string, timeout time.Duration) (*eks.Addon, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			eks.AddonStatusActive,
			eks.AddonStatusDeleting,
		},
		Target:  []string{ResourceStatusDeleted},
		Refresh: EksAddonDeletedStatus(ctx, conn, addonName, clusterName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		// EKS API returns the ResourceNotFound error in this form:
		// ResourceNotFoundException: No addon: vpc-cni found in cluster: tf-acc-test-533189557170672934
		if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
			return nil, nil
		}
	}
	if v, ok := outputRaw.(*eks.Addon); ok {
		return v, err
	}

	return nil, err
}
