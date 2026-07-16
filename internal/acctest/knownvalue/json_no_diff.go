// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
)

var _ knownvalue.Check = jsonNoDiff{}

type jsonNoDiff struct {
	value string
}

func (v jsonNoDiff) CheckValue(other any) error {
	otherVal, ok := other.(string)
	if !ok {
		return fmt.Errorf("expected string value for JSONNoDiff check, got: %T", other)
	}

	if diff := jsoncmp.Diff(otherVal, v.value); diff != "" {
		return fmt.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}

	return nil
}

func (v jsonNoDiff) String() string {
	return v.value
}

func JSONNoDiff(value string) knownvalue.Check {
	return jsonNoDiff{value: value}
}
