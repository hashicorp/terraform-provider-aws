// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testErrorcheck_missing(t *testing.T) {
	ctx := acctest.Context(t)

	// ruleid: testcase-missing-errorcheck
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "# test",
			},
		},
	})
}

func testErrorcheck_present(t *testing.T) {
	ctx := acctest.Context(t)

	// ok: testcase-missing-errorcheck
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "# test",
			},
		},
	})
}
