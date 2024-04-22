// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.crawler_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.crawler_configuration.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.provider_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.provider_identity.0.external_id", "windows-sysmon-test"),
					resource.TestCheckResourceAttr(resourceName, "event_classes.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_classes.*", "FILE_ACTIVITY"),
					resource.TestCheckResourceAttr(resourceName, "source_name", "windows-sysmon"),
					resource.TestCheckResourceAttr(resourceName, "source_version", "1.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration", "event_classes"},
			},
		},
	})
}

func testAccCustomLogSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceCustomLogSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomLogSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_custom_log_source" {
				continue
			}

			_, err := tfsecuritylake.FindCustomLogSourceBySourceName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Custom Log Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomLogSourceExists(ctx context.Context, n string, v *types.CustomLogSourceResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindCustomLogSourceBySourceName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCustomLogSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), `
resource "aws_iam_role" "test" {
  name = "AmazonSecurityLakeCustomDataGlueCrawler-windows-sysmon"
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "glue.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = "AmazonSecurityLakeCustomDataGlueCrawler-windows-sysmon"
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3:GetObject",
      "s3:PutObject"
    ],
    "Resource": "*"
  }]
}
POLICY

  depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}

resource "aws_securitylake_custom_log_source" "test" {
  source_name    = "windows-sysmon"
  source_version = "1.0"
  event_classes  = ["FILE_ACTIVITY"]

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "windows-sysmon-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}
`)
}
