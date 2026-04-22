// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testAccApplicationProviderARN = "arn:aws:sso::aws:applicationProvider/custom" //lintignore:AWSAT005
)

func TestAccSSOAdminApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName, testAccApplicationProviderARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "application_provider_arn", testAccApplicationProviderARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("sso", regexache.MustCompile(`application/ssoins-[0-9a-z]{16}/apl-[0-9a-z]{16}`))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrID), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("application_arn"), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
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

func TestAccSSOAdminApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName, testAccApplicationProviderARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplication_description(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_description(rName, testAccApplicationProviderARN, "text1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_description(rName, testAccApplicationProviderARN, "text2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "text2"),
				),
			},
		},
	})
}

func TestAccSSOAdminApplication_portalOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"
	applicationURL1 := "http://example.com"
	applicationURL2 := "http://example2.com"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_portalOptions(rName, testAccApplicationProviderARN, applicationURL1, string(types.SignInOriginApplication)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "portal_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "portal_options.0.sign_in_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "portal_options.0.sign_in_options.*", map[string]string{
						"application_url": applicationURL1,
						"origin":          string(types.SignInOriginApplication),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_portalOptions(rName, testAccApplicationProviderARN, applicationURL2, string(types.SignInOriginApplication)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "portal_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "portal_options.0.sign_in_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "portal_options.0.sign_in_options.*", map[string]string{
						"application_url": applicationURL2,
						"origin":          string(types.SignInOriginApplication),
					}),
				),
			},
		},
	})
}

func TestAccSSOAdminApplication_status(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_status(rName, testAccApplicationProviderARN, string(types.ApplicationStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ApplicationStatusEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_status(rName, testAccApplicationProviderARN, string(types.ApplicationStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ApplicationStatusDisabled)),
				),
			},
			{
				Config: testAccApplicationConfig_status(rName, testAccApplicationProviderARN, string(types.ApplicationStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ApplicationStatusEnabled)),
				),
			},
		},
	})
}

func TestAccSSOAdminApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, testAccApplicationProviderARN, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_tags2(rName, testAccApplicationProviderARN, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, testAccApplicationProviderARN, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_application" {
				continue
			}

			_, err := tfssoadmin.FindApplicationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, t *testing.T, n string, v *ssoadmin.DescribeApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		output, err := tfssoadmin.FindApplicationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationConfig_basic(rName, applicationProviderARN string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName, applicationProviderARN)
}

func testAccApplicationConfig_description(rName, applicationProviderARN, description string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  description              = %[3]q
}
`, rName, applicationProviderARN, description)
}

func testAccApplicationConfig_portalOptions(rName, applicationProviderARN, applicationURL, origin string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  portal_options {
    visibility = "ENABLED"
    sign_in_options {
      application_url = %[3]q
      origin          = %[4]q
    }
  }
}
`, rName, applicationProviderARN, applicationURL, origin)
}

func testAccApplicationConfig_status(rName, applicationProviderARN, status string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  status                   = %[3]q
}
`, rName, applicationProviderARN, status)
}

func testAccApplicationConfig_tags1(rName, applicationProviderARN, key1, value1 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, applicationProviderARN, key1, value1)
}

func testAccApplicationConfig_tags2(rName, applicationProviderARN, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, applicationProviderARN, key1, value1, key2, value2)
}
