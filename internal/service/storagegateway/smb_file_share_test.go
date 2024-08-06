// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccStorageGatewaySMBFileShare_Authentication_activeDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_authenticationActiveDirectory(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "ActiveDirectory"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexache.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "file_share_name", regexache.MustCompile(`^tf-acc-test-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "object_acl", string(awstypes.ObjectACLPrivate)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPath, regexache.MustCompile(`^/.+`)),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_authenticationGuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "GuestAccess"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "ClientSpecified"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexache.MustCompile(`^share-`)),
					resource.TestMatchResourceAttr(resourceName, "file_share_name", regexache.MustCompile(`^tf-acc-test-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "object_acl", string(awstypes.ObjectACLPrivate)),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_accessBasedEnumeration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_accessBasedEnumeration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", acctest.CtFalse),
				),
			},
			{
				Config: testAccSMBFileShareConfig_accessBasedEnumeration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "access_based_enumeration", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_notificationPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_notificationPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_authenticationGuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_notificationPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_defaultStorageClass(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_defaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_defaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
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

func TestAccStorageGatewaySMBFileShare_encryptedUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_encryptedUpdate(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtTrue),
				),
			},
			{
				Config: testAccSMBFileShareConfig_encryptedUpdate(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_fileShareName(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_name(rName, "foo_share"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "foo_share"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_name(rName, "bar_share"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSMBFileShareConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_guessMIMETypeEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_guessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccSMBFileShareConfig_guessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", acctest.CtTrue),
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
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories:acctest.ProtoV5ProviderFactories,
		CheckDestroy: testAccCheckSMBFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_opLocksEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "oplocks_enabled", "false"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_opLocksEnabled(rName, true),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_invalidUserListSingle(rName, domainName, "invaliduser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", acctest.Ct1),
				),
			},
			{
				Config: testAccSMBFileShareConfig_invalidUserListMultiple(rName, domainName, "invaliduser2", "invaliduser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", acctest.Ct2),
				),
			},
			{
				Config: testAccSMBFileShareConfig_invalidUserListSingle(rName, domainName, "invaliduser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccSMBFileShareConfig_kmsEncrypted(rName, true),
				ExpectError: regexache.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccSMBFileShareConfig_kmsEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	keyName := "aws_kms_key.test.0"
	keyUpdatedName := "aws_kms_key.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_kmsKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, keyName, names.AttrARN),
				),
			},
			{
				Config: testAccSMBFileShareConfig_kmsKeyARNUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, keyUpdatedName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_kmsEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_objectACL(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_objectACL(rName, string(awstypes.ObjectACLPublicRead)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", string(awstypes.ObjectACLPublicRead)),
				),
			},
			{
				Config: testAccSMBFileShareConfig_objectACL(rName, string(awstypes.ObjectACLPublicReadWrite)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", string(awstypes.ObjectACLPublicReadWrite)),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_readOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
				),
			},
			{
				Config: testAccSMBFileShareConfig_readOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_requesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", acctest.CtFalse),
				),
			},
			{
				Config: testAccSMBFileShareConfig_requesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_validUserListSingle(rName, domainName, "validuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", acctest.Ct1),
				),
			},
			{
				Config: testAccSMBFileShareConfig_validUserListMultiple(rName, domainName, "validuser2", "validuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", acctest.Ct2),
				),
			},
			{
				Config: testAccSMBFileShareConfig_validUserListSingle(rName, domainName, "validuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_acl(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_acl(rName, domainName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccSMBFileShareConfig_acl(rName, domainName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_audit(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	logResourceName := "aws_cloudwatch_log_group.test"
	logResourceNameSecond := "aws_cloudwatch_log_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_auditDestination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_auditDestinationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceNameSecond, names.AttrARN),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_cacheAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_cacheAttributes(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_cacheAttributes(rName, 500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "500"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_cacheAttributes(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_caseSensitivity(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_caseSensitivity(rName, "CaseSensitive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "CaseSensitive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSMBFileShareConfig_caseSensitivity(rName, "ClientSpecified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "ClientSpecified"),
				),
			},
			{
				Config: testAccSMBFileShareConfig_caseSensitivity(rName, "CaseSensitive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "CaseSensitive"),
				),
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_authenticationGuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfstoragegateway.ResourceSMBFileShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStorageGatewaySMBFileShare_adminUserList(t *testing.T) {
	ctx := acctest.Context(t)
	var smbFileShare awstypes.SMBFileShareInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_smb_file_share.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMBFileShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMBFileShareConfig_adminUserListSingle(rName, domainName, "adminuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", acctest.Ct1),
				),
			},
			{
				Config: testAccSMBFileShareConfig_adminUserListMultiple(rName, domainName, "adminuser2", "adminuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", acctest.Ct2),
				),
			},
			{
				Config: testAccSMBFileShareConfig_adminUserListSingle(rName, domainName, "adminuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSMBFileShareExists(ctx, resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", acctest.Ct1),
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

func testAccCheckSMBFileShareDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storagegateway_smb_file_share" {
				continue
			}

			_, err := tfstoragegateway.FindSMBFileShareByARN(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckSMBFileShareExists(ctx context.Context, n string, v *awstypes.SMBFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayClient(ctx)

		output, err := tfstoragegateway.FindSMBFileShareByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_smbActiveDirectorySettings(rName, domainName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccGatewayConfig_smbGuestPassword(rName, "smbguestpassword"), fmt.Sprintf(`
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

func testAccSMBFileShareConfig_authenticationActiveDirectory(rName, domainName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_ActiveDirectoryBase(rName, domainName), `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "ActiveDirectory"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`)
}

func testAccSMBFileShareConfig_authenticationGuestAccess(rName string) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`)
}

func testAccSMBFileShareConfig_accessBasedEnumeration(rName string, enabled bool) string {
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

func testAccSMBFileShareConfig_notificationPolicy(rName string) string {
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

func testAccSMBFileShareConfig_defaultStorageClass(rName, defaultStorageClass string) string {
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

func testAccSMBFileShareConfig_encryptedUpdate(rName string, readOnly bool) string {
	return acctest.ConfigCompose(testAcc_SMBFileShare_GuestAccessBase(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = true
  kms_key_arn    = aws_kms_key.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
  read_only      = %[1]t
}
`, readOnly))
}

