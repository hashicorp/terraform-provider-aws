// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_securitylake_custom_log_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.crawler_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.crawler_configuration.0.role_arn", "aws_iam_role.custom_log", "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.provider_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.provider_identity.0.external_id", "windows-sysmon-test"),
					resource.TestCheckResourceAttr(resourceName, "source_name", "windows-sysmon"),
					resource.TestCheckResourceAttr(resourceName, "source_version", "1.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{""},
			},
		},
	})
}

func testAccCustomLogSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_custom_log_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_basic(rName),
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

			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameCustomLogSource, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCustomLogSourceExists(ctx context.Context, name string, logSource *types.CustomLogSourceResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameCustomLogSource, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameCustomLogSource, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		resp, err := tfsecuritylake.FindCustomLogSourceBySourceName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameCustomLogSource, rs.Primary.ID, err)
		}

		*logSource = *resp

		return nil
	}
}

func testAccCustomLogSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(rName), `

data "aws_caller_identity" "test" {}

resource "aws_iam_role" "custom_log" {
	name               = "AmazonSecurityLakeCustomDataGlueCrawler-windows-sysmon"
	path               = "/service-role/"
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

resource "aws_iam_role_policy" "custom_log" {
	name = "AmazonSecurityLakeCustomDataGlueCrawler-windows-sysmon"
	role = aws_iam_role.custom_log.name
  
	policy = <<POLICY
{
"Version": "2012-10-17",
"Statement": [
	{
		"Sid": "S3WriteRead",
		"Effect": "Allow",
		"Action": [
			"s3:GetObject",
			"s3:PutObject"
		],
		"Resource": [
			"arn:${data.aws_partition.current.partition}:s3:::aws-security-data-lake*/*"
		]
	}
]
}
  POLICY
  depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_iam_role_policy_attachment" "glue_service_role" {
	policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
	role       = aws_iam_role.custom_log.name
  }

resource "aws_securitylake_custom_log_source" "test" {
    source_name    = "windows-sysmon"
    source_version = "1.0"
	event_classes  = ["FILE_ACTIVITY"]
	configuration {
		crawler_configuration {
			role_arn = aws_iam_role.custom_log.arn
		}

		provider_identity {
			external_id = "windows-sysmon-test"
			principal   = data.aws_caller_identity.test.account_id
		}
	}

	depends_on = [aws_securitylake_data_lake.test, aws_iam_role.custom_log]
}
`)
}
