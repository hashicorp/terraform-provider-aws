package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccAWSRDSCertificateDataSource_Id(t *testing.T) {
	dataSourceName := "data.aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccAWSRDSCertificatePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSCertificateDataSourceConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "data.aws_rds_certificate.latest", "id"),
				),
			},
		},
	})
}

func TestAccAWSRDSCertificateDataSource_LatestValidTill(t *testing.T) {
	dataSourceName := "data.aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccAWSRDSCertificatePreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSCertificateDataSourceConfigLatestValidTill(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "rds", regexp.MustCompile(`cert:rds-ca-[0-9]{4}`)),
					resource.TestCheckResourceAttr(dataSourceName, "certificate_type", "CA"),
					resource.TestCheckResourceAttr(dataSourceName, "customer_override", "false"),
					resource.TestCheckNoResourceAttr(dataSourceName, "customer_override_valid_till"),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`rds-ca-[0-9]{4}`)),
					resource.TestMatchResourceAttr(dataSourceName, "thumbprint", regexp.MustCompile(`[0-9a-f]+`)),
					resource.TestMatchResourceAttr(dataSourceName, "valid_from", regexp.MustCompile(acctest.RFC3339RegexPattern)),
					resource.TestMatchResourceAttr(dataSourceName, "valid_till", regexp.MustCompile(acctest.RFC3339RegexPattern)),
				),
			},
		},
	})
}

func testAccAWSRDSCertificatePreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	input := &rds.DescribeCertificatesInput{}

	_, err := conn.DescribeCertificates(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRDSCertificateDataSourceConfigId() string {
	return `
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

data "aws_rds_certificate" "test" {
  id = data.aws_rds_certificate.latest.id
}
`
}

func testAccAWSRDSCertificateDataSourceConfigLatestValidTill() string {
	return `
data "aws_rds_certificate" "test" {
  latest_valid_till = true
}
`
}
