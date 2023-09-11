package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	//var verifiedaccessaccessgroup ec2.DescribeVerifiedAccessGroupsOutput

	rName := "test"
	resourceName := "aws_verifiedaccess_group." + rName
	description := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "policy_document", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccVerifiedAccessGroup_policy(t *testing.T) {
	ctx := acctest.Context(t)

	description := sdkacctest.RandString(100)
	policyDoc := "permit(principal, action, resource) \nwhen {\ncontext.http_request.method == \"GET\"\n};"
	rName := "test"
	resourceName := "aws_verifiedaccess_group." + rName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_policy(rName, description, policyDoc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "policy_document", policyDoc),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccVerifiedAccessGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)

	description := sdkacctest.RandString(100)
	rName := "test"
	resourceName := "aws_verifiedaccess_group." + rName
	tag1 := sdkacctest.RandString(10)
	value1 := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVerifiedAccessGroupConfig_tags(rName, description, tag1, value1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+tag1, value1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckVerifiedAccessGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_group" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.EC2, create.ErrActionCheckingDestroyed, tfec2.ResNameVerifiedAccessGroup, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccVerifiedAccessGroupConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_group" %[1]q {
  description = %[2]q
  verified_access_instance_id = "vai-014ba03ba2d7c6a6f"
}
`, rName, description)
}

func testAccVerifiedAccessGroupConfig_tags(rName, description, tag1, val1 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_group" %[1]q {
  description = %[2]q
  verified_access_instance_id = "vai-014ba03ba2d7c6a6f"
  tags = {
	%[3]q = %[4]q
  }
}
`, rName, description, tag1, val1)
}

func testAccVerifiedAccessGroupConfig_policy(rName, description, policy string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_verifiedaccess_group" %[1]q {
	description = %[2]q
	verified_access_instance_id = "vai-014ba03ba2d7c6a6f"
	policy_document = %[3]q
}
`, rName, description, policy))
}
