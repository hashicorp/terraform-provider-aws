package ram_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ram"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRAMResourceShareDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_resource_share.test"
	datasourceName := "data.aws_ram_resource_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceShareDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`No matching resource found`),
			},
			{
				Config: testAccResourceShareDataSourceConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "owning_account_id"),
				),
			},
		},
	})
}

func TestAccRAMResourceShareDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_resource_share.test"
	datasourceName := "data.aws_ram_resource_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccRAMResourceShareDataSource_status(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ram_resource_share.test"
	datasourceName := "data.aws_ram_resource_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareDataSourceConfig_status(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "owning_account_id"),
				),
			},
		},
	})
}

func testAccResourceShareDataSourceConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "wrong" {
  name = "%[1]s-wrong"
}

resource "aws_ram_resource_share" "test" {
  name = %[1]q
}

data "aws_ram_resource_share" "test" {
  name           = aws_ram_resource_share.test.name
  resource_owner = "SELF"
}
`, rName)
}

func testAccResourceShareDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    Name = "%[1]s-Tags"
  }
}

data "aws_ram_resource_share" "test" {
  name           = aws_ram_resource_share.test.name
  resource_owner = "SELF"

  filter {
    name   = "Name"
    values = ["%[1]s-Tags"]
  }
}
`, rName)
}

const testAccResourceShareDataSourceConfig_nonExistent = `
data "aws_ram_resource_share" "test" {
  name           = "tf-acc-test-does-not-exist"
  resource_owner = "SELF"
}
`

func testAccResourceShareDataSourceConfig_status(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = "%s"
}

data "aws_ram_resource_share" "test" {
  name                  = aws_ram_resource_share.test.name
  resource_owner        = "SELF"
  resource_share_status = "ACTIVE"
}
`, rName)
}
