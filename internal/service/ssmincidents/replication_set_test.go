package ssmincidents_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmincidents "github.com/hashicorp/terraform-provider-aws/internal/service/ssmincidents"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMIncidentsReplicationSet_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Tests": {
			"basic":            testAccReplicationSet_basic,
			"updateDefaultKey": testAccReplicationSet_updateRegionsWithoutCMK,
			"updateCMK":        testAccReplicationSet_updateRegionsWithCMK,
			"updateTags":       testAccReplicationSet_updateTags,
			"updateEmptyTags":  testAccReplicationSet_updateEmptyTags,
			"disappears":       testAccReplicationSet_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccReplicationSet_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()
	region2 := acctest.AlternateRegion()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicRegions(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name":        region1,
						"kms_key_arn": "DefaultKey",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name":        region2,
						"kms_key_arn": "DefaultKey",
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
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

	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()
	region2 := acctest.AlternateRegion()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 2),
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicRegions(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name":        region1,
						"kms_key_arn": "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_basicRegions(region1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_basicRegions(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name":        region1,
						"kms_key_arn": "DefaultKey",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name":        region2,
						"kms_key_arn": "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_basicRegions(region1, region2),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_basicRegions(region1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name":        region1,
						"kms_key_arn": "DefaultKey",
					}),

					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config:            testAccReplicationSetConfig_basicRegions(region1),
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

	resourceName := "aws_ssmincidents_replication_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 2),
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_regionsWithCMK(map[string]string{
					acctest.Region(): "aws_kms_key.default.arn",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name": acctest.Region(),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_regionsWithCMK(map[string]string{
					acctest.Region(): "aws_kms_key.default.arn",
				}),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_regionsWithCMK(map[string]string{
					acctest.Region():          "aws_kms_key.default.arn",
					acctest.AlternateRegion(): "aws_kms_key.alternate.arn",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name": acctest.Region(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name": acctest.AlternateRegion(),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_regionsWithCMK(map[string]string{
					acctest.Region():          "aws_kms_key.default.arn",
					acctest.AlternateRegion(): "aws_kms_key.alternate.arn",
				}),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_regionsWithCMK(map[string]string{
					acctest.Region(): "aws_kms_key.default.arn",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "region.*", map[string]string{
						"name": acctest.Region(),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_regionsWithCMK(map[string]string{
					acctest.Region(): "aws_kms_key.default.arn",
				}),
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1(rProviderKey1, rProviderVal1Ini),
					testAccReplicationSetConfig_tags(map[string]string{
						rKey1: rVal1Ini,
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Ini),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey1, rProviderVal1Ini),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
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
					testAccReplicationSetConfig_tags(map[string]string{
						rKey1: rVal1Updated,
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey1, rProviderVal1Upd),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
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
					testAccReplicationSetConfig_tags(map[string]string{
						rKey2: rVal2,
						rKey3: rVal3,
					}),
				),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, rVal2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey3, rVal3),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey2, rProviderVal2),
					resource.TestCheckResourceAttr(resourceName, "tags_all."+rProviderKey3, rProviderVal3),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
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

	resourceName := "aws_ssmincidents_replication_set.test"

	rKey1 := sdkacctest.RandString(26)
	rKey2 := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_tags(map[string]string{
					rKey1: "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_tags(map[string]string{
					rKey1: "",
					rKey2: "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, ""),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_tags(map[string]string{
					rKey1: "",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, ""),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
		},
	})
}

func testAccReplicationSet_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := acctest.Region()
	region2 := acctest.AlternateRegion()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basicRegions(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssmincidents.ResourceReplicationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReplicationSetDestroy(tfState *terraform.State) error {
	client := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient()
	context := context.Background()

	for _, resource := range tfState.RootModule().Resources {
		if resource.Type != "aws_ssmincidents_replication_set" {
			continue
		}

		log.Printf("Checking Deletion of replication set resource: %s with ID: %s \n", resource.Type, resource.Primary.ID)

		_, err := tfssmincidents.FindReplicationSetByID(context, client, resource.Primary.ID)

		if tfresource.NotFound(err) {
			log.Printf("Replication Resource correctly returns NotFound Error... \n")
			continue
		}

		log.Printf("Replication Set Resource has incorrect Error\n")

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, resource.Primary.ID,
				errors.New("expected resource not found error, received an unexpected error"))
		}

		return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, resource.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckReplicationSetExists(name string) resource.TestCheckFunc {
	return func(tfState *terraform.State) error {
		resource, ok := tfState.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not found"))
		}

		if resource.Primary.ID == "" {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not set"))
		}

		client := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient()
		context := context.Background()

		_, err := tfssmincidents.FindReplicationSetByID(context, client, resource.Primary.ID)

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, resource.Primary.ID, err)
		}

		return nil
	}
}

func testAccReplicationSetConfig_basicRegions(regions ...string) string {
	return acctest.ConfigCompose(`
resource "aws_ssmincidents_replication_set" "test" {
`, generateReplicationSetRegionsConfig(regions...), `
}
`)
}

func generateReplicationSetRegionsConfig(regions ...string) string {
	regionSections := make([]string, 0)

	for _, region := range regions {
		regionSections = append(regionSections, fmt.Sprintf(`
		region {
			name = %[1]q
		}
		`, region))
	}
	return acctest.ConfigCompose(regionSections...)
}

func testAccReplicationSetConfig_tags(tags map[string]string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
`, acctest.Region()), generateReplicationSetTagsConfig(tags), `
}
`)
}

func generateReplicationSetTagsConfig(tags map[string]string) string {
	tagsSections := make([]string, 0)

	for key, value := range tags {
		tagsSections = append(tagsSections, fmt.Sprintf(`
			%[1]q = %[2]q`, key, value),
		)
	}

	return acctest.ConfigCompose(`
	tags = {`,
		acctest.ConfigCompose(tagsSections...), `
	}
	`)
}

func testAccReplicationSetConfigBaseKeyDefaultRegion() string {
	return `
resource "aws_kms_key" "default" {}
`
}

func testAccReplicationSetConfigBaseKeyAlternateRegion() string {
	return acctest.ConfigMultipleRegionProvider(2) + `
resource "aws_kms_key" "alternate" {
  provider = awsalternate
}
`
}

// input map maps regionName to variable containing arn of cmk key
func testAccReplicationSetConfig_regionsWithCMK(regions map[string]string) string {
	return acctest.ConfigCompose(
		testAccReplicationSetConfigBaseKeyDefaultRegion(),
		testAccReplicationSetConfigBaseKeyAlternateRegion(), `
resource "aws_ssmincidents_replication_set" "test" {
`, generateReplicationSetRegionsWithCMKConfig(regions), `
}
`)
}

func generateReplicationSetRegionsWithCMKConfig(regions map[string]string) string {
	regionsSections := make([]string, 0)

	for region, cmk := range regions {
		regionsSections = append(regionsSections, fmt.Sprintf(`
		region {
			name = %[1]q
			kms_key_arn = %[2]s
		}
		`, region, cmk))
	}
	return acctest.ConfigCompose(regionsSections...)
}
