// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSTSCallerIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.STSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCallerIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckCallerIdentityAccountID("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

func TestAccSTSCallerIdentityDataSource_alternateRegion(t *testing.T) {
	ctx := acctest.Context(t)

	defaultRegion := os.Getenv(envvar.DefaultRegion)
	if defaultRegion == "" {
		t.Skipf("Skipping test due to missing %s", envvar.DefaultRegion)
	}

	alternateRegion := os.Getenv(envvar.AlternateRegion)
	if alternateRegion == "" {
		t.Skipf("Skipping test due to missing %s", envvar.AlternateRegion)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.STSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCallerIdentityConfig_alternateRegion(defaultRegion, alternateRegion),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckCallerIdentityAccountID("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

const testAccCallerIdentityConfig_basic = `
data "aws_caller_identity" "current" {}
`

func testAccCallerIdentityConfig_alternateRegion(defaultRegion, alternateRegion string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  region = %[1]q

  sts_region = %[2]q
  endpoints {
    sts = "https://sts.%[2]s.amazonaws.com"
  }
}

data "aws_caller_identity" "current" {}
`, defaultRegion, alternateRegion)
}
