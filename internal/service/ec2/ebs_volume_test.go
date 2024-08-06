// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EBSVolume_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEBSVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSVolume_updateAttachedEBSVolume(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_attached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_attachedUpdateSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "20"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_updateSize(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_tags1("Name", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_updateSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_updateType(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_tags1("Name", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_updateType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "sc1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_UpdateIops_io1(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_iopsIo1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_iopsIo1Updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_UpdateIops_io2(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_iopsIo2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_iopsIo2Updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_noIops(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_noIOPS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
func TestAccEC2EBSVolume_invalidIopsForType(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccEBSVolumeConfig_invalidIOPSForType,
				ExpectError: regexache.MustCompile(`'iops' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccEC2EBSVolume_invalidThroughputForType(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccEBSVolumeConfig_invalidThroughputForType,
				ExpectError: regexache.MustCompile(`'throughput' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccEC2EBSVolume_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEBSVolumeConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_multiAttach_io1(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_multiAttach(rName, "io1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "io1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_multiAttach_io2(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_multiAttach(rName, "io2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "io2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_multiAttach_gp2(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccEBSVolumeConfig_invalidMultiAttachEnabledForType,
				ExpectError: regexache.MustCompile(`'multi_attach_enabled' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccEC2EBSVolume_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_GP3_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp3", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_GP3_iops(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp3", "4000", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "4000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp3", "5000", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "5000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_GP3_throughput(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp3", "", "400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "400"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp3", "", "600"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "600"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_gp3ToGP2(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp3", "3000", "400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "3000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "400"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, acctest.Ct10, "gp2", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp2"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_io1ToGP3(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, "100", "io1", "4000", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "4000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "io1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config: testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, "100", "gp3", "4000", "125"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "4000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, "125"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp3"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_snapshotID(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_snapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSnapshotID, snapshotResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_snapshotIDAndSize(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_snapshotIdAndSize(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrIOPS, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "20"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSnapshotID, snapshotResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrThroughput, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "gp2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
		},
	})
}

func TestAccEC2EBSVolume_finalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVolumeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig_finalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot"},
			},
			{
				Config:  testAccEBSVolumeConfig_finalSnapshot(rName),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeFinalSnapshotExists(ctx, &v),
				),
			},
		},
	})
}

func testAccCheckVolumeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ebs_volume" {
				continue
			}

			_, err := tfec2.FindEBSVolumeByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EBS Volume %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckVolumeExists(ctx context.Context, n string, v *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EBS Volume ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindEBSVolumeByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVolumeFinalSnapshotExists(ctx context.Context, v *awstypes.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		input := &ec2.DescribeSnapshotsInput{
			Filters: tfec2.NewAttributeFilterList(map[string]string{
				"volume-id":      aws.ToString(v.VolumeId),
				names.AttrStatus: string(awstypes.SnapshotStateCompleted),
			}),
		}

		output, err := tfec2.FindSnapshot(ctx, conn, input)

		if err != nil {
			return err
		}

		r := tfec2.ResourceEBSSnapshot()
		d := r.Data(nil)
		d.SetId(aws.ToString(output.SnapshotId))

		err = acctest.DeleteResource(ctx, r, d, acctest.Provider.Meta())

		return err
	}
}

var testAccEBSVolumeConfig_basic = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}
`)

func testAccEBSVolumeConfig_attached(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  depends_on        = [aws_instance.test]
  availability_zone = aws_instance.test.availability_zone
  type              = "gp2"
  size              = "10"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  depends_on  = [aws_ebs_volume.test]
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccEBSVolumeConfig_attachedUpdateSize(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  depends_on        = [aws_instance.test]
  availability_zone = aws_instance.test.availability_zone
  type              = "gp2"
  size              = "20"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  depends_on  = [aws_ebs_volume.test]
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccEBSVolumeConfig_updateSize(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 10

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_updateType(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "sc1"
  size              = 500

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_iopsIo1(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 4
  iops              = 100

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_iopsIo1Updated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 4
  iops              = 200

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_iopsIo2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io2"
  size              = 4
  iops              = 100

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_iopsIo2Updated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io2"
  size              = 4
  iops              = 200

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
  kms_key_id        = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccEBSVolumeConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEBSVolumeConfig_noIOPS(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  type              = "gp2"
  iops              = 0

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

var testAccEBSVolumeConfig_invalidIOPSForType = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  iops              = 100
}
`)

var testAccEBSVolumeConfig_invalidThroughputForType = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  iops              = 100
  throughput        = 500
  type              = "io1"
}
`)

var testAccEBSVolumeConfig_invalidMultiAttachEnabledForType = acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_ebs_volume" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  size                 = 10
  multi_attach_enabled = true
  type                 = "gp2"
}
`)

func testAccEBSVolumeConfig_outpost(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  size              = 1
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEBSVolumeConfig_multiAttach(rName, volumeType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  type                 = %[2]q
  multi_attach_enabled = true
  size                 = 4
  iops                 = 100

  tags = {
    Name = %[1]q
  }
}
`, rName, volumeType))
}

func testAccEBSVolumeConfig_sizeTypeIOPSThroughput(rName, size, volumeType, iops, throughput string) string {
	if volumeType == "" {
		volumeType = "null"
	}
	if iops == "" {
		iops = "null"
	}
	if throughput == "" {
		throughput = "null"
	}

	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = %[2]s
  type              = %[3]q
  iops              = %[4]s
  throughput        = %[5]s

  tags = {
    Name = %[1]q
  }
}
`, rName, size, volumeType, iops, throughput))
}

func testAccEBSVolumeConfig_snapshotID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "source" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  snapshot_id       = aws_ebs_snapshot.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSVolumeConfig_snapshotIdAndSize(rName string, size int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "source" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  snapshot_id       = aws_ebs_snapshot.test.id
  size              = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, size))
}

func testAccEBSVolumeConfig_finalSnapshot(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  final_snapshot    = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
