package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_domain", &resource.Sweeper{
		Name: "aws_sagemaker_domain",
		F:    testSweepSagemakerDomains,
	})
}

func testSweepSagemakerDomains(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListDomainsPages(&sagemaker.ListDomainsInput{}, func(page *sagemaker.ListDomainsOutput, lastPage bool) bool {
		for _, instance := range page.Domains {
			domainArn := aws.StringValue(instance.DomainArn)
			domainID, err := decodeSagemakerDomainID(domainArn)
			if err != nil {
				log.Printf("[ERROR] Error parsing sagemaker domain arn (%s): %s", domainArn, err)
			}
			input := &sagemaker.DeleteDomainInput{
				DomainId: aws.String(domainID),
			}

			log.Printf("[INFO] Deleting SageMaker domain: %s", domainArn)
			if _, err := conn.DeleteDomain(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker domain (%s): %s", domainArn, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker domains: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerDomain_basic(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("domain/%s", rName)),
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

// func TestAccAWSSagemakerDomain_gitConfig_branch(t *testing.T) {
// 	var notebook sagemaker.DescribeDomainOutput
// 	rName := acctest.RandomWithPrefix("tf-acc-test")
// 	resourceName := "aws_sagemaker_domain.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSSagemakerDomainGitConfigBranchConfig(rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
// 					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
// 					testAccCheckResourceAttrRegionalARN(resourceName0......, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/terraform-providers/terraform-provider-aws.git"),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.0.branch", "master"),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 		},
// 	})
// }

// func TestAccAWSSagemakerDomain_gitConfig_secret(t *testing.T) {
// 	var notebook sagemaker.DescribeDomainOutput
// 	rName := acctest.RandomWithPrefix("tf-acc-test")
// 	resourceName := "aws_sagemaker_domain.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAWSSagemakerDomainGitConfigSecretConfig(rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
// 					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
// 					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/terraform-providers/terraform-provider-aws.git"),
// 					resource.TestCheckResourceAttrPair(resourceName, "git_config.0.secret_arn", "aws_secretsmanager_secret.test", "arn"),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 			{
// 				Config: testAccAWSSagemakerDomainGitConfigSecretUpdatedConfig(rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
// 					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
// 					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
// 					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/terraform-providers/terraform-provider-aws.git"),
// 					resource.TestCheckResourceAttrPair(resourceName, "git_config.0.secret_arn", "aws_secretsmanager_secret.test2", "arn"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccAWSSagemakerDomain_disappears(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_domain" {
			continue
		}

		domain, err := finder.DomainByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		domainArn := aws.StringValue(domain.DomainArn)
		domainID, err := decodeSagemakerDomainID(domainArn)
		if err != nil {
			return err
		}

		if domainID == rs.Primary.ID {
			return fmt.Errorf("sagemaker domain %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerDomainExists(n string, codeRepo *sagemaker.DescribeDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.DomainByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccAWSSagemakerDomainConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccAWSSagemakerDomainBasicConfig(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}
`, rName)
}

// func testAccAWSSagemakerDomainGitConfigBranchConfig(rName string) string {
// 	return fmt.Sprintf(`
// resource "aws_sagemaker_domain" "test" {
//   domain_name = %[1]q

//   git_config {
//     repository_url = "https://github.com/terraform-providers/terraform-provider-aws.git"
//     branch         = "master"
//   }
// }
// `, rName)
// }

// func testAccAWSSagemakerDomainGitConfigSecretConfig(rName string) string {
// 	return fmt.Sprintf(`
// resource "aws_secretsmanager_secret" "test" {
//   name = %[1]q
// }

// resource "aws_secretsmanager_secret_version" "test" {
//   secret_id     = aws_secretsmanager_secret.test.id
//   secret_string = jsonencode({ username = "example", passowrd = "example" })
// }

// resource "aws_sagemaker_domain" "test" {
//   domain_name = %[1]q

//   git_config {
//     repository_url = "https://github.com/terraform-providers/terraform-provider-aws.git"
//     secret_arn     = aws_secretsmanager_secret.test.arn
//   }

//   depends_on = [aws_secretsmanager_secret_version.test]
// }
// `, rName)
// }

// func testAccAWSSagemakerDomainGitConfigSecretUpdatedConfig(rName string) string {
// 	return fmt.Sprintf(`
// resource "aws_secretsmanager_secret" "test2" {
//   name = "%[1]s-2"
// }

// resource "aws_secretsmanager_secret_version" "test2" {
//   secret_id     = aws_secretsmanager_secret.test2.id
//   secret_string = jsonencode({ username = "example", passowrd = "example" })
// }

// resource "aws_sagemaker_domain" "test" {
//   domain_name = %[1]q

//   git_config {
//     repository_url = "https://github.com/terraform-providers/terraform-provider-aws.git"
//     secret_arn     = aws_secretsmanager_secret.test2.arn
//   }

//   depends_on = [aws_secretsmanager_secret_version.test2]
// }
// `, rName)
// }
