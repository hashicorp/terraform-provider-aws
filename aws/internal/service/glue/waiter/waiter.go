package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	MLTransformDeleteTimeout = 2 * time.Minute
	TriggerCreateTimeout     = 2 * time.Minute
	TriggerDeleteTimeout     = 2 * time.Minute
)

// MLTransformDeleted waits for an MLTransform to return Deleted
func MLTransformDeleted(conn *glue.Glue, transformId string) (*glue.GetMLTransformOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.TransformStatusTypeNotReady, glue.TransformStatusTypeReady, glue.TransformStatusTypeDeleting},
		Target:  []string{},
		Refresh: MLTransformStatus(conn, transformId),
		Timeout: MLTransformDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetMLTransformOutput); ok {
		return output, err
	}

	return nil, err
}

// TriggerCreated waits for a Trigger to return Created
func TriggerCreated(conn *glue.Glue, triggerName string) (*glue.GetTriggerOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			glue.TriggerStateActivating,
			glue.TriggerStateCreating,
		},
		Target: []string{
			glue.TriggerStateActivated,
			glue.TriggerStateCreated,
		},
		Refresh: TriggerStatus(conn, triggerName),
		Timeout: TriggerCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}

// TriggerDeleted waits for a Trigger to return Deleted
func TriggerDeleted(conn *glue.Glue, triggerName string) (*glue.GetTriggerOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.TriggerStateDeleting},
		Target:  []string{},
		Refresh: TriggerStatus(conn, triggerName),
		Timeout: TriggerDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}
