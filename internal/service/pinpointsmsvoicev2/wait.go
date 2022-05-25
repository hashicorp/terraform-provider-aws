package pinpointsmsvoicev2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitPhoneNumberCreated(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string, timeout time.Duration) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{pinpointsmsvoicev2.NumberStatusActive},
		Refresh:                   statusPhoneNumber(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pinpointsmsvoicev2.PhoneNumberInformation); ok {
		return out, err
	}

	return nil, err
}

func waitPhoneNumberUpdated(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string, timeout time.Duration) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{pinpointsmsvoicev2.NumberStatusActive},
		Refresh:                   statusPhoneNumber(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pinpointsmsvoicev2.PhoneNumberInformation); ok {
		return out, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string, timeout time.Duration) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{pinpointsmsvoicev2.NumberStatusDisassociating},
		Target:  []string{},
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pinpointsmsvoicev2.PhoneNumberInformation); ok {
		return out, err
	}

	return nil, err
}

func statusPhoneNumber(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findPhoneNumberByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func checkUpdateAfterCreateNeeded(d *schema.ResourceData, schemaKeys []string) bool {
	for _, schemaKey := range schemaKeys {
		if _, ok := d.GetOk(schemaKey); ok {
			return true
		}
	}

	return false
}
