// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type identity struct {
	resourceAddress string
	values          map[string]any
}

func Identity() identity {
	return identity{}
}

// GetIdentity sets the resource address to check and stores the identity values.
// Calls to GetIdentity occur before any TestStep is run.
func (v *identity) GetIdentity(resourceAddress string) statecheck.StateCheck {
	v.resourceAddress = resourceAddress

	return newIdentityStateChecker(v)
}

type identityStateChecker struct {
	base     Base
	identity *identity
}

func newIdentityStateChecker(identity *identity) identityStateChecker {
	return identityStateChecker{
		base:     NewBase(identity.resourceAddress),
		identity: identity,
	}
}

func (vc identityStateChecker) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := vc.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	if resource.IdentitySchemaVersion == nil || len(resource.IdentityValues) == 0 {
		response.Error = fmt.Errorf("%s - Identity not found in state. Either the resource does not support identity or the Terraform version running the test does not support identity. (must be v1.12+)", vc.base.resourceAddress)

		return
	}

	vc.identity.values = maps.Collect(maps.All(resource.IdentityValues))
}

// Checks returns a function that provides the identity values as knownvalue.Checks.
// Calls to Checks occur before any TestStep is run.
func (v *identity) Checks() func() map[string]knownvalue.Check {
	return func() map[string]knownvalue.Check {
		checks := make(map[string]knownvalue.Check, len(v.values))

		for k, val := range v.values {
			// add check for optional identities that may be nil
			if val == nil {
				checks[k] = knownvalue.Null()
			} else {
				switch v := val.(type) {
				case string:
					checks[k] = knownvalue.StringExact(v)
				case json.Number:
					if i, err := v.Int64(); err == nil {
						checks[k] = knownvalue.Int64Exact(i)
					} else {
						checks[k] = knownvalue.StringExact(v.String())
					}
				default:
					checks[k] = knownvalue.StringExact(fmt.Sprintf("%v", val))
				}
			}
		}

		return checks
	}
}
