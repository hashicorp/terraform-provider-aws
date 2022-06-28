package ec2_test

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(ec2.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"VolumeTypeNotAvailableInRegion",
		"Invalid value specified for Phase",
		"You have reached the maximum allowed number of license configurations created in one day",
		"specified zone does not support multi-attach-enabled volumes",
		"Unsupported volume type",
		"HostLimitExceeded",
		"ReservationCapacityExceeded",
		"InsufficientInstanceCapacity",
	)
}

func TestFetchRootDevice(t *testing.T) {
	cases := []struct {
		label  string
		images []*ec2.Image
		name   string
	}{
		{
			"device name in mappings",
			[]*ec2.Image{{
				ImageId:        aws.String("ami-123"),
				RootDeviceType: aws.String("ebs"),
				RootDeviceName: aws.String("/dev/xvda"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					{DeviceName: aws.String("/dev/xvdb")},
					{DeviceName: aws.String("/dev/xvda")},
				},
			}},
			"/dev/xvda",
		},
		{
			"device name not in mappings",
			[]*ec2.Image{{
				ImageId:        aws.String("ami-123"),
				RootDeviceType: aws.String("ebs"),
				RootDeviceName: aws.String("/dev/xvda"),
				BlockDeviceMappings: []*ec2.BlockDeviceMapping{
					{DeviceName: aws.String("/dev/xvdb")},
					{DeviceName: aws.String("/dev/xvdc")},
				},
			}},
			"/dev/xvdb",
		},
		{
			"no images",
			[]*ec2.Image{},
			"",
		},
	}

	sess, err := session.NewSession(nil)
	if err != nil {
		t.Errorf("Error new session: %s", err)
	}

	conn := ec2.New(sess)

	for _, tc := range cases {
		t.Run(fmt.Sprintf(tc.label), func(t *testing.T) {
			conn.Handlers.Clear()
			conn.Handlers.Send.PushBack(func(r *request.Request) {
				data := r.Data.(*ec2.DescribeImagesOutput)
				data.Images = tc.images
			})
			name, _ := tfec2.FetchRootDeviceName(conn, "ami-123")
			if tc.name != aws.StringValue(name) {
				t.Errorf("Expected name %s, got %s", tc.name, aws.StringValue(name))
			}
		})
	}
}

