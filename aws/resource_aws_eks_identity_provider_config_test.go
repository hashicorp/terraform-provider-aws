package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"testing"
	"time"
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
				Config: testAccAWSEksIdentityProviderConfigProviderConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEksIdentityProviderConfigExists(resourceName, &config),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_name", eksClusterResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "oidc.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.identity_provider_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.client_id", "test-url.apps.googleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "oidc.0.issuer_url", "https://accounts.google.com/.well-known/openid-configuration"),
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

// TODO - still needs some work
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
  cluster_name           = aws_eks_cluster.test.name
  oidc {
    client_id = "test-url.apps.googleusercontent.com"
    identity_provider_config_name = %[1]q
    issuer_url = "https://accounts.google.com/.well-known/openid-configuration"
  }
}
`, rName)
}
