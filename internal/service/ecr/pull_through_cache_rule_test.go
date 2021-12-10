package ecr_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
)

func TestAccPullThroughCacheRule_basic(t *testing.T) {
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPullThroughCacheRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ecr_repository_prefix", repositoryPrefix),
					testAccCheckPullThroughCacheRuleRegistryID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream_registry_url", "public.ecr.aws"),
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

func TestAccPullThroughCacheRule_disappears(t *testing.T) {
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPullThroughCacheRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfecr.ResourcePullThroughCacheRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPullThroughCacheRule_failWhenAlreadyExists(t *testing.T) {
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecr.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPullThroughCacheRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_failWhenAlreadyExist(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(resourceName),
				),
				ExpectError: regexp.MustCompile(`PullThroughCacheRuleAlreadyExistsException`),
			},
		},
	})
}

func testAccCheckPullThroughCacheRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_pull_through_cache_rule" {
			continue
		}

		rule, err := tfecr.FindPullThroughCacheRuleByRepositoryPrefix(context.Background(), conn, rs.Primary.Attributes["ecr_repository_prefix"])
		if err != nil {
			return err
		}

		if rule != nil {
			return fmt.Errorf("ECR Pull Through Cache Rule still exists: %s", rs.Primary.Attributes["ecr_repository_prefix"])
		}

		return nil
	}

	return nil
}

func testAccCheckPullThroughCacheRuleExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

		rule, err := tfecr.FindPullThroughCacheRuleByRepositoryPrefix(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading ECR Pull Through Cache Rule (%s): %w", rs.Primary.ID, err)
		}

		if rule == nil {
			return fmt.Errorf("ECR Pull Through Cache Rule (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(rule.EcrRepositoryPrefix) != rs.Primary.ID {
			return fmt.Errorf("ECR Pull Through Cache Rule (%s) ID not consistent with repository prefix", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPullThroughCacheRuleRegistryID(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := acctest.AccountID()
		return resource.TestCheckResourceAttr(resourceName, "registry_id", attributeValue)(s)
	}
}

func testAccPullThroughCacheRuleConfig(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}
`, repositoryPrefix)
}

func testAccPullThroughCacheRuleConfig_failWhenAlreadyExist(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}

resource "aws_ecr_pull_through_cache_rule" "duplicate" {
  depends_on            = [aws_ecr_pull_through_cache_rule.test]
  ecr_repository_prefix = %[1]q
  upstream_registry_url = "public.ecr.aws"
}
`, repositoryPrefix)
}