func TestParseInstanceType(t *testing.T) {
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
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		// No subnet_id specified requires default VPC with default subnets or EC2-Classic.
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckClassicOrHasDefaultVPCDefaultSubnets(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`instance/i-[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "instance_initiated_shutdown_behavior", "stop"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Required for EC2-Classic.
				ImportStateVerifyIgnore: []string{"source_dest_check", "user_data_replace_on_change"},
			},
		},
	})
}

func TestAccEC2Instance_disappears(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		// No subnet_id specified requires default VPC with default subnets or EC2-Classic.
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckClassicOrHasDefaultVPCDefaultSubnets(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Instance_tags(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		// No subnet_id specified requires default VPC with default subnets or EC2-Classic.
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckClassicOrHasDefaultVPCDefaultSubnets(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccInstanceConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2Instance_inDefaultVPCBySgName(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_inDefaultVPCBySgName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_inDefaultVPCBySGID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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

func TestAccEC2Instance_inEC2Classic(t *testing.T) {
	resourceName := "aws_instance.test"
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_classic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceClassicExists(resourceName, &v),
				),
			},
			{
				Config:            testAccInstanceConfig_classic(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"network_interface",
					"source_dest_check",
					"user_data_replace_on_change",
				},
			},
		},
	})
}

func TestAccEC2Instance_atLeastOneOtherEBSVolume(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_atLeastOneOtherEBSVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "0"), // This is an instance store AMI
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`instance/i-[a-z0-9]+`)),
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
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSBlockDevice_kmsKeyARN(t *testing.T) {
	var instance ec2.Instance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ebsKMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"encrypted": "true",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ebs_block_device.*.kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
func TestAccEC2Instance_EBSBlockDevice_invalidIopsForVolumeType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_ebsBlockDeviceInvalidIOPS,
				ExpectError: regexp.MustCompile(`error creating resource: iops attribute not supported for ebs_block_device with volume_type gp2`),
			},
		},
	})
}

func TestAccEC2Instance_EBSBlockDevice_invalidThroughputForVolumeType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_ebsBlockDeviceInvalidThroughput,
				ExpectError: regexp.MustCompile(`error creating resource: throughput attribute not supported for ebs_block_device with volume_type gp2`),
			},
		},
	})
}

// TestAccEC2Instance_EBSBlockDevice_RootBlockDevice_removed verifies block device mappings
// removed outside terraform no longer result in a panic.
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/20821
func TestAccEC2Instance_EBSBlockDevice_RootBlockDevice_removed(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ebsAndRootBlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					// Instance must be stopped before detaching a root block device
					testAccCheckStopInstance(&instance),
					testAccCheckDetachVolumes(&instance),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2Instance_RootBlockDevice_kmsKeyARN(t *testing.T) {
	var instance ec2.Instance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceKMSKeyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "root_block_device.0.kms_key_id", kmsKeyResourceName, "arn"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64Base64EncodedFile(rName, "test-fixtures/userdata-test.sh"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64Base64EncodedFile(rName, "test-fixtures/userdata-test.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64(rName, "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/xvda"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/xvda")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_gp2IOPSDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_gp2IOPSValue(rName),
				ExpectError: regexp.MustCompile(`error creating resource: iops attribute not supported for root_block_device with volume_type gp2`),
			},
		},
	})
}

func TestAccEC2Instance_blockDevices(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDevices(rName, rootVolumeSize),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "root_block_device.0.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", rootVolumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdb",
						"volume_size": "9",
						"volume_type": "gp2",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]*regexp.Regexp{
						"volume_id": regexp.MustCompile("vol-[a-z0-9]+"),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdc",
						"volume_size": "10",
						"volume_type": "io1",
						"iops":        "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdf",
						"volume_size": "10",
						"volume_type": "gp3",
						"throughput":  "300",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdg",
						"volume_size": "10",
						"volume_type": "gp3",
						"throughput":  "300",
						"iops":        "4000",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]*regexp.Regexp{
						"volume_id": regexp.MustCompile("vol-[a-z0-9]+"),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"device_name": "/dev/sdd",
						"encrypted":   "true",
						"volume_size": "12",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]*regexp.Regexp{
						"volume_id": regexp.MustCompile("vol-[a-z0-9]+"),
					}),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name":  "/dev/sde",
						"virtual_name": "ephemeral0",
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootStore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "0"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_noAMIEphemeralDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_block_device.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name": "/dev/sdb",
						"no_device":   "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ephemeral_block_device.*", map[string]string{
						"device_name": "/dev/sdc",
						"no_device":   "true",
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
	var v ec2.Instance
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_sourceDestDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
					testAccCheckInstanceExists(resourceName, &v),
					testCheck(true),
				),
			},
			{
				Config: testAccInstanceConfig_sourceDestDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheck(false),
				),
			},
		},
	})
}

func TestAccEC2Instance_autoRecovery(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_autoRecovery(rName, "default"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_options.0.auto_recovery", "disabled"),
				),
			},
		},
	})
}

func TestAccEC2Instance_disableAPIStop(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_disableAPIStop(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_stop", "true"),
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
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_stop", "false"),
				),
			},
		},
	})
}

func TestAccEC2Instance_disableAPITerminationFinalFalse(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_disableAPITermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", "true"),
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
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", "false"),
				),
			},
		},
	})
}

func TestAccEC2Instance_disableAPITerminationFinalTrue(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_disableAPITermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "disable_api_termination", "true"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dedicated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_outpost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, "arn"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_placementGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_placementPartitionNumber(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "placement_group", rName),
					resource.TestCheckResourceAttr(resourceName, "placement_partition_number", "3"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ipv6Support(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_ipv6Error(rName),
				ExpectError: regexp.MustCompile("Conflicting configuration arguments"),
			},
		},
	})
}

func TestAccEC2Instance_IPv6_supportAddressCountWithIPv4(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ipv6Supportv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ipv6_address_count", "1"),
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

func TestAccEC2Instance_networkInstanceSecurityGroups(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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

func TestAccEC2Instance_networkInstanceRemovingAllSecurityGroups(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkVPCSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_networkVPCRemoveSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
				ExpectError: regexp.MustCompile(`VPC-based instances require at least one security group to be attached`),
			},
		},
	})
}

func TestAccEC2Instance_networkInstanceVPCSecurityGroupIDs(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkVPCSecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
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

func TestAccEC2Instance_BlockDeviceTags_volumeTags(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDeviceTagsNoVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.Name", "acceptance-test-volume-tag"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsVolumeTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.Name", "acceptance-test-volume-tag"),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.Environment", "dev"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsNoVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "volume_tags.%", "0"),
				),
			},
		},
	})
}

func TestAccEC2Instance_BlockDeviceTags_withAttachedVolume(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	ebsVolumeName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDeviceTagsAttachedVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.%", "2"),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Name", rName),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Factum", "PerAsperaAdAstra"),
				),
			},
			{
				//https://github.com/hashicorp/terraform-provider-aws/issues/17074
				Config: testAccInstanceConfig_blockDeviceTagsAttachedVolumeTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.%", "2"),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Name", rName),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Factum", "PerAsperaAdAstra"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsAttachedVolumeTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.%", "2"),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Name", rName),
					resource.TestCheckResourceAttr(ebsVolumeName, "tags.Factum", "VincitQuiSeVincit"),
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

func TestAccEC2Instance_BlockDeviceTags_ebsAndRoot(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfig_blockDeviceTagsRootTagsConflict(rName),
				ExpectError: regexp.MustCompile(`"root_block_device\.0\.tags": conflicts with volume_tags`),
			},
			{
				Config:      testAccInstanceConfig_blockDeviceTagsEBSTagsConflict(rName),
				ExpectError: regexp.MustCompile(`"ebs_block_device\.0\.tags": conflicts with volume_tags`),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsEBSTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.0.tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.1.tags.%", "0"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsEBSAndRootTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.Purpose", "test"),
				),
			},
			{
				Config: testAccInstanceConfig_blockDeviceTagsEBSAndRootTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.tags.%", "2"),
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

func TestAccEC2Instance_instanceProfileChange(t *testing.T) {
	var v ec2.Instance
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_noProfile(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
					testAccCheckInstanceExists(resourceName, &v),
					testCheckInstanceProfile(),
				),
			},
			{
				Config: testAccInstanceConfig_profile(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(&v), // GH-8262: Error on EC2 instance role change when stopped
				),
			},
			{
				Config: testAccInstanceConfig_profile(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheckInstanceProfile(),
				),
			},
		},
	})
}

func TestAccEC2Instance_withIAMInstanceProfile(t *testing.T) {
	var v ec2.Instance
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_profile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
func TestAccEC2Instance_withIAMInstanceProfilePath(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_profilePath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
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
	var v ec2.Instance
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicIPAndPrivateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if aws.StringValue(v.PrivateIpAddress) == "" {
				return fmt.Errorf("bad computed private IP: %s", aws.StringValue(v.PrivateIpAddress))
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_emptyPrivateIP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_computed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", "true"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", "true"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_PrivateDNSNameOptions_configured(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_aaaa_record", "false"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.enable_resource_name_dns_a_record", "true"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_options.0.hostname_type", "ip-name"),
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

// Guard against regression with KeyPairs
// https://github.com/hashicorp/terraform/issues/2302
func TestAccEC2Instance_keyPairCheck(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	keyPairResourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_keyPair(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", keyPairResourceName, "key_name"),
				),
			},
		},
	})
}

func TestAccEC2Instance_rootBlockDeviceMismatch(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckRegion(t, endpoints.UsWest2RegionID) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceMismatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "13"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"root_block_device", "user_data_replace_on_change"},
			},
		},
	})
}

// This test reproduces the bug here:
//   https://github.com/hashicorp/terraform/issues/1752
//
// I wish there were a way to exercise resources built with helper.Schema in a
// unit context, in which case this test could be moved there, but for now this
// will cover the bugfix.
//
// The following triggers "diffs didn't match during apply" without the fix in to
// set NewRemoved on the .# field when it changes to 0.
func TestAccEC2Instance_forceNewAndTagsDrift(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_forceNewAndTagsDrift(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					driftTags(&v),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccInstanceConfig_forceNewAndTagsDriftUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var before ec2.Instance
	var after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_smallType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.medium"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change"},
			},
			{
				Config: testAccInstanceConfig_updateType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.large"),
				),
			},
		},
	})
}

func TestAccEC2Instance_changeInstanceTypeAndUserData(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	hash := sha1.Sum([]byte("hello world"))
	expectedUserData := hex.EncodeToString(hash[:])
	hash2 := sha1.Sum([]byte("new world"))
	expectedUserDataUpdated := hex.EncodeToString(hash2[:])

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_typeAndUserData(rName, "t2.medium", "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "user_data", expectedUserData),
				),
			},
			{
				Config: testAccInstanceConfig_typeAndUserData(rName, "t2.large", "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.large"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_typeAndUserDataBase64(rName, "t2.medium", "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.medium"),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
			{
				Config: testAccInstanceConfig_typeAndUserDataBase64(rName, "t2.large", "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.large"),
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
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_ebsRootDeviceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
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
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	deleteOnTermination := "true"
	volumeType := "gp2"

	originalSize := "30"
	updatedSize := "32"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, originalSize, deleteOnTermination, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", originalSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, updatedSize, deleteOnTermination, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", updatedSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifyType(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	deleteOnTermination := "true"

	originalType := "gp2"
	updatedType := "standard"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, deleteOnTermination, originalType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", originalType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, deleteOnTermination, updatedType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", updatedType),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceModifyIOPS_io1(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	deleteOnTermination := "true"
	volumeType := "io1"

	originalIOPS := "100"
	updatedIOPS := "200"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, deleteOnTermination, volumeType, originalIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", originalIOPS),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, deleteOnTermination, volumeType, updatedIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", updatedIOPS),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceModifyIOPS_io2(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	deleteOnTermination := "true"
	volumeType := "io2"

	originalIOPS := "100"
	updatedIOPS := "200"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, deleteOnTermination, volumeType, originalIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", originalIOPS),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, volumeSize, deleteOnTermination, volumeType, updatedIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", updatedIOPS),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceModifyThroughput_gp3(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	deleteOnTermination := "true"
	volumeType := "gp3"

	originalThroughput := "250"
	updatedThroughput := "300"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDeviceThroughput(rName, volumeSize, deleteOnTermination, volumeType, originalThroughput),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.throughput", originalThroughput),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceThroughput(rName, volumeSize, deleteOnTermination, volumeType, updatedThroughput),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", deleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.throughput", updatedThroughput),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifyDeleteOnTermination(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	volumeSize := "30"
	volumeType := "gp2"

	originalDeleteOnTermination := "false"
	updatedDeleteOnTermination := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, originalDeleteOnTermination, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", originalDeleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, volumeSize, updatedDeleteOnTermination, volumeType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", volumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", updatedDeleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", volumeType),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDevice_modifyAll(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	originalSize := "30"
	updatedSize := "32"

	originalType := "gp2"
	updatedType := "io1"

	updatedIOPS := "200"

	originalDeleteOnTermination := "false"
	updatedDeleteOnTermination := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_rootBlockDevice(rName, originalSize, originalDeleteOnTermination, originalType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", originalSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", originalDeleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", originalType),
				),
			},
			{
				Config: testAccInstanceConfig_rootBlockDeviceIOPS(rName, updatedSize, updatedDeleteOnTermination, updatedType, updatedIOPS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", updatedSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", updatedDeleteOnTermination),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_type", updatedType),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.iops", updatedIOPS),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceMultipleBlockDevices_modifySize(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	deleteOnTermination := "true"

	originalRootVolumeSize := "10"
	updatedRootVolumeSize := "14"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, originalRootVolumeSize, deleteOnTermination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", originalRootVolumeSize),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "12",
					}),
				),
			},
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, updatedRootVolumeSize, deleteOnTermination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", updatedRootVolumeSize),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "12",
					}),
				),
			},
		},
	})
}

func TestAccEC2Instance_EBSRootDeviceMultipleBlockDevices_modifyDeleteOnTermination(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	rootVolumeSize := "10"

	originalDeleteOnTermination := "false"
	updatedDeleteOnTermination := "true"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, rootVolumeSize, originalDeleteOnTermination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", rootVolumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", originalDeleteOnTermination),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "12",
					}),
				),
			},
			{
				Config: testAccInstanceConfig_blockDevicesDeleteOnTerminate(rName, rootVolumeSize, updatedDeleteOnTermination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", rootVolumeSize),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.delete_on_termination", updatedDeleteOnTermination),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "10",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"volume_size": "12",
					}),
				),
			},
		},
	})
}

// Test to validate fix for GH-ISSUE #1318 (dynamic ebs_block_devices forcing replacement after state refresh)
func TestAccEC2Instance_EBSRootDevice_multipleDynamicEBSBlockDevices(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_dynamicEBSBlockDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sdc",
						"encrypted":             "false",
						"iops":                  "100",
						"volume_size":           "10",
						"volume_type":           "gp2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sdb",
						"encrypted":             "false",
						"iops":                  "100",
						"volume_size":           "10",
						"volume_type":           "gp2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_block_device.*", map[string]string{
						"delete_on_termination": "true",
						"device_name":           "/dev/sda",
						"encrypted":             "false",
						"iops":                  "100",
						"volume_size":           "10",
						"volume_type":           "gp2",
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/xvda"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/xvda")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_gp3RootBlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.volume_size", "10"),
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
	var instance ec2.Instance
	var eni ec2.NetworkInterface
	resourceName := "aws_instance.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_primaryNetworkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					testAccCheckENIExists(eniResourceName, &eni),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "network_interface.*", map[string]string{
						"device_index":       "0",
						"network_card_index": "0",
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
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html#network-cards.
	// Only specialized (and expensive) instance types support multiple network cards (and hence network_card_index > 0).
	// Don't attempt to test with such instance types.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_networkCardIndex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "network_interface.*", map[string]string{
						"device_index":       "0",
						"network_card_index": "0",
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
	var instance ec2.Instance
	var eni ec2.NetworkInterface
	resourceName := "aws_instance.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_primaryNetworkInterfaceSourceDestCheck(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					testAccCheckENIExists(eniResourceName, &eni),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
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
	var before ec2.Instance
	var after ec2.Instance
	var eniPrimary ec2.NetworkInterface
	var eniSecondary ec2.NetworkInterface
	resourceName := "aws_instance.test"
	eniPrimaryResourceName := "aws_network_interface.primary"
	eniSecondaryResourceName := "aws_network_interface.secondary"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_addSecondaryNetworkInterfaceBefore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					testAccCheckENIExists(eniPrimaryResourceName, &eniPrimary),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckENIExists(eniSecondaryResourceName, &eniSecondary),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/3205
func TestAccEC2Instance_addSecurityGroupNetworkInterface(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_addSecurityGroupBefore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7063
func TestAccEC2Instance_NewNetworkInterface_publicIPAndSecondaryPrivateIPs(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_publicAndPrivateSecondaryIPs(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_publicAndPrivateSecondaryIPs(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "2"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	secondaryIPs := fmt.Sprintf("%q, %q", "10.1.1.42", "10.1.1.43")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "2"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	secondaryIP := fmt.Sprintf("%q", "10.1.1.42")
	secondaryIPs := fmt.Sprintf("%s, %q", secondaryIP, "10.1.1.43")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "0"),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPsNullPrivate(rName, secondaryIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "1"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateIP := "10.1.1.42"
	secondaryIPs := fmt.Sprintf("%q, %q", "10.1.1.43", "10.1.1.44")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "2"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateIP := "10.1.1.42"
	secondaryIP := fmt.Sprintf("%q", "10.1.1.43")
	secondaryIPs := fmt.Sprintf("%s, %q", secondaryIP, "10.1.1.44")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIPs),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "2"),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "0"),
				),
			},
			{
				Config: testAccInstanceConfig_privateIPAndSecondaryIPs(rName, privateIP, secondaryIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "private_ip", privateIP),
					resource.TestCheckResourceAttr(resourceName, "secondary_private_ips.#", "1"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicDefaultPrivate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicDefaultPublic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicExplicitPublic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicExplicitPrivate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicOverridePublic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublicOverridePrivate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	amiDataSourceName := "data.aws_ami.amzn-ami-minimal-hvm-ebs"
	instanceTypeDataSourceName := "data.aws_ec2_instance_type_offering.available"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", launchTemplateResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
					resource.TestCheckResourceAttrPair(resourceName, "ami", amiDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", instanceTypeDataSourceName, "instance_type"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_overrideTemplate(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	amiDataSourceName := "data.aws_ami.amzn-ami-minimal-hvm-ebs"
	instanceTypeDataSourceName := "data.aws_ec2_instance_type_offering.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateOverrideTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ami", amiDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", instanceTypeDataSourceName, "instance_type"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_setSpecificVersion(t *testing.T) {
	var v1, v2 ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
				),
			},
			{
				Config: testAccInstanceConfig_templateSpecificVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", launchTemplateResourceName, "default_version"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplateModifyTemplate_defaultVersion(t *testing.T) {
	var v1, v2 ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
				),
			},
			{
				Config: testAccInstanceConfig_templateModifyTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "launch_template.0.version", "$Default"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_updateTemplateVersion(t *testing.T) {
	var v1, v2 ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateSpecificVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", launchTemplateResourceName, "default_version"),
				),
			},
			{
				Config: testAccInstanceConfig_templateUpdateVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.version", launchTemplateResourceName, "default_version"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_swapIDAndName(t *testing.T) {
	var v1, v2 ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", launchTemplateResourceName, "name"),
				),
			},
			{
				Config: testAccInstanceConfig_templateName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v2),
					testAccCheckInstanceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.name", launchTemplateResourceName, "name"),
				),
			},
		},
	})
}

func TestAccEC2Instance_LaunchTemplate_spotAndStop(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	launchTemplateResourceName := "aws_launch_template.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_templateSpotAndStop(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "launch_template.0.id", launchTemplateResourceName, "id"),
				),
			},
		},
	})
}

func TestAccEC2Instance_GetPasswordData_falseToTrue(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(rName, publicKey, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "false"),
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
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccEC2Instance_GetPasswordData_trueToFalse(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(rName, publicKey, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "true"),
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
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_data", ""),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationEmpty_nonBurstable(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationEmptyNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecifiedNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
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
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_unspecifiedDefaultsToStandard(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_unlimitedCPUCredits(t *testing.T) {
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnknownCPUCredits_t2(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t3a.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnknownCPUCredits(rName, "t4g.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
	var first, second, third ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecification_isNotAppliedToNonBurstable(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationIsNotAppliedToNonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnspecifiedT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationT3_unlimitedCPUCredits(t *testing.T) {
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationT3_updateCPUCredits(t *testing.T) {
	var first, second, third ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCreditsT3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationStandardCPUCredits_t2Tot3Taint(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationStandardCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

func TestAccEC2Instance_CreditSpecificationUnlimitedCPUCredits_t2Tot3Taint(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecificationUnlimitedCPUCredits(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
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
					testAccCheckInstanceExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

func TestAccEC2Instance_UserData(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
				),
			},
			{
				Config: testAccInstanceConfig_userData(rName, "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
				),
			},
			{
				Config: testAccInstanceConfig_userDataBase64Encoded(rName, "new world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataEmptyString(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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
	var instance1, instance2 ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData1", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance1),
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
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData2", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance2),
					testAccCheckInstanceRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_On_Base64(t *testing.T) {
	var instance1, instance2 ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955d", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance1),
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
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955e", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance2),
					testAccCheckInstanceRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_Off(t *testing.T) {
	var instance1, instance2 ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData1", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance1),
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
				Config: testAccInstanceConfig_userDataSpecifiedReplaceFlag(rName, "TestData2", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance2),
					testAccCheckInstanceNotRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_UserDataReplaceOnChange_Off_Base64(t *testing.T) {
	var instance1, instance2 ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955d", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance1),
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
				Config: testAccInstanceConfig_userData64SpecifiedReplaceFlag(rName, "3dc39dda39be1205215e776bad998da361a5955e", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance2),
					testAccCheckInstanceNotRecreated(&instance1, &instance2),
				),
			},
		},
	})
}

func TestAccEC2Instance_hibernation(t *testing.T) {
	var v1, v2 ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_hibernation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "hibernation", "true"),
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
					testAccCheckInstanceExists(resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "hibernation", "false"),
				),
			},
		},
	})
}

func TestAccEC2Instance_metadataOptions(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "optional"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "disabled"),
				),
			},
			{
				Config: testAccInstanceConfig_metadataOptionsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.instance_metadata_tags", "enabled"),
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
	var v1, v2 ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "true"),
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
					testAccCheckInstanceExists(resourceName, &v2),
					testAccCheckInstanceRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservation_unspecifiedDefaultsToOpen(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "none"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
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
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationTargetID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "1"),
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
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "open"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
				),
			},
			{Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(&original), // Stop instance to modify capacity reservation
				),
			},
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "none"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2Instance_CapacityReservation_modifyTarget(t *testing.T) {
	var original ec2.Instance
	var updated ec2.Instance
	resourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &original),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_preference", "none"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "0"),
				),
			},
			{Config: testAccInstanceConfig_capacityReservationSpecificationPreference(rName, "none"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStopInstance(&original), // Stop instance to modify capacity reservation
				),
			},
			{
				Config: testAccInstanceConfig_capacityReservationSpecificationTargetID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &updated),
					testAccCheckInstanceNotRecreated(&original, &updated),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id"),
					resource.TestCheckResourceAttr(resourceName, "capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_resource_group_arn", ""),
				),
			},
		},
	})
}

func testAccCheckInstanceNotRecreated(before, after *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.InstanceId), aws.StringValue(after.InstanceId); before != after {
			return fmt.Errorf("EC2 Instance (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccCheckInstanceRecreated(before, after *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.InstanceId), aws.StringValue(after.InstanceId); before == after {
			return fmt.Errorf("EC2 Instance (%s) not recreated", before)
		}

		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	return testAccCheckInstanceDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckInstanceDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_instance" {
			continue
		}

		_, err := tfec2.FindInstanceByID(conn, rs.Primary.ID)

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

func testAccCheckInstanceExists(n string, v *ec2.Instance) resource.TestCheckFunc {
	return testAccCheckInstanceExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckInstanceClassicExists(n string, v *ec2.Instance) resource.TestCheckFunc {
	return testAccCheckInstanceExistsWithProvider(n, v, func() *schema.Provider { return acctest.ProviderEC2Classic })
}

func testAccCheckInstanceExistsWithProvider(n string, v *ec2.Instance, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance ID is set")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindInstanceByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStopInstance(v *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		err := tfec2.StopInstance(conn, aws.StringValue(v.InstanceId), 10*time.Minute)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckDetachVolumes(instance *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*conns.AWSClient)
		conn := client.EC2Conn

		for _, v := range instance.BlockDeviceMappings {
			if v.Ebs != nil && v.Ebs.VolumeId != nil {
				deviceName := aws.StringValue(v.DeviceName)
				instanceID := aws.StringValue(instance.InstanceId)
				volumeID := aws.StringValue(v.Ebs.VolumeId)

				// Make sure in correct state before detaching.
				if _, err := tfec2.WaitVolumeAttachmentCreated(conn, volumeID, instanceID, deviceName, 5*time.Minute); err != nil {
					return err
				}

				r := tfec2.ResourceVolumeAttachment()
				d := r.Data(nil)
				d.Set("device_name", deviceName)
				d.Set("instance_id", instanceID)
				d.Set("volume_id", volumeID)

				if err := r.Delete(d, client); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func TestInstanceHostIDSchema(t *testing.T) {
	actualSchema := tfec2.ResourceInstance().Schema["host_id"]
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
	actualSchema := tfec2.ResourceInstance().Schema["cpu_core_count"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeInt,
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

func TestInstanceCPUThreadsPerCoreSchema(t *testing.T) {
	actualSchema := tfec2.ResourceInstance().Schema["cpu_threads_per_core"]
	expectedSchema := &schema.Schema{
		Type:     schema.TypeInt,
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

func driftTags(instance *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		_, err := conn.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{instance.InstanceId},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Drift"),
					Value: aws.String("Happens"),
				},
			},
		})
		return err
	}
}

// testAccPreCheckClassicOrHasDefaultVPCDefaultSubnets checks that the test region has either
// - The EC2-Classic platform available, or
// - A default VPC with default subnets.
// This check is useful to ensure that an instance can be launched without specifying a subnet.
func testAccPreCheckClassicOrHasDefaultVPCDefaultSubnets(t *testing.T) {
	client := acctest.Provider.Meta().(*conns.AWSClient)

	if !conns.HasEC2Classic(client.SupportedPlatforms) && !(hasDefaultVPC(t) && defaultSubnetCount(t) > 0) {
		t.Skipf("skipping tests; %s does not have EC2-Classic or a default VPC with default subnets", client.Region)
	}
}

// defaultVPC returns the ID of the default VPC for the current AWS Region, or "" if none exists.
func defaultVPC(t *testing.T) string {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	output, err := conn.DescribeAccountAttributes(&ec2.DescribeAccountAttributesInput{
		AttributeNames: aws.StringSlice([]string{ec2.AccountAttributeNameDefaultVpc}),
	})

	if acctest.PreCheckSkipError(err) {
		return ""
	}

	if err != nil {
		t.Fatalf("error describing EC2 account attributes: %s", err)
	}

	if len(output.AccountAttributes) > 0 && len(output.AccountAttributes[0].AttributeValues) > 0 {
		if v := aws.StringValue(output.AccountAttributes[0].AttributeValues[0].AttributeValue); v != "none" {
			return v
		}
	}

	return ""
}

func hasDefaultVPC(t *testing.T) bool {
	return defaultVPC(t) != ""
}

// defaultSubnetCount returns the number of default subnets in the current region's default VPC.
func defaultSubnetCount(t *testing.T) int {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSubnetsInput{
		Filters: tfec2.BuildAttributeFilterList(
			map[string]string{
				"defaultForAz": "true",
			},
		),
	}

	subnets, err := tfec2.FindSubnets(conn, input)

	if acctest.PreCheckSkipError(err) {
		return 0
	}

	if err != nil {
		t.Fatalf("error listing default subnets: %s", err)
	}

	return len(subnets)
}

// testAccLatestAmazonLinuxPVEBSAMIConfig returns the configuration for a data source that
// describes the latest Amazon Linux AMI using PV virtualization and an EBS root device.
// The data source is named 'amzn-ami-minimal-pv-ebs'.
func testAccLatestAmazonLinuxPVEBSAMIConfig() string {
	return `
data "aws_ami" "amzn-ami-minimal-pv-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}
`
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
//   1) a VPC without IPv6 support
//   2) a subnet in the VPC that optionally assigns public IP addresses to ENIs
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
//   1) a VPC security group
//   2) an internet gateway in the VPC
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
//   1) a VPC with IPv6 support
//   2) a subnet in the VPC with an assigned IPv6 CIDR block
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-classic-platform.html#ec2-classic-instance-types
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  # Explicitly no tags so as to test creation without tags.
}
`)
}

