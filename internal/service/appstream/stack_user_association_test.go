package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamStackUserAssociation_basic(t *testing.T) {
	resourceName := "aws_appstream_stack_user_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	rEmail := acctest.RandomEmailAddress("hashicorp.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackUserAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackUserAssociationConfig(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackUserAssociationExists(resourceName),
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

func TestAccAppStreamStackUserAssociation_disappears(t *testing.T) {
	resourceName := "aws_appstream_stack_user_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	rEmail := acctest.RandomEmailAddress("hashicorp.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackUserAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackUserAssociationConfig(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackUserAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceStackUserAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamStackUserAssociation_complete(t *testing.T) {
	resourceName := "aws_appstream_stack_user_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	rEmail := acctest.RandomEmailAddress("hashicorp.com")
	rEmailUpdated := acctest.RandomEmailAddress("hashicorp.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackUserAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccStackUserAssociationConfig(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackUserAssociationExists(resourceName),
				),
			},
			{
				Config: testAccStackUserAssociationConfig(rName, authType, rEmailUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackUserAssociationExists(resourceName),
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

func testAccCheckStackUserAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

		stackName, userName, authType, err := tfappstream.DecodeStackUserID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding id appstream stack user association (%s): %w", rs.Primary.ID, err)
		}

		resp, err := conn.DescribeUserStackAssociationsWithContext(context.Background(), &appstream.DescribeUserStackAssociationsInput{
			AuthenticationType: aws.String(authType),
			StackName:          aws.String(stackName),
			UserName:           aws.String(userName),
		})

		if err != nil {
			return err
		}

		if len(resp.UserStackAssociations) == 0 {
			return fmt.Errorf("appstream stack user association %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackUserAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_stack_user_association" {
			continue
		}

		stackName, userName, authType, err := tfappstream.DecodeStackUserID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding id appstream stack user association (%s): %w", rs.Primary.ID, err)
		}

		resp, err := conn.DescribeUserStackAssociationsWithContext(context.Background(), &appstream.DescribeUserStackAssociationsInput{
			AuthenticationType: aws.String(authType),
			StackName:          aws.String(stackName),
			UserName:           aws.String(userName),
		})

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if len(resp.UserStackAssociations) > 0 {
			return fmt.Errorf("appstream stack user association %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccStackUserAssociationConfig(name, authType, userName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_user" "test" {
  authentication_type = %[2]q
  user_name           = %[3]q
}

resource "aws_appstream_stack_user_association" "test" {
  authentication_type = aws_appstream_user.test.authentication_type
  stack_name          = aws_appstream_stack.test.name
  user_name           = aws_appstream_user.test.user_name
}
`, name, authType, userName)
}
