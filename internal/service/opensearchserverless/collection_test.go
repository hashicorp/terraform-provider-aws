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

func TestAccOpenSearchServerlessCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var collection types.CollectionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrType),
					resource.TestCheckResourceAttrSet(resourceName, "collection_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "dashboard_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
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

func TestAccOpenSearchServerlessCollection_standbyReplicas(t *testing.T) {
	ctx := acctest.Context(t)
	var collection types.CollectionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	standbyReplicas := "DISABLED"
	resourceName := "aws_opensearchserverless_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_standbyReplicas(rName, standbyReplicas),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrType),
					resource.TestCheckResourceAttrSet(resourceName, "collection_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "dashboard_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttr(resourceName, "standby_replicas", standbyReplicas),
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

func TestAccOpenSearchServerlessCollection_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var collection types.CollectionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccCollectionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCollectionConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollection_update(t *testing.T) {
	ctx := acctest.Context(t)
	var collection types.CollectionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccCollectionConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollection_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var collection types.CollectionDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_collection.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, t, resourceName, &collection),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceCollection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCollectionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_collection" {
				continue
			}

			_, err := tfopensearchserverless.FindCollectionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameCollection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCollectionExists(ctx context.Context, t *testing.T, name string, collection *types.CollectionDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollection, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollection, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)
		resp, err := tfopensearchserverless.FindCollectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameCollection, rs.Primary.ID, err)
		}

		*collection = *resp

		return nil
	}
}

func testAccPreCheckCollection(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListCollectionsInput{}
	_, err := conn.ListCollections(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCollectionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name = %[1]q
  type = "encryption"
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
          "collection/%[1]s"
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}
`, rName)
}

func testAccCollectionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCollectionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_opensearchserverless_collection" "test" {
  name = %[1]q

  depends_on = [aws_opensearchserverless_security_policy.test]
}
`, rName),
	)
}

func testAccCollectionConfig_standbyReplicas(rName string, standbyReplicas string) string {
	return acctest.ConfigCompose(
		testAccCollectionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_opensearchserverless_collection" "test" {
  name             = %[1]q
  standby_replicas = %[2]q

  depends_on = [aws_opensearchserverless_security_policy.test]
}
`, rName, standbyReplicas),
	)
}

func testAccCollectionConfig_update(rName, description string) string {
	return acctest.ConfigCompose(
		testAccCollectionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_opensearchserverless_collection" "test" {
  name        = %[1]q
  description = %[2]q

  depends_on = [aws_opensearchserverless_security_policy.test]
}
`, rName, description),
	)
}

func testAccCollectionConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccCollectionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_opensearchserverless_collection" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_opensearchserverless_security_policy.test]
}
`, rName, key1, value1),
	)
}

func testAccCollectionConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccCollectionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_opensearchserverless_collection" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_opensearchserverless_security_policy.test]
}
`, rName, key1, value1, key2, value2),
	)
}
