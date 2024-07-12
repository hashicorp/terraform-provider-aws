// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrekognition "github.com/hashicorp/terraform-provider-aws/internal/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccCollectionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "collection_id", rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "face_model_version"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
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

func TestAccRekognitionCollection_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccCollectionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceCollection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRekognitionCollection_tags(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	tags1 := `
  tags = {
    key1 = "value1"
  }
`
	tags2 := `
  tags = {
    key1 = "value1"
    key2 = "value2"
  }
`
	tags3 := `
  tags = {
    key2 = "value2"
  }
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RekognitionEndpointID)
			testAccCollectionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RekognitionServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_tags(rName, tags1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccCollectionConfig_tags(rName, tags2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCollectionConfig_tags(rName, tags3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCollectionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameCollection, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameCollection, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)
		_, err := tfrekognition.FindCollectionByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameCollection, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckCollectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_collection" {
				continue
			}

			_, err := tfrekognition.FindCollectionByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameCollection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCollectionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

	input := &rekognition.ListCollectionsInput{}
	_, err := conn.ListCollections(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCollectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id = %[1]q

  tags = {
    test = 1
  }
}
`, rName)
}

func testAccCollectionConfig_tags(rName, tags string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id = %[1]q

%[2]s
}
`, rName, tags)
}
