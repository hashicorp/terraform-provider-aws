package networkmanager

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
)

// validationExceptionMessageContains returns true if the error matches all these conditions:
//  * err is of type networkmanager.ValidationException
//  * ValidationException.Reason equals reason
//  * ValidationException.Fields.Message contains message
func validationExceptionMessageContains(err error, reason string, message string) bool {
	var validationException *networkmanager.ValidationException

	if errors.As(err, &validationException) && aws.StringValue(validationException.Reason) == reason {
		for _, v := range validationException.Fields {
			if strings.Contains(aws.StringValue(v.Message), message) {
				return true
			}
		}
	}

	return false
}