func testAccInstanceConfig_tags1(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccInstanceConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccInstanceConfig_inDefaultVPCBySgName(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = data.aws_vpc.default.id
}

resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = data.aws_vpc.default.id
}

resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  availability_zone      = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_classic() string { // nosempgrep:ec2-in-func-name
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t1.micro", "m3.medium", "m3.large", "c3.large", "r3.large"),
		`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}
`)
}

func testAccInstanceConfig_atLeastOneOtherEBSVolume(rName string) string {
	return acctest.ConfigCompose(
		testAccAMIDataSourceConfig_latestAmazonLinuxHVMInstanceStore(),
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
  ami = data.aws_ami.amzn-ami-minimal-hvm-instance-store.id

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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  subnet_id = aws_subnet.test.id

  instance_type    = "t2.small"
  user_data_base64 = filebase64(%[2]q)

  tags = {
    Name = %[1]q
  }
}
`, rName, filename))
}

func testAccInstanceConfig_smallType(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  subnet_id = aws_subnet.test.id

  instance_type = "t2.medium"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_updateType(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  subnet_id = aws_subnet.test.id

  instance_type = "t2.large"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_typeAndUserData(rName, instanceType, userData string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami       = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(testAccAMIDataSourceConfig_latestAmazonLinuxHVMInstanceStore(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-instance-store.id

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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami              = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		// Prevent frequent errors like
		//	"InsufficientInstanceCapacity: We currently do not have sufficient m1.small capacity in the Availability Zone you requested (us-west-2a)."
		testAccInstanceVPCConfig(rName, false, 1),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

  # tflint-ignore: aws_instance_previous_type
  instance_type               = "m1.small"
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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name            = %[1]q
  strategy        = "partition"
  partition_count = 4
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCIPv6Config(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccInstanceConfig_ebsKMSKeyARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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

func testAccInstanceConfig_blockDeviceTagsEBSTags(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

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
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_profile(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
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
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = "t2.small"
  iam_instance_profile = aws_iam_instance_profile.test.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_profilePath(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
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
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccInstancePrivateDNSNameOptionsBaseConfig(rName string) string {
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstancePrivateDNSNameOptionsBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_PrivateDNSNameOptions_configured(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstancePrivateDNSNameOptionsBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  private_dns_name_options {
    enable_resource_name_dns_aaaa_record = false
    enable_resource_name_dns_a_record    = true
    hostname_type                        = "ip-name"
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_networkSecurityGroups(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
  vpc        = true
  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_networkVPCSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
  vpc        = true
  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_networkVPCRemoveSecurityGroupIDs(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		testAccInstanceVPCSecurityGroupConfig(rName),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
  vpc        = true
  depends_on = [aws_internet_gateway.test]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_keyPair(rName, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  key_name      = aws_key_pair.test.key_name

  tags = {
    Name = %[1]q
  }
}
`, rName, publicKey))
}

