package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/glue"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/aws/aws-sdk-go/aws"
)

func TestAccAWSGlueCatalogDatabase_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGlueDatabaseDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccGlueCatalogDatabase_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlueCatalogDatabaseExists("aws_glue_catalog_database.test"),
				),
			},
		},
	})
}

func testAccCheckGlueDatabaseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).glueconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_catalog_database" {
			continue
		}

		input := &glue.GetDatabaseInput{
			Name: aws.String(rs.Primary.ID),
		}
		if _, err := conn.GetDatabase(input); err != nil {
			//Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "EntityNotFoundException" {
				continue
			}

			return err
		}
		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccGlueCatalogDatabase_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "my_test_catalog_database_%d"
}
`, rInt)
}

func testAccCheckGlueCatalogDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		glueconn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueconn.GetDatabase(&glue.GetDatabaseInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.Database == nil {
			return fmt.Errorf("No Glue Database Found")
		}

		if *out.Database.Name != rs.Primary.ID {
			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
				*out.Database, rs.Primary.ID)
		}

		return nil
	}
}
