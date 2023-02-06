package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspector2OrganizationConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"basic":      testAccOrganizationConfiguration_basic,
		"disappears": testAccOrganizationConfiguration_disappears,
		"ec2ECR":     testAccOrganizationConfiguration_ec2ECR,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", "false"),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceOrganizationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationConfiguration_ec2ECR(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", "true"),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		enabledDelAdAcct := false

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_organization_configuration" {
				continue
			}

			out, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})

			if errs.MessageContains(err, "AccessDenied", "Invoking account does not") {
				if err := testEnableDelegatedAdminAccount(ctx, conn, acctest.AccountID()); err != nil {
					return err
				}

				enabledDelAdAcct = true

				out, err = conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})
			}

			if err != nil {
				if enabledDelAdAcct {
					if err := testDisableDelegatedAdminAccount(ctx, conn, acctest.AccountID()); err != nil {
						return err
					}
				}

				return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameOrganizationConfiguration, rs.Primary.ID, err)
			}

			if out != nil && out.AutoEnable != nil && !aws.ToBool(out.AutoEnable.Ec2) && !aws.ToBool(out.AutoEnable.Ecr) {
				if enabledDelAdAcct {
					if err := testDisableDelegatedAdminAccount(ctx, conn, acctest.AccountID()); err != nil {
						return err
					}
				}

				return nil
			}

			if enabledDelAdAcct {
				if err := testDisableDelegatedAdminAccount(ctx, conn, acctest.AccountID()); err != nil {
					return err
				}
			}

			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameOrganizationConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testEnableDelegatedAdminAccount(ctx context.Context, conn *inspector2.Client, accountID string) error {
	_, err := conn.EnableDelegatedAdminAccount(ctx, &inspector2.EnableDelegatedAdminAccountInput{
		DelegatedAdminAccountId: aws.String(accountID),
	})
	if err != nil {
		return err
	}

	if err := tfinspector2.WaitDelegatedAdminAccountEnabled(ctx, conn, accountID, time.Minute*2); err != nil {
		return err
	}

	return nil
}

func testDisableDelegatedAdminAccount(ctx context.Context, conn *inspector2.Client, accountID string) error {
	_, err := conn.DisableDelegatedAdminAccount(ctx, &inspector2.DisableDelegatedAdminAccountInput{
		DelegatedAdminAccountId: aws.String(accountID),
	})
	if err != nil {
		return err
	}

	if err := tfinspector2.WaitDelegatedAdminAccountDisabled(ctx, conn, accountID, time.Minute*2); err != nil {
		return err
	}

	return nil
}

func testAccCheckOrganizationConfigurationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameOrganizationConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameOrganizationConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client()

		_, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})

		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameOrganizationConfiguration, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccOrganizationConfigurationConfig_basic(ec2, ecr bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_inspector2_organization_configuration" "test" {
  auto_enable {
    ec2 = %[1]t
    ecr = %[2]t
  }

  depends_on = [aws_inspector2_delegated_admin_account.test]
}
`, ec2, ecr)
}
