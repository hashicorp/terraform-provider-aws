// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// This file contains code generation customizations for each AWS Go SDK service.

package namevaluesfilters

import "fmt"

// ServiceFilterPackage determines the service filter type package.
func ServiceFilterPackage(serviceName string) string {
	switch serviceName {
	default:
		return fmt.Sprintf("%[1]s \"github.com/aws/aws-sdk-go-v2/service/%[2]s/types\"", ServiceFilterPackagePrefix(serviceName), serviceName)
	}
}

// ServiceFilterPackage determines the service filter type package.
func ServiceFilterPackagePrefix(serviceName string) string {
	switch serviceName {
	default:
		return fmt.Sprintf("%[1]stypes", serviceName)
	}
}

// ServiceFilterType determines the service filter type.
func ServiceFilterType(serviceName string) string {
	switch serviceName {
	default:
		return "Filter"
	}
}

// ServiceFilterTypeNameField determines the service filter type name field.
func ServiceFilterTypeNameField(serviceName string) string {
	switch serviceName {
	case "secretsmanager":
		return "Key"
	default:
		return "Name"
	}
}

// ServiceFilterTypeNameFunc determines the function called on the service filter type name.
func ServiceFilterTypeNameFunc(serviceName string) string {
	switch serviceName {
	case "secretsmanager":
		return fmt.Sprintf("%[1]s.FilterNameStringType", ServiceFilterPackagePrefix(serviceName))
	default:
		return "aws.String"
	}
}

// ServiceFilterTypeValuesField determines the service filter type values field.
func ServiceFilterTypeValuesField(serviceName string) string {
	switch serviceName {
	default:
		return "Values"
	}
}
