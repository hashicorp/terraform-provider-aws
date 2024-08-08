// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// func TestFetchRootDevice(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	t.Parallel()

// 	cases := []struct {
// 		label  string
// 		images []*awstypes.Image
// 		name   string
// 	}{
// 		{
// 			"device name in mappings",
// 			[]*awstypes.Image{{
// 				ImageId:        aws.String("ami-123"),
// 				RootDeviceType: awstypes.DeviceType("ebs"),
// 				RootDeviceName: aws.String("/dev/xvda"),
// 				BlockDeviceMappings: []awstypes.BlockDeviceMapping{
// 					{DeviceName: aws.String("/dev/xvdb")},
// 					{DeviceName: aws.String("/dev/xvda")},
// 				},
// 			}},
// 			"/dev/xvda",
// 		},
// 		{
// 			"device name not in mappings",
// 			[]*awstypes.Image{{
// 				ImageId:        aws.String("ami-123"),
// 				RootDeviceType: awstypes.DeviceType("ebs"),
// 				RootDeviceName: aws.String("/dev/xvda"),
// 				BlockDeviceMappings: []awstypes.BlockDeviceMapping{
// 					{DeviceName: aws.String("/dev/xvdb")},
// 					{DeviceName: aws.String("/dev/xvdc")},
// 				},
// 			}},
// 			"/dev/xvdb",
// 		},
// 		{
// 			"no images",
// 			[]*awstypes.Image{},
// 			"",
// 		},
// 	}

// 	sess, err := session.NewSession(nil)
// 	if err != nil {
// 		t.Errorf("Error new session: %s", err)
// 	}

// 	for _, tc := range cases {
// 		tc := tc
// 		t.Run(fmt.Sprintf(tc.label), func(t *testing.T) {

// 			t.Parallel()

// 			conn := ec2.New(sess)
// 			conn.Handlers.Clear()
// 			conn.Handlers.Send.PushBack(func(r *request.Request) {
// 				data := r.Data.(*ec2.DescribeImagesOutput)
// 				data.Images = tc.images
// 			})
// 			name, _ := tfec2.FetchRootDeviceName(ctx, conn, "ami-123")
// 			if tc.name != aws.ToString(name) {
// 				t.Errorf("Expected name %s, got %s", tc.name, aws.ToString(name))
// 			}
// 		})
// 	}
// }

func TestParseInstanceType(t *testing.T) {
	t.Parallel()

	invalidInstanceTypes := []string{
		"",
		"abc",
		"abc4",
		"abc4.",
		"abc.xlarge",
		"4g.3xlarge",
	}

	for _, v := range invalidInstanceTypes {
		if _, err := tfec2.ParseInstanceType(v); err == nil {
			t.Errorf("Expected error for %s", v)
		}
	}

	v, err := tfec2.ParseInstanceType("c4.large")

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if got, want := v.Type, "c4"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	if got, want := v.Family, "c"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	if got, want := v.Generation, 4; got != want {
		t.Errorf("Got: %d, want: %d", got, want)
	}

	if got, want := v.AdditionalCapabilities, ""; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	if got, want := v.Size, "large"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	v, err = tfec2.ParseInstanceType("im4gn.16xlarge")

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if got, want := v.Type, "im4gn"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	if got, want := v.Family, "im"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	if got, want := v.Generation, 4; got != want {
		t.Errorf("Got: %d, want: %d", got, want)
	}

	if got, want := v.AdditionalCapabilities, "gn"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}

	if got, want := v.Size, "16xlarge"; got != want {
		t.Errorf("Got: %s, want: %s", got, want)
	}
}

