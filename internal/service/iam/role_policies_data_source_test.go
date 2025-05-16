// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"strconv"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolePoliciesDataSource(t *testing.T) {
	ctx := acctest.Context(t)
	// Long-running test guard for tests that run longer than defined thrshhold
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	inline_policies_names := []string{}
	for i := 1; i <= 101; i++ {
		inline_policies_names = append(inline_policies_names, fmt.Sprintf("%d", i))
	}
	dataSourceName := "data.aws_iam_role_policies.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Testing Empty case handling
			{
				Config: testAccRolePoliciesDataSourceConfig_basic(rName, []string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "policy_names.#", "0"),
				),
			},
			// Testing normal case handling
			{
				Config: testAccRolePoliciesDataSourceConfig_basic(rName, inline_policies_names[0:2]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "policy_names.#", strconv.Itoa(len(inline_policies_names[0:2]))),
					resource.TestCheckResourceAttr(dataSourceName, "policy_names.0", inline_policies_names[0]),
					resource.TestCheckResourceAttr(dataSourceName, "policy_names.1", inline_policies_names[1]),
				),
			},
			// Testing correct handling of pagination
			{
				Config: testAccRolePoliciesDataSourceConfig_basic(rName, inline_policies_names),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "policy_names.#", strconv.Itoa(len(inline_policies_names))),
				),
			},
		},
	})
}

func testAccRolePoliciesDataSourceConfig_basic(rName string, inline_policies_names []string) string {
	var inline_policies string
	for _, v := range inline_policies_names {
		inline_policies = fmt.Sprintf(`
			%[2]s
			
			inline_policy {
				name = %[1]q
			
				policy = jsonencode({
				Version = "2012-10-17"
				Statement = [
					{
					Action   = "*"
					Effect   = "Allow"
					Resource = "*"
					},
				]
				})
			}`, v, inline_policies)
	}
	return fmt.Sprintf(`
		data "aws_iam_policy_document" "test" {
			statement {
			actions = ["sts:AssumeRole"]
			principals {
				type        = "Service"
				identifiers = ["ec2.amazonaws.com"]
			}
			}
		}
		resource "aws_iam_role" "test" {
			name               = %[1]q
			assume_role_policy = data.aws_iam_policy_document.test.json
			%[2]s
		}
		
		data "aws_iam_role_policies" "test" {
			role_name = aws_iam_role.test.name
		}
	`, rName, inline_policies)
}
