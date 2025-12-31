// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestUserAgentFunction_valid(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testUserAgentFunctionConfig("test-module", "0.0.1", "test comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "test-module/0.0.1 (test comment)"),
				),
			},
		},
	})
}

func TestUserAgentFunction_valid_name(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testUserAgentFunctionConfig("test-module", "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "test-module"),
				),
			},
		},
	})
}

func TestUserAgentFunction_valid_nameVersion(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testUserAgentFunctionConfig("test-module", "0.0.1", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "test-module/0.0.1"),
				),
			},
		},
	})
}

func TestUserAgentFunction_valid_nameComment(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testUserAgentFunctionConfig("test-module", "", "test comment"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "test-module (test comment)"),
				),
			},
		},
	})
}

func TestUserAgentFunction_invalid(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testUserAgentFunctionConfig("", "", ""),
				// The full message is broken across lines, complicating validation.
				// Check just the start.
				ExpectError: regexache.MustCompile("product_name must be"),
			},
		},
	})
}

func testUserAgentFunctionConfig(name, version, comment string) string {
	return fmt.Sprintf(`
output "test" {
  value = provider::aws::user_agent(%[1]q, %[2]q, %[3]q)
}
`, name, version, comment)
}
