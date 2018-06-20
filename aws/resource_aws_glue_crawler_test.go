package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAWSGlueCrawler_basic(t *testing.T) {
	const name = "aws_glue_crawler.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists(name),
					resource.TestCheckResourceAttr(name, "database_name", "test_db"),
					resource.TestCheckResourceAttr(name, "role", "AWSGlueServiceRole-tf"),
				),
			},
		},
	})
}

func TestAccAWSGlueCrawler_jdbcCrawler(t *testing.T) {
	const name = "aws_glue_crawler.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigJdbc(rName),
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists(name),
					resource.TestCheckResourceAttr(name, "database_name", "test_db"),
					resource.TestCheckResourceAttr(name, "role", "AWSGlueServiceRoleDefault"),
					resource.TestCheckResourceAttr(name, "jdbc_target.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSGlueCrawler_customCrawlers(t *testing.T) {
	const name = "aws_glue_crawler.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueCrawlerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueCrawlerConfigCustomClassifiers(rName),
				Check: resource.ComposeTestCheckFunc(
					checkGlueCatalogCrawlerExists(name),
					resource.TestCheckResourceAttr(name, "database_name", "test_db"),
					resource.TestCheckResourceAttr(name, "role", "AWSGlueServiceRoleDefault"),
					resource.TestCheckResourceAttr(name, "table_prefix", "table_prefix"),
					resource.TestCheckResourceAttr(name, "schema_change_policy.0.delete_behavior", "DELETE_FROM_DATABASE"),
					resource.TestCheckResourceAttr(name, "schema_change_policy.0.update_behavior", "UPDATE_IN_DATABASE"),
					resource.TestCheckResourceAttr(name, "s3_target.#", "2"),
					resource.TestCheckResourceAttr(name, "s3_target.#", "2"),
				),
			},
		},
	})
}

func checkGlueCatalogCrawlerExists(name string) resource.TestCheckFunc {
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
			Name: aws.String(rs.Primary.ID),
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

func testAccCheckAWSGlueCrawlerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_crawler" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := conn.GetCrawler(&glue.GetCrawlerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}

		}

		crawler := output.Crawler
		if crawler != nil && aws.StringValue(crawler.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Crawler %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccGlueCrawlerConfigBasic(rName string) string {
	return fmt.Sprintf(`
	resource "aws_glue_catalog_database" "test_db" {
  		name = "test_db"
	}

	resource "aws_glue_crawler" "test" {
	  name = "%s"
	  database_name = "${aws_glue_catalog_database.test_db.name}"
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
  		name = "AWSGlueServiceRole-tf"
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
`, rName)
}

func testAccGlueCrawlerConfigJdbc(rName string) string {
	return fmt.Sprintf(`
	resource "aws_glue_catalog_database" "test_db" {
  		name = "test_db"
	}

	resource "aws_glue_connection" "test" {
  		connection_properties = {
    		JDBC_CONNECTION_URL = "jdbc:mysql://terraformacctesting.com/testdatabase"
    		PASSWORD            = "testpassword"
    		USERNAME            = "testusername"
  		}
  		description = "tf_test_jdbc_connection_description"
  		name        = "tf_test_jdbc_connection"
	}
	
	resource "aws_iam_role_policy_attachment" "aws-glue-service-full-console-attachment" {
  		policy_arn = "arn:aws:iam::aws:policy/AWSGlueConsoleFullAccess"
  		role = "${aws_iam_role.glue.name}"
	}

	resource "aws_iam_role_policy_attachment" "aws-glue-service-role-service-attachment" {
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

	resource "aws_glue_crawler" "test" {
	  name = "%s"
	  database_name = "${aws_glue_catalog_database.test_db.name}"
	  role = "${aws_iam_role.glue.name}"
	  description = "TF-test-crawler"
	  schedule="cron(0 1 * * ? *)"
	  jdbc_target {
		path = "s3://bucket"
		connection_name = "${aws_glue_connection.test.name}"
	  }
	}
`, rName)
}

func testAccGlueCrawlerConfigCustomClassifiers(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test_db" {
  name = "test_db"
}

resource "aws_glue_classifier" "test" {
  name = "tf-example-123"
  grok_classifier {
    classification = "example"
    grok_pattern   = "example"
  }
}

resource "aws_glue_crawler" "test" {
  name = "%s"
  database_name = "${aws_glue_catalog_database.test_db.name}"
  role = "${aws_iam_role.glue.name}"
  classifiers = [
    "${aws_glue_classifier.test.id}"
  ]
  s3_target {
    path = "s3://bucket1"
    exclusions = [
      "s3://bucket1/foo"
    ]
  }
  s3_target {
    path = "s3://bucket2"
  }
  table_prefix = "table_prefix"
  schema_change_policy {
    delete_behavior = "DELETE_FROM_DATABASE"
    update_behavior = "UPDATE_IN_DATABASE"
  }

  configuration = <<EOF
{
  "Version": 1.0,
  "CrawlerOutput": {
    "Partitions": {
      "AddOrUpdateBehavior": "InheritFromTable"
    }
  }
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws-glue-service-full-console-attachment" {
  policy_arn = "arn:aws:iam::aws:policy/AWSGlueConsoleFullAccess"
  role = "${aws_iam_role.glue.name}"
}

resource "aws_iam_role_policy_attachment" "aws-glue-service-role-service-attachment" {
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
`, rName)
}
