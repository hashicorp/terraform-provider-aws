package cloudtrail_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
)

func TestAccCloudTrailServiceAccountDataSource_basic(t *testing.T) {
	expectedAccountID := tfcloudtrail.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_cloudtrail_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccCloudTrailServiceAccountDataSource_region(t *testing.T) {
	expectedAccountID := tfcloudtrail.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_cloudtrail_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_region,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccServiceAccountDataSourceConfig_basic = `
data "aws_cloudtrail_service_account" "main" {}
`

const testAccServiceAccountDataSourceConfig_region = `
data "aws_region" "current" {}

data "aws_cloudtrail_service_account" "regional" {
  region = data.aws_region.current.name
}
`
