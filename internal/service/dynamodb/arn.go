// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func arnForNewRegion(rn string, newRegion string) (string, error) {
	parsedARN, err := arn.Parse(rn)
	if err != nil {
		return "", err
	}

	parsedARN.Region = newRegion

	return parsedARN.String(), nil
}

func regionFromARN(rn string) (string, error) {
	parsedARN, err := arn.Parse(rn)
	if err != nil {
		return "", err
	}

	return parsedARN.Region, nil
}

func tableNameFromARN(rn string) (string, error) {
	parsedARN, err := arn.Parse(rn)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(parsedARN.Resource, "table/"), nil
}
