// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// This file contains code generation customizations for each AWS Go SDK service.

package namevaluesfilters

import "fmt"

// ServiceFilterPackage determines the service filter type package.
func ServiceFilterPackage(serviceName string) string {
	switch serviceName {
	case "secretsmanager":
		return fmt.Sprintf("github.com/aws/aws-sdk-go-v2/service/%s/types", serviceName)
	default:
		return "github.com/aws/aws-sdk-go/service/" + serviceName
	}
}

// ServiceFilterType determines the service filter type.
func ServiceFilterType(serviceName string) string {
	switch serviceName {
	case "resourcegroupstaggingapi":
		return "TagFilter"
	default:
		return "Filter"
	}
}

// ServiceFilterTypeNameField determines the service filter type name field.
func ServiceFilterTypeNameField(serviceName string) string {
	switch serviceName {
	case "resourcegroupstaggingapi", "secretsmanager":
		return "Key"
	default:
		return "Name"
	}
}

// ServiceFilterTypeValuesField determines the service filter type values field.
func ServiceFilterTypeValuesField(serviceName string) string {
	switch serviceName {
	default:
		return "Values"
	}
}
