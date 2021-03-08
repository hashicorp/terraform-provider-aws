package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_eks_identity_provider_config", &resource.Sweeper{
		Name: "aws_eks_identity_provider_config",
		F:    testSweepEksIdentityProviderConfigs,
	})
}

func testSweepEksIdentityProviderConfigs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).eksconn

	var errors error
	input := &eks.ListClustersInput{}
	err = conn.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		for _, cluster := range page.Clusters {
			clusterName := aws.StringValue(cluster)
			input := &eks.ListIdentityProviderConfigsInput{
				ClusterName: cluster,
			}
			err := conn.ListIdentityProviderConfigsPages(input, func(page *eks.ListIdentityProviderConfigsOutput, lastPage bool) bool {
				for _, config := range page.IdentityProviderConfigs {
					configName := aws.StringValue(config.Name)
					log.Printf("[INFO] Disassociating Identity Provider Config %q", configName)
					input := &eks.DisassociateIdentityProviderConfigInput{
						ClusterName:            cluster,
						IdentityProviderConfig: config,
					}
					_, err := conn.DisassociateIdentityProviderConfig(input)

					if err != nil && !isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
						errors = multierror.Append(errors, fmt.Errorf("error disassociating Identity Provider Config %q: %w", configName, err))
						continue
					}

					// TODO - Should I use context.Background here?
					if err := waitForEksIdentityProviderConfigDisassociation(context.Background(), conn, clusterName, config, 10*time.Minute); err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EKS Identity Provider Config %q disassociation: %w", configName, err))
						continue
					}
				}
				return true
			})
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error listing Identity Provider Configs for EKS Cluster %s: %w", clusterName, err))
			}
		}

		return true
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EKS Clusters sweep for %s: %s", region, err)
		return errors // In case we have completed some pages, but had errors
	}
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EKS Clusters: %w", err))
	}

	return errors
}

func TestAccAWSEksIdentityProviderConfig_basic(t *testing.T) {
	var config eks.IdentityProviderConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	eksClusterResourceName := "aws_eks_cluster.test"
	resourceName := "aws_eks_identity_provider_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEksIdentityProviderConfigProvider_Oidc_IssuerUrl(rName, "http://accounts.google.com/.well-known/openid-configuration"),
				ExpectError: regexp.MustCompile(`expected .* to have a url with schema of: "https", got http://accounts.google.com/.well-known/openid-configuration`),
			},
			{
				Config: testAccAWSEksIdentityProviderConfigProviderConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", eksClusterResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "oidc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.client_id", "test-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.identity_provider_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.issuer_url", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.%", "0"),
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

