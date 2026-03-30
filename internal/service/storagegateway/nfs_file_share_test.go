// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccStorageGatewayNFSFileShare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_required(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bucket_region", ""),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", rName),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexache.MustCompile(`^share-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", string(awstypes.ObjectACLPrivate)),
					resource.TestMatchResourceAttr(resourceName, names.AttrPath, regexache.MustCompile(`^/.+`)),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "squash", "RootSquash"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"
	logResourceName := "aws_cloudwatch_log_group.test"
	logResourceNameSecond := "aws_cloudwatch_log_group.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_audit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareConfig_auditUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceNameSecond, names.AttrARN),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNFSFileShareConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_fileShareName(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_name(rName, "test_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "test_1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareConfig_name(rName, "test_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "file_share_name", "test_2"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_clientList(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_clientListSingle(rName, "1.1.1.1/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "1.1.1.1/32"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_clientListMultiple(rName, "2.2.2.2/32", "3.3.3.3/32"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "client_list.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "2.2.2.2/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "client_list.*", "3.3.3.3/32"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_clientListSingle(rName, "4.4.4.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_defaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_defaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_guessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccNFSFileShareConfig_guessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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

func TestAccStorageGatewayNFSFileShare_kmsEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccNFSFileShareConfig_kmsEncrypted(rName, true),
				ExpectError: regexache.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccNFSFileShareConfig_kmsEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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

func TestAccStorageGatewayNFSFileShare_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"
	keyName := "aws_kms_key.test.0"
	keyUpdatedName := "aws_kms_key.test.1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_kmsKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, keyName, names.AttrARN),
				),
			},
			{
				Config: testAccNFSFileShareConfig_kmsKeyARNUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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
				Config: testAccNFSFileShareConfig_kmsEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_nFSFileShareDefaults(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_defaults(rName, "0700", "0600", 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.directory_mode", "0700"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.file_mode", "0600"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.group_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "nfs_file_share_defaults.0.owner_id", "2"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_defaults(rName, "0770", "0660", 3, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_objectACL(rName, string(awstypes.ObjectACLPublicRead)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", string(awstypes.ObjectACLPublicRead)),
				),
			},
			{
				Config: testAccNFSFileShareConfig_objectACL(rName, string(awstypes.ObjectACLPublicReadWrite)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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

func TestAccStorageGatewayNFSFileShare_readOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_readOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", acctest.CtFalse),
				),
			},
			{
				Config: testAccNFSFileShareConfig_readOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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

func TestAccStorageGatewayNFSFileShare_requesterPays(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_requesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", acctest.CtFalse),
				),
			},
			{
				Config: testAccNFSFileShareConfig_requesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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

func TestAccStorageGatewayNFSFileShare_squash(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_squash(rName, "NoSquash"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "squash", "NoSquash"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_squash(rName, "AllSquash"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_notificationPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNFSFileShareConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{}"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_notificationPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "notification_policy", "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_cacheAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_cacheAttributes(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
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
				Config: testAccNFSFileShareConfig_cacheAttributes(rName, 500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "500"),
				),
			},
			{
				Config: testAccNFSFileShareConfig_cacheAttributes(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
		},
	})
}

func TestAccStorageGatewayNFSFileShare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var nfsFileShare awstypes.NFSFileShareInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_nfs_file_share.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNFSFileShareDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNFSFileShareConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNFSFileShareExists(ctx, t, resourceName, &nfsFileShare),
					acctest.CheckSDKResourceDisappears(ctx, t, tfstoragegateway.ResourceNFSFileShare(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfstoragegateway.ResourceNFSFileShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNFSFileShareDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storagegateway_nfs_file_share" {
				continue
			}

			_, err := tfstoragegateway.FindNFSFileShareByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Storage Gateway NFS File Share %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckNFSFileShareExists(ctx context.Context, t *testing.T, n string, v *awstypes.NFSFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		output, err := tfstoragegateway.FindNFSFileShareByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccNFSFileShareConfig_baseS3(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_baseFileS3(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccNFSFileShareConfig_required(rName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), `
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`)
}

func testAccNFSFileShareConfig_name(rName, fsName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list     = ["0.0.0.0/0"]
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  file_share_name = %[1]q
}
`, fsName))
}

func testAccNFSFileShareConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1))
}

func testAccNFSFileShareConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
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
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccNFSFileShareConfig_clientListSingle(rName, clientList1 string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = [%q]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`, clientList1))
}

func testAccNFSFileShareConfig_clientListMultiple(rName, clientList1, clientList2 string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = [%q, %q]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
}
`, clientList1, clientList2))
}

func testAccNFSFileShareConfig_defaultStorageClass(rName, defaultStorageClass string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list           = ["0.0.0.0/0"]
  default_storage_class = %q
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
}
`, defaultStorageClass))
}

