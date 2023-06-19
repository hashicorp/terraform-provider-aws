package tfawserr

import (
	smithy "github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// AWS SDK for Go v2 variants of helpers in github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr.

// ErrCodeEquals returns true if the error matches all these conditions:
//   - err is of type smithy.APIError
//   - Error.Code() equals one of the passed codes
func ErrCodeEquals(err error, codes ...string) bool {
	if apiErr, ok := errs.As[smithy.APIError](err); ok {
		for _, code := range codes {
			if apiErr.ErrorCode() == code {
				return true
			}
		}
	}
	return false
}
