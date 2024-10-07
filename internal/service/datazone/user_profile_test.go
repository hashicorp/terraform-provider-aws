// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.DataZoneServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"IAM Identity Center application not enabled",
	)
}

func TestAccDataZoneUserProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var userprofile datazone.GetUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					resource.TestCheckResourceAttrSet(resourceName, "domain_identifier"),
					resource.TestCheckResourceAttr(resourceName, "details.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVATED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccUserProfileImportStateFunc(resourceName),
				ImportStateVerifyIgnore: []string{"user_type"},
			},
		},
	})
}

func TestAccDataZoneUserProfile_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var userprofile datazone.GetUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_user_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileConfig_update(rName, "ACTIVATED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					resource.TestCheckResourceAttrSet(resourceName, "domain_identifier"),
					resource.TestCheckResourceAttr(resourceName, "details.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVATED"),
				),
			},
			{
				Config: testAccUserProfileConfig_update(rName, "DEACTIVATED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					resource.TestCheckResourceAttrSet(resourceName, "domain_identifier"),
					resource.TestCheckResourceAttr(resourceName, "details.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DEACTIVATED"),
				),
			},
		},
	})
}

func testAccUserProfileImportStateFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return strings.Join([]string{rs.Primary.Attributes["user_identifier"], rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrType]}, ","), nil
	}
}

func testAccCheckUserProfileExists(ctx context.Context, name string, userprofile *datazone.GetUserProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameUserProfile, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameUserProfile, name, errors.New("not set"))
		}

		du := rs.Primary.Attributes["domain_identifier"]
		ui := rs.Primary.Attributes["user_identifier"]
		ut := types.UserProfileType(rs.Primary.Attributes[names.AttrType])

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := tfdatazone.FindUserProfileByID(ctx, conn, du, ui, ut)

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameUserProfile, rs.Primary.ID, err)
		}

		*userprofile = *resp

		return nil
	}
}

func testAccUserProfileConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_datazone_user_profile" "test" {
  user_identifier   = aws_iam_user.test.arn
  domain_identifier = aws_datazone_domain.test.id
  user_type         = "IAM_USER"
}
`, rName))
}

func testAccUserProfileConfig_update(rName, status string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_datazone_user_profile" "test" {
  user_identifier   = aws_iam_user.test.arn
  domain_identifier = aws_datazone_domain.test.id
  user_type         = "IAM_USER"
  status            = %[2]q
}
`, rName, status))
}
