// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testAccApplicationProviderARN = "arn:aws:sso::aws:applicationProvider/custom" //lintignore:AWSAT005
)

func TestAccSSOAdminApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName, testAccApplicationProviderARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "application_provider_arn", testAccApplicationProviderARN),
					acctest.MatchResourceAttrGlobalARN(resourceName, "application_arn", "sso", regexache.MustCompile(`application/+.`)),
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

func TestAccSSOAdminApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName, testAccApplicationProviderARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplication_description(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_description(rName, testAccApplicationProviderARN, "text1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
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
					testAccCheckApplicationExists(ctx, resourceName, &application),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"
	applicationURL1 := "http://example.com"
	applicationURL2 := "http://example2.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_portalOptions(rName, testAccApplicationProviderARN, applicationURL1, string(types.SignInOriginApplication)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "portal_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "portal_options.0.sign_in_options.#", acctest.Ct1),
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
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "portal_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "portal_options.0.sign_in_options.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_status(rName, testAccApplicationProviderARN, string(types.ApplicationStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
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
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ApplicationStatusDisabled)),
				),
			},
			{
				Config: testAccApplicationConfig_status(rName, testAccApplicationProviderARN, string(types.ApplicationStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, testAccApplicationProviderARN, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccApplicationConfig_tags2(rName, testAccApplicationProviderARN, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, testAccApplicationProviderARN, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_application" {
				continue
			}

			_, err := tfssoadmin.FindApplicationByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplication, rs.Primary.ID, err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplication, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, name string, application *ssoadmin.DescribeApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplication, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplication, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)
		resp, err := tfssoadmin.FindApplicationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplication, rs.Primary.ID, err)
		}

		*application = *resp

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
