package waiter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	DeliveryStreamCreatedTimeout = 20 * time.Minute
	DeliveryStreamDeletedTimeout = 20 * time.Minute

	DeliveryStreamEncryptionEnabledTimeout  = 10 * time.Minute
	DeliveryStreamEncryptionDisabledTimeout = 10 * time.Minute
)

func DeliveryStreamCreated(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamStatusCreating},
		Target:  []string{firehose.DeliveryStreamStatusActive},
		Refresh: DeliveryStreamStatus(conn, name),
		Timeout: DeliveryStreamCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*firehose.DeliveryStreamDescription); ok {
		if status, failureDescription := aws.StringValue(output.DeliveryStreamStatus), output.FailureDescription; status == firehose.DeliveryStreamStatusCreatingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func DeliveryStreamDeleted(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamStatusDeleting},
		Target:  []string{},
		Refresh: DeliveryStreamStatus(conn, name),
		Timeout: DeliveryStreamDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*firehose.DeliveryStreamDescription); ok {
		if status, failureDescription := aws.StringValue(output.DeliveryStreamStatus), output.FailureDescription; status == firehose.DeliveryStreamStatusDeletingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func DeliveryStreamEncryptionEnabled(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamEncryptionStatusEnabling},
		Target:  []string{firehose.DeliveryStreamEncryptionStatusEnabled},
		Refresh: DeliveryStreamEncryptionConfigurationStatus(conn, name),
		Timeout: DeliveryStreamEncryptionEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*firehose.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := aws.StringValue(output.Status), output.FailureDescription; status == firehose.DeliveryStreamEncryptionStatusEnablingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}

func DeliveryStreamEncryptionDisabled(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamEncryptionStatusDisabling},
		Target:  []string{firehose.DeliveryStreamEncryptionStatusDisabled},
		Refresh: DeliveryStreamEncryptionConfigurationStatus(conn, name),
		Timeout: DeliveryStreamEncryptionDisabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*firehose.DeliveryStreamEncryptionConfiguration); ok {
		if status, failureDescription := aws.StringValue(output.Status), output.FailureDescription; status == firehose.DeliveryStreamEncryptionStatusDisablingFailed && failureDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(failureDescription.Type), aws.StringValue(failureDescription.Details)))
		}

		return output, err
	}

	return nil, err
}
