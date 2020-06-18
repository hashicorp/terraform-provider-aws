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
func TestAccAWSServiceCatalogProduct_Basic(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	resourceName := "aws_servicecatalog_product.test"
	var describeProductOutput servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput),

					testAccCheckServiceCatalogProductStandardFields(resourceName, describeProductOutput, arbitraryProductName),

					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact.0.description"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSServiceCatalogProduct_Tags(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	tag1key := "FooKey"
	tag1value := "bar"
	tag1valueB := "bar2"
	tag1 := tag1key + " = \"" + tag1value + "\""
	tag1B := tag1key + " = \"" + tag1valueB + "\""
	tag2key := "BarKey"
	tag2value := "foo"
	tag2 := tag2key + " = \"" + tag2value + "\""
	resourceName := "aws_servicecatalog_product.test"

	var describeProductOutput1, describeProductOutput2, describeProductOutput3 servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckServiceCatalogProductStandardFields(resourceName, describeProductOutput1, arbitraryProductName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag1key, tag1value),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, tag1B, tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckServiceCatalogProductStandardFields(resourceName, describeProductOutput2, arbitraryProductName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag1key, tag1valueB),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag2key, tag2value),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", tag2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput3),
					testAccCheckServiceCatalogProductStandardFields(resourceName, describeProductOutput3, arbitraryProductName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag2key, tag2value),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_updateProvisioningArtifactBasic(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	resourceName := "aws_servicecatalog_product.test"

	newArbitraryProvisionArtifactName := fmt.Sprintf("pa-new-%s", acctest.RandString(5))

	var describeProductOutput1, describeProductOutput2 servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput1),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", arbitraryProvisionArtifactName),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, newArbitraryProvisionArtifactName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput2),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", newArbitraryProvisionArtifactName),
				),
			},
		},
	})
}

// update the LoadFromTemplateURL to force recreating new Product
func TestAccAWSServiceCatalogProduct_updateSourceBucketForceNew(t *testing.T) {
	arbitraryProductName := fmt.Sprintf("product-%s", acctest.RandString(5))
	arbitraryProvisionArtifactName := fmt.Sprintf("pa-%s", acctest.RandString(5))
	arbitraryBucketName := fmt.Sprintf("bucket-%s", acctest.RandString(16))
	resourceName := "aws_servicecatalog_product.test"

	newArbitraryBucketName := fmt.Sprintf("bucket-new-%s", acctest.RandString(16))

	var describeProductOutput1, describeProductOutput2 servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, arbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput1),
				),
			},
			{
				Config: testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, newArbitraryBucketName, arbitraryProductName, arbitraryProvisionArtifactName, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceCatalogProductExists(resourceName, &describeProductOutput2),
				),
			},
		},
	})
}

func testAccCheckServiceCatalogProductExists(resourceName string, describeProductOutput *servicecatalog.DescribeProductAsAdminOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' does not exist in local state", resourceName)
		}
		conn := testAccProvider.Meta().(*AWSClient).scconn
		resp, err := conn.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("resource '%s' id '%s' could not be read from AWS: %s", resourceName, rs.Primary.ID, err)
		}
		*describeProductOutput = *resp
		return nil
	}
}

func testAccCheckServiceCatalogProductStandardFields(resourceName string, describeProductOutput servicecatalog.DescribeProductAsAdminOutput, expectedProductName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", expectedProductName),
		resource.TestCheckResourceAttrSet(resourceName, "description"),
		resource.TestCheckResourceAttrSet(resourceName, "distributor"),
		resource.TestCheckResourceAttrSet(resourceName, "owner"),
		resource.TestCheckResourceAttrSet(resourceName, "product_type"),
		resource.TestCheckResourceAttrSet(resourceName, "support_description"),
		resource.TestCheckResourceAttrSet(resourceName, "support_email"),
		resource.TestCheckResourceAttrSet(resourceName, "support_url"),
	)
}

func testAccCheckServiceCatalogProductDestroy(s *terraform.State) error {
	count := 0
	var rs *terraform.ResourceState
	for _, rsi := range s.RootModule().Resources {
		if rsi.Type != "aws_servicecatalog_product" {
			continue
		}
		count++
		rs = rsi
	}
	if count != 1 {
		return fmt.Errorf("product count mismatch, found %d, expected 1", count)
	}

	conn := testAccProvider.Meta().(*AWSClient).scconn
	input := servicecatalog.DescribeProductInput{}
	input.Id = aws.String(rs.Primary.ID)

	_, err := conn.DescribeProduct(&input)
	if err == nil {
		return fmt.Errorf("product still exists")
	}
	if !isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
		return err
	}
	// not found at AWS, as expected once destroyed
	return nil
}

func testAccCheckAwsServiceCatalogProductResourceConfigTemplate(resourceName, bucketName, productName, provisioningArtifactName, tag1, tag2 string) string {
	thisResourceParts := strings.Split(resourceName, ".")
	return fmt.Sprintf(`
data "aws_region" "current" {}	
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  region        = data.aws_region.current.name
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
