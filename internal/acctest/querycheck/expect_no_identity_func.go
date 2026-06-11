// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package querycheck

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

var _ querycheck.QueryResultCheck = expectNoIdentityFunc{}

type expectNoIdentityFunc struct {
	listResourceAddress string
	identityFunc        func() map[string]knownvalue.Check
}

// Adapted (with updates) from github.com/hashicorp/terraform-plugin-testing/statecheck.ExpectNoIdentity
func (e expectNoIdentityFunc) CheckQuery(_ context.Context, req querycheck.CheckQueryRequest, resp *querycheck.CheckQueryResponse) {
	checks := e.identityFunc()

	for _, res := range req.Query {
		var errCollection []error

		if e.listResourceAddress != strings.TrimPrefix(res.Address, "list.") {
			continue
		}

		if len(res.Identity) != len(checks) {
			var deltaMsg string
			if len(res.Identity) > len(checks) {
				deltaMsg = statecheck.CreateDeltaString(res.Identity, checks, "actual identity has extra attribute(s): ")
			} else {
				deltaMsg = statecheck.CreateDeltaString(checks, res.Identity, "actual identity is missing attribute(s): ")
			}

			resp.Error = fmt.Errorf("%s - Expected %d attribute(s) in the actual identity object, got %d attribute(s): %s", e.listResourceAddress, len(checks), len(res.Identity), deltaMsg)
			return
		}

		var keys []string

		for k := range checks {
			keys = append(keys, k)
		}

		slices.Sort(keys)

		for _, k := range keys {
			actualIdentityVal, ok := res.Identity[k]

			if !ok {
				resp.Error = fmt.Errorf("%s - missing attribute %q in actual identity object", e.listResourceAddress, k)
				return
			}

			if err := checks[k].CheckValue(actualIdentityVal); err != nil {
				errCollection = append(errCollection, fmt.Errorf("%s - %q identity attribute: %w", e.listResourceAddress, k, err))
			}
		}

		if errCollection == nil {
			errs := []error{fmt.Errorf("an unexpected identity matching the given attributes was found")}
			// wrap errors for each check
			for attr, check := range checks {
				errs = append(errs, fmt.Errorf("attribute %q: %s", attr, check))
			}
			errs = append(errs, fmt.Errorf("address: %s\n", e.listResourceAddress))
			resp.Error = errors.Join(errs...)
		}
	}
}

// ExpectNoIdentityFunc returns a query check that asserts that the given list resource does not contain a resource with an identity matching
// the identity checks returned by the identityFunc.
//
// This query check can only be used with managed resources that support resource identity and query. Query is only supported in Terraform v1.14+
func ExpectNoIdentityFunc(resourceAddress string, identityFunc func() map[string]knownvalue.Check) querycheck.QueryResultCheck {
	return expectNoIdentityFunc{
		listResourceAddress: resourceAddress,
		identityFunc:        identityFunc,
	}
}
