package kms_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomKeyStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("CLOUD_HSM_CLUSTER_ID") == "" {
		t.Skip("CLOUD_HSM_CLUSTER_ID environment variable not set")
	}

	if os.Getenv("TRUST_ANCHOR_CERTIFICATE") == "" {
		t.Skip("TRUST_ANCHOR_CERTIFICATE environment variable not set")
	}

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customkeystore kms.CustomKeyStoresListEntry
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	clusterId := os.Getenv("CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := os.Getenv("TRUST_ANCHOR_CERTIFICATE")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kms.EndpointsID, t)
			testAccCustomKeyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName, clusterId, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, resourceName, &customkeystore),
					resource.TestCheckResourceAttr(resourceName, "cloud_hsm_cluster_id", clusterId),
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
	if os.Getenv("CLOUD_HSM_CLUSTER_ID") == "" {
		t.Skip("CLOUD_HSM_CLUSTER_ID environment variable not set")
	}

	if os.Getenv("TRUST_ANCHOR_CERTIFICATE") == "" {
		t.Skip("TRUST_ANCHOR_CERTIFICATE environment variable not set")
	}

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customkeystore kms.CustomKeyStoresListEntry
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	clusterId := os.Getenv("CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := os.Getenv("TRUST_ANCHOR_CERTIFICATE")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kms.EndpointsID, t)
			testAccCustomKeyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(fmt.Sprintf("%s-updated", rName), clusterId, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, resourceName, &customkeystore),
					resource.TestCheckResourceAttr(resourceName, "cloud_hsm_cluster_id", clusterId),
					resource.TestCheckResourceAttr(resourceName, "custom_key_store_name", fmt.Sprintf("%s-updated", rName)),
				),
			},
		},
	})
}

func testAccCustomKeyStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("CLOUD_HSM_CLUSTER_ID") == "" {
		t.Skip("CLOUD_HSM_CLUSTER_ID environment variable not set")
	}

	if os.Getenv("TRUST_ANCHOR_CERTIFICATE") == "" {
		t.Skip("TRUST_ANCHOR_CERTIFICATE environment variable not set")
	}

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customkeystore kms.CustomKeyStoresListEntry
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	clusterId := os.Getenv("CLOUD_HSM_CLUSTER_ID")
	trustAnchorCertificate := os.Getenv("TRUST_ANCHOR_CERTIFICATE")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kms.EndpointsID, t)
			testAccCustomKeyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName, clusterId, trustAnchorCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(ctx, resourceName, &customkeystore),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkms.ResourceCustomKeyStore(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomKeyStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kms_custom_key_store" {
				continue
			}

			in := &kms.DescribeCustomKeyStoresInput{
				CustomKeyStoreId: aws.String(rs.Primary.ID),
			}
			_, err := tfkms.FindCustomKeyStoreByID(ctx, conn, in)

			if tfresource.NotFound(err) {
				continue
			}

			return create.Error(names.KMS, create.ErrActionCheckingDestroyed, tfkms.ResNameCustomKeyStore, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCustomKeyStoreExists(ctx context.Context, name string, customkeystore *kms.CustomKeyStoresListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.KMS, create.ErrActionCheckingExistence, tfkms.ResNameCustomKeyStore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.KMS, create.ErrActionCheckingExistence, tfkms.ResNameCustomKeyStore, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

		in := &kms.DescribeCustomKeyStoresInput{
			CustomKeyStoreId: aws.String(rs.Primary.ID),
		}
		resp, err := tfkms.FindCustomKeyStoreByID(ctx, conn, in)

		if err != nil {
			return create.Error(names.KMS, create.ErrActionCheckingExistence, tfkms.ResNameCustomKeyStore, rs.Primary.ID, err)
		}

		*customkeystore = *resp

		return nil
	}
}

func testAccCustomKeyStoresPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn()

	input := &kms.DescribeCustomKeyStoresInput{}
	_, err := conn.DescribeCustomKeyStoresWithContext(ctx, input)

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
