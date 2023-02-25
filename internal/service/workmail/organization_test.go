package workmail_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/aws/aws-sdk-go-v2/service/workmail/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	tfworkmail "github.com/hashicorp/terraform-provider-aws/internal/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkMailOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var organization1 workmail.DescribeOrganizationOutput
	var organization2 workmail.DescribeOrganizationOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.WorkMail, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMail),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization1),
					resource.TestCheckResourceAttr(resourceName, "alias", rName1),
					resource.TestCheckResourceAttr(resourceName, "state", "Active"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "workmail", regexp.MustCompile(`organization/.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization2),
					testAccCheckOrganizationRecreated(&organization1, &organization2),
					resource.TestCheckResourceAttr(resourceName, "alias", rName2),
					resource.TestCheckResourceAttr(resourceName, "state", "Active"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "workmail", regexp.MustCompile(`organization/.`)),
				),
			},
		},
	})
}

func TestAccWorkMailOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var organization workmail.DescribeOrganizationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.WorkMail, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMail),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "alias", rName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfworkmail.ResourceOrganization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOrganizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_organization" {
				continue
			}

			input := &workmail.DescribeOrganizationInput{
				OrganizationId: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeOrganization(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			if aws.ToString(out.State) == "Deleted" {
				return nil
			}

			return create.Error(names.WorkMail, create.ErrActionCheckingDestroyed, tfworkmail.ResNameOrganization, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOrganizationExists(ctx context.Context, name string, organization *workmail.DescribeOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailClient()
		resp, err := conn.DescribeOrganization(ctx, &workmail.DescribeOrganizationInput{
			OrganizationId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, rs.Primary.ID, err)
		}

		*organization = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailClient()

	input := &workmail.ListOrganizationsInput{}
	_, err := conn.ListOrganizations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckOrganizationRecreated(before, after *workmail.DescribeOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.OrganizationId), aws.ToString(after.OrganizationId); before == after {
			return create.Error(names.WorkMail, create.ErrActionCheckingNotRecreated, tfworkmail.ResNameOrganization, before, errors.New("not recreated"))
		}

		return nil
	}
}

func testAccOrganizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  alias = %[1]q
}
`, rName)
}
