package firehose

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	deliveryStreamCreatedTimeout = 20 * time.Minute
	deliveryStreamDeletedTimeout = 20 * time.Minute

	deliveryStreamEncryptionEnabledTimeout  = 10 * time.Minute
	deliveryStreamEncryptionDisabledTimeout = 10 * time.Minute
)

func waitDeliveryStreamCreated(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamStatusCreating},
		Target:  []string{firehose.DeliveryStreamStatusActive},
		Refresh: statusDeliveryStream(conn, name),
		Timeout: deliveryStreamCreatedTimeout,
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

func waitDeliveryStreamDeleted(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamStatusDeleting},
		Target:  []string{},
		Refresh: statusDeliveryStream(conn, name),
		Timeout: deliveryStreamDeletedTimeout,
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

func waitDeliveryStreamEncryptionEnabled(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamEncryptionStatusEnabling},
		Target:  []string{firehose.DeliveryStreamEncryptionStatusEnabled},
		Refresh: statusDeliveryStreamEncryptionConfiguration(conn, name),
		Timeout: deliveryStreamEncryptionEnabledTimeout,
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

func waitDeliveryStreamEncryptionDisabled(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamEncryptionConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{firehose.DeliveryStreamEncryptionStatusDisabling},
		Target:  []string{firehose.DeliveryStreamEncryptionStatusDisabled},
		Refresh: statusDeliveryStreamEncryptionConfiguration(conn, name),
		Timeout: deliveryStreamEncryptionDisabledTimeout,
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
