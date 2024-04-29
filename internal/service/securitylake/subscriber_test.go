// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSubscriber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccSubscriber_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var subscriber types.SubscriberResource
	resourceName := "aws_securitylake_subscriber.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriber, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSubscriber_customLogSource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_customLog(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.custom_log_source_resource.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.custom_log_source_resource.0.source_name", "aws_securitylake_custom_log_source.test", "source_name"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.custom_log_source_resource.0.source_version", "aws_securitylake_custom_log_source.test", "source_version"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccSubscriber_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccSubscriberConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSubscriberConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccSubscriber_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber.test"
	var subscriber types.SubscriberResource
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccSubscriberConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberExists(ctx, resourceName, &subscriber),
					resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "VPC_FLOW"),
					resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_version", "1.0"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "updated"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckSubscriberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_subscriber" {
				continue
			}

			_, err := tfsecuritylake.FindSubscriberByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Subscriber %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSubscriberExists(ctx context.Context, n string, v *types.SubscriberResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		output, err := tfsecuritylake.FindSubscriberByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSubscriberConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.test.account_id]
    regions        = [%[1]q]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[2]q
  access_type     = "S3"
  source {
    aws_log_source_resource {
      source_name    = "ROUTE53"
      source_version = "1.0"
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.test.account_id
  }

  depends_on = [aws_securitylake_aws_log_source.test]
}
`, acctest.Region(), rName))
}

func testAccSubscriberConfig_customLog(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
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

resource "aws_securitylake_subscriber" "test" {
  subscriber_name        = %[1]q
  subscriber_description = "Example"
  source {
    custom_log_source_resource {
      source_name    = aws_securitylake_custom_log_source.test.source_name
      source_version = aws_securitylake_custom_log_source.test.source_version
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  depends_on = [aws_securitylake_custom_log_source.test]
}
`, rName))
}

func testAccSubscriberConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
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

resource "aws_securitylake_subscriber" "test" {
  subscriber_name        = %[1]q
  subscriber_description = "Example"
  source {
    custom_log_source_resource {
      source_name    = aws_securitylake_custom_log_source.test.source_name
      source_version = aws_securitylake_custom_log_source.test.source_version
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_securitylake_custom_log_source.test]
}
`, rName, tag1Key, tag1Value))
}

func testAccSubscriberConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
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

resource "aws_securitylake_subscriber" "test" {
  subscriber_name        = %[1]q
  subscriber_description = "Example"
  source {
    custom_log_source_resource {
      source_name    = aws_securitylake_custom_log_source.test.source_name
      source_version = aws_securitylake_custom_log_source.test.source_version
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_securitylake_custom_log_source.test]
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccSubscriberConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_securitylake_aws_log_source" "test" {
  source {
    accounts       = [data.aws_caller_identity.test.account_id]
    regions        = [%[1]q]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[2]q
  access_type     = "S3"
  source {
    aws_log_source_resource {
      source_name    = "VPC_FLOW"
      source_version = "1.0"
    }
  }
  subscriber_identity {
    external_id = "updated"
    principal   = data.aws_caller_identity.test.account_id
  }
}
`, acctest.Region(), rName))
}
