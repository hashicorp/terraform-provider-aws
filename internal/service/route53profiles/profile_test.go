// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/route53profiles/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"fmt"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfroute53profiles "github.com/hashicorp/terraform-provider-aws/internal/service/route53profiles"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_profile.test"
	var v awstypes.Profile

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoute53ProfilesProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ProfilesProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoute53ProfilesProfileExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccCheckRoute53ProfilesProfileExists(ctx context.Context, name string, r *awstypes.Profile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		output, err := tfroute53profiles.FindProfileByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*r = *output

		return nil
	}
}

func testAccCheckRoute53ProfilesProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53profiles_profile" {
				continue
			}

			_, err := tfroute53profiles.FindProfileByID(ctx, conn, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("route 53 Profile %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccRoute53ProfilesProfileConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}
`, rName)
}
