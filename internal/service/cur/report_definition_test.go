package cur_test

import (
	"fmt"
	"strings"
	"testing"

	cur "github.com/aws/aws-sdk-go/service/costandusagereportservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcur "github.com/hashicorp/terraform-provider-aws/internal/service/cur"
)

func testAccReportDefinition_basic(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_basic(reportName, bucketName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					//workaround for region being based on s3 bucket region
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "arn", "cur", fmt.Sprintf("definition/%s", reportName)),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReportDefinitionConfig_basic(reportName, bucketName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					//workaround for region being based on s3 bucket region
					acctest.CheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "arn", "cur", fmt.Sprintf("definition/%s", reportName)),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "compression", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", "test"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
				),
			},
		},
	})
}

func testAccReportDefinition_textOrCSV(t *testing.T) {
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
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
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{}
	refreshClosedReports := false
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
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
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := "data"
	format := "Parquet"
	compression := "Parquet"
	additionalArtifacts := []string{"ATHENA"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
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
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := true
	reportVersioning := "CREATE_NEW_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", bucketPrefix),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "true"),
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
	resourceName := "aws_cur_report_definition.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	bucketPrefix := ""
	format := "textORcsv"
	compression := "GZIP"
	additionalArtifacts := []string{"REDSHIFT", "QUICKSIGHT"}
	refreshClosedReports := false
	reportVersioning := "OVERWRITE_REPORT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_additional(reportName, bucketName, bucketPrefix, format, compression, additionalArtifacts, refreshClosedReports, reportVersioning),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "report_name", reportName),
					resource.TestCheckResourceAttr(resourceName, "time_unit", "DAILY"),
					resource.TestCheckResourceAttr(resourceName, "format", format),
					resource.TestCheckResourceAttr(resourceName, "compression", compression),
					resource.TestCheckResourceAttr(resourceName, "additional_schema_elements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_prefix", ""),
					resource.TestCheckResourceAttrPair(resourceName, "s3_region", s3BucketResourceName, "region"),
					resource.TestCheckResourceAttr(resourceName, "additional_artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "refresh_closed_reports", "false"),
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
	resourceName := "aws_cur_report_definition.test"
	reportName := sdkacctest.RandomWithPrefix("tf_acc_test")
	bucketName := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cur.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReportDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReportDefinitionConfig_basic(reportName, bucketName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReportDefinitionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcur.ResourceReportDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReportDefinitionDestroy(s *terraform.State) error {
	conn := testAccProviderCur.Meta().(*conns.AWSClient).CURConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cur_report_definition" {
			continue
		}

		matchingReportDefinition, err := tfcur.FindReportDefinitionByName(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Report Definition (%s): %w", rs.Primary.ID, err)
		}

		if matchingReportDefinition == nil {
			continue
		}

		return fmt.Errorf("Report Definition still exists: %q", rs.Primary.ID)
	}
	return nil

}

func testAccCheckReportDefinitionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderCur.Meta().(*conns.AWSClient).CURConn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		matchingReportDefinition, err := tfcur.FindReportDefinitionByName(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading Report Definition (%s): %w", rs.Primary.ID, err)
		}

		if matchingReportDefinition == nil {
			return fmt.Errorf("Report Definition does not exist: %q", rs.Primary.ID)
		}

		return nil
	}
}

func testAccReportDefinitionConfig_basic(reportName, bucketName, prefix string) string {
	return acctest.ConfigCompose(
		testAccRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = %[3]q
  s3_region                  = aws_s3_bucket.test.region
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
}
`, reportName, bucketName, prefix))
}

func testAccReportDefinitionConfig_additional(reportName string, bucketName string, bucketPrefix string, format string, compression string, additionalArtifacts []string, refreshClosedReports bool, reportVersioning string) string {
	artifactsStr := strings.Join(additionalArtifacts, "\", \"")

	if len(additionalArtifacts) > 0 {
		artifactsStr = fmt.Sprintf("additional_artifacts       = [\"%s\"]", artifactsStr)
	} else {
		artifactsStr = ""
	}

	return acctest.ConfigCompose(
		testAccRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%[2]s"
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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

  report_name                = "%[1]s"
  time_unit                  = "DAILY"
  format                     = "%[4]s"
  compression                = "%[5]s"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = aws_s3_bucket.test.id
  s3_prefix                  = "%[3]s"
  s3_region                  = aws_s3_bucket.test.region
	%[6]s
  refresh_closed_reports = %[7]t
  report_versioning      = "%[8]s"
}
`, reportName, bucketName, bucketPrefix, format, compression, artifactsStr, refreshClosedReports, reportVersioning))
}

func TestCheckDefinitionPropertyCombination(t *testing.T) {
	type propertyCombinationTestCase struct {
		additionalArtifacts []string
		compression         string
		format              string
		prefix              string
		reportVersioning    string
		shouldError         bool
	}

	testCases := map[string]propertyCombinationTestCase{
		"TestAthenaAndAdditionalArtifacts": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndEmptyPrefix": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaAndOverrideReport": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaWithNonParquetFormat": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestParquetFormatWithoutParquetCompression": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestCSVFormatWithParquetCompression": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestRedshiftWithParquetformat": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestQuicksightWithParquetformat": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      true,
		},
		"TestAthenaValidCombination": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactAthena,
			},
			compression:      cur.CompressionFormatParquet,
			format:           cur.ReportFormatParquet,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestRedshiftWithGzipedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestRedshiftWithZippedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedOverwrite": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
		"TestQuicksightWithGzipedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestQuicksightWithZippedCreateNew": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatZip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "",
			reportVersioning: cur.ReportVersioningCreateNewReport,
			shouldError:      false,
		},
		"TestMultipleArtifacts": {
			additionalArtifacts: []string{
				cur.AdditionalArtifactRedshift,
				cur.AdditionalArtifactQuicksight,
			},
			compression:      cur.CompressionFormatGzip,
			format:           cur.ReportFormatTextOrcsv,
			prefix:           "prefix/",
			reportVersioning: cur.ReportVersioningOverwriteReport,
			shouldError:      false,
		},
	}

	for name, tCase := range testCases {
		t.Run(name, func(t *testing.T) {
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
