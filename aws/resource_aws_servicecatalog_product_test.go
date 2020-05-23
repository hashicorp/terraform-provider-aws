package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1 := "FooKey = \"bar\""
	tag2 := "BarKey = \"foo\""

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, tag2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "name", arbitraryProductName),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "provisioning_artifact.0.description"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.FooKey", "bar"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.BarKey", "foo"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "description"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "distributor"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "owner"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "product_type"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "support_description"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "support_email"),
					resource.TestCheckResourceAttrSet("aws_servicecatalog_product.test", "support_url"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.FooKey", "bar"),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", tag2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "tags.BarKey", "foo"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, newArbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", newArbitraryProvisionArtifactName),
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

	newArbitraryBucketName := fmt.Sprintf("bucket-new-%s", acctest.RandString(16))
	newArbitraryProvisionArtifactName := fmt.Sprintf("pa-new-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(newArbitraryBucketName, arbitraryProductName, newArbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr("aws_servicecatalog_product.test", "provisioning_artifact.0.name", newArbitraryProvisionArtifactName),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_import(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1 := "FooKey = \"bar\""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
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

func testAccCheckAwsServiceCatalogProductResourceConfigTemplate(bucketName, productName, provisioningArtifactName, tag1, tag2 string) string {
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

resource "aws_servicecatalog_product" "test" {
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
    info {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

  tags {
     %s
     %s
  }
}`, bucketName, productName, provisioningArtifactName, tag1, tag2)
}
