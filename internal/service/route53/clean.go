// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-provider-aws/internal/dns"
)

// cleanDelegationSetID is used to remove the leading "/delegationset/" from a
// delegation set ID
func cleanDelegationSetID(id string) string {
	return strings.TrimPrefix(id, "/delegationset/")
}

// cleanZoneID is used to remove the leading "/hostedzone/" from a hosted zone ID
func cleanZoneID(id string) string {
	return strings.TrimPrefix(id, "/hostedzone/")
}

// normalizeAliasDomainName is a wrapper around the shared dns package normalization
// function which handles interface types for Plugin SDK V2 based resources
//
// The only difference between this helper and normalizeDomainName is that the single
// dot (".") domain name is not passed through as-is.
func normalizeAliasDomainName(v any) string {
	var s string
	switch v := v.(type) {
	case *string:
		s = aws.ToString(v)
	case string:
		s = v
	default:
		return ""
	}

	return dns.Normalize(strings.TrimSuffix(s, "."))
}

// normalizeDomainName is a wrapper around the shared dns package normalization
// function which handles interface values from Plugin SDK V2 based resources
func normalizeDomainName(v any) string {
	var s string
	switch v := v.(type) {
	case *string:
		s = aws.ToString(v)
	case string:
		s = v
	default:
		return ""
	}

	return dns.Normalize(s)
}

// fqdn appends a single dot (".") to the input string if necessary
func fqdn(name string) string {
	if n := len(name); n == 0 || name[n-1] == '.' {
		return name
	} else {
		return name + "."
	}
}
