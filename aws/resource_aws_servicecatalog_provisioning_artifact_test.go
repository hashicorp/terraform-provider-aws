package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// add sweeper to delete known test servicecat provisioning artifacts
func init() {
	resource.AddTestSweepers("aws_servicecatalog_provisioning_artifact", &resource.Sweeper{
		Name:         "aws_servicecatalog_provisioning_artifact",
		Dependencies: []string{},
		F:            testSweepServiceCatalogProvisioningArtifacts,
	})
}

func testSweepServiceCatalogProvisioningArtifacts(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).scconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &servicecatalog.SearchProductsAsAdminInput{}

	err = conn.SearchProductsAsAdminPages(input, func(page *servicecatalog.SearchProductsAsAdminOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pvd := range page.ProductViewDetails {
			if pvd == nil || pvd.ProductViewSummary == nil || pvd.ProductViewSummary.ProductId == nil {
				continue
			}

			productID := aws.StringValue(pvd.ProductViewSummary.ProductId)

			artInput := &servicecatalog.ListProvisioningArtifactsInput{
				ProductId: aws.String(productID),
			}

			// there's no paginator for ListProvisioningArtifacts
			for {
				output, err := conn.ListProvisioningArtifacts(artInput)

				if err != nil {
					err := fmt.Errorf("error listing Service Catalog Provisioning Artifacts for product (%s): %w", productID, err)
					log.Printf("[ERROR] %s", err)
					errs = multierror.Append(errs, err)
					break
				}

				for _, pad := range output.ProvisioningArtifactDetails {
					r := resourceAwsServiceCatalogProvisioningArtifact()
					d := r.Data(nil)

					d.SetId(aws.StringValue(pad.Id))
					d.Set("product_id", productID)

					sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
				}

				/*
					// Currently, the API has no token field on input (AWS oops)
					if aws.StringValue(output.NextPageToken) == "" {
						break
					}

					artInput.NextPageToken = output.NextPageToken
				*/
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Service Catalog Provisioning Artifacts for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Service Catalog Provisioning Artifacts for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Service Catalog Provisioning Artifacts sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSServiceCatalogProvisioningArtifact_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName, "beskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "disable_template_validation", "true"),
					resource.TestCheckResourceAttr(resourceName, "guidance", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", "aws_servicecatalog_product.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.ProductTypeCloudFormationTemplate),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"disable_template_validation",
					"template_url",
				},
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisioningArtifact_disappears(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceCatalogProvisioningArtifact(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisioningArtifact_update(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName, "beskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName, "ny beskrivning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "ny beskrivning"),
					resource.TestCheckResourceAttr(resourceName, "support_description", "ny supportbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSServiceCatalogProvisioningArtifact_physicalID(t *testing.T) {
	resourceName := "aws_servicecatalog_provisioning_artifact.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogProvisioningArtifactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogProvisioningArtifactConfig_physicalID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "catalog", regexp.MustCompile(`product/prod-.*`)),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.description", "artefaktbeskrivning"),
					resource.TestCheckResourceAttr(resourceName, "provisioning_artifact_parameters.0.name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_artifact_parameters.0.template_physical_id"),
					testAccMatchResourceAttrRegionalARN(
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

func testAccCheckAwsServiceCatalogProvisioningArtifactDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_provisioning_artifact" {
			continue
		}

		artifactID, productID, err := parseServiceCatalogProvisioningArtifactID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", rs.Primary.ID, err)
		}

		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
		}

		output, err := conn.DescribeProvisioningArtifact(input)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Provisioning Artifact (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Provisioning Artifact (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogProvisioningArtifactExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		artifactID, productID, err := parseServiceCatalogProvisioningArtifactID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", rs.Primary.ID, err)
		}

		input := &servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
		}

		_, err = conn.DescribeProvisioningArtifact(input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Provisioning Artifact (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogProvisioningArtifactConfigTemplateURLBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.json"

  content = <<EOF
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
EOF
}

data "aws_partition" "current" {}

resource "aws_servicecatalog_product" "test" {
  description         = "Produktbeskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = "support@example.com"
  support_url         = "http://example.com"

  provisioning_artifact_parameters {
    description                 = "artefaktbeskrivning"
    disable_template_validation = true
    name                        = "%[1]s-1"
    template_url                = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
    #template_url                = "https://s3.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
    type                        = "CLOUD_FORMATION_TEMPLATE"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSServiceCatalogProvisioningArtifactConfig_basic(rName, description string) string {
	return composeConfig(testAccAWSServiceCatalogProvisioningArtifactConfigTemplateURLBase(rName), fmt.Sprintf(`
resource "aws_servicecatalog_provisioning_artifact" "test" {
  accept_language             = "en"
  active                      = "true"
  description                 = %[2]q
  disable_template_validation = true
  guidance                    = "DEFAULT"
  name                        = "%[1]s-2"
  product_id                  = aws_servicecatalog_product.test.id
  template_url                = "https://s3.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket.test.id}/${aws_s3_bucket_object.test.key}"
  type                        = "CLOUD_FORMATION_TEMPLATE"
}
`, rName, description))
}

func testAccAWSServiceCatalogProvisioningArtifactConfig_physicalID(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}

resource "aws_servicecatalog_provisioning_artifact" "test" {
  description         = "beskrivning"
  distributor         = "distributör"
  name                = %[1]q
  owner               = "ägare"
  type                = "CLOUD_FORMATION_TEMPLATE"
  support_description = "supportbeskrivning"
  support_email       = "support@example.com"
  support_url         = "http://example.com"
  description          = "artefaktbeskrivning"
  name                 = %[1]q
  template_physical_id = aws_cloudformation_stack.test.id
  type                 = "CLOUD_FORMATION_TEMPLATE"
}
`, rName)
}
