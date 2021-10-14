package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	awspolicy "github.com/jen20/awspolicyequivalence"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	iamwaiter "github.com/hashicorp/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfkms "github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/kms/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for KeyState to return PendingDeletion
	KeyStatePendingDeletionTimeout = 20 * time.Minute

	KeyDeletedTimeout                = 20 * time.Minute
	KeyDescriptionPropagationTimeout = 5 * time.Minute
	KeyMaterialImportedTimeout       = 10 * time.Minute
	KeyPolicyPropagationTimeout      = 5 * time.Minute
	KeyRotationUpdatedTimeout        = 10 * time.Minute
	KeyStatePropagationTimeout       = 20 * time.Minute
	KeyTagsPropagationTimeout        = 5 * time.Minute
	KeyValidToPropagationTimeout     = 5 * time.Minute

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

func KeyDescriptionPropagated(conn *kms.KMS, id string, description string) error {
	checkFunc := func() (bool, error) {
		output, err := finder.KeyByID(conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.StringValue(output.Description) == description, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(KeyDescriptionPropagationTimeout, checkFunc, opts)
}

func KeyMaterialImported(conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kms.KeyStatePendingImport},
		Target:  []string{kms.KeyStateDisabled, kms.KeyStateEnabled},
		Refresh: KeyState(conn, id),
		Timeout: KeyMaterialImportedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kms.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

func KeyPolicyPropagated(conn *kms.KMS, id, policy string) error {
	checkFunc := func() (bool, error) {
		output, err := finder.KeyPolicyByKeyIDAndPolicyName(conn, id, tfkms.PolicyNameDefault)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(aws.StringValue(output), policy)

		if err != nil {
			return false, err
		}

		return equivalent, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(KeyPolicyPropagationTimeout, checkFunc, opts)
}

func KeyRotationEnabledPropagated(conn *kms.KMS, id string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := finder.KeyRotationEnabledByKeyID(conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.BoolValue(output) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(KeyRotationUpdatedTimeout, checkFunc, opts)
}

func KeyStatePropagated(conn *kms.KMS, id string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := finder.KeyByID(conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.BoolValue(output.Enabled) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 15,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(KeyStatePropagationTimeout, checkFunc, opts)
}

func KeyValidToPropagated(conn *kms.KMS, id string, validTo string) error {
	checkFunc := func() (bool, error) {
		output, err := finder.KeyByID(conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if output.ValidTo != nil {
			return aws.TimeValue(output.ValidTo).Format(time.RFC3339) == validTo, nil
		}

		return validTo == "", nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(KeyValidToPropagationTimeout, checkFunc, opts)
}

func TagsPropagated(conn *kms.KMS, id string, tags tftags.KeyValueTags) error {
	checkFunc := func() (bool, error) {
		output, err := tftags.KmsListTags(conn, id)

		if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return output.Equal(tags), nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(KeyTagsPropagationTimeout, checkFunc, opts)
}
