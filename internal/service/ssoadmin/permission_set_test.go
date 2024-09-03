// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminPermissionSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
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

func TestAccSSOAdminPermissionSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
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
				Config: testAccPermissionSetConfig_tags2(rName, acctest.CtKey1, "updatedvalue1", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPermissionSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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

func TestAccSSOAdminPermissionSet_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
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

func TestAccSSOAdminPermissionSet_updateRelayState(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", ""),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateRelayState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
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

func TestAccSSOAdminPermissionSet_updateSessionDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
				),
			},
			{
				Config: testAccPermissionSetConfig_updateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

// TestAccSSOAdminPermissionSet_RelayState_updateSessionDuration validates
// the resource's unchanged values (primarily relay_state) after updating the session_duration argument
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411
func TestAccSSOAdminPermissionSet_RelayState_updateSessionDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_relayState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
				),
			},
			{
				Config: testAccPermissionSetConfig_relayStateUpdateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

func TestAccSSOAdminPermissionSet_mixedPolicyAttachments(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
				),
			},
			{
				Config: testAccPermissionSetConfig_mixedPolicyAttachments(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSOAdminPermissionSetExists(ctx, resourceName),
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

func testAccCheckPermissionSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_permission_set" {
				continue
			}

			permissionSetARN, instanceARN, err := tfssoadmin.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfssoadmin.FindPermissionSet(ctx, conn, permissionSetARN, instanceARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSO Permission Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSOAdminPermissionSetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		permissionSetARN, instanceARN, err := tfssoadmin.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err = tfssoadmin.FindPermissionSet(ctx, conn, permissionSetARN, instanceARN)

		return err
	}
}

func testAccPermissionSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_updateDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccPermissionSetConfig_updateRelayState(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"
}
`, rName)
}

func testAccPermissionSetConfig_updateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  session_duration = "PT2H"
}
`, rName)
}

func testAccPermissionSetConfig_relayState(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT1H"
}
`, rName)
}

func testAccPermissionSetConfig_relayStateUpdateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT2H"
}
`, rName)
}

func testAccPermissionSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPermissionSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPermissionSetConfig_mixedPolicyAttachments(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }
}
resource "aws_ssoadmin_permission_set_inline_policy" "test" {
  inline_policy      = data.aws_iam_policy_document.test.json
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}
