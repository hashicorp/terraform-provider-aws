package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAWSGlueCrawler_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists("aws_glue_catalog_crawler.test", "test"),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_crawler.test",
						"name",
						"test",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_crawler.test",
						"database_name",
						"db_name",
					),
					resource.TestCheckResourceAttr(
						"aws_glue_catalog_crawler.test",
						"role",
						"AWSGlueServiceRoleDefault",
					),
				),
			},
		},
	})
}

//func testAccCheckGlueDatabaseDestroy(s *terraform.State) error {
//	conn := testAccProvider.Meta().(*AWSClient).glueconn
//
//	for _, rs := range s.RootModule().Resources {
//		if rs.Type != "aws_glue_catalog_database" {
//			continue
//		}
//
//		catalogId, dbName := readAwsGlueCatalogID(rs.Primary.ID)
//
//		input := &glue.GetDatabaseInput{
//			CatalogId: aws.String(catalogId),
//			Name:      aws.String(dbName),
//		}
//		if _, err := conn.GetDatabase(input); err != nil {
//			//Verify the error is what we want
//			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
//				continue
//			}
//
//			return err
//		}
//		return fmt.Errorf("still exists")
//	}
//	return nil
//}

//func testAccGlueCatalogDatabase_basic(rInt int) string {
//	return fmt.Sprintf(`
//resource "aws_glue_catalog_database" "test" {
//  name = "my_test_catalog_database_%d"
//}
//`, rInt)
//}

//func testAccGlueCatalogDatabase_full(rInt int, desc string) string {
//	return fmt.Sprintf(`
//resource "aws_glue_catalog_database" "test" {
//  name = "my_test_catalog_database_%d"
//  description = "%s"
//  location_uri = "my-location"
//  parameters {
//	param1 = "value1"
//	param2 = true
//	param3 = 50
//  }
//}
//`, rInt, desc)
//}

//func testAccCheckGlueCatalogDatabaseExists(name string) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[name]
//		if !ok {
//			return fmt.Errorf("Not found: %s", name)
//		}
//
//		if rs.Primary.ID == "" {
//			return fmt.Errorf("No ID is set")
//		}
//
//		catalogId, dbName := readAwsGlueCatalogID(rs.Primary.ID)
//
//		glueconn := testAccProvider.Meta().(*AWSClient).glueconn
//		out, err := glueconn.GetDatabase(&glue.GetDatabaseInput{
//			CatalogId: aws.String(catalogId),
//			Name:      aws.String(dbName),
//		})
//
//		if err != nil {
//			return err
//		}
//
//		if out.Database == nil {
//			return fmt.Errorf("No Glue Database Found")
//		}
//
//		if *out.Database.Name != dbName {
//			return fmt.Errorf("Glue Database Mismatch - existing: %q, state: %q",
//				*out.Database.Name, dbName)
//		}
//
//		return nil
//	}
//}

func checkGlueCatalogCrawlerExists(name string, crawlerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		glueConn := testAccProvider.Meta().(*AWSClient).glueconn
		out, err := glueConn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(crawlerName),
		})

		if err != nil {
			return err
		}

		if out.Crawler == nil {
			return fmt.Errorf("no Glue Crawler found")
		}

		return nil
	}
}

const testAccGlueCrawlerConfigBasic = `
	resource "aws_glue_catalog_crawler" "test" {
	  name = "test"
	  database_name = "db_name"
	  role = "${aws_iam_role.glue.name}"
	  description = "TF-test-crawler"
	  schedule="cron(0 1 * * ? *)"
	  s3_target {
		path = "s3://bucket"
	  }
	}
	
	resource "aws_iam_role_policy_attachment" "aws-glue-service-role-default-policy-attachment" {
  		policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  		role = "${aws_iam_role.glue.name}"
	}
	
	resource "aws_iam_role" "glue" {
  		name = "AWSGlueServiceRoleDefault"
  		assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
	}
`
