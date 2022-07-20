package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws/arn"
)

func ARNForNewRegion(rn string, newRegion string) (string, error) {
	parsedARN, err := arn.Parse(rn)
	if err != nil {
		return "", err
	}

	parsedARN.Region = newRegion

	return parsedARN.String(), nil
}
