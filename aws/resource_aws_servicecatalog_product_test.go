package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSServiceCatalogProduct_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"
	saltedName := acctest.RandomWithPrefix("tf-acc-test")
	var describeProductOutput servicecatalog.DescribeProductAsAdminOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput, saltedName),
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
	resourceName := "aws_servicecatalog_product.test"
	saltedName := acctest.RandomWithPrefix("tf-acc-test")
	var describeProductOutput servicecatalog.DescribeProductAsAdminOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_tags(t *testing.T) {
	saltedName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_servicecatalog_product.test"

	var describeProductOutput1, describeProductOutput2, describeProductOutput3 servicecatalog.DescribeProductAsAdminOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, "key1 = \"value1\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput1, saltedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, "key1 = \"value1updated\""+"\n"+"key2 = \"value2\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput2, saltedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckAwsServiceCatalogProductNotRecreated(&describeProductOutput1, &describeProductOutput2),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, "key2 = \"value2\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput3),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput3, saltedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckAwsServiceCatalogProductNotRecreated(&describeProductOutput2, &describeProductOutput3),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProduct_ProvisioningArtifact_updateInPlace(t *testing.T) {
	saltedName := acctest.RandomWithPrefix("tf-acc-test")
	saltedName2 := acctest.RandomWithPrefix("tf-acc-test2")
	resourceName := "aws_servicecatalog_product.test"

	var describeProductOutput1, describeProductOutput2 servicecatalog.DescribeProductAsAdminOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput1, saltedName),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", saltedName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput2, saltedName),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact.0.name", saltedName2),
					testAccCheckAwsServiceCatalogProductNotRecreated(&describeProductOutput1, &describeProductOutput2),
				),
			},
		},
	})
}

// update the LoadFromTemplateURL to force recreating new Product
func TestAccAWSServiceCatalogProduct_Bucket_updateForcesNew(t *testing.T) {
	saltedName := acctest.RandomWithPrefix("tf-acc-test")
	saltedName2 := acctest.RandomWithPrefix("tf-acc-test2")
	resourceName := "aws_servicecatalog_product.test"

	var describeProductOutput1, describeProductOutput2 servicecatalog.DescribeProductAsAdminOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName, saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput1),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput1, saltedName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProductConfig_basic(saltedName, saltedName2, saltedName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProductExists(resourceName, &describeProductOutput2),
					testAccCheckAwsServiceCatalogProductStandardFields(resourceName, &describeProductOutput2, saltedName),
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

func testAccAWSServiceCatalogProductConfig_basic(productSaltedName, bucketSaltedName, paSaltedName, tags string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%[2]s"
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

resource "aws_servicecatalog_product" "test" {
  description         = "arbitrary product description"
  distributor         = "arbitrary distributor"
  name                = "%[1]s"
  owner               = "arbitrary owner"
  product_type        = "CLOUD_FORMATION_TEMPLATE"
  support_description = "arbitrary support description"
  support_email       = "arbitrary@email.com"
  support_url         = "http://arbitrary_url/foo.html"

  provisioning_artifact {
    description = "arbitrary description"
    name        = "%[3]s"
    info = {
      LoadTemplateFromURL = "https://s3.amazonaws.com/${aws_s3_bucket.bucket.id}/${aws_s3_bucket_object.template1.key}"
    }
  }

  tags = {
     %[4]s
  }
}
`, productSaltedName, bucketSaltedName, paSaltedName, tags)
}
