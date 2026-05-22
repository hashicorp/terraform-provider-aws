// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessCollectionGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "standby_replicas", "ENABLED"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "aoss", "collection-group/{id}"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
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
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_capacityLimits(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_capacityLimits(rName, 2, 16, 2, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.min_indexing_capacity_in_ocu", "2"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.max_indexing_capacity_in_ocu", "16"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.min_search_capacity_in_ocu", "2"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.max_search_capacity_in_ocu", "16"),
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
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.min_indexing_capacity_in_ocu", "4"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.max_indexing_capacity_in_ocu", "32"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.min_search_capacity_in_ocu", "4"),
					resource.TestCheckResourceAttr(resourceName, "capacity_limits.0.max_search_capacity_in_ocu", "32"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroup_standbyReplicas(t *testing.T) {
	ctx := acctest.Context(t)
	var collectionGroup types.CollectionGroupDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupConfig_standbyReplicas(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionGroupExists(ctx, t, resourceName, &collectionGroup),
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

func testAccCheckCollectionGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_collection_group" {
				continue
			}

			input := opensearchserverless.BatchGetCollectionGroupInput{
				Ids: []string{rs.Primary.ID},
			}
			_, err := tfopensearchserverless.FindCollectionGroup(ctx, conn, &input)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Serverless Collection Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCollectionGroupExists(ctx context.Context, t *testing.T, name string, collectionGroup *types.CollectionGroupDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)
		input := opensearchserverless.BatchGetCollectionGroupInput{
			Ids: []string{rs.Primary.ID},
		}
		output, err := tfopensearchserverless.FindCollectionGroup(ctx, conn, &input)

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollectionGroup, rs.Primary.ID, err)
		}

		*collectionGroup = *output

		return nil
	}
}

func testAccPreCheckCollectionGroup(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

	input := opensearchserverless.ListCollectionGroupsInput{}
	_, err := conn.ListCollectionGroups(ctx, &input)

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

  capacity_limits {
    max_indexing_capacity_in_ocu = 1
    max_search_capacity_in_ocu   = 1
  }
}
`, rName, description)
}

func testAccCollectionGroupConfig_capacityLimits(rName string, minIndex, maxIndex, minSearch, maxSearch int) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  standby_replicas = "ENABLED"

  capacity_limits {
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

  capacity_limits {
    max_indexing_capacity_in_ocu = 1
    max_search_capacity_in_ocu   = 1
  }
}
`, rName, standbyReplicas)
}
