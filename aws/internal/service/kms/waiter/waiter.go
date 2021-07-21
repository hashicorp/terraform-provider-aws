package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for KeyState to return PendingDeletion
	KeyStatePendingDeletionTimeout = 20 * time.Minute

	KeyDeletedTimeout          = 20 * time.Minute
	KeyRotationUpdatedTimeout  = 10 * time.Minute
	KeyStatePropagationTimeout = 20 * time.Minute

	PropagationTimeout = 2 * time.Minute
)

// IAMPropagation retries the specified function if the returned error indicates an IAM eventual consistency issue.
// If the retries time out the specified function is called one last time.
func IAMPropagation(f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhenAwsErrCodeEquals(iamwaiter.PropagationTimeout, f, kms.ErrCodeMalformedPolicyDocumentException)
}

func KeyDeleted(conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kms.KeyStateDisabled, kms.KeyStateEnabled},
		Target:  []string{},
		Refresh: KeyState(conn, id),
		Timeout: KeyDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kms.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

// KeyStatePendingDeletion waits for KeyState to return PendingDeletion
func KeyStatePendingDeletion(conn *kms.KMS, keyID string) (*kms.DescribeKeyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			kms.KeyStateDisabled,
			kms.KeyStateEnabled,
		},
		Target:  []string{kms.KeyStatePendingDeletion},
		Refresh: KeyState(conn, keyID),
		Timeout: KeyStatePendingDeletionTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kms.DescribeKeyOutput); ok {
		return output, err
	}

	return nil, err
}
