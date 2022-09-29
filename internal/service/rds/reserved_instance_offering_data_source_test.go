package rds_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSInstanceOffering_basic(t *testing.T) {
	datasourceName := "data.aws_rds_reserved_instance_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceOfferingConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "currency_code"),
					resource.TestCheckResourceAttr(datasourceName, "db_instance_class", "db.t2.micro"),
					resource.TestCheckResourceAttr(datasourceName, "duration", "31536000"),
					resource.TestCheckResourceAttrSet(datasourceName, "fixed_price"),
					resource.TestCheckResourceAttr(datasourceName, "multi_az", "false"),
					resource.TestCheckResourceAttrSet(datasourceName, "offering_id"),
					resource.TestCheckResourceAttr(datasourceName, "offering_type", "All Upfront"),
					resource.TestCheckResourceAttr(datasourceName, "product_description", "mysql"),
				),
			},
		},
	})
}

func testAccInstanceOfferingConfig_basic() string {
	return `
data "aws_rds_reserved_instance_offering" "test" {
  db_instance_class   = "db.t2.micro"
  duration            = 31536000
  multi_az            = false
  offering_type       = "All Upfront"
  product_description = "mysql"
}
`
}
