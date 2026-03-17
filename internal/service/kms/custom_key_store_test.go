// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomKeyStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	clusterID := acctest.SkipIfEnvVarNotSet(t, "CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := acctest.SkipIfEnvVarNotSet(t, "TRUST_ANCHOR_CERTIFICATE")
	var customkeystore awstypes.CustomKeyStoresListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KMSEndpointID)
			testAccCustomKeyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName, clusterID, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, t, resourceName, &customkeystore),
					resource.TestCheckResourceAttr(resourceName, "cloud_hsm_cluster_id", clusterID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key_store_password"},
			},
		},
	})
}

func testAccCustomKeyStore_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	clusterID := acctest.SkipIfEnvVarNotSet(t, "CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := acctest.SkipIfEnvVarNotSet(t, "TRUST_ANCHOR_CERTIFICATE")
	var customkeystore awstypes.CustomKeyStoresListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KMSEndpointID)
			testAccCustomKeyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName, clusterID, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, t, resourceName, &customkeystore),
					resource.TestCheckResourceAttr(resourceName, "custom_key_store_name", rName),
				),
			},
			{
				Config: testAccCustomKeyStoreConfig_basic(fmt.Sprintf("%s-updated", rName), clusterID, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, t, resourceName, &customkeystore),
					resource.TestCheckResourceAttr(resourceName, "custom_key_store_name", fmt.Sprintf("%s-updated", rName)),
				),
			},
		},
	})
}

func testAccCustomKeyStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	clusterID := acctest.SkipIfEnvVarNotSet(t, "CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := acctest.SkipIfEnvVarNotSet(t, "TRUST_ANCHOR_CERTIFICATE")
	var customkeystore awstypes.CustomKeyStoresListEntry
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.KMSEndpointID)
			testAccCustomKeyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName, clusterID, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, t, resourceName, &customkeystore),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkms.ResourceCustomKeyStore(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomKeyStoreDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kms_custom_key_store" {
				continue
			}

			_, err := tfkms.FindCustomKeyStoreByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("KMS Custom Key Store %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomKeyStoreExists(ctx context.Context, t *testing.T, n string, v *awstypes.CustomKeyStoresListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		output, err := tfkms.FindCustomKeyStoreByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCustomKeyStoresPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

	_, err := conn.DescribeCustomKeyStores(ctx, &kms.DescribeCustomKeyStoresInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCustomKeyStoreConfig_basic(rName, clusterId, anchorCertificate string) string {
	return fmt.Sprintf(`
resource "aws_kms_custom_key_store" "test" {
  cloud_hsm_cluster_id  = %[2]q
  custom_key_store_name = %[1]q
  key_store_password    = "noplaintextpasswords1"

  trust_anchor_certificate = file(%[3]q)
}
`, rName, clusterId, anchorCertificate)
}
