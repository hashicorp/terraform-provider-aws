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

func TestAccStorageGatewaySMBFileShare_Authentication_activeDirectory(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_Authentication_ActiveDirectory(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "ActiveDirectory"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "file_share_name", regexp.MustCompile(`^tf-acc-test-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestMatchResourceAttr(resourceName, "path", regexp.MustCompile(`^/.+`)),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
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

func TestAccStorageGatewaySMBFileShare_Authentication_guestAccess(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_Authentication_GuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "GuestAccess"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "ClientSpecified"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "file_share_name", regexp.MustCompile(`^tf-acc-test-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
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

func TestAccStorageGatewaySMBFileShare_accessBasedEnumeration(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareAccessBasedEnumerationConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareAccessBasedEnumerationConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", "false"),
				),
			},
			{
				Config: testAccSMBFileShareAccessBasedEnumerationConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", "true"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_notificationPolicy(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareNotificationPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_Authentication_GuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
				),
			},
			{
				Config: testAccSMBFileShareNotificationPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_defaultStorageClass(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_DefaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_DefaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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

func TestAccStorageGatewaySMBFileShare_fileShareName(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_FileShareName(rName, "foo_share"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "foo_share"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_FileShareName(rName, "bar_share"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "bar_share"),
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

func TestAccStorageGatewaySMBFileShare_tags(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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
				Config: testAccSMBFileShareTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSMBFileShareTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_guessMIMETypeEnabled(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_GuessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "false"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_GuessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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

/*
Currently failing when enabling oplocks:

=== CONT  TestAccStorageGatewaySMBFileShare_opLocksEnabled
    resource_aws_storagegateway_smb_file_share_test.go:350: Step 2/3 error: Error running apply: exit status 1

        Error: error updating Storage Gateway SMB File Share (arn:aws:storagegateway:us-west-2:123456789012:share/share-86C5A6E3): InvalidGatewayRequestException: The specified gateway is out of date.
        {
          RespMetadata: {
            StatusCode: 400,
            RequestID: "56a23d7f-b8c3-420a-ba06-a4fde82b4092"
          },
          Error_: {
            ErrorCode: "OutdatedGateway"
          },
          Message_: "The specified gateway is out of date."
        }

func TestAccStorageGatewaySMBFileShare_opLocksEnabled(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories:acctest.ProviderFactories,
		CheckDestroy: testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_OpLocksEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "oplocks_enabled", "false"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_OpLocksEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "oplocks_enabled", "true"),
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
*/

func TestAccStorageGatewaySMBFileShare_invalidUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_InvalidUserList_Single(rName, domainName, "invaliduser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "1"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_InvalidUserList_Multiple(rName, domainName, "invaliduser2", "invaliduser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "2"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_InvalidUserList_Single(rName, domainName, "invaliduser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "1"),
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

func TestAccStorageGatewaySMBFileShare_kmsEncrypted(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSMBFileShareConfig_KMSEncrypted(rName, true),
				ExpectError: regexp.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccSMBFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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

func TestAccStorageGatewaySMBFileShare_kmsKeyARN(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	keyName := "aws_kms_key.test.0"
	keyUpdatedName := "aws_kms_key.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_KMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyName, "arn"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_KMSKeyARN_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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
				Config: testAccSMBFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_objectACL(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicRead),
				),
			},
			{
				Config: testAccSMBFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicReadWrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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

func TestAccStorageGatewaySMBFileShare_readOnly(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_ReadOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_ReadOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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

func TestAccStorageGatewaySMBFileShare_requesterPays(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_RequesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_RequesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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

func TestAccStorageGatewaySMBFileShare_validUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_ValidUserList_Single(rName, domainName, "validuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "1"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_ValidUserList_Multiple(rName, domainName, "validuser2", "validuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "2"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_ValidUserList_Single(rName, domainName, "validuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "1"),
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

func TestAccStorageGatewaySMBFileShare_SMB_acl(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareSMBACLConfig(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareSMBACLConfig(rName, domainName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", "false"),
				),
			},
			{
				Config: testAccSMBFileShareSMBACLConfig(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", "true"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_audit(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	logResourceName := "aws_cloudwatch_log_group.test"
	logResourceNameSecond := "aws_cloudwatch_log_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareAuditDestinationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareAuditDestinationUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceNameSecond, "arn"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_cacheAttributes(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareCacheAttributesConfig(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
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
				Config: testAccSMBFileShareCacheAttributesConfig(rName, 500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "500"),
				),
			},
			{
				Config: testAccSMBFileShareCacheAttributesConfig(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_caseSensitivity(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareCaseSensitivityConfig(rName, "CaseSensitive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "CaseSensitive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareCaseSensitivityConfig(rName, "ClientSpecified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "ClientSpecified"),
				),
			},
			{
				Config: testAccSMBFileShareCaseSensitivityConfig(rName, "CaseSensitive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "CaseSensitive"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_disappears(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_Authentication_GuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					acctest.CheckResourceDisappears(acctest.Provider, tfstoragegateway.ResourceSMBFileShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_adminUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_AdminUserList_Single(rName, domainName, "adminuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "1"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_AdminUserList_Multiple(rName, domainName, "adminuser2", "adminuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "2"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_AdminUserList_Single(rName, domainName, "adminuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "1"),
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

func testAccCheckSMBFileShareDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_smb_file_share" {
			continue
		}

		_, err := tfstoragegateway.FindSMBFileShareByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Storage Gateway SMB File Share %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckSMBFileShareExists(resourceName string, smbFileShare *storagegateway.SMBFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

		output, err := tfstoragegateway.FindSMBFileShareByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*smbFileShare = *output

		return nil
	}
}

func testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_SMBActiveDirectorySettings(rName, domainName), fmt.Sprintf(`
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
`, rName))
}

func testAcc_SMBFileShare_GuestAccessBase(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_SMBGuestPassword(rName, "smbguestpassword"), fmt.Sprintf(`
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
`, rName))
}

func testAccSMBFileShareConfig_Authentication_ActiveDirectory(rName, domainName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "ActiveDirectory"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`)
}

func testAccSMBFileShareConfig_Authentication_GuestAccess(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`)
}

func testAccSMBFileShareAccessBasedEnumerationConfig(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  authentication           = "GuestAccess"
  gateway_arn              = aws_storagegateway_gateway.test.arn
  location_arn             = aws_s3_bucket.test.arn
  role_arn                 = aws_iam_role.test.arn
  access_based_enumeration = %[1]t
}
`, enabled))
}

func testAccSMBFileShareNotificationPolicyConfig(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication      = "GuestAccess"
  gateway_arn         = aws_storagegateway_gateway.test.arn
  location_arn        = aws_s3_bucket.test.arn
  role_arn            = aws_iam_role.test.arn
  notification_policy = "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"
}
`)
}

func testAccSMBFileShareConfig_DefaultStorageClass(rName, defaultStorageClass string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  default_storage_class = %q
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
}
`, defaultStorageClass))
}

func testAccSMBFileShareConfig_FileShareName(rName, fileShareName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication  = "GuestAccess"
  file_share_name = %q
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
}
`, fileShareName))
}

func testAccSMBFileShareConfig_GuessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication          = "GuestAccess"
  gateway_arn             = aws_storagegateway_gateway.test.arn
  guess_mime_type_enabled = %t
  location_arn            = aws_s3_bucket.test.arn
  role_arn                = aws_iam_role.test.arn
}
`, guessMimeTypeEnabled))
}

/*
func testAccSMBFileShareConfig_OpLocksEnabled(rName string, opLocksEnabled bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication  = "GuestAccess"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  oplocks_enabled = %t
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
}
`, opLocksEnabled))
}
*/

func testAccSMBFileShareConfig_InvalidUserList_Single(rName, domainName, invalidUser1 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn       = aws_storagegateway_gateway.test.arn
  invalid_user_list = [%q]
  location_arn      = aws_s3_bucket.test.arn
  role_arn          = aws_iam_role.test.arn
}
`, invalidUser1))
}

func testAccSMBFileShareConfig_InvalidUserList_Multiple(rName, domainName, invalidUser1, invalidUser2 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn       = aws_storagegateway_gateway.test.arn
  invalid_user_list = [%q, %q]
  location_arn      = aws_s3_bucket.test.arn
  role_arn          = aws_iam_role.test.arn
}
`, invalidUser1, invalidUser2))
}

func testAccSMBFileShareConfig_KMSEncrypted(rName string, kmsEncrypted bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = %t
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`, kmsEncrypted))
}

func testAccSMBFileShareConfig_KMSKeyARN(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = true
  kms_key_arn    = aws_kms_key.test[0].arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`)
}

func testAccSMBFileShareConfig_KMSKeyARN_Update(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = true
  kms_key_arn    = aws_kms_key.test[1].arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`)
}

func testAccSMBFileShareConfig_ObjectACL(rName, objectACL string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  object_acl     = %q
  role_arn       = aws_iam_role.test.arn
}
`, objectACL))
}

