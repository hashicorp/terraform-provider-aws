// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2AccountSuppressionAttributes_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccAccountSuppressionAttributes_basic,
		"update":        testAccAccountSuppressionAttributes_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSuppressionAttributes_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSuppressionAttributesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact(string(types.SuppressionListReasonComplaint)),
						}),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSuppressionAttributes_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sesv2_account_suppression_attributes.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSuppressionAttributesConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetExact(
						[]knownvalue.Check{
							knownvalue.StringExact(string(types.SuppressionListReasonComplaint)),
						}),
					),
				},
			},
			{
				Config: testAccAccountSuppressionAttributesConfig_updated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSuppressionAttributesExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("suppressed_reasons"), knownvalue.SetSizeExact(0)),
				},
			},
		},
	})
}

func testAccCheckAccountSuppressionAttributesExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindAccountSuppressionAttributes(ctx, conn)

		return err
	}
}

const testAccAccountSuppressionAttributesConfig_basic = `
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = ["COMPLAINT"]
}
`

const testAccAccountSuppressionAttributesConfig_updated = `
resource "aws_sesv2_account_suppression_attributes" "test" {
  suppressed_reasons = []
}
`
