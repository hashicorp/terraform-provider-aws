// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestARNParseFunction_known(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testARNParseFunctionConfig("arn:aws:iam::444455556666:role/example"),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// TODO: figure out how to test object output
				//resource.TestCheckOutput("test", "arn:aws:iam::444455556666:role/example"),
				),
			},
		},
	})
}

func TestARNParseFunction_invalid(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testARNParseFunctionConfig("invalid"),
				ExpectError: regexp.MustCompile("arn parsing failed"),
			},
		},
	})
}

func testARNParseFunctionConfig(arg string) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		aws = {
			source = "hashicorp/aws"
		}
	}
}

output "test" {
	value = provider::aws::arn_parse(%[1]q)
}
`, arg)
}
