package athena_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
)

func TestAccAthenaDataCatalog_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
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
					resource.TestCheckResourceAttr(resourceName, "parameters.function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"), //lintignore:AWSAT003,AWSAT005
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		},
	})
}

func TestAccAthenaDataCatalog_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
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

func TestAccAthenaDataCatalog_tags(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccDataCatalogConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDataCatalogConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAthenaDataCatalog_type_lambda(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogTypeLambdaConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test data catalog using Lambda"),
					resource.TestCheckResourceAttr(resourceName, "type", "LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.metadata-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"), //lintignore:AWSAT003,AWSAT005
					resource.TestCheckResourceAttr(resourceName, "parameters.record-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"),   //lintignore:AWSAT003,AWSAT005
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		},
	})
}

func TestAccAthenaDataCatalog_type_hive(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogTypeHiveConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test data catalog using Hive"),
					resource.TestCheckResourceAttr(resourceName, "type", "HIVE"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.metadata-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"), //lintignore:AWSAT003,AWSAT005
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		},
	})
}

func TestAccAthenaDataCatalog_type_glue(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogTypeGlueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "A test data catalog using Glue"),
					resource.TestCheckResourceAttr(resourceName, "type", "GLUE"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.catalog-id", "123456789012"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		},
	})
}

func TestAccAthenaDataCatalog_parameters(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogParametersConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.function", "arn:aws:lambda:us-east-1:123456789012:function:test-function-1"), //lintignore:AWSAT003,AWSAT005
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccDataCatalogParametersUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.metadata-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"), //lintignore:AWSAT003,AWSAT005
					resource.TestCheckResourceAttr(resourceName, "parameters.record-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"),   //lintignore:AWSAT003,AWSAT005
				),
			},
		},
	})
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

func testAccDataCatalogConfig(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog"
  type        = "LAMBDA"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }
}
`, rName)
}

func testAccDataCatalogConfigTags1(rName, tagKey1, tagValue1 string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name = %[1]q
  type = "LAMBDA"

  description = "Testing tags"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDataCatalogConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name = %[1]q
  type = "LAMBDA"

  description = "Testing tags"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDataCatalogTypeLambdaConfig(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog using Lambda"
  type        = "LAMBDA"

  parameters = {
    "metadata-function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
    "record-function"   = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }
}
`, rName)
}

func testAccDataCatalogTypeHiveConfig(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog using Hive"
  type        = "HIVE"

  parameters = {
    "metadata-function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }
}
`, rName)
}

func testAccDataCatalogTypeGlueConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog using Glue"
  type        = "GLUE"

  parameters = {
    "catalog-id" = "123456789012"
  }
}
`, rName)
}

func testAccDataCatalogParametersConfig(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "Testing parameters attribute"
  type        = "LAMBDA"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function-1"
  }
}
`, rName)
}

func testAccDataCatalogParametersUpdatedConfig(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "Testing parameters attribute"
  type        = "LAMBDA"

  parameters = {
    "metadata-function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"
    "record-function"   = "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"
  }
}
`, rName)
}
