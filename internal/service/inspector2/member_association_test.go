package inspector2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccInspector2MemberAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_member_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sns.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInspector2MemberAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInspector2MemberAssociationConfig_basic(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInspector2MemberAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_member_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sns.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInspector2MemberAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInspector2MemberAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceMemberAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckInspector2MemberAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_member_association" {
				continue
			}

			getMemberInput := &inspector2.GetMemberInput{
				AccountId: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetMember(ctx, getMemberInput)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SNS Data Protection Topic Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInspector2MemberAssociationConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_inspector2_member_association" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`
}
