package kms_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

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

func TestAccKMSCustomKeyStore_basic(t *testing.T) {
	if os.Getenv("CLOUD_HSM_CLUSTER_ID") == "" {
		t.Skip("CLOUD_HSM_CLUSTER_ID environment variable not set")
	}

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customkeystore kms.CustomKeyStoresListEntry
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kms.EndpointsID, t)
			testAccCustomKeyStoresPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(resourceName, &customkeystore),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kms", regexp.MustCompile(`customkeystore:+.`)),
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

func TestAccKMSCustomKeyStore_disappears(t *testing.T) {
	if os.Getenv("CLOUD_HSM_CLUSTER_ID") == "" {
		t.Skip("CLOUD_HSM_CLUSTER_ID environment variable not set")
	}

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var customkeystore kms.CustomKeyStoresListEntry
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_custom_key_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(kms.EndpointsID, t)
			testAccCustomKeyStoresPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomKeyStoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomKeyStoreConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomKeyStoreExists(resourceName, &customkeystore),
					acctest.CheckResourceDisappears(acctest.Provider, tfkms.ResourceCustomKeyStore(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomKeyStoreDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_custom_key_store" {
			continue
		}

		_, err := tfkms.FindCustomKeyStoreByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		return create.Error(names.KMS, create.ErrActionCheckingDestroyed, tfkms.ResNameCustomKeyStore, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckCustomKeyStoreExists(name string, customkeystore *kms.CustomKeyStoresListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.KMS, create.ErrActionCheckingExistence, tfkms.ResNameCustomKeyStore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.KMS, create.ErrActionCheckingExistence, tfkms.ResNameCustomKeyStore, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn
		ctx := context.Background()
		resp, err := tfkms.FindCustomKeyStoreByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.KMS, create.ErrActionCheckingExistence, tfkms.ResNameCustomKeyStore, rs.Primary.ID, err)
		}

		*customkeystore = *resp

		return nil
	}
}

func testAccCustomKeyStoresPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn
	ctx := context.Background()

	input := &kms.DescribeCustomKeyStoresInput{}
	_, err := conn.DescribeCustomKeyStoresWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCustomKeyStoreConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_custom_key_store" "test" {
  custom_key_store_name             = %[1]q
  engine_type             = "ActiveKMS"
  host_instance_type      = "kms.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}
