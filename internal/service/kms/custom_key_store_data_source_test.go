package kms_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMSCustomKeyStoreDataSource_basic(t *testing.T) {
	if os.Getenv("CLOUD_HSM_CLUSTER_ID") == "" {
		t.Skip("CLOUD_HSM_CLUSTER_ID environment variable not set")
	}

	if os.Getenv("TRUST_ANCHOR_CERTIFICATE") == "" {
		t.Skip("TRUST_ANCHOR_CERTIFICATE environment variable not set")
	}

	resourceName := "aws_kms_custom_key_store.test"
	dataSourceName := "data.aws_kms_custom_key_store.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterId := os.Getenv("CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := os.Getenv("TRUST_ANCHOR_CERTIFICATE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreDataSourceConfig_basic(rName, clusterId, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_key_store_name", resourceName, "custom_key_store_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_key_store_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "trust_anchor_certificate", resourceName, "trust_anchor_certificate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cloud_hsm_cluster_id", resourceName, "cloud_hsm_cluster_id"),
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