func TestAccEC2Instance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		// No subnet_id specified requires default VPC with default subnets.
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckHasDefaultVPCDefaultSubnets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`instance/i-[0-9a-z]+`)),
					resource.TestCheckResourceAttr(resourceName, "instance_initiated_shutdown_behavior", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		// No subnet_id specified requires default VPC with default subnets.
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckHasDefaultVPCDefaultSubnets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Instance_inDefaultVPCBySgName(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_inDefaultVPCBySgName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_inDefaultVPCBySgID(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_inDefaultVPCBySGID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_atLeastOneOtherEBSVolume(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_atLeastOneOtherEBSVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct0), // This is an instance store AMI
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`instance/i-[0-9a-z]+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			// We repeat the exact same test so that we can be sure
			// that the user data hash stuff is working without generating
			// an incorrect diff.
			{
				Config: testAccInstanceConfig_atLeastOneOtherEBSVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSBlockDevice_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ebsKMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrEncrypted: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.kms_key_id", kmsKeyResourceName, names.AttrARN),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
func TestAccEC2Instance_EBSBlockDevice_invalidIopsForVolumeType(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_ebsBlockDeviceInvalidIOPS,
				ExpectError: regexache.MustCompile(`creating resource: iops attribute not supported for ebs_block_device with volume_type gp2`),
			},
		},
	})
}

func TestAccEC2Instance_EBSBlockDevice_invalidThroughputForVolumeType(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_ebsBlockDeviceInvalidThroughput,
				ExpectError: regexache.MustCompile(`creating resource: throughput attribute not supported for ebs_block_device with volume_type gp2`),
			},
		},
	})
}

// TestAccEC2Instance_EBSBlockDevice_RootBlockDevice_removed verifies block device mappings
// removed outside terraform no longer result in a panic.
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20821
func TestAccEC2Instance_EBSBlockDevice_RootBlockDevice_removed(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ebsAndRootBlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					// Instance must be stopped before detaching a root block device
					testAccCheckStopInstance(ctx, &instance),
					testAccCheckDetachVolumes(ctx, &instance),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Instance_RootBlockDevice_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceKMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "root_block_device.0.kms_key_id", kmsKeyResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_userDataBase64(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_userDataBase64_updateWithBashFile(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64Base64EncodedFile(rName, "test-fixtures/userdata-test.sh"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_userDataBase64_updateWithZipFile(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64Base64EncodedFile(rName, "test-fixtures/userdata-test.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_userDataBase64_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "bmV3IHdvcmxk"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_gp2IopsDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]awstypes.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[aws.ToString(blockDevice.DeviceName)] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/xvda"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/xvda")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_gp2IOPSDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", "100"),
					testCheck(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// TestAccEC2Instance_gp2WithIopsValue updated in v3.0.0
// to account for apply-time validation of the root_block_device.iops attribute for supported volume types
// Reference: https://github.com/hashicorp/terraform-provider-aws/pull/14310
func TestAccEC2Instance_gp2WithIopsValue(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_gp2IOPSValue(rName),
				ExpectError: regexache.MustCompile(`creating resource: iops attribute not supported for root_block_device with volume_type gp2`),
			},
		},
	})
}

func TestAccEC2Instance_blockDevices(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]awstypes.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[aws.ToString(blockDevice.DeviceName)] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/xvda"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/xvda")
			}

			// Check if the secondary block device exists.
			if _, ok := blockDevices["/dev/sdb"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdb")
			}

			// Check if the third block device exists.
			if _, ok := blockDevices["/dev/sdc"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdc")
			}

			// Check if the encrypted block device exists
			if _, ok := blockDevices["/dev/sdd"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdd")
			}

			if _, ok := blockDevices["/dev/sdf"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdf")
			}

			if _, ok := blockDevices["/dev/sdg"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sdg")
			}

			return nil
		}
	}

	rootVolumeSize := "11"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDevices(rName, rootVolumeSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "root_block_device.0.volume_id", regexache.MustCompile("vol-[0-9a-z]+")),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", rootVolumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdb",
						names.AttrVolumeSize: "9",
						names.AttrVolumeType: "gp2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]*regexp.Regexp{
						"volume_id": regexache.MustCompile("vol-[0-9a-z]+"),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdc",
						names.AttrVolumeSize: acctest.Ct10,
						names.AttrVolumeType: "io1",
						names.AttrIOPS:       "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdf",
						names.AttrVolumeSize: acctest.Ct10,
						names.AttrVolumeType: "gp3",
						names.AttrThroughput: "300",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdg",
						names.AttrVolumeSize: acctest.Ct10,
						names.AttrVolumeType: "gp3",
						names.AttrThroughput: "300",
						names.AttrIOPS:       "4000",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]*regexp.Regexp{
						"volume_id": regexache.MustCompile("vol-[0-9a-z]+"),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdd",
						names.AttrEncrypted:  acctest.CtTrue,
						names.AttrVolumeSize: "12",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]*regexp.Regexp{
						"volume_id": regexache.MustCompile("vol-[0-9a-z]+"),
					}),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						names.AttrDeviceName:  "/dev/sde",
						names.AttrVirtualName: "ephemeral0",
					}),
					testCheck(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_rootInstanceStore(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_noAMIEphemeralDevices(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]awstypes.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[aws.ToString(blockDevice.DeviceName)] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/xvda"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/xvda")
			}

			// Check if the secondary block not exists.
			if _, ok := blockDevices["/dev/sdb"]; ok {
				return fmt.Errorf("block device exist: /dev/sdb")
			}

			// Check if the third block device not exists.
			if _, ok := blockDevices["/dev/sdc"]; ok {
				return fmt.Errorf("block device exist: /dev/sdc")
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_noAMIEphemeralDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdb",
						"no_device":          acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						names.AttrDeviceName: "/dev/sdc",
						"no_device":          acctest.CtTrue,
					}),
					testCheck(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_sourceDestCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func(enabled bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.SourceDestCheck == nil {
				return fmt.Errorf("bad source_dest_check: got nil")
			}
			if *v.SourceDestCheck != enabled {
				return fmt.Errorf("bad source_dest_check: %#v", *v.SourceDestCheck)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_sourceDestDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheck(false),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_sourceDestEnable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheck(true),
				),
			},
			{
				Config: testAccInstanceConfig_sourceDestDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheck(false),
				),
			},
		},
	})
}

func TestAccEC2Instance_autoRecovery(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_autoRecovery(rName, "default"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.0.auto_recovery", "default"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_autoRecovery(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.0.auto_recovery", "disabled"),
				),
			},
		},
	})
}

func TestAccEC2Instance_disableAPIStop(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_disableAPIStop(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_stop", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_disableAPIStop(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_stop", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccEC2Instance_disableAPITerminationFinalFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_disableAPITermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_disableAPITermination(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccEC2Instance_disableAPITerminationFinalTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_disableAPITermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_dedicatedInstance(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dedicated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "dedicated"),
					resource.TestCheckResourceAttr(resourceName, "user_data", "562a3e32810edf6ff09994f050f12e799452379d"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"associate_public_ip_address",
					"user_data",
					"user_data_replace_on_change",
				},
			},
		},
	})
}

func TestAccEC2Instance_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_placementGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_placementGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "placement_group", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_placementPartitionNumber(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_placementPartitionNumber(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "placement_group", rName),
					resource.TestCheckResourceAttr(resourceName, "placement_partition_number", acctest.Ct3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_IPv6_supportAddressCount(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ipv6Support(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_ipv6AddressCountAndSingleAddressCausesError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_ipv6Error(rName),
				ExpectError: regexache.MustCompile("Conflicting configuration arguments"),
			},
		},
	})
}

func TestAccEC2Instance_IPv6_supportAddressCountWithIPv4(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ipv6Supportv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_IPv6AddressCount(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	originalCount := 0
	updatedCount := 2
	shrunkenCount := 1

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstance_ipv6AddressCount(rName, originalCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", fmt.Sprint(originalCount)),
				),
			},
			{
				Config: testAccInstance_ipv6AddressCount(rName, updatedCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", fmt.Sprint(updatedCount)),
				),
			},
			{
				Config: testAccInstance_ipv6AddressCount(rName, shrunkenCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", fmt.Sprint(shrunkenCount)),
				),
			},
		},
	})
}

func TestAccEC2Instance_networkInstanceSecurityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_ip", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_networkInstanceRemovingAllSecurityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkVPCSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"associate_public_ip_address",
					"public_ip",
					"user_data_replace_on_change",
				},
			},
			{
				Config: testAccInstanceConfig_networkVPCRemoveSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
				ExpectError: regexache.MustCompile(`VPC-based instances require at least one security group to be attached`),
			},
		},
	})
}

func TestAccEC2Instance_networkInstanceVPCSecurityGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkVPCSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"associate_public_ip_address",
					"public_ip",
					"user_data_replace_on_change",
				},
			},
		},
	})
}

func TestAccEC2Instance_BlockDeviceTags_volumeTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDeviceTagsNoVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, "volume_tags"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device", "user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.Name", "acceptance-test-volume-tag"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsVolumeTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.Name", "acceptance-test-volume-tag"),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.Environment", "dev"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsNoVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2Instance_BlockDeviceTags_attachedVolume(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	ebsVolumeName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDeviceTagsAttachedVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(ebsVolumeName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Name", rName),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Factum", "PerAsperaAdAstra"),
				),
			},
			{
				//https://github.com/hashicorp/terraform-provider-aws/issues/17074
				Config: testAccInstanceConfig_blockDeviceTagsAttachedVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(ebsVolumeName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Name", rName),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Factum", "PerAsperaAdAstra"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsAttachedVolumeTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(ebsVolumeName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Name", rName),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Factum", "VincitQuiSeVincit"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ebs_block_device.0.tags", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_BlockDeviceTags_ebsAndRoot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_blockDeviceTagsRootTagsConflict(rName),
				ExpectError: regexache.MustCompile(`"root_block_device\.0\.tags": conflicts with volume_tags`),
			},
			{
				Config:      testAccInstanceConfig_blockDeviceTagsEBSTagsConflict(rName),
				ExpectError: regexache.MustCompile(`"ebs_block_device\.0\.tags": conflicts with volume_tags`),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsEBSTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.1.tags.%", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsEBSAndRootTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.Purpose", "test"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsEBSAndRootTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.Env", "dev"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_BlockDeviceTags_defaultTagsVolumeTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"

	emptyMap := map[string]string{}
	mapWithOneKey1 := map[string]string{"brodo": "baggins"}
	mapWithOneKey2 := map[string]string{"every": "gnomes"}
	mapWithTwoKeys := map[string]string{"brodo": "baggins", "jelly": "bean"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{ // 1 defaultTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, emptyMap, emptyMap, emptyMap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.every", "gnomes"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags_all.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags_all.every", "gnomes"),
				),
			},
			{ // 1 defaultTags + 1 volumeTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, mapWithOneKey1, emptyMap, emptyMap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.brodo", "baggins"),
				),
			},
			{ // 1 defaultTags + 2 volumeTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, mapWithTwoKeys, emptyMap, emptyMap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.brodo", "baggins"),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.jelly", "bean"),
				),
			},
			{ // 1 defaultTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, emptyMap, emptyMap, emptyMap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
				),
			},
			{ // no tags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(emptyMap, emptyMap, emptyMap, emptyMap),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2Instance_BlockDeviceTags_defaultTagsEBSRoot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"

	emptyMap := map[string]string{}
	mapWithOneKey1 := map[string]string{"gigi": "kitty"}
	mapWithOneKey2 := map[string]string{"every": "gnomes"}
	mapWithTwoKeys1 := map[string]string{"brodo": "baggins", "jelly": "bean"}
	mapWithTwoKeys2 := map[string]string{"brodo": "baggins", "jelly": "andrew"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{ // 1 defaultTags + 0 rootTags + 1 ebsTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, emptyMap, emptyMap, mapWithOneKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags_all.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags_all.gigi", "kitty"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags_all.every", "gnomes"),
				),
			},
			{ // 1 defaultTags + 2 rootTags + 1 ebsTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, emptyMap, mapWithTwoKeys1, mapWithOneKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.every", "gnomes"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.brodo", "baggins"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.jelly", "bean"),
				),
			},
			{ // 1 defaultTags + 2 rootTags (1 update) + 1 ebsTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(mapWithOneKey2, emptyMap, mapWithTwoKeys2, mapWithOneKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.every", "gnomes"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.brodo", "baggins"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.jelly", "andrew"),
				),
			},
			{ // 0 defaultTags + 2 rootTags + 1 ebsTags
				Config: testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(emptyMap, emptyMap, mapWithTwoKeys2, mapWithOneKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.brodo", "baggins"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags_all.jelly", "andrew"),
				),
			},
		},
	})
}

func TestAccEC2Instance_instanceProfileChange(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheckInstanceProfile := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.IamInstanceProfile == nil {
				return fmt.Errorf("Instance Profile is nil - we expected an InstanceProfile associated with the Instance")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_noProfile(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_profile(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheckInstanceProfile(),
				),
			},
			{
				Config: testAccInstanceConfig_profile(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(ctx, &v), // GH-8262: Error on EC2 instance role change when stopped
				),
			},
			{
				Config: testAccInstanceConfig_profile(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheckInstanceProfile(),
				),
			},
		},
	})
}

func TestAccEC2Instance_iamInstanceProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheckInstanceProfile := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.IamInstanceProfile == nil {
				return fmt.Errorf("Instance Profile is nil - we expected an InstanceProfile associated with the Instance")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_profile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheckInstanceProfile(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17719
func TestAccEC2Instance_iamInstanceProfilePath(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_profilePath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_privateIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *v.PrivateIpAddress != "10.1.1.42" {
				return fmt.Errorf("bad private IP: %s", *v.PrivateIpAddress)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheckPrivateIP(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_associatePublicIPAndPrivateIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *v.PrivateIpAddress != "10.1.1.42" {
				return fmt.Errorf("bad private IP: %s", *v.PrivateIpAddress)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicIPAndPrivateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheckPrivateIP(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// Allow Empty Private IP
// https://github.com/hashicorp/terraform-provider-aws/issues/13626
func TestAccEC2Instance_Empty_privateIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if aws.ToString(v.PrivateIpAddress) == "" {
				return fmt.Errorf("bad computed private IP: %s", aws.ToString(v.PrivateIpAddress))
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_emptyPrivateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					testCheckPrivateIP(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_PrivateDNSNameOptions_computed(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_computed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.hostname_type", "resource-name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_PrivateDNSNameOptions_configured(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_configured(rName, false, true, "ip-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.hostname_type", "ip-name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_configured(rName, true, true, "ip-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.hostname_type", "ip-name"),
				),
			},
			// "InvalidParameterValue: Cannot modify hostname-type for running instances".
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_configured(rName, true, true, "ip-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(ctx, &v2),
				),
			},
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_configured(rName, true, true, "resource-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v3),
					testAccCheckInstanceNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.hostname_type", "resource-name"),
				),
			},
		},
	})
}

// Guard against regression with KeyPairs
// https://github.com/hashicorp/terraform/issues/2302
func TestAccEC2Instance_keyPairCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	keyPairResourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_keyPair(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", keyPairResourceName, "key_name"),
				),
			},
		},
	})
}

// This test reproduces the bug here:
//
//	https://github.com/hashicorp/terraform/issues/1752
//
// I wish there were a way to exercise resources built with helper.Schema in a
// unit context, in which case this test could be moved there, but for now this
// will cover the bugfix.
//
// The following triggers "diffs didn't match during apply" without the fix in to
// set NewRemoved on the .# field when it changes to 0.
func TestAccEC2Instance_forceNewAndTagsDrift(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_forceNewAndTagsDrift(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					driftTags(ctx, &v),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccInstanceConfig_forceNewAndTagsDriftUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_changeInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_type(rName, "t2.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.medium"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_type(rName, "t2.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.large"),
				),
			},
		},
	})
}

func TestAccEC2Instance_changeInstanceTypeReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_typeReplace(rName, "m5.2xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "m5.2xlarge"),
				),
			},
			{
				Config: testAccInstanceConfig_typeReplace(rName, "m6g.2xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckInstanceRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "m6g.2xlarge"),
				),
			},
		},
	})
}

func TestAccEC2Instance_changeInstanceTypeAndUserData(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	hash := sha1.Sum([]byte("hello world"))
	expectedUserData := hex.EncodeToString(hash[:])
	hash2 := sha1.Sum([]byte("new world"))
	expectedUserDataUpdated := hex.EncodeToString(hash2[:])

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_typeAndUserData(rName, "t2.medium", "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "user_data", expectedUserData),
				),
			},
			{
				Config: testAccInstanceConfig_typeAndUserData(rName, "t2.large", "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.large"),
					resource.TestCheckResourceAttr(resourceName, "user_data", expectedUserDataUpdated),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_changeInstanceTypeAndUserDataBase64(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_typeAndUserDataBase64(rName, "t2.medium", "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_typeAndUserDataBase64(rName, "t2.large", "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, "t2.large"),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "bmV3IHdvcmxk"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ebsRootDeviceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.delete_on_termination"),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.encrypted"),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.iops"),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.volume_size"),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.volume_type"),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.volume_id"),
					resource.TestCheckResourceAttrSet(resourceName, "root_block_device.0.device_name"),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifySize(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeType := "gp2"

	originalSize := "30"
	updatedSize := "32"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, originalSize, acctest.CtTrue, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", originalSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, updatedSize, acctest.CtTrue, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", updatedSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifyType(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"

	originalType := "gp2"
	updatedType := "standard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, acctest.CtTrue, originalType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", originalType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, acctest.CtTrue, updatedType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", updatedType),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceModifyIOPS_io1(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"

	volumeType := "io1"

	originalIOPS := "100"
	updatedIOPS := "200"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, acctest.CtTrue, volumeType, originalIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", originalIOPS),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, acctest.CtTrue, volumeType, updatedIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", updatedIOPS),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceModifyIOPS_io2(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	volumeType := "io2"

	originalIOPS := "100"
	updatedIOPS := "200"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, acctest.CtTrue, volumeType, originalIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", originalIOPS),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, acctest.CtTrue, volumeType, updatedIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", updatedIOPS),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceModifyThroughput_gp3(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"

	volumeType := "gp3"

	originalThroughput := "250"
	updatedThroughput := "300"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceThroughput(rName, volumeSize, acctest.CtTrue, volumeType, originalThroughput),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.throughput", originalThroughput),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceThroughput(rName, volumeSize, acctest.CtTrue, volumeType, updatedThroughput),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.throughput", updatedThroughput),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifyDeleteOnTermination(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	volumeType := "gp2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, acctest.CtFalse, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, acctest.CtTrue, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifyAll(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	originalSize := "30"
	updatedSize := "32"

	originalType := "gp2"
	updatedType := "io1"

	updatedIOPS := "200"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, originalSize, acctest.CtFalse, originalType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", originalSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", originalType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, updatedSize, acctest.CtTrue, updatedType, updatedIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", updatedSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", updatedType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", updatedIOPS),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceMultipleBlockDevices_modifySize(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	updatedRootVolumeSize := "14"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, acctest.Ct10, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", acctest.Ct10),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: acctest.Ct10,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "12",
					}),
				),
			},
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, updatedRootVolumeSize, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", updatedRootVolumeSize),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: acctest.Ct10,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "12",
					}),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceMultipleBlockDevices_modifyDeleteOnTermination(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, acctest.Ct10, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtFalse),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: acctest.Ct10,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "12",
					}),
				),
			},
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, acctest.Ct10, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", acctest.CtTrue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: acctest.Ct10,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrVolumeSize: "12",
					}),
				),
			},
		},
	})
}

// Test to validate fix for GH-ISSUE #1318 (dynamic ebs_block_devices forcing replacement after state refresh)
func TestAccEC2Instance_EBSRootDevice_multipleDynamicEBSBlockDevices(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dynamicEBSBlockDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeleteOnTermination: acctest.CtTrue,
						names.AttrDeviceName:          "/dev/sdd",
						names.AttrEncrypted:           acctest.CtFalse,
						names.AttrIOPS:                "100",
						names.AttrVolumeSize:          acctest.Ct10,
						names.AttrVolumeType:          "gp2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeleteOnTermination: acctest.CtTrue,
						names.AttrDeviceName:          "/dev/sdc",
						names.AttrEncrypted:           acctest.CtFalse,
						names.AttrIOPS:                "100",
						names.AttrVolumeSize:          acctest.Ct10,
						names.AttrVolumeType:          "gp2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						names.AttrDeleteOnTermination: acctest.CtTrue,
						names.AttrDeviceName:          "/dev/sdb",
						names.AttrEncrypted:           acctest.CtFalse,
						names.AttrIOPS:                "100",
						names.AttrVolumeSize:          acctest.Ct10,
						names.AttrVolumeType:          "gp2",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_gp3RootBlockDevice(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]awstypes.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[aws.ToString(blockDevice.DeviceName)] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/xvda"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/xvda")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_gp3RootBlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", "gp3"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", "4000"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.throughput", "300"),
					testCheck(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_primaryNetworkInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	var eni awstypes.NetworkInterface
	resourceName := "aws_instance.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_primaryNetworkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					testAccCheckENIExists(ctx, eniResourceName, &eni),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "network_interface.*", map[string]string{
						"device_index":       acctest.Ct0,
						"network_card_index": acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_networkCardIndex(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#network-cards.
	// Only specialized (and expensive) instance types support multiple network cards (and hence network_card_index > 0).
	// Don't attempt to test with such instance types.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkCardIndex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "network_interface.*", map[string]string{
						"device_index":       acctest.Ct0,
						"network_card_index": acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_primaryNetworkInterfaceSourceDestCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	var eni awstypes.NetworkInterface
	resourceName := "aws_instance.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_primaryNetworkInterfaceSourceDestCheck(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					testAccCheckENIExists(ctx, eniResourceName, &eni),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_addSecondaryInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	var eniPrimary awstypes.NetworkInterface
	var eniSecondary awstypes.NetworkInterface
	resourceName := "aws_instance.test"
	eniPrimaryResourceName := "aws_network_interface.primary"
	eniSecondaryResourceName := "aws_network_interface.secondary"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_addSecondaryNetworkInterfaceBefore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					testAccCheckENIExists(ctx, eniPrimaryResourceName, &eniPrimary),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_addSecondaryNetworkInterfaceAfter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckENIExists(ctx, eniSecondaryResourceName, &eniSecondary),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", acctest.Ct1),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/3205
func TestAccEC2Instance_addSecurityGroupNetworkInterface(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_addSecurityGroupBefore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_addSecurityGroupAfter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7063
func TestAccEC2Instance_NewNetworkInterface_publicIPAndSecondaryPrivateIPs(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_publicAndPrivateSecondaryIPs(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct2),
				),
			},
			{
				Config: testAccInstanceConfig_publicAndPrivateSecondaryIPs(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7063
func TestAccEC2Instance_NewNetworkInterface_emptyPrivateIPAndSecondaryPrivateIPs(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	secondaryIPs := fmt.Sprintf("%q, %q", "10.1.1.42", "10.1.1.43")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7063
func TestAccEC2Instance_NewNetworkInterface_emptyPrivateIPAndSecondaryPrivateIPsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	secondaryIP := fmt.Sprintf("%q", "10.1.1.42")
	secondaryIPs := fmt.Sprintf("%s, %q", secondaryIP, "10.1.1.43")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct2),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7063
func TestAccEC2Instance_NewNetworkInterface_privateIPAndSecondaryPrivateIPs(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateIP := "10.1.1.42"
	secondaryIPs := fmt.Sprintf("%q, %q", "10.1.1.43", "10.1.1.44")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7063
func TestAccEC2Instance_NewNetworkInterface_privateIPAndSecondaryPrivateIPsUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateIP := "10.1.1.42"
	secondaryIP := fmt.Sprintf("%q", "10.1.1.43")
	secondaryIPs := fmt.Sprintf("%s, %q", secondaryIP, "10.1.1.44")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct2),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct0),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/227
func TestAccEC2Instance_AssociatePublic_defaultPrivate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicDefaultPrivate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/227
func TestAccEC2Instance_AssociatePublic_defaultPublic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicDefaultPublic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/227
func TestAccEC2Instance_AssociatePublic_explicitPublic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicExplicitPublic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/227
func TestAccEC2Instance_AssociatePublic_explicitPrivate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicExplicitPrivate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/227
func TestAccEC2Instance_AssociatePublic_overridePublic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicOverridePublic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/227
func TestAccEC2Instance_AssociatePublic_overridePrivate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicOverridePrivate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	amiDataSourceName := "data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64"
	instanceTypeDataSourceName := "data.aws_ec2_instance_type_offering.available"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", launchTemplateResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
					resource.TestCheckResourceAttrPair(resourceName, "ami", amiDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceType, instanceTypeDataSourceName, names.AttrInstanceType),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_overrideTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	amiDataSourceName := "data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64"
	instanceTypeDataSourceName := "data.aws_ec2_instance_type_offering.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateOverrideTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ami", amiDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceType, instanceTypeDataSourceName, names.AttrInstanceType),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_setSpecificVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
				),
			},
			{
				Config: testAccInstanceConfig_templateSpecificVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", launchTemplateResourceName, "default_version"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplateModifyTemplate_defaultVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
				),
			},
			{
				Config: testAccInstanceConfig_templateModifyTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_updateTemplateVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateSpecificVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", launchTemplateResourceName, "default_version"),
				),
			},
			{
				Config: testAccInstanceConfig_templateUpdateVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", launchTemplateResourceName, "default_version"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_swapIDAndName(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", launchTemplateResourceName, names.AttrName),
				),
			},
			{
				Config: testAccInstanceConfig_templateName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", launchTemplateResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_iamInstanceProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateWithIAMInstanceProfile(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "iam_instance_profile", launchTemplateResourceName, "iam_instance_profile.0.name"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_spotAndStop(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateSpotAndStop(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_vpcSecurityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateWithVPCSecurityGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccEC2Instance_GetPasswordData_falseToTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "password_data", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_getPasswordData(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccEC2Instance_GetPasswordData_trueToFalse(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password_data",
					"get_password_data",
					"user_data_replace_on_change",
				},
			},
			{
				Config: testAccInstanceConfig_getPasswordData(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "password_data", ""),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsAmdSevSnpUnspecifiedToDisabledToEnabledToUnspecified(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// AMD SEV-SNP currently only supported in us-east-2
			// Ref: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html
			acctest.PreCheckRegion(t, names.USEast2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnpUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					// empty string set if amd_sev_snp is not specified
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				// test DiffSuppressFunc to suppress "" to "disabled"
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					// Since read is not triggered, empty string is returned
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", ""),
				),
			},
			{
				// expect recreation when it is enabled
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v3),
					testAccCheckInstanceRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
				),
			},
			{
				// expect no recreation if the cpu options block is removed - amd_sev_snp should still be enabled
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnpUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v4),
					testAccCheckInstanceNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsAmdSevSnpUnspecifiedToEnabledToDisabledToUnspecified(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// AMD SEV-SNP currently only supported in us-east-2
			// Ref: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html
			acctest.PreCheckRegion(t, names.USEast2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnpUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					// API returns empty string if amd_sev_snp is not specified
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				// expect recreation when it is enabled
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
				),
			},
			{
				// expect recreation when it is disabled
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v3),
					testAccCheckInstanceRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationDisabled)),
				),
			},
			{
				// expect no recreation if the cpu options block is removed - amd_sev_snp should still be disabled
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnpUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v4),
					testAccCheckInstanceNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationDisabled)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsAmdSevSnpEnabledToDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// AMD SEV-SNP currently only supported in us-east-2
			// Ref: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html
			acctest.PreCheckRegion(t, names.USEast2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationDisabled)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsAmdSevSnpDisabledToEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// AMD SEV-SNP currently only supported in us-east-2
			// Ref: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html
			acctest.PreCheckRegion(t, names.USEast2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationDisabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, string(awstypes.AmdSevSnpSpecificationEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsAmdSevSnpCoreThreads(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	originalCoreCount := 2
	updatedCoreCount := 3
	originalThreadsPerCore := 2
	updatedThreadsPerCore := 1

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// AMD SEV-SNP currently only supported in us-east-2
			// Ref: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html
			acctest.PreCheckRegion(t, names.USEast2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnpCoreThreads(rName, string(awstypes.AmdSevSnpSpecificationEnabled), originalCoreCount, originalThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationEnabled)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(originalCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(originalThreadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(originalCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(originalThreadsPerCore)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_cpuOptionsAmdSevSnpCoreThreads(rName, string(awstypes.AmdSevSnpSpecificationDisabled), updatedCoreCount, updatedThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.amd_sev_snp", string(awstypes.AmdSevSnpSpecificationDisabled)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(updatedCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(updatedThreadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(updatedCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(updatedThreadsPerCore)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsCoreThreads(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	originalCoreCount := 2
	updatedCoreCount := 3
	originalThreadsPerCore := 2
	updatedThreadsPerCore := 1

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsCoreThreads(rName, originalCoreCount, originalThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(originalCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(originalThreadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(originalCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(originalThreadsPerCore)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_cpuOptionsCoreThreads(rName, updatedCoreCount, updatedThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(updatedCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(updatedThreadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(updatedCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(updatedThreadsPerCore)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsCoreThreadsMigration(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	coreCount := 2
	threadsPerCore := 2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsCoreThreadsDeprecated(rName, coreCount, threadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(coreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(threadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(coreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(threadsPerCore)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				// EC2 instance should not be recreated
				Config: testAccInstanceConfig_cpuOptionsCoreThreads(rName, coreCount, threadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(coreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(threadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(coreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(threadsPerCore)),
				),
			},
		},
	})
}

func TestAccEC2Instance_cpuOptionsCoreThreadsUnspecifiedToSpecified(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	defaultCoreCount := 4
	defaultThreadsPerCore := 2

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_cpuOptionsUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(defaultCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(defaultThreadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(defaultCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(defaultThreadsPerCore)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				// EC2 instance should not be recreated
				Config: testAccInstanceConfig_cpuOptionsCoreThreads(rName, defaultCoreCount, defaultThreadsPerCore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.core_count", strconv.Itoa(defaultCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_options.0.threads_per_core", strconv.Itoa(defaultThreadsPerCore)),
					resource.TestCheckResourceAttr(resourceName, "cpu_core_count", strconv.Itoa(defaultCoreCount)),
					resource.TestCheckResourceAttr(resourceName, "cpu_threads_per_core", strconv.Itoa(defaultThreadsPerCore)),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationEmpty_nonBurstable(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationEmptyNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credit_specification", "user_data_replace_on_change"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10203
func TestAccEC2Instance_CreditSpecificationUnspecifiedToEmpty_nonBurstable(t *testing.T) {
	ctx := acctest.Context(t)
	var instance awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecifiedNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationEmptyNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_unspecifiedDefaultsToStandard(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_standardCPUCredits(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_unlimitedCPUCredits(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnknownCPUCredits_t2(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnknownCPUCredits_t3(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnknownCPUCredits_t3a(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t3a.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnknownCPUCredits_t4g(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t4g.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_updateCPUCredits(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second, third awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_isNotAppliedToNonBurstable(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationIsNotAppliedToNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credit_specification", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationT3_unspecifiedDefaultsToUnlimited(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecifiedT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationT3_standardCPUCredits(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecifiedT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationT3_unlimitedCPUCredits(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecifiedT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationT3_updateCPUCredits(t *testing.T) {
	ctx := acctest.Context(t)
	var first, second, third awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationStandardCPUCredits_t2Tot3Taint(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnlimitedCPUCredits_t2Tot3Taint(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

func TestAccEC2Instance_UserData(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_UserData_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccInstanceConfig_userData(rName, "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_UserData_stringToEncodedString(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64Encoded(rName, "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_UserData_emptyStringToUnspecified(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataEmptyString(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data", "user_data_replace_on_change"},
			},
			// Switching should show no difference
			{
				Config:             testAccInstanceConfig_userDataUnspecified(rName),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccEC2Instance_UserData_unspecifiedToEmptyString(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			// Switching should show no difference
			{
				Config:             testAccInstanceConfig_userDataEmptyString(rName),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_On(t *testing.T) {
	ctx := acctest.Context(t)
	var instance1, instance2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData1", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			// Switching should force a recreate
			{
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData2", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance2),
					testAccCheckInstanceRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_On_Base64(t *testing.T) {
	ctx := acctest.Context(t)
	var instance1, instance2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955d", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "user_data"},
			},
			// Switching should force a recreate
			{
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955e", acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance2),
					testAccCheckInstanceRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_Off(t *testing.T) {
	ctx := acctest.Context(t)
	var instance1, instance2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData1", acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			// Switching should not force a recreate
			{
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData2", acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance2),
					testAccCheckInstanceNotRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_Off_Base64(t *testing.T) {
	ctx := acctest.Context(t)
	var instance1, instance2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955d", acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "user_data"},
			},
			// Switching should not force a recreate
			{
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955e", acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance2),
					testAccCheckInstanceNotRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_hibernation(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_hibernation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "hibernation", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_hibernation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "hibernation", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccEC2Instance_metadataOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_metadataOptionsDefaults(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "optional"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "disabled"),
				),
			},
			{
				Config: testAccInstanceConfig_metadataOptionsDisabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "optional"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "disabled"),
				),
			},
			{
				Config: testAccInstanceConfig_metadataOptionsUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", names.AttrEnabled),
				),
			},
			{
				Config: testAccInstanceConfig_metadataOptionsUpdatedAgain(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_protocol_ipv6", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "optional"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "disabled"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_enclaveOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_enclaveOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservation_unspecifiedDefaultsToOpen(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			// Adding 'open' preference should show no difference
			{
				Config:             testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservationPreference_open(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservationPreference_none(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "none"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservation_targetID(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationTargetID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_resource_group_arn", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservation_modifyPreference(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
				),
			},
			{Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(ctx, &original), // Stop instance to modify capacity reservation
				),
			},
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "none"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservation_modifyTarget(t *testing.T) {
	ctx := acctest.Context(t)
	var original, updated awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "none"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct0),
				),
			},
			{Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(ctx, &original), // Stop instance to modify capacity reservation
				),
			},
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationTargetID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_resource_group_arn", ""),
				),
			},
		},
	})
}

func TestAccEC2Instance_basicWithSpot(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		// No subnet_id specified requires default VPC with default subnets.
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basicWithSpot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "spot_instance_request_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_lifecycle", "spot"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.market_type", "spot"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.0.instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_market_options.0.spot_options.0.max_price"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.0.spot_instance_type", "one-time"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.0.valid_until", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
		},
	})
}

func testAccCheckInstanceNotRecreated(before, after *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.InstanceId), aws.ToString(after.InstanceId); before != after {
			return fmt.Errorf("EC2 Instance (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccCheckInstanceRecreated(before, after *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.InstanceId), aws.ToString(after.InstanceId); before == after {
			return fmt.Errorf("EC2 Instance (%s) not recreated", before)
		}

		return nil
	}
}

func testAccCheckInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_instance" {
				continue
			}

			_, err := tfec2.FindInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckInstanceExists(ctx context.Context, n string, v *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStopInstance(ctx context.Context, v *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		return tfec2.StopInstance(ctx, conn, aws.ToString(v.InstanceId), false, 10*time.Minute)
	}
}

func testAccCheckDetachVolumes(ctx context.Context, instance *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, v := range instance.BlockDeviceMappings {
			if v.Ebs != nil && v.Ebs.VolumeId != nil {
				deviceName := aws.ToString(v.DeviceName)
				instanceID := aws.ToString(instance.InstanceId)
				volumeID := aws.ToString(v.Ebs.VolumeId)

				// Make sure in correct state before detaching.
				if _, err := tfec2.WaitVolumeAttachmentCreated(ctx, conn, volumeID, instanceID, deviceName, 5*time.Minute); err != nil {
					return err
				}

				r := tfec2.ResourceVolumeAttachment()
				d := r.Data(nil)
				d.Set(names.AttrDeviceName, deviceName)
				d.Set(names.AttrInstanceID, instanceID)
				d.Set("volume_id", volumeID)

				if err := acctest.DeleteResource(ctx, r, d, acctest.Provider.Meta()); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func TestInstanceHostIDSchema(t *testing.T) {
	t.Parallel()

	actualSchema := tfec2.ResourceInstance().SchemaMap()["host_id"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ForceNew: true,
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func TestInstanceCPUCoreCountSchema(t *testing.T) {
	t.Parallel()

	actualSchema := tfec2.ResourceInstance().SchemaMap()["cpu_core_count"]
	expectedSchema := &schema.Schema{
		Type:          schema.TypeInt,
		Optional:      true,
		Computed:      true,
		ForceNew:      true,
		Deprecated:    "use 'cpu_options' argument instead",
		ConflictsWith: []string{"cpu_options.0.core_count"},
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func TestInstanceCPUThreadsPerCoreSchema(t *testing.T) {
	t.Parallel()

	actualSchema := tfec2.ResourceInstance().SchemaMap()["cpu_threads_per_core"]
	expectedSchema := &schema.Schema{
		Type:          schema.TypeInt,
		Optional:      true,
		Computed:      true,
		ForceNew:      true,
		Deprecated:    "use 'cpu_options' argument instead",
		ConflictsWith: []string{"cpu_options.0.threads_per_core"},
	}
	if !reflect.DeepEqual(actualSchema, expectedSchema) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			actualSchema,
			expectedSchema)
	}
}

func driftTags(ctx context.Context, instance *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		_, err := conn.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{aws.ToString(instance.InstanceId)},
			Tags: []awstypes.Tag{
				{
					Key:   aws.String("Drift"),
					Value: aws.String("Happens"),
				},
			},
		})
		return err
	}
}

// testAccPreCheckHasDefaultVPCDefaultSubnets checks that the test region has
// - A default VPC with default subnets.
// This check is useful to ensure that an instance can be launched without specifying a subnet.
func testAccPreCheckHasDefaultVPCDefaultSubnets(ctx context.Context, t *testing.T) {
	client := acctest.Provider.Meta().(*conns.AWSClient)

	if !(hasDefaultVPC(ctx, t) && defaultSubnetCount(ctx, t) > 0) {
		t.Skipf("skipping tests; %s does not have a default VPC with default subnets", client.Region)
	}
}

// defaultVPC returns the ID of the default VPC for the current AWS Region, or "" if none exists.
func defaultVPC(ctx context.Context, t *testing.T) string {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	output, err := conn.DescribeAccountAttributes(ctx, &ec2.DescribeAccountAttributesInput{
		AttributeNames: flex.ExpandStringyValueList[awstypes.AccountAttributeName]([]any{string(awstypes.AccountAttributeNameDefaultVpc)}),
	})

	if acctest.PreCheckSkipError(err) {
		return ""
	}

	if err != nil {
		t.Fatalf("error describing EC2 account attributes: %s", err)
	}

	if len(output.AccountAttributes) > 0 && len(output.AccountAttributes[0].AttributeValues) > 0 {
		if v := aws.ToString(output.AccountAttributes[0].AttributeValues[0].AttributeValue); v != "none" {
			return v
		}
	}

	return ""
}

func hasDefaultVPC(ctx context.Context, t *testing.T) bool {
	return defaultVPC(ctx, t) != ""
}

// defaultSubnetCount returns the number of default subnets in the current region's default VPC.
func defaultSubnetCount(ctx context.Context, t *testing.T) int {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeSubnetsInput{
		Filters: tfec2.NewAttributeFilterList(
			map[string]string{
				"defaultForAz": acctest.CtTrue,
			},
		),
	}

	subnets, err := tfec2.FindSubnets(ctx, conn, input)

	if acctest.PreCheckSkipError(err) {
		return 0
	}

	if err != nil {
		t.Fatalf("error listing default subnets: %s", err)
	}

	return len(subnets)
}

// testAccLatestWindowsServer2016CoreAMIConfig returns the configuration for a data source that
// describes the latest Microsoft Windows Server 2016 Core AMI.
// The data source is named 'win2016core-ami'.
func testAccLatestWindowsServer2016CoreAMIConfig() string {
	return `
data "aws_ami" "win2016core-ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["Windows_Server-2016-English-Core-Base-*"]
  }
}
`
}

// testAccLatestAmazonLinux2023AMIConfig returns the configuration for a data source that
// describes the latest Amazon Linux 2023 AMI for x86.
// The data source is named 'amzn-linux-2023-ami'.
func testAccLatestAmazonLinux2023AMIConfig() string {
	return testAccLatestAmazonLinux2023AMIConfigWithProvider(acctest.ProviderName)
}

func testAccLatestAmazonLinux2023AMIConfigWithProvider(provider string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-linux-2023-ami" {
  provider    = %[1]q
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}
`, provider)
}

