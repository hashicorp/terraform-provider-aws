// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)

	// var collection rekognition.CreateCollectionOutput
	rCollectionId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// CheckDestroy:             testAccCheckBotVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rCollectionId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "collection_id", rCollectionId),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "face_model_version"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCollectionConfig_basic(rCollectionId string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id             = "%s"
}
`, rCollectionId)
}
