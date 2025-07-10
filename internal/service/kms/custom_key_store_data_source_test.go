// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomKeyStoreDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	clusterID := acctest.SkipIfEnvVarNotSet(t, "CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := acctest.SkipIfEnvVarNotSet(t, "TRUST_ANCHOR_CERTIFICATE")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kms_custom_key_store.test"
	resourceName := "aws_kms_custom_key_store.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreDataSourceConfig_basic(rName, clusterID, trustAnchorCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cloud_hsm_cluster_id", resourceName, "cloud_hsm_cluster_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_key_store_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_key_store_name", resourceName, "custom_key_store_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "trust_anchor_certificate", resourceName, "trust_anchor_certificate"),
				),
			},
		},
	})
}

func testAccCustomKeyStoreDataSourceConfig_basic(rName, clusterId, anchorCertificate string) string {
	return fmt.Sprintf(`
resource "aws_kms_custom_key_store" "test" {
  cloud_hsm_cluster_id  = %[2]q
  custom_key_store_name = %[1]q
  key_store_password    = "noplaintextpasswords1"

  trust_anchor_certificate = file(%[3]q)
}

data "aws_kms_custom_key_store" "test" {
  custom_key_store_id = aws_kms_custom_key_store.test.id
}
`, rName, clusterId, anchorCertificate)
}
