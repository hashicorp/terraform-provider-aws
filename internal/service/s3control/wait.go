package s3control

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Minimum amount of times to verify change propagation
	propagationContinuousTargetOccurence = 2

	// Minimum amount of time to wait between S3control change polls
	propagationMinTimeout = 5 * time.Second

	// Maximum amount of time to wait for S3control changes to propagate
	propagationTimeout = 1 * time.Minute
)

func waitPublicAccessBlockConfigurationBlockPublicACLsUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   statusPublicAccessBlockConfigurationBlockPublicACLs(conn, accountID),
		Timeout:                   propagationTimeout,
		MinTimeout:                propagationMinTimeout,
		ContinuousTargetOccurence: propagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitPublicAccessBlockConfigurationBlockPublicPolicyUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   statusPublicAccessBlockConfigurationBlockPublicPolicy(conn, accountID),
		Timeout:                   propagationTimeout,
		MinTimeout:                propagationMinTimeout,
		ContinuousTargetOccurence: propagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitPublicAccessBlockConfigurationIgnorePublicACLsUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   statusPublicAccessBlockConfigurationIgnorePublicACLs(conn, accountID),
		Timeout:                   propagationTimeout,
		MinTimeout:                propagationMinTimeout,
		ContinuousTargetOccurence: propagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitPublicAccessBlockConfigurationRestrictPublicBucketsUpdated(conn *s3control.S3Control, accountID string, expectedValue bool) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Target:                    []string{strconv.FormatBool(expectedValue)},
		Refresh:                   statusPublicAccessBlockConfigurationRestrictPublicBuckets(conn, accountID),
		Timeout:                   propagationTimeout,
		MinTimeout:                propagationMinTimeout,
		ContinuousTargetOccurence: propagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}