func testAccSMBFileShareConfig_name(rName, fileShareName string) string {
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

func testAccSMBFileShareConfig_guessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
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
func testAccSMBFileShareConfig_opLocksEnabled(rName string, opLocksEnabled bool) string {
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

func testAccSMBFileShareConfig_invalidUserListSingle(rName, domainName, invalidUser1 string) string {
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

func testAccSMBFileShareConfig_invalidUserListMultiple(rName, domainName, invalidUser1, invalidUser2 string) string {
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

func testAccSMBFileShareConfig_kmsEncrypted(rName string, kmsEncrypted bool) string {
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

func testAccSMBFileShareConfig_kmsKeyARN(rName string) string {
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

func testAccSMBFileShareConfig_kmsKeyARNUpdate(rName string) string {
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

func testAccSMBFileShareConfig_objectACL(rName, objectACL string) string {
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

func testAccSMBFileShareConfig_readOnly(rName string, readOnly bool) string {
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

func testAccSMBFileShareConfig_requesterPays(rName string, requesterPays bool) string {
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

func testAccSMBFileShareConfig_validUserListSingle(rName, domainName, validUser1 string) string {
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

func testAccSMBFileShareConfig_validUserListMultiple(rName, domainName, validUser1, validUser2 string) string {
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

func testAccSMBFileShareConfig_adminUserListSingle(rName, domainName, adminUser1 string) string {
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

func testAccSMBFileShareConfig_adminUserListMultiple(rName, domainName, adminUser1, adminUser2 string) string {
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

func testAccSMBFileShareConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

func testAccSMBFileShareConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccSMBFileShareConfig_acl(rName, domainName string, enabled bool) string {
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

func testAccSMBFileShareConfig_auditDestination(rName string) string {
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

func testAccSMBFileShareConfig_auditDestinationUpdated(rName string) string {
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

func testAccSMBFileShareConfig_cacheAttributes(rName string, timeout int) string {
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

func testAccSMBFileShareConfig_caseSensitivity(rName, option string) string {
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
