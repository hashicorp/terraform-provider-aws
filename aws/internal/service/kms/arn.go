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

// KeyARNOrIDEqual returns whether two CMK ARNs or IDs are equal.
func KeyARNOrIDEqual(arnOrID1, arnOrID2 string) bool {
	if arnOrID1 == arnOrID2 {
		return true
	}

	// Key ARN: arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
	// Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab
	arn1, err := arn.Parse(arnOrID1)
	firstIsARN := err == nil
	arn2, err := arn.Parse(arnOrID2)
	secondIsARN := err == nil

	if firstIsARN && !secondIsARN {
		return arn1.Resource == "key/"+arnOrID2
	}

	if secondIsARN && !firstIsARN {
		return arn2.Resource == "key/"+arnOrID1
	}

	return false
}
