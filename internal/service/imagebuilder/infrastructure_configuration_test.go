// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderInfrastructureConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamInstanceProfileResourceName := "aws_iam_instance_profile.test"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "imagebuilder", fmt.Sprintf("infrastructure-configuration/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_updated", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSNSTopicARN, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSubnetID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", acctest.CtFalse),
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

func TestAccImageBuilderInfrastructureConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceInfrastructureConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_instanceMetadataOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_instanceMetadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.0.http_put_response_hop_limit", "64"),
					resource.TestCheckResourceAttr(resourceName, "instance_metadata_options.0.http_tokens", "required"),
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

func TestAccImageBuilderInfrastructureConfiguration_instanceProfileName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamInstanceProfileResourceName := "aws_iam_instance_profile.test"
	iamInstanceProfileResourceName2 := "aws_iam_instance_profile.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_instanceProfileName1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_instanceProfileName2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_profile_name", iamInstanceProfileResourceName2, names.AttrName),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_instanceTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_instanceTypes1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_instanceTypes2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_keyPair(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	keyPairResourceName := "aws_key_pair.test"
	keyPairResourceName2 := "aws_key_pair.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	publicKey1, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	publicKey2, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_keyPair1(rName, publicKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "key_pair", keyPairResourceName, "key_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_keyPair2(rName, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, "key_pair", keyPairResourceName2, "key_name"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_LoggingS3Logs_s3BucketName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketResourceName2 := "aws_s3_bucket.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_loggingS3LogsS3BucketName1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.s3_logs.0.s3_bucket_name", s3BucketResourceName, names.AttrBucket),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_loggingS3LogsS3BucketName2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "logging.0.s3_logs.0.s3_bucket_name", s3BucketResourceName2, names.AttrBucket),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_LoggingS3Logs_s3KeyPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_loggingS3LogsS3KeyPrefix(rName, "/prefix1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "/prefix1/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_loggingS3LogsS3KeyPrefix(rName, "/prefix2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "logging.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "/prefix2/"),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_resourceTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_resourceTags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_resourceTags(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.key2", acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_securityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	securityGroupResourceName := "aws_security_group.test"
	securityGroupResourceName2 := "aws_security_group.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_securityGroupIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_securityGroupIDs2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", securityGroupResourceName2, names.AttrID),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_snsTopicARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snsTopicResourceName := "aws_sns_topic.test"
	snsTopicResourceName2 := "aws_sns_topic.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_snsTopicARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, snsTopicResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_snsTopicARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, snsTopicResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_subnetID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	subnetResourceName := "aws_subnet.test"
	subnetResourceName2 := "aws_subnet.test2"
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_subnetID1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_subnetID2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSubnetID, subnetResourceName2, names.AttrID),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
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
				Config: testAccInfrastructureConfigurationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccInfrastructureConfigurationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderInfrastructureConfiguration_terminateInstanceOnFailure(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInfrastructureConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInfrastructureConfigurationConfig_terminateInstanceOnFailure(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInfrastructureConfigurationConfig_terminateInstanceOnFailure(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInfrastructureConfigurationExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckInfrastructureConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_infrastructure_configuration" {
				continue
			}

			input := &imagebuilder.GetInfrastructureConfigurationInput{
				InfrastructureConfigurationArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetInfrastructureConfigurationWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Image Builder Infrastructure Configuration (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckInfrastructureConfigurationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		input := &imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetInfrastructureConfigurationWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccInfrastructureConfigurationBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.role.name
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}
`, rName)
}

func testAccInfrastructureConfigurationConfig_description(rName string, description string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  description           = %[2]q
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName, description))
}

func testAccInfrastructureConfigurationConfig_instanceMetadataOptions(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  instance_metadata_options {
    http_put_response_hop_limit = 64
    http_tokens                 = "required"
  }
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_instanceProfileName1(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_instanceProfileName2(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_iam_instance_profile" "test2" {
  name = aws_iam_role.role2.name
  role = aws_iam_role.role2.name
}

resource "aws_iam_role" "role2" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test2.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_instanceTypes1(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		acctest.AvailableEC2InstanceTypeForRegion("t3.medium", "t2.medium"),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_types        = [data.aws_ec2_instance_type_offering.available.instance_type]
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_instanceTypes2(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		acctest.AvailableEC2InstanceTypeForRegion("t3.large", "t2.large"),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_types        = [data.aws_ec2_instance_type_offering.available.instance_type]
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_keyPair1(rName, publicKey string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  key_pair              = aws_key_pair.test.key_name
  name                  = %[1]q
}
`, rName, publicKey))
}

func testAccInfrastructureConfigurationConfig_keyPair2(rName, publicKey string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_key_pair" "test2" {
  key_name   = "%[1]s-2"
  public_key = %[2]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  key_pair              = aws_key_pair.test2.key_name
  name                  = %[1]q
}
`, rName, publicKey))
}

func testAccInfrastructureConfigurationConfig_loggingS3LogsS3BucketName1(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test.bucket
    }
  }
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_loggingS3LogsS3BucketName2(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test2" {
  bucket = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test2.bucket
    }
  }
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_loggingS3LogsS3KeyPrefix(rName string, s3KeyPrefix string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  logging {
    s3_logs {
      s3_bucket_name = aws_s3_bucket.test.bucket
      s3_key_prefix  = %[2]q
    }
  }
}
`, rName, s3KeyPrefix))
}

func testAccInfrastructureConfigurationConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_resourceTags(rName string, resourceTagKey string, resourceTagValue string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  resource_tags = {
    %[2]q = %[3]q
  }
}
`, rName, resourceTagKey, resourceTagValue))
}

func testAccInfrastructureConfigurationConfig_securityGroupIDs1(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id]
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_securityGroupIDs2(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test2.id]
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_subnetID1(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id     = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id] # Required with subnet_id
  subnet_id             = aws_subnet.test.id
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_subnetID2(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test2" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 2, 2)
  vpc_id     = aws_vpc.test.id
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_security_group.test.id] # Required with subnet_id
  subnet_id             = aws_subnet.test2.id
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_snsTopicARN1(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  sns_topic_arn         = aws_sns_topic.test.arn
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_snsTopicARN2(rName string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test2" {
  name = "%[1]s-2"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  sns_topic_arn         = aws_sns_topic.test2.arn
}
`, rName))
}

func testAccInfrastructureConfigurationConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccInfrastructureConfigurationConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccInfrastructureConfigurationConfig_terminateInstanceOnFailure(rName string, terminateInstanceOnFailure bool) string {
	return acctest.ConfigCompose(
		testAccInfrastructureConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name         = aws_iam_instance_profile.test.name
  name                          = %[1]q
  terminate_instance_on_failure = %[2]t
}
`, rName, terminateInstanceOnFailure))
}
