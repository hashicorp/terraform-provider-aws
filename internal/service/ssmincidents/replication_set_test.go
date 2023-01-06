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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	tfssmincidents "github.com/hashicorp/terraform-provider-aws/internal/service/ssmincidents"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMIncidentsReplicationSet_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Tests": {
			"basic":            testAccReplicationSet_basic,
			"updateDefaultKey": testAccReplicationSet_updateDefaultKey,
			"updateCMK":        testAccReplicationSet_updateCustomerManagedKey,
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
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
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

func testAccReplicationSet_updateDefaultKey(t *testing.T) {
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
				Config: testAccReplicationSetConfig_basicOneRegion(region1),
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
				Config:            testAccReplicationSetConfig_basicOneRegion(region1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
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
				Config:            testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_basicOneRegion(region1),
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
				Config:            testAccReplicationSetConfig_basicOneRegion(region1),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccReplicationSet_updateCustomerManagedKey(t *testing.T) {
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
				Config: testAccReplicationSetConfig_oneRegionCMK(),
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
				Config:            testAccReplicationSetConfig_oneRegionCMK(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_twoRegionCMK(),
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
				Config:            testAccReplicationSetConfig_twoRegionCMK(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_oneRegionCMK(),
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
				Config:            testAccReplicationSetConfig_oneRegionCMK(),
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
				Config: testAccReplicationSetConfig_oneTag(rKey1, rVal1Ini),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Ini),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_oneTag(rKey1, rVal1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey1, rVal1Updated),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccReplicationSetConfig_twoTags(rKey2, rVal2, rKey3, rVal3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, rVal2),
					resource.TestCheckResourceAttr(resourceName, "tags."+rKey2, rVal2),
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
				Config: testAccReplicationSetConfig_oneTag(rKey1, ""),
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
				Config: testAccReplicationSetConfig_twoTags(rKey1, "", rKey2, ""),
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
				Config: testAccReplicationSetConfig_oneTag(rKey1, ""),
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
				Config: testAccReplicationSetConfig_basicTwoRegion(region1, region2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssmincidents.ResourceReplicationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReplicationSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssmincidents_replication_set" {
			continue
		}

		log.Printf("Checking Deletion of replication set resource: %s with ID: %s \n", rs.Type, rs.Primary.ID)

		_, err := tfssmincidents.FindReplicationSetByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			log.Printf("Replication Resource correctly returns NotFound Error... \n")
			continue
		}

		log.Printf("Replication Set Resource has incorrect Error\n")

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, rs.Primary.ID,
				errors.New("expected resource not found error, received an unexpected error"))
		}

		return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckReplicationSetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient()
		ctx := context.Background()

		_, err := tfssmincidents.FindReplicationSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, rs.Primary.ID, err)
		}

		return nil
	}
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

func testAccReplicationSetConfig_basicOneRegion(region1 string) string {
	return fmt.Sprintf(`

resource "aws_ssmincidents_replication_set" "test" {
  region {
	name = %[1]q
  }

}
`, region1)
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
`, tfssmincidents.Trim(tagKey), tagVal, acctest.Region())
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
`, tfssmincidents.Trim(tag1Key), tag1Val, tfssmincidents.Trim(tag2Key), tag2Val, acctest.Region())
}

func testAccReplicationSetConfigBaseKey1() string {
	return `
resource "aws_kms_key" "default" {}	
`
}

func testAccReplicationSetConfigBaseKey2() string {
	return acctest.ConfigMultipleRegionProvider(2) + `
resource "aws_kms_key" "alternate" {
	provider = awsalternate
}
`
}

func testAccReplicationSetConfig_oneRegionCMK() string {
	return acctest.ConfigCompose(
		testAccReplicationSetConfigBaseKey1(),
		testAccReplicationSetConfigBaseKey2(),
		fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
	region {
		name = %[1]q
		kms_key_arn = aws_kms_key.default.arn
	}
}
`, acctest.Region()))
}

func testAccReplicationSetConfig_twoRegionCMK() string {
	return acctest.ConfigCompose(
		testAccReplicationSetConfigBaseKey1(),
		testAccReplicationSetConfigBaseKey2(),
		fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
	region {
		name = %[1]q
		kms_key_arn = aws_kms_key.default.arn
	}

	region {
		name = %[2]q
		kms_key_arn = aws_kms_key.alternate.arn
	}
}
`, acctest.Region(), acctest.AlternateRegion()))
}
