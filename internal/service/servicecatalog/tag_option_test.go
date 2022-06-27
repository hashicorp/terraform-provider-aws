package servicecatalog_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
)

// add sweeper to delete known test servicecat tag options

func TestAccServiceCatalogTagOption_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagOptionConfig_basic(rName, "värde", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagOptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceCatalogTagOption_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagOptionConfig_basic(rName, "värde", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagOptionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfservicecatalog.ResourceTagOption(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogTagOption_update(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// UpdateTagOption() is very particular about what it receives. Only fields that change should
	// be included or it will throw servicecatalog.ErrCodeDuplicateResourceException, "already exists"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagOptionConfig_basic(rName, "värde ett", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde ett"),
				),
			},
			{
				Config: testAccTagOptionConfig_basic(rName, "värde två", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde två"),
				),
			},
			{
				Config: testAccTagOptionConfig_basic(rName, "värde två", "active = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "key", rName), // cannot be updated in place
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde två"),
				),
			},
			{
				Config: testAccTagOptionConfig_basic(rName, "värde två", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName), // cannot be updated in place
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde två"),
				),
			},
			{
				Config: testAccTagOptionConfig_basic(rName2, "värde ett", "active = true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "key", rName2),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde ett"),
				),
			},
		},
	})
}

func TestAccServiceCatalogTagOption_notActive(t *testing.T) {
	resourceName := "aws_servicecatalog_tag_option.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTagOptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagOptionConfig_basic(rName, "värde ett", "active = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "active", "false"),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "value", "värde ett"),
				),
			},
		},
	})
}

func testAccCheckTagOptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_tag_option" {
			continue
		}

		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeTagOption(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Tag Option (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Tag Option (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTagOptionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		input := &servicecatalog.DescribeTagOptionInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTagOption(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Tag Option (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTagOptionConfig_basic(key, value, active string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_tag_option" "test" {
  key   = %[1]q
  value = %[2]q
  %[3]s
}
`, key, value, active)
}
