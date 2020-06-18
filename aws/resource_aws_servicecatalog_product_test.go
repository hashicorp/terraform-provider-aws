package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// TODO basic test with no tags
// TODO other tests with tags (update) - all tests - as per https://github.com/terraform-providers/terraform-provider-aws/blob/master/docs/contributing/contribution-checklists.md#resource-tagging-acceptance-testing-implementation
// TODO import test (all tests)
func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1 := "FooKey = \"bar\""
	tag2 := "BarKey = \"foo\""
	thisResourceFqn := "aws_servicecatalog_product.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "name", arbitraryProductName),
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "provisioning_artifact.0.description"),
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.%", "2"),
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.FooKey", "bar"),
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.BarKey", "foo"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "description"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "distributor"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "owner"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "product_type"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "support_description"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "support_email"),
					resource.TestCheckResourceAttrSet(thisResourceFqn, "support_url"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_updateTags(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1 := "FooKey = \"bar\""
	tag2 := "BarKey = \"foo\""
	thisResourceFqn := "aws_servicecatalog_product.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.%", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.FooKey", "bar"),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", tag2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.%", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "tags.BarKey", "foo"),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_updateProvisioningArtifactBasic(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	newArbitraryProvisionArtifactName := fmt.Sprintf("pa-new-%s", acctest.RandString(5))
	tag1 := "FooKey = \"bar\""
	thisResourceFqn := "aws_servicecatalog_product.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, newArbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.0.name", newArbitraryProvisionArtifactName),
				),
			},
		},
	})
}

// update the LoadFromTemplateURL to force recreating new Product
func TestAccAWSServiceCatalogProduct_updateProvisioningArtifactForceNew(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1 := "FooKey = \"bar\""
	thisResourceFqn := "aws_servicecatalog_product.test"

	newArbitraryBucketName := fmt.Sprintf("bucket-new-%s", acctest.RandString(16))
	newArbitraryProvisionArtifactName := fmt.Sprintf("pa-new-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, newArbitraryBucketName, arbitraryProductName, newArbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(thisResourceFqn, "provisioning_artifact.0.name", newArbitraryProvisionArtifactName),
				),
			},
		},
	})
}

// tests import, but function name can't include that word!
func TestAccAWSServiceCatalogProduct_read_in_existing(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1 := "FooKey = \"bar\""
	thisResourceFqn := "aws_servicecatalog_product.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckServiceCatalogProductDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product" {
			continue
		}
		input := servicecatalog.DescribeProductInput{}
		input.Id = aws.String(rs.Primary.ID)

		_, err := conn.DescribeProduct(&input)
		if err != nil {
			if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("product still exists")
	}

	return nil
}

func testAccCheckAwsServiceCatalogProductResourceConfigTemplate(thisResourceFqn, bucketName, productName, provisioningArtifactName, tag1, tag2 string) string {
	thisResourceParts := strings.Split(thisResourceFqn, ".")
	return fmt.Sprintf(`
data "aws_region" "current" { }

resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  region        = "${data.aws_region.current.name}"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = "${aws_s3_bucket.bucket.id}"
  key     = "test_templates_for_terraform_sc_dev1.json"
  content = <<EOF
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Test CF teamplate for Service Catalog terraform dev",
  "Resources": {
    "Empty": {
      "Type": "AWS::CloudFormation::WaitConditionHandle"
    }
  }
}
EOF
}

resource "%s" "%s" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "%s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "%s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

  tags = {
     %s
     %s
  }
}`, bucketName, thisResourceParts[0], thisResourceParts[1], productName, provisioningArtifactName, tag1, tag2)
}
