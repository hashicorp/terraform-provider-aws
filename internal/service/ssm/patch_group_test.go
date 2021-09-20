package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestAccAWSSSMPatchGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_patch_group.patchgroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSSMPatchGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchGroupExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSSSMPatchGroup_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_patch_group.patchgroup"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssm.ResourcePatchGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMPatchGroup_multipleBaselines(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName1 := "aws_ssm_patch_group.test1"
	resourceName2 := "aws_ssm_patch_group.test2"
	resourceName3 := "aws_ssm_patch_group.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ssm.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSSMPatchGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMPatchGroupConfigMultipleBaselines(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMPatchGroupExists(resourceName1),
					testAccCheckAWSSSMPatchGroupExists(resourceName2),
					testAccCheckAWSSSMPatchGroupExists(resourceName3),
				),
			},
		},
	})
}

func testAccCheckAWSSSMPatchGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_patch_group" {
			continue
		}

		patchGroup, baselineId, err := tfssm.ParsePatchGroupID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing SSM Patch Group ID (%s): %w", rs.Primary.ID, err)
		}

		group, err := tfssm.FindPatchGroup(conn, patchGroup, baselineId)

		if err != nil {
			return fmt.Errorf("error describing SSM Patch Group ID (%s): %w", rs.Primary.ID, err)
		}

		if group != nil {
			return fmt.Errorf("SSM Patch Group %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSSMPatchGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Patch Baseline ID is set")
		}

		patchGroup, baselineId, err := tfssm.ParsePatchGroupID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing SSM Patch Group ID (%s): %w", rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		group, err := tfssm.FindPatchGroup(conn, patchGroup, baselineId)

		if err != nil {
			return fmt.Errorf("error reading SSM Patch Group (%s): %w", rs.Primary.ID, err)
		}

		if group == nil {
			return fmt.Errorf("No SSM Patch Group found")
		}

		return nil
	}
}

func testAccAWSSSMPatchGroupBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "foo" {
  name             = %[1]q
  approved_patches = ["KB123456"]
}

resource "aws_ssm_patch_group" "patchgroup" {
  baseline_id = aws_ssm_patch_baseline.foo.id
  patch_group = %[1]q
}
`, rName)
}

func testAccAWSSSMPatchGroupConfigMultipleBaselines(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test1" {
  approved_patches = ["KB123456"]
  name             = %[1]q
  operating_system = "CENTOS"
}

resource "aws_ssm_patch_baseline" "test2" {
  approved_patches = ["KB123456"]
  name             = %[1]q
  operating_system = "AMAZON_LINUX_2"
}

resource "aws_ssm_patch_baseline" "test3" {
  approved_patches = ["KB123456"]
  name             = %[1]q
  operating_system = "AMAZON_LINUX"
}

resource "aws_ssm_patch_group" "test1" {
  baseline_id = aws_ssm_patch_baseline.test1.id
  patch_group = %[1]q
}

resource "aws_ssm_patch_group" "test2" {
  baseline_id = aws_ssm_patch_baseline.test2.id
  patch_group = %[1]q
}

resource "aws_ssm_patch_group" "test3" {
  baseline_id = aws_ssm_patch_baseline.test3.id
  patch_group = %[1]q
}
`, rName)
}
