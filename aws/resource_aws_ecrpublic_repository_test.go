package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ecrpublic_repository", &resource.Sweeper{
		Name: "aws_ecrpublic_repository",
		F:    testSweepEcrPublicRepositories,
	})
}

func testSweepEcrPublicRepositories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).ecrpublicconn

	var errors error
	err = conn.DescribeRepositoriesPages(&ecrpublic.DescribeRepositoriesInput{}, func(page *ecrpublic.DescribeRepositoriesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, repository := range page.Repositories {
			repositoryName := aws.StringValue(repository.RepositoryName)
			log.Printf("[INFO] Deleting ECR Public repository: %s", repositoryName)

			_, err = conn.DeleteRepository(&ecrpublic.DeleteRepositoryInput{
				// We should probably sweep repositories even if there are images.
				Force:          aws.Bool(true),
				RegistryId:     repository.RegistryId,
				RepositoryName: repository.RepositoryName,
			})
			if err != nil {
				if !isAWSErr(err, ecrpublic.ErrCodeRepositoryNotFoundException, "") {
					sweeperErr := fmt.Errorf("Error deleting ECR Public repository (%s): %w", repositoryName, err)
					log.Printf("[ERROR] %s", sweeperErr)
					errors = multierror.Append(errors, sweeperErr)
				}
				continue
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ECR Public repository sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("Error retreiving ECR Public repositories: %w", err))
	}

	return errors
}

func TestAccAWSEcrPublicRepository_basic(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rName),
					testAccCheckAWSEcrPublicRepositoryRegistryID(resourceName),
					testAccCheckAWSEcrPublicRepositoryRepositoryARN(resourceName, rName),
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

func TestAccAWSEcrPublicRepository_catalogdata_abouttext(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigAboutText(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "About Text"),
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

func TestAccAWSEcrPublicRepository_catalogdata_architectures(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigArchitectures(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Linux"),
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

func TestAccAWSEcrPublicRepository_catalogdata_description(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "Description"),
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

func TestAccAWSEcrPublicRepository_catalogdata_operatingsystems(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigOperatingSystems(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "ARM"),
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

func TestAccAWSEcrPublicRepository_catalogdata_usagetext(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigUsageText(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "Usage Text"),
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

func TestAccAWSEcrPublicRepository_catalogdata_logoimageblob(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigLogoImageBlob(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "catalog_data.0.logo_image_url"),
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

func TestAccAWSEcrPublicRepository_disappears(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcrPublicRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEcrPublicRepositoryExists(name string, res *ecrpublic.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Public repository ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ecrpublicconn

		output, err := conn.DescribeRepositories(&ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		if len(output.Repositories) == 0 {
			return fmt.Errorf("ECR Public repository %s not found", rs.Primary.ID)
		}

		*res = *output.Repositories[0]

		return nil
	}
}

func testAccCheckAWSEcrPublicRepositoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrpublicconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecrpublic_repository" {
			continue
		}

		input := ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: []*string{aws.String(rs.Primary.Attributes["repository_name"])},
		}

		out, err := conn.DescribeRepositories(&input)

		if isAWSErr(err, ecrpublic.ErrCodeRepositoryNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		for _, repository := range out.Repositories {
			if aws.StringValue(repository.RepositoryName) == rs.Primary.Attributes["repository_name"] {
				return fmt.Errorf("ECR Public repository still exists: %s", rs.Primary.Attributes["repository_name"])
			}
		}
	}

	return nil
}

func testAccCheckAWSEcrPublicRepositoryRegistryID(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := testAccGetAccountID()
		return resource.TestCheckResourceAttr(resourceName, "registry_id", attributeValue)(s)
	}
}

func testAccCheckAWSEcrPublicRepositoryRepositoryARN(resourceName string, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := fmt.Sprintf("arn:%s:ecr-public::%s:repository/%s", testAccGetPartition(), testAccGetAccountID(), rName)
		return resource.TestCheckResourceAttr(resourceName, "repository_arn", attributeValue)(s)
	}
}

func testAccAWSEcrPublicRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigAboutText(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    about_text = "About Text"
  }
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigArchitectures(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    architectures = ["Linux"]
  }
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    description = "Description"
  }
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigOperatingSystems(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    operating_systems = ["ARM"]
  }
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigUsageText(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    usage_text = "Usage Text"
  }
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigLogoImageBlob(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    logo_image_blob = filebase64("test-fixtures/terraform_logo.png")
  }
}
`, rName)
}
