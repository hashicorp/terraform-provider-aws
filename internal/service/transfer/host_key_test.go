// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccHostKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_basic(privateKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("transfer", regexache.MustCompile(`host-key/.+/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("host_key_fingerprint"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("host_key_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "host_key_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "server_id", "host_key_id"),
				ImportStateVerifyIgnore:              []string{"host_key_body"},
			},
		},
	})
}

func testAccHostKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_basic(privateKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tftransfer.ResourceHostKey, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccHostKey_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_tags1(privateKey, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "host_key_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "server_id", "host_key_id"),
				ImportStateVerifyIgnore:              []string{"host_key_body"},
			},
			{
				Config: testAccHostKeyConfig_tags2(privateKey, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccHostKeyConfig_tags1(privateKey, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccHostKey_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_description(privateKey, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description1")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "host_key_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "server_id", "host_key_id"),
				ImportStateVerifyIgnore:              []string{"host_key_body"},
			},
			{
				Config: testAccHostKeyConfig_description(privateKey, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description2")),
				},
			},
		},
	})
}

func testAccHostKey_updateHostKeyBody(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey1, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	_, privateKey2, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_basic(privateKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHostKeyConfig_basic(privateKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccHostKey_hostKeyBodyWO(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_hostKeyBodyWO(privateKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("transfer", regexache.MustCompile(`host-key/.+/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("host_key_fingerprint"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("host_key_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "host_key_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "server_id", "host_key_id"),
				ImportStateVerifyIgnore:              []string{"host_key_body"},
			},
		},
	})
}

func testAccHostKey_updateHostKeyBodyWO(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DescribedHostKey
	resourceName := "aws_transfer_host_key.test"
	_, privateKey1, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	_, privateKey2, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostKeyConfig_hostKeyBodyWO(privateKey1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHostKeyConfig_hostKeyBodyWO(privateKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHostKeyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func testAccCheckHostKeyExists(ctx context.Context, t *testing.T, n string, v *awstypes.DescribedHostKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

		output, err := tftransfer.FindHostKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes["host_key_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckHostKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_host_key" {
				continue
			}

			_, err := tftransfer.FindHostKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes["host_key_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Host Key %s still exists", rs.Primary.Attributes["host_key_id"])
		}

		return nil
	}
}

func testAccHostKeyConfig_basic(privateKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_transfer_host_key" "test" {
  server_id     = aws_transfer_server.test.id
  host_key_body = <<EOT
%[1]s
EOT
}
`, privateKey)
}

func testAccHostKeyConfig_description(privateKey, description string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_transfer_host_key" "test" {
  server_id     = aws_transfer_server.test.id
  description   = %[2]q
  host_key_body = <<EOT
%[1]s
EOT
}
`, privateKey, description)
}

func testAccHostKeyConfig_tags1(privateKey, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_transfer_host_key" "test" {
  server_id     = aws_transfer_server.test.id
  host_key_body = <<EOT
%[1]s
EOT

  tags = {
    %[2]q = %[3]q
  }
}
`, privateKey, tag1Key, tag1Value)
}

func testAccHostKeyConfig_tags2(privateKey, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_transfer_host_key" "test" {
  server_id     = aws_transfer_server.test.id
  host_key_body = <<EOT
%[1]s
EOT

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, privateKey, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccHostKeyConfig_hostKeyBodyWO(privateKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_transfer_host_key" "test" {
  server_id        = aws_transfer_server.test.id
  host_key_body_wo = <<EOT
%[1]s
EOT
}
`, privateKey)
}