func testAccInstanceConfig_rootBlockDeviceMismatch(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceVPCConfig(rName, false, 0), fmt.Sprintf(`
resource "aws_instance" "test" {
  # This is an AMI in UsWest2 with RootDeviceName: "/dev/sda1"; actual root: "/dev/sda"
  ami = "ami-ef5b69df"

  # tflint-ignore: aws_instance_previous_type
  instance_type = "t1.micro"
  subnet_id     = aws_subnet.test.id

  root_block_device {
    volume_size = 13
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)) //lintignore:AWSAT002
}

func testAccInstanceConfig_forceNewAndTagsDrift(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  vpc_id      = aws_vpc.test.id
  description = "%[1]s_1"
  name        = "%[1]s_1"
}

resource "aws_security_group" "test2" {
  vpc_id      = aws_vpc.test.id
  description = "%[1]s_2"
  name        = "%[1]s_2"
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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
  vpc_id      = aws_vpc.test.id
  description = "%[1]s_1"
  name        = "%[1]s_1"
}

resource "aws_security_group" "test2" {
  vpc_id      = aws_vpc.test.id
  description = "%[1]s_2"
  name        = "%[1]s_2"
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = %[1]q
  name        = %[1]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = %[1]q
  name        = %[1]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  vpc_id      = aws_vpc.test.id
  description = %[1]q
  name        = %[1]q
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, true, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccInstanceConfig_creditSpecificationEmptyNonBurstable(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
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

# must be >= m3 and have an encrypted root volume to eanble hibernation
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccInstanceConfig_metadataOptions(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
    instance_metadata_tags      = "enabled"
  }
}
`, rName))
}

