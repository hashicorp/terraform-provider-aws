package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
	resourceName := "aws_verifiedaccess_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccVerifiedAccessGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttr(resourceName, "deletion_time", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
					resource.TestCheckResourceAttr(resourceName, "policy_document", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "verifiedaccess_instance_id"),
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

func TestAccVerifiedAccessGroup_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.VerifiedAccessGroup
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
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
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
	var v types.VerifiedAccessGroup
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
					testAccCheckVerifiedAccessGroupExists(ctx, resourceName, &v),
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

func testAccCheckVerifiedAccessGroupExists(ctx context.Context, n string, v *types.VerifiedAccessGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
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

			return fmt.Errorf("Verified Access Group %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccVerifiedAccessGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_instance" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_verifiedaccess_group" "test" {
  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, rName)
}

func testAccVerifiedAccessGroupConfig_tags(rName, description, tag1, val1 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedaccess_group" %[1]q {
  description                = %[2]q
  verifiedaccess_instance_id = "vai-014ba03ba2d7c6a6f"
  tags = {
	%[3]q = %[4]q
  }
}
`, rName, description, tag1, val1)
}

func testAccVerifiedAccessGroupConfig_policy(rName, description, policy string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_verifiedaccess_group" %[1]q {
  description                = %[2]q
  verifiedaccess_instance_id = "vai-014ba03ba2d7c6a6f"
  policy_document            = %[3]q
}
`, rName, description, policy))
}
