package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSRedshiftServiceAccount_basic(t *testing.T) {
	expectedAccountID := redshiftServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_redshift_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsRedshiftServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "user/logs"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftServiceAccount_Region(t *testing.T) {
	expectedAccountID := redshiftServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_redshift_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, redshift.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "user/logs"),
				),
			},
		},
	})
}

const testAccCheckAwsRedshiftServiceAccountConfig = `
data "aws_redshift_service_account" "main" {}
`

const testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig = `
data "aws_region" "current" {}

data "aws_redshift_service_account" "regional" {
  region = data.aws_region.current.name
}
`
