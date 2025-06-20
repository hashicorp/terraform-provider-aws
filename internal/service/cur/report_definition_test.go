// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cur_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/costandusagereportservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcur "github.com/hashicorp/terraform-provider-aws/internal/service/cur"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccReportDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"
	s3PrefixUpdated := "testUpdated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_basic(rName, s3Prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					//workaround for region being based on s3 bucket region
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, names.AttrARN, "cur", fmt.Sprintf("definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportDefinitionConfig_basic(rName, s3PrefixUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					//workaround for region being based on s3 bucket region
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, names.AttrARN, "cur", fmt.Sprintf("definition/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3PrefixUpdated),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
		},
	})
}

func testAccReportDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_tags1(rName, s3Prefix, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccReportDefinitionConfig_tags2(rName, s3Prefix, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccReportDefinitionConfig_tags1(rName, s3Prefix, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccReportDefinition_textOrCSV(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(rName, s3Prefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReportDefinition_parquet(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(rName, s3Prefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReportDefinition_athena(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{"ATHENA"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(rName, s3Prefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReportDefinition_refresh(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := true
	reportVersioning := "CREATE_NEW_REPORT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(rName, s3Prefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReportDefinition_overwrite(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(rName, s3Prefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", rName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3Bucket, rName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", s3Prefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, names.AttrRegion),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "report_versioning", reportVersioning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReportDefinition_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cur_report_definition.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3Prefix := "test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CURServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReportDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_basic(rName, s3Prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcur.ResourceReportDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckReportDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CURClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cur_report_definition" {
				continue
			}

			_, err := tfcur.FindReportDefinitionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cost And Usage Report Definition %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckReportDefinitionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CURClient(ctx)
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		_, err := tfcur.FindReportDefinitionByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccReportDefinitionConfig_basic(rName, s3Prefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[2]q
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
}
`, rName, s3Prefix)
}

func testAccReportDefinitionConfig_additional(rName, s3Prefix, format, compression string, additionalArtifacts []string, refreshClosedReports bool, reportVersioning string) string {
	artifactsStr := strings.Join(additionalArtifacts, "\", \"")

	if len(additionalArtifacts) > 0 {
		artifactsStr = fmt.Sprintf("additional_artifacts       = [\"%s\"]", artifactsStr)
	} else {
		artifactsStr = ""
	}

	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = %[3]q
  compression                = %[4]q
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[2]q
  s3_region                  = aws_s3_bucket.test.region
	%[5]s
  refresh_closed_reports = %[6]t
  report_versioning      = %[7]q
}
`, rName, s3Prefix, format, compression, artifactsStr, refreshClosedReports, reportVersioning)
}

func testAccReportDefinitionConfig_tags1(rName, s3Prefix, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[2]q
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, s3Prefix, tagKey1, tagValue1)
}

func testAccReportDefinitionConfig_tags2(rName, s3Prefix, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Id": "s3policy",
  "Statement": [
    {
      "Sid": "AllowCURBillingACLPolicy",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": [
        "s3:GetBucketAcl",
        "s3:GetBucketPolicy"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    },
    {
      "Sid": "AllowCURPutObject",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::386209384616:root"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    }
  ]
}
POLICY
}

resource "aws_cur_report_definition" "test" {
  depends_on = [aws_s3_bucket_policy.test] # needed to avoid "ValidationException: Failed to verify customer bucket permission."

  report_name                = %[1]q
  time_unit                  = "DAILY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES", "SPLIT_COST_ALLOCATION_DATA"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[2]q
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, s3Prefix, tagKey1, tagValue1, tagKey2, tagValue2)
}

func TestCheckDefinitionPropertyCombination(t *testing.T) {
	t.Parallel()

	type propertyCombinationTestCase struct {
		additionalArtifacts []types.AdditionalArtifact
		compression         types.CompressionFormat
		format              types.ReportFormat
		prefix              string
		reportVersioning    types.ReportVersioning
		shouldError         bool
	}

	testCases := map[string]propertyCombinationTestCase{
		"TestAthenaAndAdditionalArtifacts": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
				types.AdditionalArtifactRedshift,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndEmptyPrefix": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndOverrideReport": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaWithNonParquetFormat": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
			},
			compression:      types.CompressionFormatParquet,
			format:           types.ReportFormatCsv,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestParquetFormatWithoutParquetCompression": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestCSVFormatWithParquetCompression": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
			},
			compression:      types.CompressionFormatParquet,
			format:           types.ReportFormatCsv,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestRedshiftWithParquetformat": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactRedshift,
			},
			compression:      types.CompressionFormatParquet,
			format:           types.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestQuicksightWithParquetformat": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactQuicksight,
			},
			compression:      types.CompressionFormatParquet,
			format:           types.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaValidCombination": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactAthena,
			},
			compression:      types.CompressionFormatParquet,
			format:           types.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedOverwrite": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactRedshift,
			},
			compression:      types.CompressionFormatGzip,
			format:           types.ReportFormatCsv,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedOverwrite": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactRedshift,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedCreateNew": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactRedshift,
			},
			compression:      types.CompressionFormatGzip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedCreateNew": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactRedshift,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedOverwrite": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactQuicksight,
			},
			compression:      types.CompressionFormatGzip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedOverwrite": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactQuicksight,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedCreateNew": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactQuicksight,
			},
			compression:      types.CompressionFormatGzip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedCreateNew": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactQuicksight,
			},
			compression:      types.CompressionFormatZip,
			format:           types.ReportFormatCsv,
			prefix:           "",
			reportVersioning: types.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestMultipleArtifacts": {
			additionalArtifacts: []types.AdditionalArtifact{
				types.AdditionalArtifactRedshift,
				types.AdditionalArtifactQuicksight,
			},
			compression:      types.CompressionFormatGzip,
			format:           types.ReportFormatCsv,
			prefix:           "prefix/",
			reportVersioning: types.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
	}

	for name, tCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tfcur.CheckReportDefinitionPropertyCombination(
				tCase.additionalArtifacts,
				tCase.compression,
				tCase.format,
				tCase.prefix,
				tCase.reportVersioning,
			)

			if tCase.shouldError && err == nil {
				t.Fail()
			} else if !tCase.shouldError && err != nil {
				t.Fail()
			}
		})
	}
}
