// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfroute53profiles "github.com/hashicorp/terraform-provider-aws/internal/service/route53profiles"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ProfilesAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association awstypes.ProfileAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName, &association),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func TestAccRoute53ProfilesAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var association awstypes.ProfileAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53profiles_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ProfilesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(ctx, resourceName, &association),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfroute53profiles.Route53ProfileAssocation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53profiles_association" {
				continue
			}

			_, err := tfroute53profiles.FindAssociationByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Route53Profiles, create.ErrActionCheckingDestroyed, tfroute53profiles.ResNameAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.Route53Profiles, create.ErrActionCheckingDestroyed, tfroute53profiles.ResNameAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssociationExists(ctx context.Context, name string, association *awstypes.ProfileAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ProfilesClient(ctx)
		resp, err := tfroute53profiles.FindAssociationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Route53Profiles, create.ErrActionCheckingExistence, tfroute53profiles.ResNameAssociation, rs.Primary.ID, err)
		}

		*association = *resp

		return nil
	}
}

func testAccAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53profiles_profile" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_route53profiles_association" "test" {
  name        = %[1]q
  profile_id  = aws_route53profiles_profile.test.id
  resource_id = aws_vpc.test.id
}
`, rName)
}