func testAccNFSFileShareConfig_guessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list             = ["0.0.0.0/0"]
  gateway_arn             = aws_storagegateway_gateway.test.arn
  guess_mime_type_enabled = %t
  location_arn            = aws_s3_bucket.test.arn
  role_arn                = aws_iam_role.test.arn
}
`, guessMimeTypeEnabled))
}

func testAccNFSFileShareConfig_kmsEncrypted(rName string, kmsEncrypted bool) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list   = ["0.0.0.0/0"]
  gateway_arn   = aws_storagegateway_gateway.test.arn
  kms_encrypted = %t
  location_arn  = aws_s3_bucket.test.arn
  role_arn      = aws_iam_role.test.arn
}
`, kmsEncrypted))
}

func testAccNFSFileShareConfig_kmsKeyARN(rName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  enable_key_rotation     = true
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
`)
}

func testAccNFSFileShareConfig_kmsKeyARNUpdate(rName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  enable_key_rotation     = true
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
`)
}

func testAccNFSFileShareConfig_defaults(rName, directoryMode, fileMode string, groupID, ownerID int) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
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
`, directoryMode, fileMode, groupID, ownerID))
}

func testAccNFSFileShareConfig_objectACL(rName, objectACL string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  object_acl   = %q
  role_arn     = aws_iam_role.test.arn
}
`, objectACL))
}

func testAccNFSFileShareConfig_readOnly(rName string, readOnly bool) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  read_only    = %t
  role_arn     = aws_iam_role.test.arn
}
`, readOnly))
}

func testAccNFSFileShareConfig_requesterPays(rName string, requesterPays bool) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list    = ["0.0.0.0/0"]
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  requester_pays = %t
  role_arn       = aws_iam_role.test.arn
}
`, requesterPays))
}

func testAccNFSFileShareConfig_squash(rName, squash string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn
  squash       = %q
}
`, squash))
}

func testAccNFSFileShareConfig_cacheAttributes(rName string, timeout int) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list  = ["0.0.0.0/0"]
  gateway_arn  = aws_storagegateway_gateway.test.arn
  location_arn = aws_s3_bucket.test.arn
  role_arn     = aws_iam_role.test.arn

  cache_attributes {
    cache_stale_timeout_in_seconds = %[1]d
  }
}
`, timeout))
}

func testAccNFSFileShareConfig_notificationPolicy(rName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), `
resource "aws_storagegateway_nfs_file_share" "test" {
  client_list         = ["0.0.0.0/0"]
  gateway_arn         = aws_storagegateway_gateway.test.arn
  location_arn        = aws_s3_bucket.test.arn
  role_arn            = aws_iam_role.test.arn
  notification_policy = "{\"Upload\": {\"SettlingTimeInSeconds\": 60}}"
}
`)
}

func testAccNFSFileShareConfig_audit(rName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccNFSFileShareConfig_auditUpdated(rName string) string {
	return acctest.ConfigCompose(testAccNFSFileShareConfig_baseS3(rName), fmt.Sprintf(`
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
`, rName))
}
