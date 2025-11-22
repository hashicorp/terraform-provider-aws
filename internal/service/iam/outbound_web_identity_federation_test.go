// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOutboundWebIdentityFederation_serial(t *testing.T) {
	t.Helper()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:  testAccIAMOutboundWebIdentityFederation_basic,
		"alreadyEnabled": testAccIAMOutboundWebIdentityFederation_alreadyEnabled,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccIAMOutboundWebIdentityFederation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_iam_outbound_web_identity_federation.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "jwt_vending_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "issuer_identifier"),
				),
			},
		},
	})
}

func testAccIAMOutboundWebIdentityFederation_alreadyEnabled(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_iam_outbound_web_identity_federation.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

					_, err := conn.EnableOutboundWebIdentityFederation(ctx, &iam.EnableOutboundWebIdentityFederationInput{})
					if err != nil {
						t.Fatalf("error enabling outbound web identity federation: %s", err)
					}
				},
				Config: testAccOutboundWebIdentityFederationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "jwt_vending_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "issuer_identifier"),
				),
			},
		},
	})
}

func testAccCheckOutboundWebIdentityFederationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_outbound_web_identity_federation" {
				continue
			}

			out, err := tfiam.GetOutboundWebIdentityFederation(ctx, conn)

			if out == nil {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Outbound Web Identity Federation still exists")
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

	_, err := tfiam.GetOutboundWebIdentityFederation(ctx, conn)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOutboundWebIdentityFederationConfig_basic() string {
	return `
resource "aws_iam_outbound_web_identity_federation" "test" {
}
`
}
