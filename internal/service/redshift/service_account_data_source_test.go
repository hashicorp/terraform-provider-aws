package redshift_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
)

func TestAccRedshiftServiceAccountDataSource_basic(t *testing.T) {
	expectedAccountID := tfredshift.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_redshift_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "user/logs"),
				),
			},
		},
	})
}

func TestAccRedshiftServiceAccountDataSource_region(t *testing.T) {
	expectedAccountID := tfredshift.ServiceAccountPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_redshift_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceAccountDataSourceConfig_explicitRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "user/logs"),
				),
			},
		},
	})
}

const testAccServiceAccountDataSourceConfig_basic = `
data "aws_redshift_service_account" "main" {}
`

const testAccServiceAccountDataSourceConfig_explicitRegion = `
data "aws_region" "current" {}

data "aws_redshift_service_account" "regional" {
  region = data.aws_region.current.name
}
`
