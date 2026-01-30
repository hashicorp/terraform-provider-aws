// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsPolicyStoreDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	dataSourceName := "data.aws_verifiedpermissions_policy_store.test"
	resourceName := "aws_verifiedpermissions_policy_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreDataSourceConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, dataSourceName, &policystore),
					resource.TestCheckResourceAttrPair(resourceName, "validation_settings.0.mode", dataSourceName, "validation_settings.0.mode"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "DISABLED"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStoreDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	dataSourceName := "data.aws_verifiedpermissions_policy_store.test"
	resourceName := "aws_verifiedpermissions_policy_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreDataSourceConfig_tags("OFF", acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, dataSourceName, &policystore),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsKey1, dataSourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsKey2, dataSourceName, acctest.CtTagsKey2),
				),
			},
		},
	})
}

func testAccPolicyStoreDataSourceConfig_basic(mode string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = %[1]q
  }
}

data "aws_verifiedpermissions_policy_store" "test" {
  id = aws_verifiedpermissions_policy_store.test.id
}
`, mode)
}

func testAccPolicyStoreDataSourceConfig_tags(mode string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = %[1]q
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

data "aws_verifiedpermissions_policy_store" "test" {
  id = aws_verifiedpermissions_policy_store.test.id
}
`, mode, tagKey1, tagValue1, tagKey2, tagValue2)
}
