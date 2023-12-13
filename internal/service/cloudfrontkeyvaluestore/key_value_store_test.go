// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfrontkeyvaluestore_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCloudFrontKeyValueStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, cloudfront.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudfront.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// CheckDestroy:             testAccCheckFunctionDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccCloudFrontKeyValueStoreConfig_basic(rName),
			},
		},
	})
}

func testAccCloudFrontKeyValueStoreConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_key_value_store" "test" {
 key_value_store_name    = %[1]q
 key     				 = "hello"
 value   				 = "world"
}
`, rName)
}
