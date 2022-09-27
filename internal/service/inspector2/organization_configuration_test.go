package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/inspector2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspector2OrganizationConfiguration_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccOrganizationConfiguration_basic,
		"disappears": testAccOrganizationConfiguration_disappears,
		"ec2ECR":     testAccOrganizationConfiguration_ec2ECR,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccOrganizationConfiguration_basic(t *testing.T) {
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", "false"),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_disappears(t *testing.T) {
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfinspector2.ResourceOrganizationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationConfiguration_ec2ECR(t *testing.T) {
	resourceName := "aws_inspector2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.Inspector2EndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ec2", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable.0.ecr", "true"),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Conn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_inspector2_organization_configuration" {
			continue
		}

		out, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})
		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameOrganizationConfiguration, rs.Primary.ID, err)
		}

		if out != nil && out.AutoEnable != nil && !aws.ToBool(out.AutoEnable.Ec2) && !aws.ToBool(out.AutoEnable.Ecr) {
			return nil
		}

		return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameOrganizationConfiguration, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckOrganizationConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameOrganizationConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameOrganizationConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Conn
		ctx := context.Background()
		_, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})

		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameOrganizationConfiguration, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Conn
	ctx := context.Background()

	_, err := conn.DescribeOrganizationConfiguration(ctx, &inspector2.DescribeOrganizationConfigurationInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && strings.Contains(err.Error(), "Invoking account does not") {
		// does not have code AccessDeniedException despite having that in the text
		t.Skipf("to run this test, enable this account as a Delegated Admin Account: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOrganizationConfigurationConfig_basic(ec2, ecr bool) string {
	return fmt.Sprintf(`
resource "aws_inspector2_organization_configuration" "test" {
  auto_enable {
    ec2 = %[1]t
    ecr = %[2]t
  }
}
`, ec2, ecr)
}
