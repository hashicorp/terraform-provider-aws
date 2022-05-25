package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findOptOutListByID(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string) (*pinpointsmsvoicev2.OptOutListInformation, error) {
	in := &pinpointsmsvoicev2.DescribeOptOutListsInput{
		OptOutListNames: aws.StringSlice([]string{id}),
	}

	out, err := conn.DescribeOptOutListsWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.OptOutLists) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.OptOutLists[0], nil
}

func findPhoneNumberByID(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	in := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
		PhoneNumberIds: aws.StringSlice([]string{id}),
	}

	out, err := conn.DescribePhoneNumbersWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.PhoneNumbers == nil || len(out.PhoneNumbers) != 1 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.PhoneNumbers[0], nil
}