func testAccAvailableAZsWavelengthZonesExcludeConfig(excludeZoneIds ...string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  exclude_zone_ids = [%[1]q]
  state            = "available"

  filter {
    name   = "zone-type"
    values = ["wavelength-zone"]
  }

  filter {
    name   = "opt-in-status"
    values = ["opted-in"]
  }
}
`, strings.Join(excludeZoneIds, "\", \""))
}

func testAccAvailableAZsWavelengthZonesDefaultExcludeConfig() string {
	// Exclude usw2-wl1-den-wlz1 as there may be problems allocating carrier IP addresses.
	return testAccAvailableAZsWavelengthZonesExcludeConfig("usw2-wl1-den-wlz1")
}

// testAccInstanceVPCConfig returns the configuration for tests that create
//  1. a VPC without IPv6 support
//  2. a subnet in the VPC that optionally assigns public IP addresses to ENIs
//
// The resources are named 'test'.
func testAccInstanceVPCConfig(rName string, mapPublicIpOnLaunch bool, azIndex int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = aws_vpc.test.id
  availability_zone       = data.aws_availability_zones.available.names[%[3]d]
  map_public_ip_on_launch = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, mapPublicIpOnLaunch, azIndex))
}

// testAccInstanceVPCSecurityGroupConfig returns the configuration for tests that create
//  1. a VPC security group
//  2. an internet gateway in the VPC
//
// The resources are named 'test'.
func testAccInstanceVPCSecurityGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "test"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

// testAccInstanceVPCIPv6Config returns the configuration for tests that create
//  1. a VPC with IPv6 support
//  2. a subnet in the VPC with an assigned IPv6 CIDR block
//
// The resources are named 'test'.
func testAccInstanceVPCIPv6Config(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_basic() string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  # Explicitly no tags so as to test creation without tags.
}
`)
}

