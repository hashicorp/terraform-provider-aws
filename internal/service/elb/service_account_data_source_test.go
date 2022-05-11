package elb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
)

func TestAccELBServiceAccountDataSource_basic(t *testing.T) {
	expectedAccountID := tfelb.AccountIdPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_elb_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSElbServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccELBServiceAccountDataSource_region(t *testing.T) {
	expectedAccountID := tfelb.AccountIdPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_elb_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSElbServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccCheckAWSElbServiceAccountConfig = `
data "aws_elb_service_account" "main" {}
`

const testAccCheckAWSElbServiceAccountExplicitRegionConfig = `
data "aws_region" "current" {}

data "aws_elb_service_account" "regional" {
  region = data.aws_region.current.name
}
`
