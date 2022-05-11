package storagegateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccStorageGatewayNFSFileShare_basic(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_Required(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bucket_region", ""),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", rName),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestMatchResourceAttr(resourceName, "path", regexp.MustCompile(`^/.+`)),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "squash", "RootSquash"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_dns_name", ""),
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

func TestAccStorageGatewayNFSFileShare_audit(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"
	logResourceName := "aws_cloudwatch_log_group.test"
	logResourceNameSecond := "aws_cloudwatch_log_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareAuditConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareAuditUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceNameSecond, "arn"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_tags(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNFSFileShareTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_fileShareName(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareFileShareNameConfig(rName, "test_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "test_1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareFileShareNameConfig(rName, "test_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "test_2"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_clientList(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_ClientList_Single(rName, "1.1.1.1/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "1.1.1.1/32"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_ClientList_Multiple(rName, "2.2.2.2/32", "3.3.3.3/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "2.2.2.2/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "3.3.3.3/32"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_ClientList_Single(rName, "4.4.4.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "4.4.4.4"),
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

func TestAccStorageGatewayNFSFileShare_defaultStorageClass(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_DefaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_DefaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_ONEZONE_IA"),
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

func TestAccStorageGatewayNFSFileShare_guessMIMETypeEnabled(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_GuessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "false"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_GuessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
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

func TestAccStorageGatewayNFSFileShare_kmsEncrypted(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNFSFileShareConfig_KMSEncrypted(rName, true),
				ExpectError: regexp.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccNFSFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
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

func TestAccStorageGatewayNFSFileShare_kmsKeyARN(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"
	keyName := "aws_kms_key.test.0"
	keyUpdatedName := "aws_kms_key.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_KMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyName, "arn"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_KMSKeyARN_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyUpdatedName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_nFSFileShareDefaults(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_NFSFileShareDefaults(rName, "0700", "0600", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.directory_mode", "0700"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.file_mode", "0600"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.group_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.owner_id", "2"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_NFSFileShareDefaults(rName, "0770", "0660", 3, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.directory_mode", "0770"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.file_mode", "0660"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.group_id", "3"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.owner_id", "4"),
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

func TestAccStorageGatewayNFSFileShare_objectACL(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicRead),
				),
			},
			{
				Config: testAccNFSFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicReadWrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicReadWrite),
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

func TestAccStorageGatewayNFSFileShare_readOnly(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_ReadOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_ReadOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "true"),
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

func TestAccStorageGatewayNFSFileShare_requesterPays(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_RequesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_RequesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "true"),
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

func TestAccStorageGatewayNFSFileShare_squash(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_Squash(rName, "NoSquash"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "squash", "NoSquash"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_Squash(rName, "AllSquash"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "squash", "AllSquash"),
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

func TestAccStorageGatewayNFSFileShare_notificationPolicy(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareNotificationPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
				),
			},
			{
				Config: testAccNFSFileShareNotificationPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_cacheAttributes(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareCacheAttributesConfig(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareCacheAttributesConfig(rName, 500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "500"),
				),
			},
			{
				Config: testAccNFSFileShareCacheAttributesConfig(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_disappears(t *testing.T) {
	var nfsFileShare storagegateway.NFSFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNFSFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(resourceName, &nfsFileShare),
					acctest.CheckResourceDisappears(acctest.Provider, tfstoragegateway.ResourceNFSFileShare(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfstoragegateway.ResourceNFSFileShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNFSFileShareDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_nfs_file_share" {
			continue
		}

		_, err := tfstoragegateway.FindNFSFileShareByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Storage Gateway NFS File Share %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckNFSFileShareExists(resourceName string, nfsFileShare *storagegateway.NFSFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

		output, err := tfstoragegateway.FindNFSFileShareByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*nfsFileShare = *output

		return nil
	}
}

func testAcc_S3FileShareBase(rName string) string {
	return testAcc_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "storagegateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Effect": "Allow"
    }
  ]
}
POLICY

}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_storagegateway_gateway" "test" {
  depends_on = [aws_iam_role_policy.test]

  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccNFSFileShareConfig_Required(rName string) string {
	return testAcc_S3FileShareBase(rName) + `
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`
}

func testAccNFSFileShareFileShareNameConfig(rName, fsName string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list     = ["0.0.0.0/0"]
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  file_share_name = %[1]q
}
`, fsName)
}

func testAccNFSFileShareTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccNFSFileShareTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn

  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccNFSFileShareConfig_ClientList_Single(rName, clientList1 string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = [%q]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`, clientList1)
}

func testAccNFSFileShareConfig_ClientList_Multiple(rName, clientList1, clientList2 string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = [%q, %q]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`, clientList1, clientList2)
}

func testAccNFSFileShareConfig_DefaultStorageClass(rName, defaultStorageClass string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = ["0.0.0.0/0"]
  default_storage_class = %q
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
}
`, defaultStorageClass)
}

func testAccNFSFileShareConfig_GuessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list             = ["0.0.0.0/0"]
  gateway_arn             = aws_storagegateway_gateway.test.arn
  guess_mime_type_enabled = %t
  location_arn            = aws_s3_bucket.test.arn
  role_arn                = aws_iam_role.test.arn
}
`, guessMimeTypeEnabled)
}

func testAccNFSFileShareConfig_KMSEncrypted(rName string, kmsEncrypted bool) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list   = ["0.0.0.0/0"]
  gateway_arn   = aws_storagegateway_gateway.test.arn
  kms_encrypted = %t
  location_arn  = aws_s3_bucket.test.arn
  role_arn      = aws_iam_role.test.arn
}
`, kmsEncrypted)
}

func testAccNFSFileShareConfig_KMSKeyARN(rName string) string {
	return testAcc_S3FileShareBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_nfs_file_share" "test" {
  client_list   = ["0.0.0.0/0"]
  gateway_arn   = aws_storagegateway_gateway.test.arn
  kms_encrypted = true
  kms_key_arn   = aws_kms_key.test[0].arn
  location_arn  = aws_s3_bucket.test.arn
  role_arn      = aws_iam_role.test.arn
}
`
}

func testAccNFSFileShareConfig_KMSKeyARN_Update(rName string) string {
	return testAcc_S3FileShareBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_nfs_file_share" "test" {
  client_list   = ["0.0.0.0/0"]
  gateway_arn   = aws_storagegateway_gateway.test.arn
  kms_encrypted = true
  kms_key_arn   = aws_kms_key.test[1].arn
  location_arn  = aws_s3_bucket.test.arn
  role_arn      = aws_iam_role.test.arn
}
`
}

