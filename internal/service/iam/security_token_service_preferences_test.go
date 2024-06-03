// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMSecurityTokenServicePreferences_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_security_token_service_preferences.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityTokenServicePreferencesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "global_endpoint_token_version", "v2Token"),
				),
			},
		},
	})
}

const testAccSecurityTokenServicePreferencesConfig_basic = `
resource "aws_iam_security_token_service_preferences" "test" {
  global_endpoint_token_version = "v2Token"
}
`
