// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	arnResourceSeparator = "/"
	arnService           = "kms"
)

// aliasARNToKeyARN converts an alias ARN to a CMK ARN.
func aliasARNToKeyARN(inputARN, keyID string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, arnService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join([]string{names.AttrKey, keyID}, arnResourceSeparator),
	}.String()

	return outputARN, nil
}

// keyARNOrIDEqual returns whether two CMK ARNs or IDs are equal.
func keyARNOrIDEqual(arnOrID1, arnOrID2 string) bool {
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
