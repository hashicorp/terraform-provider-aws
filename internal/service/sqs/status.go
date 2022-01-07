package sqs

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/sqs"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusQueueState(conn *sqs.SQS, url string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindQueueAttributesByURL(conn, url)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, queueStateExists, nil
	}
}

func statusQueueAttributeState(conn *sqs.SQS, url string, expected map[string]string) resource.StateRefreshFunc {
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

					return queuePolicyStateNotEqual
				}

				switch k {
				case sqs.QueueAttributeNamePolicy:
					equivalent, err := awspolicy.PoliciesAreEquivalent(g, e)

					if err != nil {
						return queuePolicyStateNotEqual
					}

					if !equivalent {
						return queuePolicyStateNotEqual
					}
				case sqs.QueueAttributeNameRedriveAllowPolicy, sqs.QueueAttributeNameRedrivePolicy:
					if !StringsEquivalent(g, e) {
						return queuePolicyStateNotEqual
					}
				default:
					if g != e {
						return queuePolicyStateNotEqual
					}
				}
			}

			return queuePolicyStateEqual
		}

		got, err := FindQueueAttributesByURL(conn, url)

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