func testAccInstanceConfig_inDefaultVPCBySgName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type     = "t2.micro"
  security_groups   = [aws_security_group.test.name]
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_inDefaultVPCBySGID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  availability_zone      = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_atLeastOneOtherEBSVolume(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIDataSourceConfig_latestUbuntuBionicHVMInstanceStore(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
# Ensure that there is at least 1 EBS volume in the current region.
# See https://github.com/hashicorp/terraform/issues/1249.
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 5

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami = data.aws_ami.ubuntu-bionic-ami-hvm-instance-store.id

  # tflint-ignore: aws_instance_previous_type
  instance_type = "m1.small"
  subnet_id     = aws_subnet.test.id
  user_data     = "foo:-with-character's"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ebs_volume.test]
}
`, rName))
}

func testAccInstanceConfig_userData(rName, userData string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type = "t2.small"
  user_data     = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, userData))
}

func testAccInstanceConfig_userDataBase64Encoded(rName, userData string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type = "t2.small"
  user_data     = base64encode(%[2]q)

  tags = {
    Name = %[1]q
  }
}
`, rName, userData))
}

func testAccInstanceConfig_userDataBase64(rName, userData string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type    = "t2.small"
  user_data_base64 = base64encode(%[2]q)

  tags = {
    Name = %[1]q
  }
}
`, rName, userData))
}

func testAccInstanceConfig_userDataBase64Base64EncodedFile(rName, filename string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type    = "t2.small"
  user_data_base64 = filebase64(%[2]q)

  tags = {
    Name = %[1]q
  }
}
`, rName, filename))
}

