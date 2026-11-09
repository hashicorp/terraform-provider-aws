// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ruleid: testcase-missing-context
func TestAccFoo_noContext(t *testing.T) {
	acctest.ParallelTest(nil, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(nil, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "# test",
			},
		},
	})
}

// ok: testcase-missing-context
func TestAccFoo_withContext(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

// ok: testcase-missing-context
func TestAccFoo_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccBar_withContext,
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

// ok: testcase-missing-context
func TestAccFoo_serial1Level(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccBar_withContext,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

// ok: testcase-missing-context
func TestAccFoo_serialSemaphore(t *testing.T) {
	t.Parallel()

	semaphore := tfsync.GetSemaphore("Foo", "AWS_FOO_LIMIT", 5)
	testCases := map[string]map[string]func(*testing.T, tfsync.Semaphore){
		acctest.CtBasic: {
			acctest.CtBasic: testAccFooSemaphore_basic,
		},
	}

	acctest.RunLimitedConcurrencyTests2Levels(t, semaphore, testCases)
}

// ruleid: testcase-missing-context
func testAccBar_noContext(t *testing.T) {
	acctest.Test(nil, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(nil, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "# test",
			},
		},
	})
}

// ok: testcase-missing-context
func testAccBar_withContext(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.Test(ctx, t, resource.TestCase{
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

// ok: testcase-missing-context
func testAccBar_utilityHelper(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"some error message",
	)
}
