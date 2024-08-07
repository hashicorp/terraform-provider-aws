// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRBlockPublicAccessConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, emr.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockPublicAccessConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationAttributes_enabledOnly(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccBlockPublicAccessConfigurationConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationAttributes_disabled(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccEMRBlockPublicAccessConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, emr.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockPublicAccessConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationAttributes_enabledOnly(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceBlockPublicAccessConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccEMRBlockPublicAccessConfiguration_default(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, emr.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: blockPublicAccessConfigurationConfig_defaultString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationAttributes_default(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.min_range", "22"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.max_range", "22"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccEMRBlockPublicAccessConfiguration_enabledMultiRange(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, emr.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: blockPublicAccessConfigurationConfig_enabledMultiRangeString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationAttributes_enabledMultiRange(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.min_range", "22"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.max_range", "22"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.1.min_range", "100"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.1.max_range", "101"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func testAccCheckBlockPublicAccessConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emr_block_public_access_configuration" {
				continue
			}

			resp, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

			if err != nil {
				return err
			}

			blockPublicSecurityGroupRules := resp.BlockPublicAccessConfiguration.BlockPublicSecurityGroupRules
			permittedPublicSecurityGroupRuleRanges := resp.BlockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges

			if *blockPublicSecurityGroupRules != true {
				return fmt.Errorf("Block Public Security Group Rules is not enabled")
			}

			if length := len(permittedPublicSecurityGroupRuleRanges); length != 1 {
				return fmt.Errorf("The incorrect number (%v) of permitted public security group rule ranges exist, should be 1 by default", length)
			}
			if p := permittedPublicSecurityGroupRuleRanges; !((*p[0].MinRange == 22 && *p[0].MaxRange == 22) || (*p[1].MinRange == 22 || *p[1].MaxRange == 22)) {
				return fmt.Errorf("Port 22 is not open as a permitted public security group rule")
			}
		}

		return nil
	}
}

func testAccCheckBlockPublicAccessConfigurationAttributes_enabledOnly(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)
		resp, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

		if err != nil {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, rs.Primary.ID, err)
		}

		blockPublicSecurityGroupRules := resp.BlockPublicAccessConfiguration.BlockPublicSecurityGroupRules
		permittedPublicSecurityGroupRuleRanges := resp.BlockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges

		if *blockPublicSecurityGroupRules != true {
			return fmt.Errorf("Block Public Security Group Rules is not enabled")
		}

		if length := len(permittedPublicSecurityGroupRuleRanges); length != 0 {
			return fmt.Errorf("The incorrect number (%v) of permitted public security group rule ranges have been created, should be 0", length)
		}

		return nil
	}
}

func testAccCheckBlockPublicAccessConfigurationAttributes_default(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, name, errors.New("not found"))
		}

		resp, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

		if err != nil {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, rs.Primary.ID, err)
		}

		blockPublicSecurityGroupRules := resp.BlockPublicAccessConfiguration.BlockPublicSecurityGroupRules
		permittedPublicSecurityGroupRuleRanges := resp.BlockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges

		if *blockPublicSecurityGroupRules != true {
			return fmt.Errorf("Block Public Security Group Rules is not enabled")
		}

		if length := len(permittedPublicSecurityGroupRuleRanges); length != 1 {
			return fmt.Errorf("The incorrect number (%v) of permitted public security group rule ranges exist, should be 1 by default", length)
		}
		if p := permittedPublicSecurityGroupRuleRanges; !((*p[0].MinRange == 22 && *p[0].MaxRange == 22) || (*p[1].MinRange == 22 || *p[1].MaxRange == 22)) {
			return fmt.Errorf("Port 22 is not open as a permitted public security group rule")
		}

		return nil
	}
}

func testAccCheckBlockPublicAccessConfigurationAttributes_enabledMultiRange(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)
		resp, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

		if err != nil {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, rs.Primary.ID, err)
		}

		blockPublicSecurityGroupRules := resp.BlockPublicAccessConfiguration.BlockPublicSecurityGroupRules
		permittedPublicSecurityGroupRuleRanges := resp.BlockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges

		if *blockPublicSecurityGroupRules != true {
			return fmt.Errorf("Block Public Security Group Rules is not enabled")
		}

		if length := len(permittedPublicSecurityGroupRuleRanges); length != 2 {
			return fmt.Errorf("The incorrect number (%v) of permitted public security group rule ranges have been created, should be 2", length)
		}
		if p := permittedPublicSecurityGroupRuleRanges; !((*p[0].MinRange == 22 && *p[0].MaxRange == 22) || (*p[1].MinRange == 22 || *p[1].MaxRange == 22)) {
			return fmt.Errorf("Port 22 has not been opened as a permitted_public_security_group_rule")
		}
		if p := permittedPublicSecurityGroupRuleRanges; !((*p[0].MinRange == 100 && *p[0].MaxRange == 101) || (*p[1].MinRange == 100 || *p[1].MaxRange == 101)) {
			return fmt.Errorf("Ports 100-101 have not been opened as a permitted_public_security_group_rule")
		}

		return nil
	}
}

func testAccCheckBlockPublicAccessConfigurationAttributes_disabled(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)
		resp, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

		if err != nil {
			return create.Error(names.EMR, create.ErrActionCheckingExistence, tfemr.ResNameBlockPublicAccessConfiguration, rs.Primary.ID, err)
		}

		blockPublicSecurityGroupRules := resp.BlockPublicAccessConfiguration.BlockPublicSecurityGroupRules
		permittedPublicSecurityGroupRuleRanges := resp.BlockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges

		if *blockPublicSecurityGroupRules != false {
			return fmt.Errorf("Block Public Security Group Rules is not disabled")
		}

		if length := len(permittedPublicSecurityGroupRuleRanges); length != 0 {
			return fmt.Errorf("The incorrect number (%v) of permitted public security group rule ranges have been created, should be 0", length)
		}
		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

	input := &emr.GetBlockPublicAccessConfigurationInput{}
	_, err := conn.GetBlockPublicAccessConfigurationWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBlockPublicAccessConfigurationConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_emr_block_public_access_configuration" "test" {
  block_public_security_group_rules = %[1]t
}
`, enabled)
}

const blockPublicAccessConfigurationConfig_defaultString = `
resource "aws_emr_block_public_access_configuration" "test" {
  block_public_security_group_rules = true

  permitted_public_security_group_rule_range {
    min_range = 22
    max_range = 22
  }
}
`

const blockPublicAccessConfigurationConfig_enabledMultiRangeString = `
resource "aws_emr_block_public_access_configuration" "test" {
  block_public_security_group_rules = true

  permitted_public_security_group_rule_range {
    min_range = 22
    max_range = 22
  }

  permitted_public_security_group_rule_range {
    min_range = 100
    max_range = 101
  }
}
`