func testAccInstanceConfig_enclaveOptions(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		testAccInstanceVPCConfig(rName, false, 0),
		acctest.AvailableEC2InstanceTypeForRegion("c5a.xlarge", "c5.xlarge"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
	return acctest.ConfigCompose(testAccLatestAmazonLinuxPVEBSAMIConfig(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami = data.aws_ami.amzn-ami-minimal-pv-ebs.id
  # tflint-ignore: aws_instance_previous_type
  instance_type = "m3.medium"

  tags = {
    Name = %[1]q
  }

  dynamic "ebs_block_device" {
    for_each = ["a", "b", "c"]
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccInstanceConfig_capacityReservationSpecificationPreference(rName, crPreference string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForRegion("t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
`, rName, ec2.CapacityReservationInstancePlatformLinuxUnix))
}

func testAccInstanceConfig_templateBasic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegionNamed("micro", "t3.micro", "t2.micro", "t1.micro", "m1.small"),
		acctest.AvailableEC2InstanceTypeForRegionNamed("small", "t3.small", "t2.small", "t1.small", "m1.medium"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  instance_type = data.aws_ec2_instance_type_offering.micro.instance_type
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.small", "t2.small", "t1.small", "m1.medium"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.small", "t2.small", "t1.small", "m1.medium"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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

func testAccInstanceConfig_templateSpotAndStop(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro", "t1.micro", "m1.small"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
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
