// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmincidents "github.com/hashicorp/terraform-provider-aws/internal/service/ssmincidents"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region2,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicOneRegion(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_basicOneRegion(region1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region2,
						names.AttrKMSKeyARN: "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_basicOneRegion(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName:      region1,
						names.AttrKMSKeyARN: "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_basicOneRegion(region1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_oneRegionWithCMK(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName: acctest.Region(),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_oneRegionWithCMK(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_twoRegionWithCMK(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName: acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName: acctest.AlternateRegion(),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_twoRegionWithCMK(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_oneRegionWithCMK(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						names.AttrName: acctest.Region(),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_oneRegionWithCMK(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReplicationSet_updateTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"

	rKey1 := sdkacctest.RandString(26)
	rVal1Ini := sdkacctest.RandString(26)
	rVal1Updated := sdkacctest.RandString(26)
	rKey2 := sdkacctest.RandString(26)
	rVal2 := sdkacctest.RandString(26)
	rKey3 := sdkacctest.RandString(26)
	rVal3 := sdkacctest.RandString(26)

	rProviderKey1 := sdkacctest.RandString(26)
	rProviderVal1Ini := sdkacctest.RandString(26)
	rProviderVal1Upd := sdkacctest.RandString(26)
	rProviderKey2 := sdkacctest.RandString(26)
	rProviderVal2 := sdkacctest.RandString(26)
	rProviderKey3 := sdkacctest.RandString(26)
	rProviderVal3 := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(rProviderKey1, rProviderVal1Ini),
					testAccReplicationSetConfig_oneTag(rKey1, rVal1Ini),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Ini),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey1, rProviderVal1Ini),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(rProviderKey1, rProviderVal1Upd),
					testAccReplicationSetConfig_oneTag(rKey1, rVal1Updated),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey1, rProviderVal1Upd),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2(rProviderKey2, rProviderVal2, rProviderKey3, rProviderVal3),
					testAccReplicationSetConfig_twoTags(rKey2, rVal2, rKey3, rVal3),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, rVal2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey3, rVal3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey2, rProviderVal2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey3, rProviderVal3),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
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

func testAccReplicationSet_updateEmptyTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_ssmincidents_replication_set.test"

	rKey1 := sdkacctest.RandString(26)
	rKey2 := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_oneTag(rKey1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_twoTags(rKey1, "", rKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, ""),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_oneTag(rKey1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ssm-incidents", regexache.MustCompile(`replication-set\/+.`)),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfssmincidents.ResourceReplicationSet(),
						resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReplicationSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient(ctx)

		for _, resource := range s.RootModule().Resources {
			if resource.Type != "aws_ssmincidents_replication_set" {
				continue
			}

			log.Printf("Checking Deletion of replication set resource: %s with ID: %s \n", resource.Type, resource.Primary.ID)

			_, err := tfssmincidents.FindReplicationSetByID(ctx, client, resource.Primary.ID)

			if tfresource.NotFound(err) {
				log.Printf("Replication Resource correctly returns NotFound Error... \n")
				continue
			}

			if err != nil {
				return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, resource.Primary.ID,
					errors.New("expected resource not found error, received an unexpected error"))
			}

			return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, resource.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckReplicationSetExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not found"))
		}

		if resource.Primary.ID == "" {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not set"))
		}

		client := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient(ctx)

		_, err := tfssmincidents.FindReplicationSetByID(ctx, client, resource.Primary.ID)

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, resource.Primary.ID, err)
		}

		return nil
	}
}

func testAccReplicationSetConfig_basicOneRegion(region1 string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, region1)
}

func testAccReplicationSetConfig_basicTwoRegion(region1, region2 string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
  region {
    name = %[2]q
  }
}
`, region1, region2)
}

func testAccReplicationSetConfig_oneTag(tagKey, tagVal string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[3]q
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey, tagVal, acctest.Region())
}

func testAccReplicationSetConfig_twoTags(tag1Key, tag1Val, tag2Key, tag2Val string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[5]q
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Val, tag2Key, tag2Val, acctest.Region())
}

func testAccReplicationSetConfig_baseKeyDefaultRegion() string {
	return `
resource "aws_kms_key" "default" {}
`
}

func testAccReplicationSetConfig_baseKeyAlternateRegion() string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), `
resource "aws_kms_key" "alternate" {
  provider = awsalternate
}
`)
}

func testAccReplicationSetConfig_oneRegionWithCMK() string {
	return acctest.ConfigCompose(
		testAccReplicationSetConfig_baseKeyDefaultRegion(),
		testAccReplicationSetConfig_baseKeyAlternateRegion(),
		fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
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
  region {
    name        = %[1]q
    kms_key_arn = aws_kms_key.default.arn
  }
  region {
    name        = %[2]q
    kms_key_arn = aws_kms_key.alternate.arn
  }
}
`, acctest.Region(), acctest.AlternateRegion()))
}