func testAccInstanceConfig_type(rName, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type = %[1]q

  tags = {
    Name = %[2]q
  }
}
`, instanceType, rName))
}

func testAccInstanceConfig_typeReplace(rName, instanceType string) string {
	arch := acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI()
	archs := "x86_64"
	if strings.HasPrefix(instanceType, "m6g.") {
		arch = acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI()
		archs = "arm64"
	}
	return acctest.ConfigCompose(
		arch,
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-%[3]s.id
  subnet_id = aws_subnet.test.id

  instance_type = %[1]q

  tags = {
    Name = %[2]q
  }

  lifecycle {
    ignore_changes = [ami]
  }
}
`, instanceType, rName, archs))
}

func testAccInstanceConfig_typeAndUserData(rName, instanceType, userData string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type = %[2]q
  user_data     = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceType, userData))
}

func testAccInstanceConfig_typeAndUserDataBase64(rName, instanceType, userData string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  subnet_id = aws_subnet.test.id

  instance_type    = %[2]q
  user_data_base64 = base64encode(%[3]q)

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceType, userData))
}

func testAccInstanceConfig_gp2IOPSDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_gp2IOPSValue(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
    # configured explicitly
    iops = 10
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_rootStore(rName string) string {
	return acctest.ConfigCompose(testAccAMIDataSourceConfig_latestUbuntuBionicHVMInstanceStore(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.ubuntu-bionic-ami-hvm-instance-store.id

  # Only certain instance types support ephemeral root instance stores.
  # http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html
  # tflint-ignore: aws_instance_previous_type
  instance_type = "m3.medium"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_noAMIEphemeralDevices(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ephemeral_block_device {
    device_name = "/dev/sdb"
    no_device   = true
  }

  ephemeral_block_device {
    device_name = "/dev/sdc"
    no_device   = true
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_ebsRootDeviceBasic(rName string) string {
	return acctest.ConfigCompose(testAccInstanceAMIWithEBSRootVolume, fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.ami.id

  instance_type = "t2.medium"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_rootBlockDevice(rName, size, delete, volumeType string) string {
	return testAccInstanceConfig_rootBlockDeviceIOPS(rName, size, delete, volumeType, "")
}

func testAccInstanceConfig_rootBlockDeviceIOPS(rName, size, delete, volumeType, iops string) string {
	if iops == "" {
		iops = "null"
	}

	return acctest.ConfigCompose(testAccInstanceAMIWithEBSRootVolume, fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.ami.id

  instance_type = "t2.medium"

  root_block_device {
    volume_size           = %[2]s
    delete_on_termination = %[3]s
    volume_type           = %[4]q
    iops                  = %[5]s
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, size, delete, volumeType, iops))
}

func testAccInstanceConfig_rootBlockDeviceThroughput(rName, size, delete, volumeType, throughput string) string {
	if throughput == "" {
		throughput = "null"
	}

	return acctest.ConfigCompose(testAccInstanceAMIWithEBSRootVolume, fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.ami.id

  instance_type = "t2.medium"

  root_block_device {
    volume_size           = %[2]s
    delete_on_termination = %[3]s
    volume_type           = %[4]q
    throughput            = %[5]s
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, size, delete, volumeType, throughput))
}

func testAccInstanceConfig_gp3RootBlockDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size = 10
    volume_type = "gp3"
    throughput  = 300
    iops        = 4000
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

const testAccInstanceAMIWithEBSRootVolume = `
data "aws_ami" "ami" {
  owners      = ["amazon"]
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }
}
`

func testAccInstanceConfig_blockDevices(rName, size string) string {
	return testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, size, "")
}

func testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, size, delete string) string {
	if delete == "" {
		delete = "null"
	}

	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type           = "gp2"
    volume_size           = %[2]s
    delete_on_termination = %[3]s
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops        = 100
  }

  # Encrypted ebs block device

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 12
    encrypted   = true
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  ebs_block_device {
    device_name = "/dev/sdf"
    volume_size = 10
    volume_type = "gp3"
    throughput  = 300
  }

  ebs_block_device {
    device_name = "/dev/sdg"
    volume_size = 10
    volume_type = "gp3"
    throughput  = 300
    iops        = 4000
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, size, delete))
}

func testAccInstanceConfig_sourceDestEnable(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.small"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_sourceDestDisable(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type     = "t2.small"
  subnet_id         = aws_subnet.test.id
  source_dest_check = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_autoRecovery(rName string, val string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  maintenance_options {
    auto_recovery = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, val))
}

func testAccInstanceConfig_disableAPIStop(rName string, val bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami              = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type    = "t2.small"
  subnet_id        = aws_subnet.test.id
  disable_api_stop = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, val))
}

func testAccInstanceConfig_disableAPITermination(rName string, val bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type           = "t2.small"
  subnet_id               = aws_subnet.test.id
  disable_api_termination = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, val))
}

