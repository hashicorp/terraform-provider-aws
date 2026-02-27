// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kms", regexache.MustCompile(`alias/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, tfkms.AliasNamePrefix+rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", keyResourceName, names.AttrID),
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

func TestAccKMSAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkms.ResourceAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSAlias_Name_generated(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					resource.TestMatchResourceAttr(resourceName, names.AttrName, regexache.MustCompile(fmt.Sprintf("%s[[:xdigit:]]{%d}", tfkms.AliasNamePrefix, sdkid.UniqueIDSuffixLength))),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, tfkms.AliasNamePrefix),
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

func TestAccKMSAlias_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_namePrefix(rName, tfkms.AliasNamePrefix+"tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, tfkms.AliasNamePrefix+"tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, tfkms.AliasNamePrefix+"tf-acc-test-prefix-"),
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

func TestAccKMSAlias_updateKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"
	key1ResourceName := "aws_kms_key.test"
	key2ResourceName := "aws_kms_key.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", key1ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", key1ResourceName, names.AttrID),
				),
			},
			{
				Config: testAccAliasConfig_updatedKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", key2ResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", key2ResourceName, names.AttrID),
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

func TestAccKMSAlias_multipleAliasesForSameKey(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"
	alias2ResourceName := "aws_kms_alias.test2"
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_arn", keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "target_key_id", keyResourceName, names.AttrID),
					testAccCheckAliasExists(ctx, t, alias2ResourceName, &alias),
					resource.TestCheckResourceAttrPair(alias2ResourceName, "target_key_arn", keyResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(alias2ResourceName, "target_key_id", keyResourceName, names.AttrID),
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

func TestAccKMSAlias_arnDiffSuppress(t *testing.T) {
	ctx := acctest.Context(t)
	var alias awstypes.AliasListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig_diffSuppress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(ctx, t, resourceName, &alias),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("target_key_arn"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasConfig_diffSuppress(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kms_alias" {
				continue
			}

			_, err := tfkms.FindAliasByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("KMS Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAliasExists(ctx context.Context, t *testing.T, n string, v *awstypes.AliasListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		output, err := tfkms.FindAliasByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAliasConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.id
}
`, rName)
}

func testAccAliasConfig_nameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "test" {
  target_key_id = aws_kms_key.test.id
}
`, rName)
}

func testAccAliasConfig_namePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "test" {
  name_prefix   = %[2]q
  target_key_id = aws_kms_key.test.id
}
`, rName, namePrefix)
}

func testAccAliasConfig_updatedKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_key" "test2" {
  description             = "%[1]s-2"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test2.id
}
`, rName)
}

func testAccAliasConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s-1"
  target_key_id = aws_kms_key.test.key_id
}

resource "aws_kms_alias" "test2" {
  name          = "alias/%[1]s-2"
  target_key_id = aws_kms_key.test.key_id
}
`, rName)
}

func testAccAliasConfig_diffSuppress(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.arn
}
`, rName)
}
