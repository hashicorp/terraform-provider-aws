package servicecatalog_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
)

// add sweeper to delete known test servicecat products

func TestAccServiceCatalogProduct_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "distributor", "distributör"),
					resource.TestCheckResourceAttr(resourceName, "has_default_path", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", "ägare"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.disable_template_validation", "true"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_url"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
					resource.TestCheckResourceAttr(resourceName, "status", tfservicecatalog.StatusCreated),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_email", acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "support_url", domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func TestAccServiceCatalogProduct_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfig_basic(rName, rName, rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfservicecatalog.ResourceProduct(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogProduct_update(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccProductConfig_basic(rName, "ny beskrivning", "ny supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "ny beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "ny supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccServiceCatalogProduct_updateTags(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccProductConfig_updateTags(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "natural"),
				),
			},
		},
	})
}

func TestAccServiceCatalogProduct_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, servicecatalog.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfig_physicalID(rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_physical_id"),
					acctest.MatchResourceAttrRegionalARN(
						resourceName,
						"provisioning_artifact_parameters.0.template_physical_id",
						"cloudformation",
						regexp.MustCompile(fmt.Sprintf(`stack/%s/.*`, rName)),
					),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.type", servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func testAccCheckProductDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_product" {
			continue
		}

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeProductAsAdmin(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Product (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckProductExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn

		input := &servicecatalog.DescribeProductAsAdminInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeProductAsAdmin(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Product (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccProductTemplateURLBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}
`, rName)
}

func testAccProductConfig_basic(rName, description, supportDescription, domain, email string) string {
	return acctest.ConfigCompose(testAccProductTemplateURLBaseConfig(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = %[2]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[3]q
  support_email       = %[5]q
  support_url         = %[4]q

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, description, supportDescription, domain, email))
}

func testAccProductConfig_updateTags(rName, description, supportDescription, domain, email string) string {
	return acctest.ConfigCompose(testAccProductTemplateURLBaseConfig(rName), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = %[2]q
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = %[3]q
  support_email       = %[5]q
  support_url         = %[4]q

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = %[1]q
    template_url                = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Yak         = %[1]q
    Environment = "natural"
  }
}
`, rName, description, supportDescription, domain, email))
}

func testAccProductConfig_physicalID(rName, domain, email string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Resources = {
      MyVPC = {
        Type = "AWS::EC2::VPC"
        Properties = {
          CidrBlock = "10.1.0.0/16"
        }
      }
    }

    Outputs = {
      VpcID = {
        Description = "VPC ID"
        Value = {
          Ref = "MyVPC"
        }
      }
    }
  })
}

resource "aws_servicecatalog_product" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = %[3]q
  support_url         = %[2]q

  provisioning_artifact_parameters {
    description          = "artefaktbeskrivning"
    name                 = %[1]q
    template_physical_id = aws_cloudformation_stack.test.id
    type                 = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, domain, email)
}
