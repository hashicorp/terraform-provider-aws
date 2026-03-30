// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssmincidents "github.com/hashicorp/terraform-provider-aws/internal/service/ssmincidents"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccReplicationSet_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()
	region2 := acctest.AlternateRegion()

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region2,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
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

func testAccReplicationSet_updateRegionsWithoutCMK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()
	region2 := acctest.AlternateRegion()

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicOneRegion(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region2,
						names.AttrKMSKeyARN: "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_basicOneRegion(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
				),
			},
		},
	})
}

func testAccReplicationSet_updateRegionsWithCMK(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_oneRegionWithCMK(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName: acctest.Region(),
					}),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_twoRegionWithCMK(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName: acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName: acctest.AlternateRegion(),
					}),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_oneRegionWithCMK(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName: acctest.Region(),
					}),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
				),
			},
		},
	})
}

func testAccReplicationSet_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccReplicationSetConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
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
				Config: acctest.ConfigCompose(
					testAccReplicationSetConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccReplicationSetConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccReplicationSet_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()
	region2 := acctest.AlternateRegion()

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfssmincidents.ResourceReplicationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccReplicationSet_deprecatedRegion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_deprecatedRegion(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set/.+$`)),
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

func testAccCheckReplicationSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMIncidentsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmincidents_replication_set" {
				continue
			}

			_, err := tfssmincidents.FindReplicationSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SSMIncidents Replication Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckReplicationSetExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SSMIncidentsClient(ctx)

		_, err := tfssmincidents.FindReplicationSetByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccReplicationSetConfig_basicOneRegion(region1 string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  regions {
    name = %[1]q
  }
}
`, region1)
}

func testAccReplicationSetConfig_basicTwoRegion(region1, region2 string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  regions {
    name = %[1]q
  }
  regions {
    name = %[2]q
  }
}
`, region1, region2)
}

func testAccReplicationSetConfig_tags1(tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  regions {
    name = %[3]q
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value, acctest.Region())
}

func testAccReplicationSetConfig_tags2(tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  regions {
    name = %[5]q
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value, acctest.Region())
}

func testAccReplicationSetConfig_baseKeyDefaultRegion() string {
	return `
resource "aws_kms_key" "default" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`
}

func testAccReplicationSetConfig_baseKeyAlternateRegion() string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), `
resource "aws_kms_key" "alternate" {
  provider                = awsalternate
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`)
}

func testAccReplicationSetConfig_oneRegionWithCMK() string {
	return acctest.ConfigCompose(
		testAccReplicationSetConfig_baseKeyDefaultRegion(),
		testAccReplicationSetConfig_baseKeyAlternateRegion(),
		fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  regions {
    name        = %[1]q
    kms_key_arn = aws_kms_key.default.arn
  }
}
`, acctest.Region()))
}

func testAccReplicationSetConfig_twoRegionWithCMK() string {
	return acctest.ConfigCompose(
		testAccReplicationSetConfig_baseKeyDefaultRegion(),
		testAccReplicationSetConfig_baseKeyAlternateRegion(),
		fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  regions {
    name        = %[1]q
    kms_key_arn = aws_kms_key.default.arn
  }
  regions {
    name        = %[2]q
    kms_key_arn = aws_kms_key.alternate.arn
  }
}
`, acctest.Region(), acctest.AlternateRegion()))
}

func testAccReplicationSetConfig_deprecatedRegion(region1 string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, region1)
}
