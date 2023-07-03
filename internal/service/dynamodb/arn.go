// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"strings"

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

func RegionFromARN(rn string) (string, error) {
	parsedARN, err := arn.Parse(rn)
	if err != nil {
		return "", err
	}

	return parsedARN.Region, nil
}

func TableNameFromARN(rn string) (string, error) {
	parsedARN, err := arn.Parse(rn)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(parsedARN.Resource, "table/"), nil
}
