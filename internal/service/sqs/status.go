// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/service/sqs"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusQueueState(ctx context.Context, conn *sqs.SQS, url string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindQueueAttributesByURL(ctx, conn, url)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, queueStateExists, nil
	}
}

func statusQueueAttributeState(ctx context.Context, conn *sqs.SQS, url string, expected map[string]string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributesMatch := func(got map[string]string) string {
			for k, e := range expected {
				g, ok := got[k]

				if !ok {
					// Missing attribute equivalent to empty expected value.
					if e == "" {
						continue
					}

					// Backwards compatibility: https://github.com/hashicorp/terraform-provider-aws/issues/19786.
					if k == sqs.QueueAttributeNameKmsDataKeyReusePeriodSeconds && e == strconv.Itoa(DefaultQueueKMSDataKeyReusePeriodSeconds) {
						continue
					}

					return queueAttributeStateNotEqual
				}

				switch k {
				case sqs.QueueAttributeNamePolicy:
					equivalent, err := awspolicy.PoliciesAreEquivalent(g, e)

					if err != nil {
						return queueAttributeStateNotEqual
					}

					if !equivalent {
						return queueAttributeStateNotEqual
					}
				case sqs.QueueAttributeNameRedriveAllowPolicy, sqs.QueueAttributeNameRedrivePolicy:
					if !StringsEquivalent(g, e) {
						return queueAttributeStateNotEqual
					}
				default:
					if g != e {
						return queueAttributeStateNotEqual
					}
				}
			}

			return queueAttributeStateEqual
		}

		got, err := FindQueueAttributesByURL(ctx, conn, url)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := attributesMatch(got)

		return got, status, nil
	}
}
