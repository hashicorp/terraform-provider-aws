// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"regexp"
	"strings"
)

var (
	accountIDRe      = regexp.MustCompile(`^\d{12}$`)
	s3CatalogRe      = regexp.MustCompile(`s3tablescatalog/([a-z0-9][a-z0-9-]{1,61}[a-z0-9])$`)
	reservedPrefixes = []string{"xn--", "sthree-", "amzn-s3-demo-"}
	reservedSuffixes = []string{"-s3alias", "--ol-s3", "--x-s3", "--table-s3"}
)

// IsAWSDataCatalogID returns whether or not the specified string is a valid AWS Data Catalog ID.
func IsAWSDataCatalogID(s string) bool {
	if accountIDRe.MatchString(s) {
		return true
	}
	m := s3CatalogRe.FindStringSubmatch(s)
	if m == nil {
		return false
	}

	catalog := m[1]

	// Check reserved prefixes
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(catalog, prefix) {
			return false
		}
	}
	// Check reserved suffixes
	for _, suffix := range reservedSuffixes {
		if strings.HasSuffix(catalog, suffix) {
			return false
		}
	}

	return true
}
