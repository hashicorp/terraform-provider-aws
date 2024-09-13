// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findPhoneNumberByID(ctx context.Context, conn *pinpointsmsvoicev2.PinpointSMSVoiceV2, id string) (*pinpointsmsvoicev2.PhoneNumberInformation, error) {
	in := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
		PhoneNumberIds: aws.StringSlice([]string{id}),
	}

	out, err := conn.DescribePhoneNumbersWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
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
