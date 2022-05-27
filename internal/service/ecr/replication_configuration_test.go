package ecr_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccECRReplicationConfiguration_serial(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"basic":            testAccReplicationConfiguration_basic,
		"repositoryFilter": testAccReplicationConfiguration_repositoryFilter,
	}

	for name, testFunc := range testFuncs {
		testFunc := testFunc

		t.Run(name, func(t *testing.T) {
			testFunc(t)
		})
	}
}

func testAccReplicationConfiguration_basic(t *testing.T) {
	resourceName := "aws_ecr_replication_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfiguration(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					acctest.CheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationMultipleRegionConfiguration(acctest.AlternateRegion(), acctest.ThirdRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					acctest.CheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.1.region", acctest.ThirdRegion()),
					acctest.CheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.1.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
			{
				Config: testAccReplicationConfiguration(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", acctest.AlternateRegion()),
					acctest.CheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "0"),
				),
			},
		},
	})
}

func testAccReplicationConfiguration_repositoryFilter(t *testing.T) {
	resourceName := "aws_ecr_replication_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckReplicationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationRepositoryFilter(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "a-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter_type", "PREFIX_MATCH"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationConfigurationRepositoryFilterMultiple(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "a-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter_type", "PREFIX_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.1.filter", "a-second-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.1.filter_type", "PREFIX_MATCH"),
				),
			},
			{
				Config: testAccReplicationConfigurationRepositoryFilter(acctest.AlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "a-prefix"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter_type", "PREFIX_MATCH"),
				),
			},
		},
	})
}

func testAccCheckReplicationConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn
		out, err := conn.DescribeRegistry(&ecr.DescribeRegistryInput{})
		if err != nil {
			return fmt.Errorf("ECR replication rules not found: %w", err)
		}

		if len(out.ReplicationConfiguration.Rules) == 0 {
			return fmt.Errorf("ECR replication rules not found")
		}

		return nil
	}
}

func testAccCheckReplicationConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_replication_configuration" {
			continue
		}

		out, err := conn.DescribeRegistry(&ecr.DescribeRegistryInput{})
		if err != nil {
			return err
		}

		if len(out.ReplicationConfiguration.Rules) != 0 {
			return fmt.Errorf("ECR replication rules found")
		}
	}

	return nil
}

func testAccReplicationConfiguration(region string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`, region)
}

func testAccReplicationMultipleRegionConfiguration(region1, region2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }


      destination {
        region      = %[2]q
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`, region1, region2)
}

func testAccReplicationConfigurationRepositoryFilter(region string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }

      repository_filter {
        filter      = "a-prefix"
        filter_type = "PREFIX_MATCH"
      }
    }
  }
}
`, region)
}

func testAccReplicationConfigurationRepositoryFilterMultiple(region string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }

      repository_filter {
        filter      = "a-prefix"
        filter_type = "PREFIX_MATCH"
      }

      repository_filter {
        filter      = "a-second-prefix"
        filter_type = "PREFIX_MATCH"
      }
    }
  }
}
`, region)
}
