package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamUserStackAssociation_basic(t *testing.T) {
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserStackAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", authType),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rEmail),
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

func TestAccAppStreamUserStackAssociation_disappears(t *testing.T) {
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserStackAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceUserStackAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamUserStackAssociation_complete(t *testing.T) {
	resourceName := "aws_appstream_user_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	authType := "USERPOOL"
	domain := acctest.RandomDomainName()
	rEmail := acctest.RandomEmailAddress(domain)
	rEmailUpdated := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserStackAssociationDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(resourceName),
				),
			},
			{
				Config: testAccUserStackAssociationConfig_basic(rName, authType, rEmailUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserStackAssociationExists(resourceName),
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

func testAccCheckUserStackAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

		userName, authType, stackName, err := tfappstream.DecodeUserStackAssociationID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding AppStream User Stack Association ID (%s): %w", rs.Primary.ID, err)
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
			return fmt.Errorf("AppStream User Stack Association %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserStackAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_user_stack_association" {
			continue
		}

		userName, authType, stackName, err := tfappstream.DecodeUserStackAssociationID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding AppStream User Stack Association ID (%s): %w", rs.Primary.ID, err)
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
			return fmt.Errorf("AppStream User Stack Association %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccUserStackAssociationConfig_basic(name, authType, userName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_user" "test" {
  authentication_type = %[2]q
  user_name           = %[3]q
}

resource "aws_appstream_user_stack_association" "test" {
  authentication_type = aws_appstream_user.test.authentication_type
  stack_name          = aws_appstream_stack.test.name
  user_name           = aws_appstream_user.test.user_name
}
`, name, authType, userName)
}
