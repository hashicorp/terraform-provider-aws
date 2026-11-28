// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"time"

	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	propagationTimeout = 2 * time.Minute
)

// Parameter data types.
// See https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-data-types.html.
const (
	parameterDataTypeAccountID   = "aws:account:id"
	parameterDataTypeBinary      = "binary"
	parameterDataTypeBoolean     = "boolean"
	parameterDataTypeEC2Image    = "aws:ec2:image"
	parameterDataTypeEmail       = names.AttrEmail
	parameterDataTypeFloat       = "float"
	parameterDataTypeHostname    = "hostname"
	parameterDataTypeIntegration = "aws:ssm:integration"
	parameterDataTypeInteger     = "integer"
	parameterDataTypeIPv4Address = "ipv4-address"
	parameterDataTypeIPv6Address = "ipv6-address"
	parameterDataTypeJSON        = names.AttrJSON
	parameterDataTypeJSONC       = "jsonc"
	parameterDataTypeRegion      = "aws:region"
	parameterDataTypeText        = "text"
	parameterDataTypeURL         = names.AttrURL
)

func parameterDataType_Values() []string {
	return []string{
		parameterDataTypeAccountID,
		parameterDataTypeBinary,
		parameterDataTypeBoolean,
		parameterDataTypeEC2Image,
		parameterDataTypeEmail,
		parameterDataTypeFloat,
		parameterDataTypeHostname,
		parameterDataTypeIntegration,
		parameterDataTypeInteger,
		parameterDataTypeIPv4Address,
		parameterDataTypeIPv6Address,
		parameterDataTypeJSON,
		parameterDataTypeJSONC,
		parameterDataTypeRegion,
		parameterDataTypeText,
		parameterDataTypeURL,
	}
}
