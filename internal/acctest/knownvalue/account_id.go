// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

var _ knownvalue.Check = accountID{}

type accountID struct {
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v accountID) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for AccountID check, got: %T", other)
	}

	if a, e := otherVal, acctest.AccountID(context.Background()); a != e {
		return fmt.Errorf("expected value %s for AccountID check, got: %s", e, a)
	}

	return nil
}

// String returns the string representation of the value.
func (v accountID) String() string {
	return "Who Knows"
}

func AccountID() knownvalue.Check {
	return accountID{}
}
