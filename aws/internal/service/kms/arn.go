package kms

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
)

const (
	ARNSeparator = "/"
	ARNService   = "kms"
)

// AliasARNToKeyARN converts an alias ARN to a CMK ARN.
func AliasARNToKeyARN(inputARN, keyID string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("error parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, ARNService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join([]string{"key", keyID}, ARNSeparator),
	}.String()

	return outputARN, nil
}
