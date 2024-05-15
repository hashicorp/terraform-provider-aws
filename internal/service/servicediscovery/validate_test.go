// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidNamespaceName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"ValidName",
		"V_-.dN01e",
		acctest.Ct0,
		".",
		"-",
		"_",
		strings.Repeat("x", 1024),
	}
	for _, v := range validNames {
		_, errors := tfservicediscovery.ValidNamespaceName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid namespace name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Inval:dName",
		"Invalid Name",
		"*",
		"",
		// length > 512
		strings.Repeat("x", 1025),
	}
	for _, v := range invalidNames {
		_, errors := tfservicediscovery.ValidNamespaceName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid namespace name", v)
		}
	}
}
