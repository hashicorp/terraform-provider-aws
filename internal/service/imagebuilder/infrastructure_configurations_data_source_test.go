package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccImageBuilderInfrastructureConfigurationsDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_infrastructure_configurations.test"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arns.0", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "names.0", resourceName, "name"),
				),
			},
		},
	})
}

func testAccInfrastructureConfigurationsDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  name                  = %[1]q
  instance_profile_name = aws_iam_instance_profile.test.name
}

data "aws_imagebuilder_infrastructure_configurations" "test" {
  filter {
    name   = "name"
    values = [aws_imagebuilder_infrastructure_configuration.test.name]
  }
}
`, rName)
}