func testAccNFSFileShareConfig_NFSFileShareDefaults(rName, directoryMode, fileMode string, groupID, ownerID int) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn

  nfs_file_share_defaults {
    directory_mode = %q
    file_mode      = %q
    group_id       = %d
    owner_id       = %d
  }
}
`, directoryMode, fileMode, groupID, ownerID)
}

func testAccNFSFileShareConfig_ObjectACL(rName, objectACL string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  object_acl   = %q
  role_arn     = aws_iam_role.test.arn
}
`, objectACL)
}

func testAccNFSFileShareConfig_ReadOnly(rName string, readOnly bool) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  read_only    = %t
  role_arn     = aws_iam_role.test.arn
}
`, readOnly)
}

func testAccNFSFileShareConfig_RequesterPays(rName string, requesterPays bool) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list    = ["0.0.0.0/0"]
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  requester_pays = %t
  role_arn       = aws_iam_role.test.arn
}
`, requesterPays)
}

func testAccNFSFileShareConfig_Squash(rName, squash string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
  squash       = %q
}
`, squash)
}

func testAccNFSFileShareCacheAttributesConfig(rName string, timeout int) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn

  cache_attributes {
    cache_stale_timeout_in_seconds = %[1]d
  }
}
`, timeout)
}

func testAccNFSFileShareNotificationPolicyConfig(rName string) string {
	return testAcc_S3FileShareBase(rName) + `
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list         = ["0.0.0.0/0"]
  gateway_arn         = aws_storagegateway_gateway.test.arn
  location_arn        = aws_s3_bucket.test.arn
  role_arn            = aws_iam_role.test.arn
  notification_policy = "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"
}
`
}

func testAccNFSFileShareAuditConfig(rName string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = ["0.0.0.0/0"]
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
  audit_destination_arn = aws_cloudwatch_log_group.test.arn
}
`, rName)
}

func testAccNFSFileShareAuditUpdatedConfig(rName string) string {
	return testAcc_S3FileShareBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-updated"
}

resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = ["0.0.0.0/0"]
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
  audit_destination_arn = aws_cloudwatch_log_group.test2.arn
}
`, rName)
}
