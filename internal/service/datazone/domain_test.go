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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var domain datazone.GetDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "portal_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
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

func TestAccDataZoneDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var domain datazone.GetDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceDomain, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneDomain_kms_key_identifier(t *testing.T) {
	ctx := acctest.Context(t)

	var domain datazone.GetDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_kms_key_identifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_identifier"),
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

func TestAccDataZoneDomain_description(t *testing.T) {
	ctx := acctest.Context(t)

	var domain datazone.GetDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"
	description := "This is a description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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

func TestAccDataZoneDomain_single_sign_on(t *testing.T) {
	ctx := acctest.Context(t)

	var domain datazone.GetDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_single_sign_on(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// we do not set single_sign_on if it's the default value
				ImportStateVerifyIgnore: []string{"single_sign_on"},
			},
		},
	})
}

func TestAccDataZoneDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var domain datazone.GetDomainOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDomainConfig_tags(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_domain" {
				continue
			}

			_, err := conn.GetDomain(ctx, &datazone.GetDomainInput{
				Identifier: aws.String(rs.Primary.ID),
			})

			if tfdatazone.IsResourceMissing(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameDomain, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameDomain, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, name string, domain *datazone.GetDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameDomain, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameDomain, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetDomain(ctx, &datazone.GetDomainInput{
			Identifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameDomain, rs.Primary.ID, err)
		}

		*domain = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	input := &datazone.ListDomainsInput{}
	_, err := conn.ListDomains(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDomainConfigDomainExecutionRole(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccDomainConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}
`, rName),
	)
}

func testAccDomainConfig_kms_key_identifier(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
  kms_key_identifier    = aws_kms_key.test.arn
}
`, rName),
	)
}

func testAccDomainConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
  description           = %[2]q
}
`, rName, description),
	)
}

func testAccDomainConfig_single_sign_on(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn
  single_sign_on {
    type = "DISABLED"
  }
}
`, rName),
	)
}

func testAccDomainConfig_tags(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey, tagValue),
	)
}

func testAccDomainConfig_tags2(rName, tagKey, tagValue, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccDomainConfigDomainExecutionRole(rName),
		fmt.Sprintf(`
resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.domain_execution_role.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey, tagValue, tagKey2, tagValue2),
	)
}