func testAccSMBFileShareConfig_ReadOnly(rName string, readOnly bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  read_only      = %t
  role_arn       = aws_iam_role.test.arn
}
`, readOnly))
}

func testAccSMBFileShareConfig_RequesterPays(rName string, requesterPays bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  requester_pays = %t
  role_arn       = aws_iam_role.test.arn
}
`, requesterPays))
}

func testAccSMBFileShareConfig_ValidUserList_Single(rName, domainName, validUser1 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  valid_user_list = [%q]
}
`, validUser1))
}

func testAccSMBFileShareConfig_ValidUserList_Multiple(rName, domainName, validUser1, validUser2 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  valid_user_list = [%q, %q]
}
`, validUser1, validUser2))
}

func testAccSMBFileShareConfig_AdminUserList_Single(rName, domainName, adminUser1 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  admin_user_list = [%q]
}
`, adminUser1))
}

func testAccSMBFileShareConfig_AdminUserList_Multiple(rName, domainName, adminUser1, adminUser2 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  admin_user_list = [%q, %q]
}
`, adminUser1, adminUser2))
}

func testAccSMBFileShareTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1))
}

func testAccSMBFileShareTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn

  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccSMBFileShareSMBACLConfig(rName, domainName string, enabled bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  smb_acl_enabled = %[1]t
}
`, enabled))
}

func testAccSMBFileShareAuditDestinationConfig(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
  audit_destination_arn = aws_cloudwatch_log_group.test.arn
}
`, rName))
}

func testAccSMBFileShareAuditDestinationUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-updated"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
  audit_destination_arn = aws_cloudwatch_log_group.test2.arn
}
`, rName))
}

func testAccSMBFileShareCacheAttributesConfig(rName string, timeout int) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn

  cache_attributes {
    cache_stale_timeout_in_seconds = %[1]d
  }
}
`, timeout))
}

func testAccSMBFileShareCaseSensitivityConfig(rName, option string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication   = "GuestAccess"
  gateway_arn      = aws_storagegateway_gateway.test.arn
  location_arn     = aws_s3_bucket.test.arn
  role_arn         = aws_iam_role.test.arn
  case_sensitivity = %[1]q
}
`, option))
}
