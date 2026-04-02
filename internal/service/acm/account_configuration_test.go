// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMAccountConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"ACMAccountConfiguration": {
			acctest.CtBasic: testAccACMAccountConfiguration_basic,
			"update":        testAccACMAccountConfiguration_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccACMAccountConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var accountConfiguration acm.GetAccountConfigurationOutput
	resourceName := "aws_acm_account_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfigurationConfig_basic(14),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountConfigurationExists(ctx, resourceName, &accountConfiguration),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.0.days_before_expiry", "14"),
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

func testAccACMAccountConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)

	var accountConfiguration acm.GetAccountConfigurationOutput
	resourceName := "aws_acm_account_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfigurationConfig_basic(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountConfigurationExists(ctx, resourceName, &accountConfiguration),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.0.days_before_expiry", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountConfigurationConfig_default(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountConfigurationExists(ctx, resourceName, &accountConfiguration),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.0.days_before_expiry", "45"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountConfigurationConfig_basic(30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountConfigurationExists(ctx, resourceName, &accountConfiguration),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "expiry_events.0.days_before_expiry", "30"),
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

func testAccCheckAccountConfigurationExists(ctx context.Context, name string, accountConfiguration *acm.GetAccountConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ACMServiceID, create.ErrActionCheckingExistence, tfacm.ResNameAccountConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ACMServiceID, create.ErrActionCheckingExistence, tfacm.ResNameAccountConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

		input := acm.GetAccountConfigurationInput{}
		resp, err := conn.GetAccountConfiguration(ctx, &input)
		if err != nil {
			return create.Error(names.ACMServiceID, create.ErrActionCheckingExistence, tfacm.ResNameAccountConfiguration, rs.Primary.ID, err)
		}

		*accountConfiguration = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

	input := acm.GetAccountConfigurationInput{}

	_, err := conn.GetAccountConfiguration(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAccountConfigurationConfig_basic(daysBeforeExpiry int) string {
	return fmt.Sprintf(`
resource "aws_acm_account_configuration" "test" {
  expiry_events {
    days_before_expiry = %d
  }
}
`, daysBeforeExpiry)
}

func testAccAccountConfigurationConfig_default() string {
	return `
resource "aws_acm_account_configuration" "test" {
  expiry_events {}
}
`
}
