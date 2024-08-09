// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestAccDataZoneUserProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var userprofile datazone.GetUserProfileOutput
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_user_profile.test"
	domainName := "aws_datazone_domain.test"
	callerName := "data.caller_identity.test"

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
				Config: testAccUserProfileConfig_iamUser(dName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					resource.TestCheckResourceAttrPair(resourceName, "user_identifier", callerName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "details.arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
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

func TestAccDataZoneUserProfile_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var userprofile datazone.GetUserProfileOutput
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_user_profile.test"
	domainName := "aws_datazone_domain.test"
	callerName := "data.caller_identity.test"

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
				Config: testAccUserProfileConfig_iamUser(dName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					resource.TestCheckResourceAttrPair(resourceName, "user_identifier", callerName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "user_type", "SSO"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "details.sso_user_profiles_details.first_name"),
					resource.TestCheckResourceAttrSet(resourceName, "details.sso_user_profiles_details.last_name"),
					resource.TestCheckResourceAttrSet(resourceName, "details.sso_user_profiles_details.user_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccUserProfileConfig_iamUser(dName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					resource.TestCheckResourceAttrPair(resourceName, "user_identifier", callerName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "user_type", "SSO"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "details.sso_user_profiles_details.first_name"),
					resource.TestCheckResourceAttrSet(resourceName, "details.sso_user_profiles_details.last_name"),
					resource.TestCheckResourceAttrSet(resourceName, "details.sso_user_profiles_details.user_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ENABLED"),
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

func TestAccDataZoneUserProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var userprofile datazone.GetUserProfileOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccUserProfileConfig_iamUser(rName, uName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(ctx, resourceName, &userprofile),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceUserProfile, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
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

		td := rs.Primary.Attributes["domain_identifier"]
		tui := rs.Primary.Attributes["user_identifier"]
		tt := types.UserProfileType(rs.Primary.Attributes["user_type"])

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetUserProfile(ctx, &datazone.GetUserProfileInput{
			DomainIdentifier: &td,
			UserIdentifier:   &tui,
			Type:             tt,
		})

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameUserProfile, rs.Primary.ID, err)
		}

		*userprofile = *resp

		return nil
	}
}

/*
func testAccUserProfilePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	input := &datazone.GetUserProfileInput{}
	_, err := conn.ListUserProfiles(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
*/

func testAccCheckUserProfileNotRecreated(before, after *datazone.GetUserProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.DataZone, create.ErrActionCheckingNotRecreated, tfdatazone.ResNameUserProfile, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccUserProfileConfig_iamUser(dName, uName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_domain(dName, uName), (`

resource "aws_datazone_user_profile" "test" {
	user_identifier = data.aws_caller_identity.test.account_id
	domain_identifier = aws_datazone_domain.test.id
}
`))
}

func testAccDomainConfig_domain(rName, uName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigUserProfileExecutionRole(rName, uName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}
`, rName),
	)
}

func testAccUserProfileIamUser(uName string) string {
	return fmt.Sprintf(
		`
		resource "aws_iam_user" "test" {
		name = %[1]q
		}
		`, uName)
}

func testAccDomainConfigUserProfileExecutionRole(rName, uName string) string {
	return acctest.ConfigCompose(testAccUserProfileIamUser(uName), fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_iam_role" "domain_execution_role" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
	  {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          AWS = "${data.aws_caller_identity.test.arn}"
        }
      },
    ]
  })

  inline_policy {
    name = %[1]q
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

`, rName))
}
