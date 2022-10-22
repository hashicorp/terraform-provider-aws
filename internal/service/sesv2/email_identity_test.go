package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentity_basic(t *testing.T) {
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "email_identity", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(`identity/.+`)),
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

func TestAccSESV2EmailIdentity_disappears(t *testing.T) {
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsesv2.ResourceEmailIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEmailIdentityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sesv2_email_identity" {
			continue
		}

		_, err := tfsesv2.FindEmailIdentityByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			var nfe *types.NotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameEmailIdentity, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckEmailIdentityExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentity, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentity, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

		_, err := tfsesv2.FindEmailIdentityByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameEmailIdentity, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccEmailIdentityConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}
`, rName)
}
