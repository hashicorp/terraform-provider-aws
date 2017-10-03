package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform/helper/resource"
)

func isAWSErr(err error, code string, message string) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == code && strings.Contains(err.Message(), message)
	}
	return false
}

func retryOnAwsCode(code string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == code {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	return resp, err
}

var encodedFailureMessagePattern = regexp.MustCompile(`(?i)(.*) Encoded authorization failure message: ([\w-]+) ?( .*)?`)

type stsDecoder interface {
	DecodeAuthorizationMessage(input *sts.DecodeAuthorizationMessageInput) (*sts.DecodeAuthorizationMessageOutput, error)
}

// decodeError replaces encoded authorization messages with the
// decoded results
func decodeAWSError(decoder stsDecoder, err error) error {

	if err != nil && decoder != nil {
		groups := encodedFailureMessagePattern.FindStringSubmatch(err.Error())
		if groups != nil && len(groups) > 1 {
			result, decodeErr := decoder.DecodeAuthorizationMessage(&sts.DecodeAuthorizationMessageInput{
				EncodedMessage: aws.String(groups[2]),
			})
			if decodeErr == nil {
				msg := aws.StringValue(result.DecodedMessage)
				return fmt.Errorf("%s Authorization failure message: '%s'%s", groups[1], msg, groups[3])
			}
			log.Printf("[WARN] Attempted to decode authorization message, but received: %v", decodeErr)
		}
	}
	return err
}
