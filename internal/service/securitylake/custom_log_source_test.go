// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	rName := randomCustomLogSourceName()
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "attributes.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "attributes.0.crawler_arn", "glue", fmt.Sprintf("crawler/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(resourceName, "attributes.0.database_arn", "glue", fmt.Sprintf("database/amazon_security_lake_glue_db_%s", strings.Replace(acctest.Region(), "-", "_", -1))),
					acctest.CheckResourceAttrRegionalARN(resourceName, "attributes.0.table_arn", "glue", fmt.Sprintf("table/amazon_security_lake_table_%s_ext_%s", strings.Replace(acctest.Region(), "-", "_", -1), strings.Replace(rName, "-", "_", -1))),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.crawler_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.crawler_configuration.0.role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.provider_identity.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.provider_identity.0.external_id", fmt.Sprintf("%s-test", rName)),
					resource.TestCheckNoResourceAttr(resourceName, "event_classes"),
					resource.TestCheckResourceAttr(resourceName, "provider_details.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "provider_details.0.location", regexache.MustCompile(fmt.Sprintf(`^s3://aws-security-data-lake-%s-[a-z0-9]{30}/ext/%s/$`, acctest.Region(), rName))),
					acctest.CheckResourceAttrGlobalARN(resourceName, "provider_details.0.role_arn", "iam", fmt.Sprintf("role/AmazonSecurityLake-Provider-%s-%s", rName, acctest.Region())),
					resource.TestCheckResourceAttr(resourceName, "source_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "source_version"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration},
			},
		},
	})
}

func testAccCustomLogSource_sourceVersion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	rName := randomCustomLogSourceName()
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_sourceVersion(rName, "1.5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "source_version", "1.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration},
			},
			{
				Config: testAccCustomLogSourceConfig_sourceVersion(rName, "2.5"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "source_version", "2.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration},
			},
		},
	})
}

func testAccCustomLogSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	resourceName2 := "aws_securitylake_custom_log_source.test2"
	rName := randomCustomLogSourceName()
	rName2 := randomCustomLogSourceName()
	var customLogSource, customLogSource2 types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_multiple(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					testAccCheckCustomLogSourceExists(ctx, resourceName2, &customLogSource2),

					resource.TestCheckResourceAttr(resourceName, "source_name", rName),

					resource.TestCheckResourceAttr(resourceName2, "source_name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration},
			},
		},
	})
}

func testAccCustomLogSource_eventClasses(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	rName := randomCustomLogSourceName()
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLogSourceConfig_eventClasses(rName, "FILE_ACTIVITY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "event_classes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_classes.*", "FILE_ACTIVITY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration, "event_classes"},
			},
			{
				Config: testAccCustomLogSourceConfig_eventClasses(rName, "MEMORY_ACTIVITY", "FILE_ACTIVITY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "event_classes.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_classes.*", "MEMORY_ACTIVITY"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_classes.*", "FILE_ACTIVITY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration, "event_classes"},
			},
			{
				Config: testAccCustomLogSourceConfig_eventClasses(rName, "MEMORY_ACTIVITY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomLogSourceExists(ctx, resourceName, &customLogSource),
					resource.TestCheckResourceAttr(resourceName, "event_classes.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_classes.*", "MEMORY_ACTIVITY"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrConfiguration, "event_classes"},
			},
		},
	})
}

func testAccCustomLogSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_custom_log_source.test"
	rName := randomCustomLogSourceName()
	var customLogSource types.CustomLogSourceResource

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
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

func randomCustomLogSourceName() string {
	return fmt.Sprintf("%s-%s", acctest.ResourcePrefix, sdkacctest.RandString(20-(len(acctest.ResourcePrefix)+1)))
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

func testAccCustomLogSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_custom_log_source" "test" {
  source_name = %[1]q

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[1]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_iam_role" "test" {
  name = %[1]q
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
  name = %[1]q
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
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}
`, rName))
}

func testAccCustomLogSourceConfig_sourceVersion(rName, version string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_custom_log_source" "test" {
  source_name    = %[1]q
  source_version = %[2]q

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[1]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_iam_role" "test" {
  name = %[1]q
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
  name = %[1]q
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
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}
`, rName, version))
}

func testAccCustomLogSourceConfig_eventClasses(rName string, eventClasses ...string) string {
	eventClasses = slices.ApplyToAll(eventClasses, func(s string) string {
		return fmt.Sprintf(`"%s"`, s)
	})
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_custom_log_source" "test" {
  source_name   = %[1]q
  event_classes = [%[2]s]

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[1]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_iam_role" "test" {
  name = %[1]q
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
  name = %[1]q
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
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}
`, rName, strings.Join(eventClasses, ", ")))
}

func testAccCustomLogSourceConfig_multiple(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_custom_log_source" "test" {
  source_name = %[1]q

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[1]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_securitylake_custom_log_source" "test2" {
  source_name = %[2]q

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[2]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_iam_role" "test" {
  name = %[1]q
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
  name = %[1]q
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
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}
`, rName, rName2))
}
