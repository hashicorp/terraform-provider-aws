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

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	salt := acctest.RandString(5)
	resourceName := "aws_servicecatalog_product.test"
	var describeProductOutput servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", "", salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput, "product-"+salt),
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

func TestAccAWSServiceCatalogProduct_disappears(t *testing.T) {
	salt := acctest.RandString(5)
	resourceName := "aws_servicecatalog_product.test"
	var describeProductOutput servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", "", salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput),
					testAccCheckAwsServiceCatalogProductDisappears(&describeProductOutput),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_tags(t *testing.T) {
	salt := acctest.RandString(5)
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
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, tag1, "", salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput1, "product-"+salt),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag1key, tag1value),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, tag1B, tag2, salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput2, "product-"+salt),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag1key, tag1valueB),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag2key, tag2value),
					testAccCheckAwsServiceCatalogProductNotRecreated(&describeProductOutput1, &describeProductOutput2),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", tag2, salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput3),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput3, "product-"+salt),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag2key, tag2value),
					testAccCheckAwsServiceCatalogProductNotRecreated(&describeProductOutput2, &describeProductOutput3),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_ProvisioningArtifact_updateInPlace(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	resourceName := "aws_servicecatalog_product.test"

	var describeProductOutput1, describeProductOutput2 servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", "", salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput1, "product-"+salt),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", "pa-"+salt),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", "", salt, salt, salt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput2, "product-"+salt),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", "pa-"+salt2),
					testAccCheckAwsServiceCatalogProductNotRecreated(&describeProductOutput1, &describeProductOutput2),
				),
			},
		},
	})
}

// update the LoadFromTemplateURL to force recreating new Product
func TestAccAWSServiceCatalogProduct_Bucket_updateForcesNew(t *testing.T) {
	salt := acctest.RandString(5)
	salt2 := acctest.RandString(5)
	resourceName := "aws_servicecatalog_product.test"

	var describeProductOutput1, describeProductOutput2 servicecatalog.DescribeProductAsAdminOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", "", salt, salt, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput1, "product-"+salt),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(resourceName, "", "", salt, salt2, salt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput2, "product-"+salt),
					testAccCheckAwsServiceCatalogProductHasBeenRecreated(&describeProductOutput1, &describeProductOutput2),
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogProductExists(resourceName string, describeProductOutput *servicecatalog.DescribeProductAsAdminOutput) resource.TestCheckFunc {
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

func testAccCheckAwsServiceCatalogProductStandardFields(resourceName string, describeProductOutput *servicecatalog.DescribeProductAsAdminOutput, expectedProductName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", expectedProductName),
		resource.TestCheckResourceAttrSet(resourceName, "description"),
		resource.TestCheckResourceAttrSet(resourceName, "distributor"),
		resource.TestCheckResourceAttrSet(resourceName, "owner"),
		resource.TestCheckResourceAttrSet(resourceName, "product_type"),
		resource.TestCheckResourceAttrSet(resourceName, "support_description"),
		resource.TestCheckResourceAttrSet(resourceName, "support_email"),
		resource.TestCheckResourceAttrSet(resourceName, "support_url"),
		func(s *terraform.State) error {
			if *describeProductOutput.ProductViewDetail.ProductViewSummary.Name != expectedProductName {
				return fmt.Errorf("resource '%s' does not have expected name: '%s' vs '%s'", resourceName, *describeProductOutput.ProductViewDetail.ProductViewSummary.Name, expectedProductName)
			}
			return nil
		},
	)
}

func testAccCheckAwsServiceCatalogProductHasBeenRecreated(describeProductOutput1, describeProductOutput2 *servicecatalog.DescribeProductAsAdminOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *describeProductOutput1.ProductViewDetail.ProductARN == *describeProductOutput2.ProductViewDetail.ProductARN {
			return fmt.Errorf("Product ARN has not changed from %s, but it should have been re-created",
				*describeProductOutput1.ProductViewDetail.ProductARN)
		}
		return nil
	}
}

func testAccCheckAwsServiceCatalogProductNotRecreated(describeProductOutput1, describeProductOutput2 *servicecatalog.DescribeProductAsAdminOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *describeProductOutput1.ProductViewDetail.ProductARN != *describeProductOutput2.ProductViewDetail.ProductARN {
			return fmt.Errorf("Product ARN has changed from %s to %s, but it should not have been re-created",
				*describeProductOutput1.ProductViewDetail.ProductARN, *describeProductOutput2.ProductViewDetail.ProductARN)
		}
		return nil
	}
}

func testAccCheckAwsServiceCatalogProductDisappears(describeProductOutput *servicecatalog.DescribeProductAsAdminOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).scconn
		input := servicecatalog.DeleteProductInput{
			Id: describeProductOutput.ProductViewDetail.ProductViewSummary.ProductId,
		}
		_, err := conn.DeleteProduct(&input)
		if err != nil {
			return fmt.Errorf("could not delete product: %s", err)
		}
		return nil
	}
}

func testAccCheckAwsServiceCatalogProductDestroy(s *terraform.State) error {
	count := 0
	var rs *terraform.ResourceState
	for _, rsi := range s.RootModule().Resources {
		if rsi.Type != "aws_servicecatalog_product" {
			continue
		}
		count++
		rs = rsi
	}
	if count == 0 {
		// disappears test
		return nil
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

func testAccAWSServiceCatalogProductConfig_basic(resourceName, tag1, tag2, productSalt, bucketSalt, paSalt string) string {
	thisResourceParts := strings.Split(resourceName, ".")
	return fmt.Sprintf(`
data "aws_region" "current" {}	

resource "aws_s3_bucket" "bucket" {
  bucket        = "bucket-%[6]s"
  region        = data.aws_region.current.name
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "template1" {
  bucket  = aws_s3_bucket.bucket.id
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

resource "%[1]s" "%[2]s" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "product-%[5]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "pa-%[7]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

  tags = {
     %[3]s
     %[4]s
  }
}
`, thisResourceParts[0], thisResourceParts[1], tag1, tag2, productSalt, bucketSalt, paSalt)
}
