package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGlueUserDefinedFunction_basic(t *testing.T) {
	var udf glue.UserDefinedFunction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	updated := "test"
	resourceName := "aws_glue_user_defined_function.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueUDFDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccGlueUserDefinedFunctionBasicConfig(rName, rName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueUserDefinedFunctionExists(resourceName, &udf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "class_name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner_name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner_type", "GROUP"),
					resource.TestCheckResourceAttr(resourceName, "resource_uris.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:  testAccGlueUserDefinedFunctionBasicConfig(rName, updated),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueUserDefinedFunctionExists(resourceName, &udf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "class_name", updated),
					resource.TestCheckResourceAttr(resourceName, "owner_name", updated),
					resource.TestCheckResourceAttr(resourceName, "owner_type", "GROUP"),
					resource.TestCheckResourceAttr(resourceName, "resource_uris.#", "1"),
				),
			},
		},
	})
}

func testAccCheckGlueUDFDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_user_defined_function" {
			continue
		}

		catalogId, dbName, funcName, err := readAwsGlueUDFID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &glue.GetUserDefinedFunctionInput{
			CatalogId:    aws.String(catalogId),
			DatabaseName: aws.String(dbName),
			FunctionName: aws.String(funcName),
		}
		if _, err := conn.GetUserDefinedFunction(input); err != nil {
			//Verify the error is what we want
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccGlueUserDefinedFunctionBasicConfig(rName string, name string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_user_defined_function" "test" {
  name          = %[1]q
  catalog_id    = "${aws_glue_catalog_database.test.catalog_id}"
  database_name = "${aws_glue_catalog_database.test.name}"
  class_name    = %[2]q
  owner_name    = %[2]q
  owner_type    = "GROUP"

  resource_uris {
    resource_type = "ARCHIVE"
    uri           = %[2]q
  }
}
`, rName, name)
}

func testAccCheckGlueUserDefinedFunctionExists(name string, udf *glue.UserDefinedFunction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, funcName, err := readAwsGlueUDFID(rs.Primary.ID)
		if err != nil {
			return err
		}

		glueconn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueconn.GetUserDefinedFunction(&glue.GetUserDefinedFunctionInput{
			CatalogId:    aws.String(catalogId),
			DatabaseName: aws.String(dbName),
			FunctionName: aws.String(funcName),
		})

		if err != nil {
			return err
		}

		if out.UserDefinedFunction == nil {
			return fmt.Errorf("No Glue Database Found")
		}

		if *out.UserDefinedFunction.FunctionName != funcName {
			return fmt.Errorf("Glue UDF Mismatch - existing: %q, state: %q",
				*out.UserDefinedFunction.FunctionName, funcName)
		}

		*udf = *out.UserDefinedFunction

		return nil
	}
}
