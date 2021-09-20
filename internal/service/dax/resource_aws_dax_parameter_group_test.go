package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAwsDaxParameterGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dax_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSDax(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dax.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsDaxParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDaxParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDaxParameterGroupConfig_parameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAwsDaxParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_parameter_group" {
			continue
		}

		_, err := conn.DescribeParameterGroups(&dax.DescribeParameterGroupsInput{
			ParameterGroupNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, dax.ErrCodeParameterGroupNotFoundFault, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsDaxParameterGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn

		_, err := conn.DescribeParameterGroups(&dax.DescribeParameterGroupsInput{
			ParameterGroupNames: []*string{aws.String(rs.Primary.ID)},
		})

		return err
	}
}

func testAccDaxParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_dax_parameter_group" "test" {
  name = "%s"
}
`, rName)
}

func testAccDaxParameterGroupConfig_parameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_dax_parameter_group" "test" {
  name = "%s"

  parameters {
    name  = "query-ttl-millis"
    value = "100000"
  }

  parameters {
    name  = "record-ttl-millis"
    value = "100000"
  }
}
`, rName)
}
