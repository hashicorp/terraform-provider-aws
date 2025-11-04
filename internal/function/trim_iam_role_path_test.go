// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

var (
	// ExpectError parses the human readable output of a terraform apply run, in which
	// formatting (including line breaks) may change over time. For extra safety, we add
	// optional whitespace between each word in the expected error text.

	expectedErrorInvalidARN      = regexache.MustCompile(`invalid[\s\n]*prefix`)
	expectedErrorInvalidService  = regexache.MustCompile(`service[\s\n]*must`)
	expectedErrorInvalidRegion   = regexache.MustCompile(`region[\s\n]*must`)
	expectedErrorInvalidResource = regexache.MustCompile(`resource[\s\n]*must`)
)

func TestTrimIAMRolePathFunction_valid(t *testing.T) {
	t.Parallel()
	arg := "arn:aws:iam::444455556666:role/example"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config: testTrimIAMRolePathFunctionConfig(arg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", arg),
				),
			},
		},
	})
}

func TestTrimIAMRolePathFunction_validWithPath(t *testing.T) {
	t.Parallel()
	arg := "arn:aws:iam::444455556666:role/with/some/path/parts/example"
	expected := "arn:aws:iam::444455556666:role/example"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config: testTrimIAMRolePathFunctionConfig(arg),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", expected),
				),
			},
		},
	})
}

func TestTrimIAMRolePathFunction_invalidARN(t *testing.T) {
	t.Parallel()
	arg := "foo"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config:      testTrimIAMRolePathFunctionConfig(arg),
				ExpectError: expectedErrorInvalidARN,
			},
		},
	})
}

func TestTrimIAMRolePathFunction_invalidService(t *testing.T) {
	t.Parallel()
	arg := "arn:aws:s3:::bucket/foo"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config:      testTrimIAMRolePathFunctionConfig(arg),
				ExpectError: expectedErrorInvalidService,
			},
		},
	})
}

func TestTrimIAMRolePathFunction_invalidRegion(t *testing.T) {
	t.Parallel()
	arg := "arn:aws:iam:us-east-1:444455556666:role/example"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config:      testTrimIAMRolePathFunctionConfig(arg),
				ExpectError: expectedErrorInvalidRegion,
			},
		},
	})
}

func TestTrimIAMRolePathFunction_invalidResource(t *testing.T) {
	t.Parallel()
	arg := "arn:aws:iam::444455556666:policy/example"

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config:      testTrimIAMRolePathFunctionConfig(arg),
				ExpectError: expectedErrorInvalidResource,
			},
		},
	})
}

func testTrimIAMRolePathFunctionConfig(arg string) string {
	return fmt.Sprintf(`
output "test" {
  value = provider::aws::trim_iam_role_path(%[1]q)
}`, arg)
}