func testAccInstanceConfig_dedicated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 1),
		// Prevent frequent errors like
		//	"InsufficientInstanceCapacity: We currently do not have sufficient m1.small capacity in the Availability Zone you requested (us-west-2a)."
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[1]", "t3.small", "t3.micro", "m1.small", "a1.medium"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  tenancy                     = "dedicated"
  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_outpost(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_outposts_outpost_instance_types" "test" {
  arn = data.aws_outposts_outpost.test.arn
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  cidr_block        = "10.1.1.0/24"
  outpost_arn       = data.aws_outposts_outpost.test.arn
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
  subnet_id     = aws_subnet.test.id

  root_block_device {
    volume_type = "gp2"
    volume_size = 8
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_placementGroup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "c5.large"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  placement_group             = aws_placement_group.test.name

  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_placementPartitionNumber(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name            = %[1]q
  strategy        = "partition"
  partition_count = 4
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "c5.large"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  placement_group             = aws_placement_group.test.name
  placement_partition_number  = 3

  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_ipv6Error(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type      = "t2.micro"
  subnet_id          = aws_subnet.test.id
  ipv6_addresses     = ["2600:1f14:bb2:e501::10"]
  ipv6_address_count = 1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_ipv6Support(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type      = "t2.micro"
  subnet_id          = aws_subnet.test.id
  ipv6_address_count = 1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_ipv6Supportv4(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  ipv6_address_count          = 1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstance_ipv6AddressCount(rName string, ipv6AddressCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type      = "t2.medium"
  subnet_id          = aws_subnet.test.id
  ipv6_address_count = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6AddressCount))
}

func testAccInstanceConfig_ebsKMSKeyARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  # Encrypted ebs block device

  ebs_block_device {
    device_name = "/dev/sdd"
    encrypted   = true
    kms_key_id  = aws_kms_key.test.arn
    volume_size = 12
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_rootBlockDeviceKMSKeyARN(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.nano"
  subnet_id     = aws_subnet.test.id

  root_block_device {
    delete_on_termination = true
    encrypted             = true
    kms_key_id            = aws_kms_key.test.arn
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsAttachedVolumeTags(rName string) string {
	// https://github.com/hashicorp/terraform-provider-aws/issues/17074
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name   = %[1]q
    Factum = "PerAsperaAdAstra"
  }
}

resource "aws_volume_attachment" "test" {
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsAttachedVolumeTagsUpdate(rName string) string {
	// https://github.com/hashicorp/terraform-provider-aws/issues/17074
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name   = %[1]q
    Factum = "VincitQuiSeVincit"
  }
}

resource "aws_volume_attachment" "test" {
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsRootTagsConflict(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11

    tags = {
      Name = "root-tag"
    }
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  volume_tags = {
    Name = "volume-tags"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsEBSTagsConflict(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9

    tags = {
      Name = "ebs-volume"
    }
  }

  volume_tags = {
    Name = "volume-tags"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsNoVolumeTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops        = 100
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 12
    encrypted   = true
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func mapToTagConfig(m map[string]string, indent int) string {
	if len(m) == 0 {
		return ""
	}

	var tags []string
	for k, v := range m {
		tags = append(tags, fmt.Sprintf("%q = %q", k, v))
	}

	return fmt.Sprintf("%s\n", strings.Join(tags, fmt.Sprintf("\n%s", strings.Repeat(" ", indent))))
}

func testAccInstanceConfig_blockDeviceTagsDefaultVolumeRBDEBS(defTg, volTg, rbdTg, ebsTg map[string]string) string {
	defTgCfg := ""
	if len(defTg) > 0 {
		//lintignore:AT004
		defTgCfg = fmt.Sprintf(`
provider "aws" {
  default_tags {
    tags = {
      %[1]s
    }
  }
}`, mapToTagConfig(defTg, 6))
	}

	volTgCfg := ""
	if len(volTg) > 0 {
		volTgCfg = fmt.Sprintf(`
  volume_tags = {
    %[1]s
  }`, mapToTagConfig(volTg, 4))
	}

	rbdTgCfg := ""
	if len(rbdTg) > 0 {
		rbdTgCfg = fmt.Sprintf(`
    tags = {
      %[1]s
    }`, mapToTagConfig(rbdTg, 6))
	}

	ebsTgCfg := ""
	if len(ebsTg) > 0 {
		ebsTgCfg = fmt.Sprintf(`
    tags = {
      %[1]s
    }`, mapToTagConfig(ebsTg, 6))
	}

	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
%[1]s

resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  %[2]s

  root_block_device {
    volume_type = "gp2"

    %[3]s
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 1

    %[4]s
  }
}
`, defTgCfg, volTgCfg, rbdTgCfg, ebsTgCfg))
}

func testAccInstanceConfig_blockDeviceTagsEBSTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 1

    tags = {
      Name = %[1]q
    }
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 1
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsEBSAndRootTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"

    tags = {
      Name    = %[1]q
      Purpose = "test"
    }
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 1

    tags = {
      Name = %[1]q
    }
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 1
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsEBSAndRootTagsUpdate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"

    tags = {
      Name = %[1]q
      Env  = "dev"
    }
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 1

    tags = {
      Name = %[1]q
    }
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 1
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

var testAccInstanceConfig_ebsBlockDeviceInvalidIOPS = acctest.ConfigCompose(testAccInstanceAMIWithEBSRootVolume, `
resource "aws_instance" "test" {
  ami = data.aws_ami.ami.id

  instance_type = "t2.medium"

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "gp2"
    iops        = 100
  }
}
`)

var testAccInstanceConfig_ebsBlockDeviceInvalidThroughput = acctest.ConfigCompose(testAccInstanceAMIWithEBSRootVolume, `
resource "aws_instance" "test" {
  ami = data.aws_ami.ami.id

  instance_type = "t2.medium"

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "gp2"
    throughput  = 300
  }
}
`)

func testAccInstanceConfig_ebsAndRootBlockDevice(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type           = "gp2"
    volume_size           = 9
    delete_on_termination = true
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsVolumeTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops        = 100
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 12
    encrypted   = true
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  volume_tags = {
    Name = "acceptance-test-volume-tag"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_blockDeviceTagsVolumeTagsUpdate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops        = 100
  }

  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 12
    encrypted   = true
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }

  volume_tags = {
    Name        = "acceptance-test-volume-tag"
    Environment = "dev"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_noProfile(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.small"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_profile(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

resource "aws_instance" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type        = "t2.small"
  iam_instance_profile = aws_iam_instance_profile.test.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_profilePath(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  path = "/test/"
  role = aws_iam_role.test.name
}

resource "aws_instance" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type        = "t2.small"
  iam_instance_profile = aws_iam_instance_profile.test.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_privateIP(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
  private_ip    = "10.1.1.42"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_emptyPrivateIP(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
  private_ip    = null

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_associatePublicIPAndPrivateIP(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  private_ip                  = "10.1.1.42"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstancePrivateDNSNameOptionsConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block                      = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id                          = aws_vpc.test.id
  availability_zone               = data.aws_availability_zones.available.names[2]
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  enable_resource_name_dns_aaaa_record_on_launch = true
  enable_resource_name_dns_a_record_on_launch    = true
  private_dns_hostname_type_on_launch            = "resource-name"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_PrivateDNSNameOptions_computed(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstancePrivateDNSNameOptionsConfig_base(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_PrivateDNSNameOptions_configured(rName string, enableAAAA, enableA bool, hostNameType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstancePrivateDNSNameOptionsConfig_base(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  private_dns_name_options {
    enable_resource_name_dns_aaaa_record = %[2]t
    enable_resource_name_dns_a_record    = %[3]t
    hostname_type                        = %[4]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, enableAAAA, enableA, hostNameType))
}

func testAccInstanceConfig_networkSecurityGroups(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true
  depends_on                  = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  instance   = aws_instance.test.id
  domain     = "vpc"
  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_networkVPCSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id
  depends_on             = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  instance   = aws_instance.test.id
  domain     = "vpc"
  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_networkVPCRemoveSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = []
  subnet_id              = aws_subnet.test.id
  depends_on             = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  instance   = aws_instance.test.id
  domain     = "vpc"
  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_keyPair(rName, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  key_name      = aws_key_pair.test.key_name

  tags = {
    Name = %[1]q
  }
}
`, rName, publicKey))
}

func testAccInstanceConfig_forceNewAndTagsDrift(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.nano"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_forceNewAndTagsDriftUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_primaryNetworkInterface(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.42"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = aws_network_interface.test.id
    device_index         = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_networkCardIndex(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.42"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = aws_network_interface.test.id
    device_index         = 0
    network_card_index   = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_primaryNetworkInterfaceSourceDestCheck(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_network_interface" "test" {
  subnet_id         = aws_subnet.test.id
  private_ips       = ["10.1.1.42"]
  source_dest_check = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = aws_network_interface.test.id
    device_index         = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_addSecondaryNetworkInterfaceBefore(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_network_interface" "primary" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.42"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "secondary" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.43"]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = aws_network_interface.primary.id
    device_index         = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_addSecondaryNetworkInterfaceAfter(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_network_interface" "primary" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.42"]

  tags = {
    Name = %[1]q
  }
}

# Attach previously created network interface, observe no state diff on instance resource
resource "aws_network_interface" "secondary" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["10.1.1.43"]

  tags = {
    Name = %[1]q
  }

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = aws_network_interface.primary.id
    device_index         = 0
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_addSecurityGroupBefore(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "%[1]s_1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
  name   = "%[1]s_2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  associate_public_ip_address = false

  vpc_security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.1.1.42"]
  security_groups = [aws_security_group.test.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_addSecurityGroupAfter(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "%[1]s_1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.test.id
  name   = "%[1]s_2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  associate_public_ip_address = false

  vpc_security_group_ids = [
    aws_security_group.test.id,
    aws_security_group.test2.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.test.id
  private_ips     = ["10.1.1.42"]
  security_groups = [aws_security_group.test.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_publicAndPrivateSecondaryIPs(rName string, isPublic bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.small"
  subnet_id     = aws_subnet.test.id

  associate_public_ip_address = %[2]t

  secondary_private_ips = ["10.1.1.42", "10.1.1.43"]

  vpc_security_group_ids = [
    aws_security_group.test.id
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName, isPublic))
}

func testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIPs string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.small"
  subnet_id     = aws_subnet.test.id

  private_ip            = %[2]q
  secondary_private_ips = [%[3]s]

  vpc_security_group_ids = [
    aws_security_group.test.id
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName, privateIP, secondaryIPs))
}

func testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIPs string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.small"
  subnet_id     = aws_subnet.test.id

  private_ip            = null
  secondary_private_ips = [%[2]s]

  vpc_security_group_ids = [
    aws_security_group.test.id
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName, secondaryIPs))
}

func testAccInstanceConfig_associatePublicDefaultPrivate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_associatePublicDefaultPublic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_associatePublicExplicitPublic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_associatePublicExplicitPrivate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_associatePublicOverridePublic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = true

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_associatePublicOverridePrivate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  associate_public_ip_address = false

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_getPasswordData(rName, publicKey string, val bool) string {
	return acctest.ConfigCompose(testAccLatestWindowsServer2016CoreAMIConfig(), fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.win2016core-ami.id
  instance_type = "t2.medium"
  key_name      = aws_key_pair.test.key_name

  get_password_data = %[3]t

  tags = {
    Name = %[1]q
  }
}
`, rName, publicKey, val))
}

func testAccInstanceConfig_cpuOptionsAmdSevSnpUnspecified(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0),
		testAccLatestAmazonLinux2023AMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("c6a.2xlarge", "m6a.2xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-linux-2023-ami.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_cpuOptionsAmdSevSnp(rName, amdSevSnp string) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0),
		testAccLatestAmazonLinux2023AMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("c6a.2xlarge", "m6a.2xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-linux-2023-ami.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  cpu_options {
    amd_sev_snp = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, amdSevSnp))
}

func testAccInstanceConfig_cpuOptionsAmdSevSnpCoreThreads(rName, amdSevSnp string, coreCount, threadsPerCore int) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0),
		testAccLatestAmazonLinux2023AMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("c6a.2xlarge", "m6a.2xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-linux-2023-ami.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  cpu_options {
    amd_sev_snp      = %[2]q
    core_count       = %[3]d
    threads_per_core = %[4]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, amdSevSnp, coreCount, threadsPerCore))
}

func testAccInstanceConfig_cpuOptionsUnspecified(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0),
		testAccLatestAmazonLinux2023AMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("c6a.2xlarge", "m6a.2xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-linux-2023-ami.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_cpuOptionsCoreThreads(rName string, coreCount, threadsPerCore int) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0),
		testAccLatestAmazonLinux2023AMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("c6a.2xlarge", "m6a.2xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-linux-2023-ami.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  cpu_options {
    core_count       = %[2]d
    threads_per_core = %[3]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, coreCount, threadsPerCore))
}

func testAccInstanceConfig_cpuOptionsCoreThreadsDeprecated(rName string, coreCount, threadsPerCore int) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0),
		testAccLatestAmazonLinux2023AMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("c6a.2xlarge", "m6a.2xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-linux-2023-ami.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  cpu_core_count       = %[2]d
  cpu_threads_per_core = %[3]d

  tags = {
    Name = %[1]q
  }
}
`, rName, coreCount, threadsPerCore))
}

func testAccInstanceConfig_creditSpecificationEmptyNonBurstable(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m5.large"
  subnet_id     = aws_subnet.test.id

  credit_specification {}

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationUnspecifiedNonBurstable(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "m5.large"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationUnspecified(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationUnspecifiedT3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  credit_specification {
    cpu_credits = "standard"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.test.id

  credit_specification {
    cpu_credits = "standard"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationUnlimitedCPUCredits(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  credit_specification {
    cpu_credits = "unlimited"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationUnlimitedCPUCreditsT3(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.test.id

  credit_specification {
    cpu_credits = "unlimited"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationIsNotAppliedToNonBurstable(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "c6i.large"
  subnet_id     = aws_subnet.test.id

  credit_specification {
    cpu_credits = "standard"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, instanceType string) string {
	var amiConfig, amiIDRef string

	if v, err := tfec2.ParseInstanceType(instanceType); err == nil && v.Type == "t4g" {
		amiConfig = acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI()
		amiIDRef = "data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id"
	} else {
		amiConfig = acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI()
		amiIDRef = "data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id"
	}

	return acctest.ConfigCompose(
		amiConfig,
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = %[3]s
  instance_type = %[2]q
  subnet_id     = aws_subnet.test.id

  credit_specification {}

  tags = {
    Name = %[1]q
  }
}
`, rName, instanceType, amiIDRef))
}

func testAccInstanceConfig_userDataUnspecified(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_userDataEmptyString(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
  user_data     = ""

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName string, userData string, replaceOnChange string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  user_data                   = %[2]q
  user_data_replace_on_change = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, userData, replaceOnChange))
}

func testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName string, userData string, replaceOnChange string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type               = "t2.micro"
  subnet_id                   = aws_subnet.test.id
  user_data_base64            = base64encode(%[2]q)
  user_data_replace_on_change = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, userData, replaceOnChange))
}

func testAccInstanceConfig_hibernation(rName string, hibernation bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

# must be >= m3 and have an encrypted root volume to enable hibernation
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  hibernation   = %[2]t
  instance_type = "m5.large"
  subnet_id     = aws_subnet.test.id

  root_block_device {
    encrypted   = true
    volume_size = 20
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, hibernation))
}

func testAccInstanceConfig_metadataOptionsDefaults(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  metadata_options {}
}
`, rName))
}

func testAccInstanceConfig_metadataOptionsDisabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  metadata_options {
    http_endpoint = "disabled"
  }
}
`, rName))
}

func testAccInstanceConfig_metadataOptionsUpdated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    instance_metadata_tags      = "enabled"
  }
}
`, rName))
}

func testAccInstanceConfig_metadataOptionsUpdatedAgain(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_tokens                 = "optional"
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "disabled"
  }
}
`, rName))
}

func testAccInstanceConfig_enclaveOptions(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("c5a.xlarge", "c5.xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  enclave_options {
    enabled = %[2]t
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, enabled))
}

func testAccInstanceConfig_dynamicEBSBlockDevices(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  instance_type = "t2.medium"

  tags = {
    Name = %[1]q
  }

  dynamic "ebs_block_device" {
    for_each = ["b", "c", "d"]
    iterator = device

    content {
      device_name = format("/dev/sd%%s", device.value)
      volume_size = "10"
      volume_type = "gp2"
    }
  }
}
`, rName))
}

func testAccInstanceConfig_capacityReservationSpecificationUnspecified(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_capacityReservationSpecificationPreference(rName, crPreference string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  capacity_reservation_specification {
    capacity_reservation_preference = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, crPreference))
}

func testAccInstanceConfig_capacityReservationSpecificationTargetID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  capacity_reservation_specification {
    capacity_reservation_target {
      capacity_reservation_id = aws_ec2_capacity_reservation.test.id
    }
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_capacity_reservation" "test" {
  instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
  instance_platform = %[2]q
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 10

  tags = {
    Name = %[1]q
  }
}
`, rName, awstypes.CapacityReservationInstancePlatformLinuxUnix))
}

func testAccInstanceConfig_templateBasic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_instance" "test" {
  launch_template {
    id = aws_launch_template.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateOverrideTemplate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegionNamed("micro", "t3.micro", "t2.micro", "t1.micro", "m1.small"),
		acctest.AvailableEC2InstanceTypeForRegionNamed("small", "t3.small", "t2.small", "t1.small", "m1.medium"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  instance_type = data.aws_ec2_instance_type_offering.micro.instance_type
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.small.instance_type

  launch_template {
    id = aws_launch_template.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateSpecificVersion(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_instance" "test" {
  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateModifyTemplate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.small", "t2.small", "t1.small", "m1.medium"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_instance" "test" {
  launch_template {
    id = aws_launch_template.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateUpdateVersion(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.small", "t2.small", "t1.small", "m1.medium"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  update_default_version = true
}

resource "aws_instance" "test" {
  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_instance" "test" {
  launch_template {
    name = aws_launch_template.test.name
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateWithIAMInstanceProfile(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  iam_instance_profile {
    name = aws_iam_instance_profile.test.name
  }
}

resource "aws_instance" "test" {
  launch_template {
    name = aws_launch_template.test.name
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateSpotAndStop(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  instance_market_options {
    market_type = "spot"

    spot_options {
      instance_interruption_behavior = "stop"
      spot_instance_type             = "persistent"
    }
  }
}

resource "aws_instance" "test" {
  launch_template {
    name = aws_launch_template.test.name
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_basicWithSpot(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSARM64AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t4g.nano"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn2-ami-minimal-hvm-ebs-arm64.id

  instance_market_options {
    market_type = "spot"
  }

  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_templateWithVPCSecurityGroups(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  iam_instance_profile {
    name = aws_iam_instance_profile.test.name
  }

  vpc_security_group_ids = [aws_security_group.test.id]
}

resource "aws_instance" "test" {
  subnet_id = aws_subnet.test.id

  launch_template {
    name = aws_launch_template.test.name
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
