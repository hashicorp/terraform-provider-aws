// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// This file contains code generation customizations for each AWS Go SDK service.

package namevaluesfiltersv2

import "fmt"

// ServiceFilterPackage determines the service filter type package.
func ServiceFilterPackage(serviceName string) string {
	switch serviceName {
	default:
		return fmt.Sprintf("%[1]stypes \"github.com/aws/aws-sdk-go-v2/service/%[1]s/types\"", serviceName)
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
	default:
		return "Key"
	}
}

// ServiceFilterTypeValuesField determines the service filter type values field.
func ServiceFilterTypeValuesField(serviceName string) string {
	switch serviceName {
	default:
		return "Values"
	}
}
