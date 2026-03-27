// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessCollectionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "standby_replicas", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
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

func TestAccOpenSearchServerlessCollectionGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceCollectionGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_update(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCollectionGroupConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_capacityLimits(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_capacityLimits(rName, 2, 16, 2, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.min_indexing_capacity_in_ocu", "2"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.max_indexing_capacity_in_ocu", "16"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.min_search_capacity_in_ocu", "2"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.max_search_capacity_in_ocu", "16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCollectionGroupConfig_capacityLimits(rName, 4, 32, 4, 32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.min_indexing_capacity_in_ocu", "4"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.max_indexing_capacity_in_ocu", "32"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.min_search_capacity_in_ocu", "4"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.max_search_capacity_in_ocu", "32"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_standbyReplicas(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_standbyReplicas(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, "standby_replicas", "DISABLED"),
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

func TestAccOpenSearchServerlessCollectionGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
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
				Config: testAccCollectionGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCollectionGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCollectionGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_collection_group" {
				continue
			}

			input := &opensearchserverless.BatchGetCollectionGroupInput{
				Ids: []string{rs.Primary.ID},
			}
			output, err := conn.BatchGetCollectionGroup(ctx, input)

			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			if output != nil && len(output.CollectionGroupDetails) > 0 {
				return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameCollectionGroup, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckCollectionGroupExists(ctx context.Context, name string, collectionGroup *types.CollectionGroupDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

		input := &opensearchserverless.BatchGetCollectionGroupInput{
			Ids: []string{rs.Primary.ID},
		}
		output, err := conn.BatchGetCollectionGroup(ctx, input)

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, rs.Primary.ID, err)
		}

		if output == nil || len(output.CollectionGroupDetails) == 0 {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, rs.Primary.ID, errors.New("not found"))
		}

		*collectionGroup = output.CollectionGroupDetails[0]

		return nil
	}
}

func testAccPreCheckCollectionGroup(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListCollectionGroupsInput{}
	_, err := conn.ListCollectionGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCollectionGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  standby_replicas = "ENABLED"
}
`, rName)
}

func testAccCollectionGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  description      = %[2]q
  standby_replicas = "ENABLED"
}
`, rName, description)
}

func testAccCollectionGroupConfig_capacityLimits(rName string, minIndex, maxIndex, minSearch, maxSearch int) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  standby_replicas = "ENABLED"

  capacity_limits = {
    min_indexing_capacity_in_ocu = %[2]d
    max_indexing_capacity_in_ocu = %[3]d
    min_search_capacity_in_ocu   = %[4]d
    max_search_capacity_in_ocu   = %[5]d
  }
}
`, rName, minIndex, maxIndex, minSearch, maxSearch)
}

func testAccCollectionGroupConfig_standbyReplicas(rName, standbyReplicas string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  standby_replicas = %[2]q
}
`, rName, standbyReplicas)
}

func testAccCollectionGroupConfig_tags1(rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  standby_replicas = "ENABLED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1)
}

func testAccCollectionGroupConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  standby_replicas = "ENABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2)
}

func TestAccOpenSearchServerlessCollectionGroup_nameValidation(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccCollectionGroupConfig_basic("ab"),
				ExpectError: regexache.MustCompile(`Attribute name string length must be between 3 and 32`),
			},
			{
				Config:      testAccCollectionGroupConfig_basic("a" + sdkacctest.RandString(32)),
				ExpectError: regexache.MustCompile(`Attribute name string length must be between 3 and 32`),
			},
			{
				Config:      testAccCollectionGroupConfig_basic("Abc"),
				ExpectError: regexache.MustCompile(`must start with any lower case letter`),
			},
			{
				Config:      testAccCollectionGroupConfig_basic("abc_def"),
				ExpectError: regexache.MustCompile(`must start with any lower case letter`),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_capacityLimitsValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccCollectionGroupConfig_capacityLimits(rName, 16, 2, 2, 16),
				ExpectError: regexache.MustCompile(`ValidationException`),
			},
			{
				Config:      testAccCollectionGroupConfig_capacityLimits(rName, 2, 16, 16, 2),
				ExpectError: regexache.MustCompile(`ValidationException`),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_descriptionValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	longDescription := sdkacctest.RandString(1001)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccCollectionGroupConfig_description(rName, longDescription),
				ExpectError: regexache.MustCompile(`Attribute description string length must be between 0 and 1000`),
			},
		},
	})
}
