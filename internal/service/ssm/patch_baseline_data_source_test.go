package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSMPatchBaselineDataSource_existingBaseline(t *testing.T) {
	dataSourceName := "data.aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPatchBaselineDataSourceConfig_existingBaseline(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "approved_patches.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "approved_patches_compliance_level", "UNSPECIFIED"),
					resource.TestCheckResourceAttr(dataSourceName, "approved_patches_enable_non_security", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "approval_rule.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "default_baseline", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "description", "Default Patch Baseline for CentOS Provided by AWS."),
					resource.TestCheckResourceAttr(dataSourceName, "global_filter.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "AWS-CentOSDefaultPatchBaseline"),
					resource.TestCheckResourceAttr(dataSourceName, "rejected_patches.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "rejected_patches_action", "ALLOW_AS_DEPENDENCY"),
					resource.TestCheckResourceAttr(dataSourceName, "source.#", "0"),
				),
			},
		},
	})
}

func TestAccSSMPatchBaselineDataSource_newBaseline(t *testing.T) {
	dataSourceName := "data.aws_ssm_patch_baseline.test"
	resourceName := "aws_ssm_patch_baseline.test"
	rName := sdkacctest.RandomWithPrefix("tf-bl-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPatchBaselineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPatchBaselineDataSourceConfig_newBaseline(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "approved_patches", resourceName, "approved_patches"),
					resource.TestCheckResourceAttrPair(dataSourceName, "approved_patches_compliance_level", resourceName, "approved_patches_compliance_level"),
					resource.TestCheckResourceAttrPair(dataSourceName, "approved_patches_enable_non_security", resourceName, "approved_patches_enable_non_security"),
					resource.TestCheckResourceAttrPair(dataSourceName, "approval_rule", resourceName, "approval_rule"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_filter.#", resourceName, "global_filter.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "operating_system", resourceName, "operating_system"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rejected_patches", resourceName, "rejected_patches"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rejected_patches_action", resourceName, "rejected_patches_action"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source", resourceName, "source"),
				),
			},
		},
	})
}

// Test against one of the default baselines created by AWS
func testAccCheckPatchBaselineDataSourceConfig_existingBaseline() string {
	return `
data "aws_ssm_patch_baseline" "test" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
`
}

// Create a new baseline and pull it back
func testAccCheckPatchBaselineDataSourceConfig_newBaseline(name string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = "%s"
  operating_system = "AMAZON_LINUX_2"
  description      = "Test"

  approval_rule {
    approve_after_days = 5
    patch_filter {
      key    = "CLASSIFICATION"
      values = ["*"]
    }
  }
}

data "aws_ssm_patch_baseline" "test" {
  owner            = "Self"
  name_prefix      = aws_ssm_patch_baseline.test.name
  operating_system = "AMAZON_LINUX_2"
}
`, name)
}
