package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSElbServiceAccount_basic(t *testing.T) {
	expectedAccountID := elbAccountIdPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_elb_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, elb.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElbServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccAWSElbServiceAccount_Region(t *testing.T) {
	expectedAccountID := elbAccountIdPerRegionMap[acctest.Region()]

	dataSourceName := "data.aws_elb_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, elb.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElbServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					acctest.CheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccCheckAwsElbServiceAccountConfig = `
data "aws_elb_service_account" "main" {}
`

const testAccCheckAwsElbServiceAccountExplicitRegionConfig = `
data "aws_region" "current" {}

data "aws_elb_service_account" "regional" {
  region = data.aws_region.current.name
}
`
