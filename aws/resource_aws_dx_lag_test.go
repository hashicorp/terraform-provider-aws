package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxLag_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxLagExists("aws_dx_lag.hoge"),
				),
			},
		},
	})
}

func testAccCheckAwsDxLagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_lag" {
			continue
		}

		input := &directconnect.DescribeLagsInput{
			LagId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeLags(input)
		if err != nil {
			return err
		}
		for _, v := range resp.Lags {
			if *v.LagId == rs.Primary.ID && !(*v.LagState == directconnect.LagStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Dx Lag (%s) found", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsDxLagExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxLagConfig(rName string) string {
	return fmt.Sprintf(`
    resource "aws_dx_lag" "hoge" {
      name = "tf-dx-lag-%s"
      connections_bandwidth = "1Gbps"
      location = "EqSe2"
      number_of_connections = 2
      force_destroy = true
    }
    `, rName)
}
