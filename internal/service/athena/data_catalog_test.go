package athena_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
)

func TestAccDataCatalog_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "athena", fmt.Sprintf("datacatalog/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "description", "A test data catalog"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name",
					"parameters",
				},
			},
		},
	})
}

func TestAccDataCatalog_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfathena.ResourceDataCatalog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataCatalogConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
	name = %[1]q
	description = "A test data catalog"
	type = "LAMBDA"

	parameters = {
	  "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	}

	tags = {
	  Test = "test"
	}
}
`, rName)
}

func testAccCheckDataCatalogExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Athena Data Catalog name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.GetDataCatalogInput{
			Name: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDataCatalog(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckDataCatalogDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_data_catalog" {
			continue
		}

		input := &athena.GetDataCatalogInput{
			Name: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDataCatalog(input)

		if tfawserr.ErrMessageContains(err, athena.ErrCodeInvalidRequestException, "was not found") {
			continue
		}

		if err != nil {
			return err
		}

		if output.DataCatalog != nil {
			return fmt.Errorf("Athena Data Catalog (%s) found", rs.Primary.ID)
		}
	}

	return nil
}
