package sqs

import (
	"fmt"
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
		attributesMatch := func(got map[string]string) error {
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

					return fmt.Errorf("SQS Queue attribute (%s) not available", k)
				}

				switch k {
				case sqs.QueueAttributeNamePolicy:
					equivalent, err := awspolicy.PoliciesAreEquivalent(g, e)

					if err != nil {
						return err
					}

					if !equivalent {
						return fmt.Errorf("SQS Queue policies are not equivalent")
					}
				case sqs.QueueAttributeNameRedrivePolicy:
					if !StringsEquivalent(g, e) {
						return fmt.Errorf("SQS Queue redrive policies are not equivalent")
					}
				default:
					if g != e {
						return fmt.Errorf("SQS Queue attribute (%s) got: %s, expected: %s", k, g, e)
					}
				}
			}

			return nil
		}

		got, err := FindQueueAttributesByURL(conn, url)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		err = attributesMatch(got)

		if err != nil {
			return got, queuePolicyStateNotEqual, nil
		}

		return got, queuePolicyStateEqual, nil
	}
}
