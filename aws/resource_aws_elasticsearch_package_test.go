package aws

import (
	"fmt"
	"testing"

	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSElasticsearchPackage_basic(t *testing.T) {
	rString := acctest.RandString(7)
	prefix := "tf-acctest-espackage"
	name := fmt.Sprintf("%s-%s", prefix, rString)
	resourceName := "aws_elasticsearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticsearchPackageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchPackageConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticsearchPackageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", elasticsearch.PackageTypeTxtDictionary),
					resource.TestCheckResourceAttr(resourceName, "description", "terraform acctest"),
					resource.TestCheckResourceAttr(resourceName, "source.0.s3_bucket_name", name+"-bucket"),
					resource.TestCheckResourceAttr(resourceName, "source.0.s3_key", "synonyms.txt"),
				),
			},
		},
	})
}

func TestAccAWSElasticsearchPackage_disappears(t *testing.T) {
	rString := acctest.RandString(7)
	prefix := "tf-acctest-espackage"
	name := fmt.Sprintf("%s-%s", prefix, rString)
	resourceName := "aws_elasticsearch_package.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticsearchPackageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticsearchPackageConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticsearchPackageExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsElasticsearchPackage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSElasticsearchPackageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elasticsearch Package ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).esconn

		details, err := getElasticsearchPackage(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if details == nil {
			return fmt.Errorf("Elasticsearch Package (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSElasticsearchPackageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).esconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_package" {
			continue
		}

		details, err := getElasticsearchPackage(conn, rs.Primary.ID)

		if isAWSErr(err, "ResourceNotFoundException", "") {
			continue
		}

		if err != nil {
			return err
		}

		if details == nil {
			continue
		}

		return fmt.Errorf("Elasticsearch Package (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSElasticsearchPackageConfig_basic(name string) string {
	return fmt.Sprintf(`
	resource "aws_s3_bucket" "test" {
		bucket = "%[1]s-bucket"
	}

	resource "aws_s3_bucket_object" "test" {
		bucket  = "${aws_s3_bucket.test.bucket}"
		key     = "synonyms.txt"
		content = "foo, bar"
	}

	resource "aws_elasticsearch_package" "test" {
		name         = "%[1]s"
		type         = "TXT-DICTIONARY"
		description = "terraform acctest"
		source  {
			s3_bucket_name = "${aws_s3_bucket.test.bucket}"
			s3_key         = "${aws_s3_bucket_object.test.key}"
		}
	}
	`, name)
}