func TestAccAWSEksIdentityProviderConfig_disappears(t *testing.T) {
	var config eks.IdentityProviderConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfigProviderConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					testAccCheckAWSEksIdentityProviderConfigDisappears(rName, &config),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEksIdentityProviderConfig_Oidc_Group(t *testing.T) {
	var config eks.IdentityProviderConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfigProvider_Oidc_Groups(rName, "groups", "oidc:"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "oidc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.groups_claim", "groups"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.groups_prefix", "oidc:"),
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

func TestAccAWSEksIdentityProviderConfig_Oidc_Username(t *testing.T) {
	var config eks.IdentityProviderConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfigProvider_Oidc_Username(rName, "email", "-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "oidc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.username_claim", "email"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.username_prefix", "-"),
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

func TestAccAWSEksIdentityProviderConfig_Oidc_RequiredClaims(t *testing.T) {
	var config eks.IdentityProviderConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEksIdentityProviderConfig_Oidc_RequiredClaims(rName, "4qkvw9k2RbpSO6wgCbynY10T6Rc2n89PQblyi6bZ5VhfpMr6V7FVvrA12FiJxarh", "valueOne", "keyTwo", "valueTwo"),
				ExpectError: regexp.MustCompile("Bad map key length"),
			},
			{
				Config:      testAccAWSEksIdentityProviderConfig_Oidc_RequiredClaims(rName, "keyOne", "bUUUiuIXeFGw0M2VwiCVjR8oIIavv0PF49Ba6yNwOOC7IcoLawczSeb6MpEIhqtXKcf9aogW4uc4smLGvdTQ8uTTkVFvQTPyWXQ3F0uZP02YyoSw0d9MZ7laGRjpXSph9oFE2UlT5IyRaXIsTwl1qvItvVXLN40Pd3PDyPa6de4nlYcRNy6YIikZz2P1QUSYuvMGSJxGUzhTKYRUniolIt1vjHsXt3MAsaJtCcWz0tjLWalvG27pQ3Gl5Cs7K1", "keyTwo", "valueTwo"),
				ExpectError: regexp.MustCompile("Bad map value length"),
			},
			{
				Config: testAccAWSEksIdentityProviderConfig_Oidc_RequiredClaims(rName, "keyOne", "valueOne", "keyTwo", "valueTwo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.keyOne", "valueOne"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.required_claims.keyTwo", "valueTwo"),
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

func TestAccAWSEksIdentityProviderConfig_Tags(t *testing.T) {
	var config eks.IdentityProviderConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eks_identity_provider_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEksIdentityProviderConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksIdentityProviderConfig_Tags(rName, "keyOne", "valueOne", "keyTwo", "valueTwo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.keyOne", "valueOne"),
					resource.TestCheckResourceAttr(resourceName, "tags.keyTwo", "valueTwo"),
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

func testAccCheckAWSEksIdentityProviderConfigExists(resourceName string, config *eks.IdentityProviderConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EKS Identity Profile Config is set")
		}

		clusterName, configName, err := resourceAwsEksIdentityProviderConfigParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DescribeIdentityProviderConfigInput{
			ClusterName: aws.String(clusterName),
			IdentityProviderConfig: &eks.IdentityProviderConfig{
				Name: aws.String(configName),
				Type: aws.String(typeOidc),
			},
		}

		output, err := conn.DescribeIdentityProviderConfig(input)

		if err != nil {
			return err
		}

		if output == nil || output.IdentityProviderConfig == nil {
			return fmt.Errorf("EKS Identity Provider Config (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.IdentityProviderConfig.Oidc.IdentityProviderConfigName) != configName {
			return fmt.Errorf("EKS OIDC Identity Provider Config (%s) not found", rs.Primary.ID)
		}

		if got, want := aws.StringValue(output.IdentityProviderConfig.Oidc.Status), eks.ConfigStatusActive; got != want {
			return fmt.Errorf("EKS OIDC Identity Provider Config (%s) not in %s status, got: %s", rs.Primary.ID, want, got)
		}

		*config = eks.IdentityProviderConfig{
			Name: output.IdentityProviderConfig.Oidc.IdentityProviderConfigName,
			Type: aws.String(typeOidc),
		}

		return nil
	}
}

func testAccCheckAWSEksIdentityProviderConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).eksconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eks_identity_provider_config" {
			continue
		}

		clusterName, configName, err := resourceAwsEksIdentityProviderConfigParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &eks.DescribeIdentityProviderConfigInput{
			ClusterName: aws.String(clusterName),
			IdentityProviderConfig: &eks.IdentityProviderConfig{
				Name: aws.String(configName),
				Type: aws.String(typeOidc),
			},
		}

		output, err := conn.DescribeIdentityProviderConfig(input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if output != nil && output.IdentityProviderConfig != nil && aws.StringValue(output.IdentityProviderConfig.Oidc.IdentityProviderConfigName) == configName {
			return fmt.Errorf("EKS Identity Provider Config (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSEksIdentityProviderConfigDisappears(clusterName string, config *eks.IdentityProviderConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).eksconn

		input := &eks.DisassociateIdentityProviderConfigInput{
			ClusterName:            aws.String(clusterName),
			IdentityProviderConfig: config,
		}

		_, err := conn.DisassociateIdentityProviderConfig(input)

		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		return waitForEksIdentityProviderConfigDisassociation(context.Background(), conn, clusterName, config, 25*time.Minute)
	}
}

func testAccAWSEksIdentityProviderConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "eks.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                          = "tf-acc-test-eks-identity-provider-config"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name                          = "tf-acc-test-eks-identity-provider-config"
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy]
}
`, rName)
}

func testAccAWSEksIdentityProviderConfigProviderConfigName(rName string) string {
	return testAccAWSEksIdentityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name
  oidc {
    client_id                     = "test-url.apps.googleusercontent.com"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://accounts.google.com/.well-known/openid-configuration"
  }
}
`, rName)
}

func testAccAWSEksIdentityProviderConfigProvider_Oidc_IssuerUrl(rName, issuerUrl string) string {
	return testAccAWSEksIdentityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name
  oidc {
    client_id                     = "test-url.apps.googleusercontent.com"
    identity_provider_config_name = %[1]q
    issuer_url                    = %[2]q
  }
}
`, rName, issuerUrl)
}

func testAccAWSEksIdentityProviderConfigProvider_Oidc_Groups(rName, groupsClaim, groupsPrefix string) string {
	return testAccAWSEksIdentityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name
  oidc {
    client_id                     = "test-url.apps.googleusercontent.com"
    groups_claim                  = %[2]q
    groups_prefix                 = %[3]q
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://accounts.google.com/.well-known/openid-configuration"
  }
}
`, rName, groupsClaim, groupsPrefix)
}

func testAccAWSEksIdentityProviderConfigProvider_Oidc_Username(rName, usernameClaim, usernamePrefix string) string {
	return testAccAWSEksIdentityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name
  oidc {
    client_id                     = "test-url.apps.googleusercontent.com"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://accounts.google.com/.well-known/openid-configuration"
    username_claim                = %[2]q
    username_prefix               = %[3]q
  }
}
`, rName, usernameClaim, usernamePrefix)
}

func testAccAWSEksIdentityProviderConfig_Oidc_RequiredClaims(rName, claimsKeyOne, claimsValueOne, claimsKeyTwo, claimsValueTwo string) string {
	return testAccAWSEksIdentityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name
  oidc {
    client_id                     = "test-url.apps.googleusercontent.com"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://accounts.google.com/.well-known/openid-configuration"
    required_claims = {
      %[2]q = %[3]q
      %[4]q = %[5]q
    }
  }
}
`, rName, claimsKeyOne, claimsValueOne, claimsKeyTwo, claimsValueTwo)
}

func testAccAWSEksIdentityProviderConfig_Tags(rName, tagsKeyOne, tagsValueOne, tagsKeyTwo, tagsValueTwo string) string {
	return testAccAWSEksIdentityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_eks_identity_provider_config" "test" {
  cluster_name = aws_eks_cluster.test.name
  oidc {
    client_id                     = "test-url.apps.googleusercontent.com"
    identity_provider_config_name = %[1]q
    issuer_url                    = "https://accounts.google.com/.well-known/openid-configuration"
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagsKeyOne, tagsValueOne, tagsKeyTwo, tagsValueTwo)
}
