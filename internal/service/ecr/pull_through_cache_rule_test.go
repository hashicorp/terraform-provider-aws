package ecr_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccECRPullThroughCacheRule_basic(t *testing.T) {
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPullThroughCacheRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_basic(repositoryPrefix),
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

func TestAccECRPullThroughCacheRule_disappears(t *testing.T) {
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPullThroughCacheRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleConfig_basic(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPullThroughCacheRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfecr.ResourcePullThroughCacheRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRPullThroughCacheRule_failWhenAlreadyExists(t *testing.T) {
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_pull_through_cache_rule.test"

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("ECR Pull Through Cache Rule is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPullThroughCacheRuleDestroy,
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

		_, err := tfecr.FindPullThroughCacheRuleByRepositoryPrefix(context.Background(), conn, rs.Primary.Attributes["ecr_repository_prefix"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("ECR Pull Through Cache Rule %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckPullThroughCacheRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Pull Through Cache Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRConn

		_, err := tfecr.FindPullThroughCacheRuleByRepositoryPrefix(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
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

func testAccPullThroughCacheRuleConfig_basic(repositoryPrefix string) string {
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
