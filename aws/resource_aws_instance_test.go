package aws

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_instance", &resource.Sweeper{
		Name: "aws_instance",
		F:    testSweepInstances,
	})
}

func testSweepInstances(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeInstancesPages(&ec2.DescribeInstancesInput{}, func(page *ec2.DescribeInstancesOutput, isLast bool) bool {
		if len(page.Reservations) == 0 {
			log.Print("[DEBUG] No EC2 Instances to sweep")
			return false
		}

		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				id := aws.StringValue(instance.InstanceId)

				if instance.State != nil && aws.StringValue(instance.State.Name) == ec2.InstanceStateNameTerminated {
					log.Printf("[INFO] Skipping terminated EC2 Instance: %s", id)
					continue
				}

				log.Printf("[INFO] Terminating EC2 Instance: %s", id)
				err := awsTerminateInstance(conn, id, 5*time.Minute)

				if isAWSErr(err, "OperationNotPermitted", "Modify its 'disableApiTermination' instance attribute and try again.") {
					log.Printf("[INFO] Enabling API Termination on EC2 Instance: %s", id)

					input := &ec2.ModifyInstanceAttributeInput{
						InstanceId: instance.InstanceId,
						DisableApiTermination: &ec2.AttributeBooleanValue{
							Value: aws.Bool(false),
						},
					}

					_, err = conn.ModifyInstanceAttribute(input)

					if err == nil {
						err = awsTerminateInstance(conn, id, 5*time.Minute)
					}
				}

				if err != nil {
					log.Printf("[ERROR] Error terminating EC2 Instance (%s): %s", id, err)
				}
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving EC2 Instances: %s", err)
	}

	return nil
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
			name, _ := fetchRootDeviceName("ami-123", conn)
			if tc.name != aws.StringValue(name) {
				t.Errorf("Expected name %s, got %s", tc.name, aws.StringValue(name))
			}
		})
	}
}

func TestAccAWSInstance_inDefaultVpcBySgName(t *testing.T) {
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInDefaultVpcBySgName(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
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

func TestAccAWSInstance_inDefaultVpcBySgId(t *testing.T) {
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()
	var v ec2.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInDefaultVpcBySgId(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
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

func TestAccAWSInstance_inEc2Classic(t *testing.T) {
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()
	var v ec2.Instance

	// EC2 Classic enabled
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigInEc2Classic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
				),
			},
			{
				Config:                  testAccInstanceConfigInEc2Classic(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface", "source_dest_check"},
			},
		},
	})
}

func TestAccAWSInstance_basic(t *testing.T) {
	var v ec2.Instance
	var vol *ec2.Volume
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	testCheck := func(rInt int) func(*terraform.State) error {
		return func(*terraform.State) error {
			if *v.Placement.AvailabilityZone != "us-west-2a" {
				return fmt.Errorf("bad availability zone: %#v", *v.Placement.AvailabilityZone)
			}

			if len(v.SecurityGroups) == 0 {
				return fmt.Errorf("no security groups: %#v", v.SecurityGroups)
			}
			if *v.SecurityGroups[0].GroupName != fmt.Sprintf("tf_test_%d", rInt) {
				return fmt.Errorf("no security groups: %#v", v.SecurityGroups)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			// Create a volume to cover #1249
			{
				// Need a resource in this config so the provisioner will be available
				Config: testAccInstanceConfig_pre(rInt),
				Check: func(*terraform.State) error {
					conn := testAccProvider.Meta().(*AWSClient).ec2conn
					var err error
					vol, err = conn.CreateVolume(&ec2.CreateVolumeInput{
						AvailabilityZone: aws.String("us-west-2a"),
						Size:             aws.Int64(int64(5)),
					})
					return err
				},
			},
			{
				Config: testAccInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					testCheck(rInt),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data",
						"3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.#", "0"),
					resource.TestMatchResourceAttr(
						resourceName,
						"arn",
						regexp.MustCompile(`^arn:[^:]+:ec2:[^:]+:\d{12}:instance/i-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// We repeat the exact same test so that we can be sure
			// that the user data hash stuff is working without generating
			// an incorrect diff.
			{
				Config: testAccInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					testCheck(rInt),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data",
						"3dc39dda39be1205215e776bad998da361a5955d"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.#", "0"),
				),
			},
			// Clean up volume created above
			{
				Config: testAccInstanceConfig(rInt),
				Check: func(*terraform.State) error {
					conn := testAccProvider.Meta().(*AWSClient).ec2conn
					_, err := conn.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: vol.VolumeId})
					return err
				},
			},
		},
	})
}

func TestAccAWSInstance_EbsBlockDevice_KmsKeyArn(t *testing.T) {
	var instance ec2.Instance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigEbsBlockDeviceKmsKeyArn,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_block_device.2634515331.encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "ebs_block_device.2634515331.kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSInstance_RootBlockDevice_KmsKeyArn(t *testing.T) {
	var instance ec2.Instance
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigRootBlockDeviceKmsKeyArn,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_block_device.0.encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "root_block_device.0.kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccAWSInstance_userDataBase64(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithUserDataBase64(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data_base64",
						"aGVsbG8gd29ybGQ="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data"},
			},
		},
	})
}

func TestAccAWSInstance_GP2IopsDevice(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {

			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/sda1"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sda1")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"ephemeral_block_device", "user_data"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGP2IopsDevice,
				//Config: testAccInstanceConfigBlockDevices,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.iops", "100"),
					testCheck(),
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

func TestAccAWSInstance_GP2WithIopsValue(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"ephemeral_block_device", "user_data"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGP2WithIopsValue,
				Check:  testAccCheckInstanceExists(resourceName, &v),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccInstanceGP2WithIopsValue,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSInstance_blockDevices(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {

			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/sda1"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sda1")
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

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"ephemeral_block_device"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigBlockDevices,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.#", "1"),
					resource.TestMatchResourceAttr(
						resourceName, "root_block_device.0.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.#", "3"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2576023345.device_name", "/dev/sdb"),
					resource.TestMatchResourceAttr(
						resourceName, "ebs_block_device.2576023345.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2576023345.volume_size", "9"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2576023345.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2554893574.device_name", "/dev/sdc"),
					resource.TestMatchResourceAttr(
						resourceName, "ebs_block_device.2554893574.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2554893574.volume_size", "10"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2554893574.volume_type", "io1"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2554893574.iops", "100"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2634515331.device_name", "/dev/sdd"),
					resource.TestMatchResourceAttr(
						resourceName, "ebs_block_device.2634515331.volume_id", regexp.MustCompile("vol-[a-z0-9]+")),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2634515331.encrypted", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.2634515331.volume_size", "12"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.1692014856.device_name", "/dev/sde"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.1692014856.virtual_name", "ephemeral0"),
					testCheck(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device"},
			},
		},
	})
}

func TestAccAWSInstance_rootInstanceStore(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "aws_instance" "test" {
						# us-west-2
						# Amazon Linux HVM Instance Store 64-bit (2016.09.0)
						# https://aws.amazon.com/amazon-linux-ami
						ami = "ami-44c36524"

						# Only certain instance types support ephemeral root instance stores.
						# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html
						instance_type = "m3.medium"
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ami", "ami-44c36524"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "instance_type", "m3.medium"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.#", "0"),
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

func TestAccAWSInstance_noAMIEphemeralDevices(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	testCheck := func() resource.TestCheckFunc {
		return func(*terraform.State) error {

			// Map out the block devices by name, which should be unique.
			blockDevices := make(map[string]*ec2.InstanceBlockDeviceMapping)
			for _, blockDevice := range v.BlockDeviceMappings {
				blockDevices[*blockDevice.DeviceName] = blockDevice
			}

			// Check if the root block device exists.
			if _, ok := blockDevices["/dev/sda1"]; !ok {
				return fmt.Errorf("block device doesn't exist: /dev/sda1")
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
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"ephemeral_block_device"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "aws_instance" "test" {
						# us-west-2
						ami = "ami-01f05461"  // This AMI (Ubuntu) contains two ephemerals

						instance_type = "c3.large"

						root_block_device {
							volume_type = "gp2"
							volume_size = 11
						}
						ephemeral_block_device {
							device_name = "/dev/sdb"
							no_device = true
						}
						ephemeral_block_device {
							device_name = "/dev/sdc"
							no_device = true
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "ami", "ami-01f05461"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(
						resourceName, "instance_type", "c3.large"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr(
						resourceName, "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.172787947.device_name", "/dev/sdb"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.172787947.no_device", "true"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.3336996981.device_name", "/dev/sdc"),
					resource.TestCheckResourceAttr(
						resourceName, "ephemeral_block_device.3336996981.no_device", "true"),
					testCheck(),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device"},
			},
		},
	})
}

func TestAccAWSInstance_sourceDestCheck(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

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
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigSourceDestDisable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheck(false),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfigSourceDestEnable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheck(true),
				),
			},
			{
				Config: testAccInstanceConfigSourceDestDisable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheck(false),
				),
			},
		},
	})
}

func TestAccAWSInstance_disableApiTermination(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	checkDisableApiTermination := func(expected bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			conn := testAccProvider.Meta().(*AWSClient).ec2conn
			r, err := conn.DescribeInstanceAttribute(&ec2.DescribeInstanceAttributeInput{
				InstanceId: v.InstanceId,
				Attribute:  aws.String("disableApiTermination"),
			})
			if err != nil {
				return err
			}
			got := *r.DisableApiTermination.Value
			if got != expected {
				return fmt.Errorf("expected: %t, got: %t", expected, got)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigDisableAPITermination(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					checkDisableApiTermination(true),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfigDisableAPITermination(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					checkDisableApiTermination(false),
				),
			},
		},
	})
}

func TestAccAWSInstance_vpc(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName,
						"user_data",
						"562a3e32810edf6ff09994f050f12e799452379d"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address", "user_data"},
			},
		},
	})
}

func TestAccAWSInstance_placementGroup(t *testing.T) {
	var v ec2.Instance
	rStr := acctest.RandString(5)
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPlacementGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName,
						"placement_group",
						fmt.Sprintf("testAccInstanceConfigPlacementGroup_%s", rStr)),
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

func TestAccAWSInstance_ipv6_supportAddressCount(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigIpv6Support,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName,
						"ipv6_address_count",
						"1"),
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

func TestAccAWSInstance_ipv6AddressCountAndSingleAddressCausesError(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccInstanceConfigIpv6ErrorConfig,
				ExpectError: regexp.MustCompile("Only 1 of `ipv6_address_count` or `ipv6_addresses` can be specified"),
			},
		},
	})
}

func TestAccAWSInstance_ipv6_supportAddressCountWithIpv4(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigIpv6SupportWithIpv4,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName,
						"ipv6_address_count",
						"1"),
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

func TestAccAWSInstance_multipleRegions(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckInstanceDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigMultipleRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExistsWithProvider(resourceName, &v,
						testAccAwsRegionProviderFunc("us-west-2", &providers)),
					testAccCheckInstanceExistsWithProvider("aws_instance.test2", &v,
						testAccAwsRegionProviderFunc("us-east-1", &providers)),
				),
			},
		},
	})
}

func TestAccAWSInstance_NetworkInstanceSecurityGroups(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceNetworkInstanceSecurityGroups(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
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

func TestAccAWSInstance_NetworkInstanceRemovingAllSecurityGroups(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceNetworkInstanceVPCSecurityGroupIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceNetworkInstanceVPCRemoveSecurityGroupIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "vpc_security_group_ids.#", "1"),
				),
				ExpectError: regexp.MustCompile(`VPC-based instances require at least one security group to be attached`),
			},
		},
	})
}

func TestAccAWSInstance_NetworkInstanceVPCSecurityGroupIDs(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceNetworkInstanceVPCSecurityGroupIDs(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "security_groups.#", "0"),
					resource.TestCheckResourceAttr(
						resourceName, "vpc_security_group_ids.#", "1"),
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

func TestAccAWSInstance_tags(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckTags(&v.Tags, "test", "test2"),
					// Guard against regression of https://github.com/hashicorp/terraform/issues/914
					testAccCheckTags(&v.Tags, "#", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckInstanceConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testAccCheckTags(&v.Tags, "test", ""),
					testAccCheckTags(&v.Tags, "test2", "test3"),
				),
			},
		},
	})
}

func TestAccAWSInstance_volumeTags(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigNoVolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckNoResourceAttr(
						resourceName, "volume_tags"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ephemeral_block_device"},
			},
			{
				Config: testAccCheckInstanceConfigWithVolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "volume_tags.%", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "volume_tags.Name", "acceptance-test-volume-tag"),
				),
			},
			{
				Config: testAccCheckInstanceConfigWithVolumeTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "volume_tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "volume_tags.Name", "acceptance-test-volume-tag"),
					resource.TestCheckResourceAttr(
						resourceName, "volume_tags.Environment", "dev"),
				),
			},
			{
				Config: testAccCheckInstanceConfigNoVolumeTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckNoResourceAttr(
						resourceName, "volume_tags"),
				),
			},
		},
	})
}

func TestAccAWSInstance_volumeTagsComputed(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckInstanceConfigWithAttachedVolume,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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

func TestAccAWSInstance_instanceProfileChange(t *testing.T) {
	var v ec2.Instance
	rName := acctest.RandString(5)
	resourceName := "aws_instance.test"

	testCheckInstanceProfile := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.IamInstanceProfile == nil {
				return fmt.Errorf("Instance Profile is nil - we expected an InstanceProfile associated with the Instance")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithoutInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfigWithInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheckInstanceProfile(),
				),
			},
		},
	})
}

func TestAccAWSInstance_withIamInstanceProfile(t *testing.T) {
	var v ec2.Instance
	rName := acctest.RandString(5)
	resourceName := "aws_instance.test"

	testCheckInstanceProfile := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.IamInstanceProfile == nil {
				return fmt.Errorf("Instance Profile is nil - we expected an InstanceProfile associated with the Instance")
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheckInstanceProfile(),
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

func TestAccAWSInstance_privateIP(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *v.PrivateIpAddress != "10.1.1.42" {
				return fmt.Errorf("bad private IP: %s", *v.PrivateIpAddress)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPrivateIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheckPrivateIP(),
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

func TestAccAWSInstance_associatePublicIPAndPrivateIP(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	testCheckPrivateIP := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *v.PrivateIpAddress != "10.1.1.42" {
				return fmt.Errorf("bad private IP: %s", *v.PrivateIpAddress)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"associate_public_ip_address"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigAssociatePublicIPAndPrivateIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheckPrivateIP(),
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

// Guard against regression with KeyPairs
// https://github.com/hashicorp/terraform/issues/2302
func TestAccAWSInstance_keyPairCheck(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	testCheckKeyPair := func(keyName string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if v.KeyName == nil {
				return fmt.Errorf("No Key Pair found, expected(%s)", keyName)
			}
			if v.KeyName != nil && *v.KeyName != keyName {
				return fmt.Errorf("Bad key name, expected (%s), got (%s)", keyName, *v.KeyName)
			}

			return nil
		}
	}

	keyPairName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   resourceName,
		IDRefreshIgnore: []string{"source_dest_check"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigKeyPair(keyPairName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					testCheckKeyPair(keyPairName),
				),
			},
		},
	})
}

func TestAccAWSInstance_rootBlockDeviceMismatch(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigRootBlockDeviceMismatch,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "root_block_device.0.volume_size", "13"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"root_block_device"},
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
func TestAccAWSInstance_forceNewAndTagsDrift(t *testing.T) {
	var v ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigForceNewAndTagsDrift,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
					driftTags(&v),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccInstanceConfigForceNewAndTagsDrift_Update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &v),
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

func TestAccAWSInstance_changeInstanceType(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithSmallInstanceType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfigUpdateInstanceType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(
						t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSInstance_primaryNetworkInterface(t *testing.T) {
	var instance ec2.Instance
	var ini ec2.NetworkInterface
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPrimaryNetworkInterface,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					testAccCheckAWSENIExists("aws_network_interface.test", &ini),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface"},
			},
		},
	})
}

func TestAccAWSInstance_primaryNetworkInterfaceSourceDestCheck(t *testing.T) {
	var instance ec2.Instance
	var ini ec2.NetworkInterface
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigPrimaryNetworkInterfaceSourceDestCheck,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					testAccCheckAWSENIExists("aws_network_interface.test", &ini),
					resource.TestCheckResourceAttr(resourceName, "source_dest_check", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface"},
			},
		},
	})
}

func TestAccAWSInstance_addSecondaryInterface(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	var iniPrimary ec2.NetworkInterface
	var iniSecondary ec2.NetworkInterface
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigAddSecondaryNetworkInterfaceBefore,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					testAccCheckAWSENIExists("aws_network_interface.primary", &iniPrimary),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"network_interface"},
			},
			{
				Config: testAccInstanceConfigAddSecondaryNetworkInterfaceAfter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckAWSENIExists("aws_network_interface.secondary", &iniSecondary),
					resource.TestCheckResourceAttr(resourceName, "network_interface.#", "1"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/3205
func TestAccAWSInstance_addSecurityGroupNetworkInterface(t *testing.T) {
	var before ec2.Instance
	var after ec2.Instance
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigAddSecurityGroupBefore,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfigAddSecurityGroupAfter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "2"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_defaultPrivate(t *testing.T) {
	var before ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_defaultPrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
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

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_defaultPublic(t *testing.T) {
	var before ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_defaultPublic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
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

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_explicitPublic(t *testing.T) {
	var before ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_explicitPublic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
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

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_explicitPrivate(t *testing.T) {
	var before ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_explicitPrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
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

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_overridePublic(t *testing.T) {
	var before ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_overridePublic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
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

// https://github.com/terraform-providers/terraform-provider-aws/issues/227
func TestAccAWSInstance_associatePublic_overridePrivate(t *testing.T) {
	var before ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_associatePublic_overridePrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_ip", ""),
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

func TestAccAWSInstance_getPasswordData_falseToTrue(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_data", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_getPasswordData(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(t, &before, &after),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
		},
	})
}

func TestAccAWSInstance_getPasswordData_trueToFalse(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_getPasswordData(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password_data", "get_password_data"},
			},
			{
				Config: testAccInstanceConfig_getPasswordData(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					testAccCheckInstanceNotRecreated(t, &before, &after),
					resource.TestCheckResourceAttr(resourceName, "get_password_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "password_data", ""),
				),
			},
		},
	})
}

func TestAccAWSInstance_CreditSpecification_Empty_NonBurstable(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CreditSpecification_Empty_NonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credit_specification"},
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/10203
func TestAccAWSInstance_CreditSpecification_UnspecifiedToEmpty_NonBurstable(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_CreditSpecification_Unspecified_NonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_CreditSpecification_Empty_NonBurstable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_unspecifiedDefaultsToStandard(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
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

func TestAccAWSInstance_creditSpecification_standardCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_unlimitedCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_unknownCpuCredits_t2(t *testing.T) {
	var instance ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unknownCpuCredits(rInt, "t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
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

func TestAccAWSInstance_creditSpecification_unknownCpuCredits_t3(t *testing.T) {
	var instance ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unknownCpuCredits(rInt, "t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
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

func TestAccAWSInstance_creditSpecification_updateCpuCredits(t *testing.T) {
	var first, second, third ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_isNotAppliedToNonBurstable(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_isNotAppliedToNonBurstable(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credit_specification"},
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_unspecifiedDefaultsToUnlimited(t *testing.T) {
	var instance ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
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

func TestAccAWSInstance_creditSpecificationT3_standardCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_unlimitedCpuCredits(t *testing.T) {
	var first, second ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unspecified_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecificationT3_updateCpuCredits(t *testing.T) {
	var first, second, third ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &first),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &second),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &third),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
		},
	})
}

func TestAccAWSInstance_creditSpecification_standardCpuCredits_t2Tot3Taint(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "standard"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt),
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

func TestAccAWSInstance_creditSpecification_unlimitedCpuCredits_t2Tot3Taint(t *testing.T) {
	var before, after ec2.Instance
	resourceName := "aws_instance.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt),
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

func TestAccAWSInstance_disappears(t *testing.T) {
	var conf ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &conf),
					testAccCheckInstanceDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSInstance_UserData_EmptyStringToUnspecified(t *testing.T) {
	var instance ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_UserData_EmptyString(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data"},
			},
			// Switching should show no difference
			{
				Config:             testAccInstanceConfig_UserData_Unspecified(rInt),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccAWSInstance_UserData_UnspecifiedToEmptyString(t *testing.T) {
	var instance ec2.Instance
	rInt := acctest.RandInt()
	resourceName := "aws_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_UserData_Unspecified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &instance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Switching should show no difference
			{
				Config:             testAccInstanceConfig_UserData_EmptyString(rInt),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func testAccCheckInstanceNotRecreated(t *testing.T,
	before, after *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.InstanceId != *after.InstanceId {
			t.Fatalf("AWS Instance IDs have changed. Before %s. After %s", *before.InstanceId, *after.InstanceId)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	return testAccCheckInstanceDestroyWithProvider(s, testAccProvider)
}

func testAccCheckInstanceDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_instance" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			for _, r := range resp.Reservations {
				for _, i := range r.Instances {
					if i.State != nil && *i.State.Name != "terminated" {
						return fmt.Errorf("Found unterminated instance: %s", i)
					}
				}
			}
		}

		// Verify the error is what we want
		if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidInstanceID.NotFound" {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckInstanceExists(n string, i *ec2.Instance) resource.TestCheckFunc {
	return testAccCheckInstanceExistsWithProvider(n, i, func() *schema.Provider { return testAccProvider })
}

func testAccCheckInstanceExistsWithProvider(n string, i *ec2.Instance, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		provider := providerF()

		conn := provider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.Reservations) > 0 {
			*i = *resp.Reservations[0].Instances[0]
			return nil
		}

		return fmt.Errorf("Instance not found")
	}
}

func testAccCheckInstanceDisappears(conf *ec2.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		params := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{conf.InstanceId},
		}

		if _, err := conn.TerminateInstances(params); err != nil {
			return err
		}

		return waitForInstanceDeletion(conn, *conf.InstanceId, 10*time.Minute)
	}
}

func TestInstanceTenancySchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["tenancy"]
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

func TestInstanceHostIDSchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["host_id"]
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

func TestInstanceCpuCoreCountSchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["cpu_core_count"]
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

func TestInstanceCpuThreadsPerCoreSchema(t *testing.T) {
	actualSchema := resourceAwsInstance().Schema["cpu_threads_per_core"]
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
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
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

func testAccInstanceConfigInDefaultVpcBySgName(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "sg" {
  name        = "tf_acc_test_%d"
  description = "Test security group"
  vpc_id      = "${data.aws_vpc.default.id}"
}

resource "aws_instance" "test" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "t2.micro"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}

func testAccInstanceConfigInDefaultVpcBySgId(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "sg" {
  name        = "tf_acc_test_%d"
  description = "Test security group"
  vpc_id      = "${data.aws_vpc.default.id}"
}

resource "aws_instance" "test" {
  ami                    = "${data.aws_ami.ubuntu.id}"
  instance_type          = "t2.micro"
  vpc_security_group_ids = ["${aws_security_group.sg.id}"]
}
`, rInt)
}

func testAccInstanceConfigInEc2Classic(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["paravirtual"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_security_group" "sg" {
  name        = "tf_acc_test_%d"
  description = "Test security group"
}

resource "aws_instance" "test" {
  ami             = "${data.aws_ami.ubuntu.id}"
  instance_type   = "m3.medium"
  security_groups = ["${aws_security_group.sg.name}"]
}
`, rInt)
}

func testAccInstanceConfig_pre(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_test_test" {
  name        = "tf_test_%d"
  description = "test"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rInt)
}

func testAccInstanceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_test_test" {
  name        = "tf_test_%d"
  description = "test"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "test" {
  # us-west-2
  ami               = "ami-4fccb37f"
  availability_zone = "us-west-2a"

  instance_type   = "m1.small"
  security_groups = ["${aws_security_group.tf_test_test.name}"]
  user_data       = "foo:-with-character's"
}
`, rInt)
}

func testAccInstanceConfigWithUserDataBase64(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf_test_test" {
  name        = "tf_test_%d"
  description = "test"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "test" {
  # us-west-2
  ami               = "ami-4fccb37f"
  availability_zone = "us-west-2a"

  instance_type    = "m1.small"
  security_groups  = ["${aws_security_group.tf_test_test.name}"]
  user_data_base64 = "${base64encode("hello world")}"
}
`, rInt)
}

const testAccInstanceConfigWithSmallInstanceType = `
resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-55a7ea65"
	availability_zone = "us-west-2a"

	instance_type = "m3.medium"

	tags = {
	    Name = "tf-acctest"
	}
}
`

const testAccInstanceConfigUpdateInstanceType = `
resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-55a7ea65"
	availability_zone = "us-west-2a"

	instance_type = "m3.large"

	tags = {
	    Name = "tf-acctest"
	}
}
`

const testAccInstanceGP2IopsDevice = `
resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-55a7ea65"

	# In order to attach an encrypted volume to an instance you need to have an
	# m3.medium or larger. See "Supported Instance Types" in:
	# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
	}
}
`

const testAccInstanceGP2WithIopsValue = `
resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-55a7ea65"

	# In order to attach an encrypted volume to an instance you need to have an
	# m3.medium or larger. See "Supported Instance Types" in:
	# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
	instance_type = "m3.medium"

	root_block_device {
		volume_type = "gp2"
		volume_size = 11
        # configured explicitly
		iops        = 10
	}
}
`

const testAccInstanceConfigBlockDevices = `
resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-55a7ea65"

	# In order to attach an encrypted volume to an instance you need to have an
	# m3.medium or larger. See "Supported Instance Types" in:
	# http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
	instance_type = "m3.medium"

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
		iops = 100
	}

	# Encrypted ebs block device
	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted   = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}
}
`

const testAccInstanceConfigSourceDestEnable = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-source-dest-enable"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-source-dest-enable"
	}
}

resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.test.id}"
}
`

const testAccInstanceConfigSourceDestDisable = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-source-dest-disable"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-source-dest-disable"
	}
}

resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.test.id}"
	source_dest_check = false
}
`

func testAccInstanceConfigDisableAPITermination(val bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-disable-api-termination"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-instance-disable-api-termination"
  }
}

resource "aws_instance" "test" {
  # us-west-2
  ami                     = "ami-4fccb37f"
  instance_type           = "m1.small"
  subnet_id               = "${aws_subnet.test.id}"
  disable_api_termination = %t
}
`, val)
}

const testAccInstanceConfigVPC = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-vpc"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-vpc"
	}
}

resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.test.id}"
	associate_public_ip_address = true
	tenancy = "dedicated"
	# pre-encoded base64 data
	user_data = "3dc39dda39be1205215e776bad998da361a5955d"
}
`

func testAccInstanceConfigPlacementGroup(rStr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-placement-group"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-instance-placement-group"
  }
}

resource "aws_placement_group" "test" {
  name     = "testAccInstanceConfigPlacementGroup_%s"
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "test" {
  # us-west-2
  ami                         = "ami-55a7ea65"
  instance_type               = "c3.large"
  subnet_id                   = "${aws_subnet.test.id}"
  associate_public_ip_address = true
  placement_group             = "${aws_placement_group.test.name}"

  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"
}
`, rStr)
}

const testAccInstanceConfigIpv6ErrorConfig = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-instance-ipv6-err"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)}"
	tags = {
		Name = "tf-acc-instance-ipv6-err"
	}
}

resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"
	ipv6_addresses = ["2600:1f14:bb2:e501::10"]
	ipv6_address_count = 1
	tags = {
		Name = "tf-ipv6-instance-acc-test"
	}
}
`

const testAccInstanceConfigIpv6Support = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-instance-ipv6-support"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)}"
	tags = {
		Name = "tf-acc-instance-ipv6-support"
	}
}

resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"

	ipv6_address_count = 1
	tags = {
		Name = "tf-ipv6-instance-acc-test"
	}
}
`

const testAccInstanceConfigIpv6SupportWithIpv4 = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-instance-ipv6-support-with-ipv4"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)}"
	tags = {
		Name = "tf-acc-instance-ipv6-support-with-ipv4"
	}
}

resource "aws_instance" "test" {
	# us-west-2
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"

	associate_public_ip_address = true
	ipv6_address_count = 1
	tags = {
		Name = "tf-ipv6-instance-acc-test"
	}
}
`

const testAccInstanceConfigMultipleRegions = `
provider "aws" {
	alias = "west"
	region = "us-west-2"
}

provider "aws" {
	alias = "east"
	region = "us-east-1"
}
resource "aws_instance" "test" {
	# us-west-2
	provider = "aws.west"
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
}

resource "aws_instance" "test2" {
	# us-east-1
	provider = "aws.east"
	ami = "ami-8c6ea9e4"
	instance_type = "m1.small"
}
`

const testAccCheckInstanceConfigTags = `
resource "aws_instance" "test" {
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	tags = {
		test = "test2"
	}
}
`

const testAccInstanceConfigEbsBlockDeviceKmsKeyArn = `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  # us-west-2
  ami = "ami-55a7ea65"

  # In order to attach an encrypted volume to an instance you need to have an
  # m3.medium or larger. See "Supported Instance Types" in:
  # http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html
  instance_type = "m3.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  # Encrypted ebs block device
  ebs_block_device {
    device_name = "/dev/sdd"
    encrypted   = true
    kms_key_id  = "${aws_kms_key.test.arn}"
    volume_size = 12
  }
}
`

const testAccInstanceConfigRootBlockDeviceKmsKeyArn = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-source-dest-enable"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-instance-source-dest-enable"
  }
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_instance" "test" {
  ami           = "ami-08692d171e3cf02d6"
  instance_type = "t3.nano"
  subnet_id     = "${aws_subnet.test.id}"

  root_block_device {
    delete_on_termination = true
    encrypted             = true
    kms_key_id            = "${aws_kms_key.test.arn}"
  }
}
`

const testAccCheckInstanceConfigWithAttachedVolume = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["debian-jessie-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  owners = ["379101102735"] # Debian
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.debian_jessie_latest.id}"
  instance_type = "t2.medium"

  root_block_device {
    delete_on_termination = true
    volume_size           = "10"
    volume_type           = "standard"
  }

  tags = {
    Name    = "test-terraform"
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = "test-terraform"
  }
}

resource "aws_volume_attachment" "test" {
  device_name = "/dev/xvdg"
  volume_id   = "${aws_ebs_volume.test.id}"
  instance_id = "${aws_instance.test.id}"
}
`

const testAccCheckInstanceConfigNoVolumeTags = `
resource "aws_instance" "test" {
	ami = "ami-55a7ea65"

	instance_type = "m3.medium"

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
		iops = 100
	}

	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}
}
`

const testAccCheckInstanceConfigWithVolumeTags = `
resource "aws_instance" "test" {
	ami = "ami-55a7ea65"

	instance_type = "m3.medium"

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
		iops = 100
	}

	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}

	volume_tags = {
		Name = "acceptance-test-volume-tag"
	}
}
`

const testAccCheckInstanceConfigWithVolumeTagsUpdate = `
resource "aws_instance" "test" {
	ami = "ami-55a7ea65"

	instance_type = "m3.medium"

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
		iops = 100
	}

	ebs_block_device {
		device_name = "/dev/sdd"
		volume_size = 12
		encrypted = true
	}

	ephemeral_block_device {
		device_name = "/dev/sde"
		virtual_name = "ephemeral0"
	}

	volume_tags = {
		Name = "acceptance-test-volume-tag"
		Environment = "dev"
	}
}
`

const testAccCheckInstanceConfigTagsUpdate = `
resource "aws_instance" "test" {
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	tags = {
		test2 = "test3"
	}
}
`

func testAccInstanceConfigWithoutInstanceProfile(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = "test-%s"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_instance" "test" {
  ami           = "ami-4fccb37f"
  instance_type = "m1.small"

  tags = {
    test = "test3"
  }
}
`, rName)
}

func testAccInstanceConfigWithInstanceProfile(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = "test-%s"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_instance_profile" "test" {
  name  = "test-%s"
  roles = ["${aws_iam_role.test.name}"]
}

resource "aws_instance" "test" {
  ami                  = "ami-4fccb37f"
  instance_type        = "m1.small"
  iam_instance_profile = "${aws_iam_instance_profile.test.name}"

  tags = {
    test = "test3"
  }
}
`, rName, rName)
}

const testAccInstanceConfigPrivateIP = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-private-ip"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-private-ip"
	}
}

resource "aws_instance" "test" {
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"
	private_ip = "10.1.1.42"
}
`

const testAccInstanceConfigAssociatePublicIPAndPrivateIP = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-public-ip-and-private-ip"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-public-ip-and-private-ip"
	}
}

resource "aws_instance" "test" {
	ami = "ami-c5eabbf5"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"
	associate_public_ip_address = true
	private_ip = "10.1.1.42"
}
`

func testAccInstanceNetworkInstanceSecurityGroups(rInt int) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-network-security-groups"
  }
}

resource "aws_security_group" "test" {
  name        = "tf_test_%d"
  description = "test"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-instance-network-security-groups"
  }
}

resource "aws_instance" "test" {
  ami                         = "ami-21f78e11"
  instance_type               = "t1.micro"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"
  associate_public_ip_address = true
  depends_on                  = ["aws_internet_gateway.gw"]
}

resource "aws_eip" "test_eip" {
  instance   = "${aws_instance.test.id}"
  vpc        = true
  depends_on = ["aws_internet_gateway.gw"]
}
`, rInt)
}

func testAccInstanceNetworkInstanceVPCSecurityGroupIDs(rInt int) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-network-vpc-sg-ids"
  }
}

resource "aws_security_group" "tf_test_test" {
  name        = "tf_test_%d"
  description = "test"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-instance-network-vpc-sg-ids"
  }
}

resource "aws_instance" "test" {
  ami                    = "ami-21f78e11"
  instance_type          = "t1.micro"
  vpc_security_group_ids = ["${aws_security_group.tf_test_test.id}"]
  subnet_id              = "${aws_subnet.test.id}"
  depends_on             = ["aws_internet_gateway.gw"]
}

resource "aws_eip" "test_eip" {
  instance   = "${aws_instance.test.id}"
  vpc        = true
  depends_on = ["aws_internet_gateway.gw"]
}
`, rInt)
}

func testAccInstanceNetworkInstanceVPCRemoveSecurityGroupIDs(rInt int) string {
	return fmt.Sprintf(`
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-network-vpc-sg-ids"
  }
}

resource "aws_security_group" "tf_test_test" {
  name        = "tf_test_%d"
  description = "test"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-instance-network-vpc-sg-ids"
  }
}

resource "aws_instance" "test" {
  ami                    = "ami-21f78e11"
  instance_type          = "t1.micro"
  vpc_security_group_ids = []
  subnet_id              = "${aws_subnet.test.id}"
  depends_on             = ["aws_internet_gateway.gw"]
}

resource "aws_eip" "test_eip" {
  instance   = "${aws_instance.test.id}"
  vpc        = true
  depends_on = ["aws_internet_gateway.gw"]
}
`, rInt)
}

func testAccInstanceConfigKeyPair(keyPairName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_key_pair" "debugging" {
  key_name   = "%s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}
resource "aws_instance" "test" {
  ami           = "ami-408c7f28"
  instance_type = "t1.micro"
  key_name      = "${aws_key_pair.debugging.key_name}"
  
  tags = {
    Name = "testAccInstanceConfigKeyPair_TestAMI"
  }
}
`, keyPairName)
}

const testAccInstanceConfigRootBlockDeviceMismatch = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-root-block-device-mismatch"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-root-block-device-mismatch"
	}
}

resource "aws_instance" "test" {
	// This is an AMI with RootDeviceName: "/dev/sda1"; actual root: "/dev/sda"
	ami = "ami-ef5b69df"
	instance_type = "t1.micro"
	subnet_id = "${aws_subnet.test.id}"
	root_block_device {
		volume_size = 13
	}
}
`

const testAccInstanceConfigForceNewAndTagsDrift = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-force-new-and-tags-drift"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-force-new-and-tags-drift"
	}
}

resource "aws_instance" "test" {
	ami = "ami-22b9a343"
	instance_type = "t2.nano"
	subnet_id = "${aws_subnet.test.id}"
}
`

const testAccInstanceConfigForceNewAndTagsDrift_Update = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-instance-force-new-and-tags-drift"
	}
}

resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.test.id}"
	tags = {
		Name = "tf-acc-instance-force-new-and-tags-drift"
	}
}

resource "aws_instance" "test" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"
}
`

const testAccInstanceConfigPrimaryNetworkInterface = `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-primary-network-iface"
  }
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-primary-network-iface"
  }
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["172.16.10.100"]
  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_instance" "test" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.test.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigPrimaryNetworkInterfaceSourceDestCheck = `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-primary-network-iface-source-dest-check"
  }
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-primary-network-iface-source-dest-check"
  }
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["172.16.10.100"]
  source_dest_check = false
  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_instance" "test" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.test.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigAddSecondaryNetworkInterfaceBefore = `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-add-secondary-network-iface"
  }
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-add-secondary-network-iface"
  }
}

resource "aws_network_interface" "primary" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["172.16.10.100"]
  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_network_interface" "secondary" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["172.16.10.101"]
  tags = {
    Name = "secondary_network_interface"
  }
}

resource "aws_instance" "test" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.primary.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigAddSecondaryNetworkInterfaceAfter = `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-add-secondary-network-iface"
  }
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.16.10.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-instance-add-secondary-network-iface"
  }
}

resource "aws_network_interface" "primary" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["172.16.10.100"]
  tags = {
    Name = "primary_network_interface"
  }
}

// Attach previously created network interface, observe no state diff on instance resource
resource "aws_network_interface" "secondary" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["172.16.10.101"]
  tags = {
    Name = "secondary_network_interface"
  }
  attachment {
    instance = "${aws_instance.test.id}"
    device_index = 1
  }
}

resource "aws_instance" "test" {
	ami = "ami-22b9a343"
	instance_type = "t2.micro"
	network_interface {
	 network_interface_id = "${aws_network_interface.primary.id}"
	 device_index = 0
  }
}
`

const testAccInstanceConfigAddSecurityGroupBefore = `
resource "aws_vpc" "test" {
    cidr_block = "172.16.0.0/16"
  tags = {
        Name = "terraform-testacc-instance-add-security-group"
    }
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-test"
    }
}

resource "aws_subnet" "test2" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.11.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-test2"
    }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test"
  name = "test"
}

resource "aws_security_group" "test2" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test2"
  name = "test2"
}

resource "aws_instance" "test" {
    ami = "ami-c5eabbf5"
    instance_type = "t2.micro"
    subnet_id = "${aws_subnet.test.id}"
    associate_public_ip_address = false
    vpc_security_group_ids = [
      "${aws_security_group.test.id}"
    ]
  tags = {
        Name = "test-instance-sg-add-test"
    }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.test.id}"]
    attachment {
        instance = "${aws_instance.test.id}"
        device_index = 1
    }
  tags = {
        Name = "test_interface"
    }
}
`

const testAccInstanceConfigAddSecurityGroupAfter = `
resource "aws_vpc" "test" {
    cidr_block = "172.16.0.0/16"
  tags = {
        Name = "terraform-testacc-instance-add-security-group"
    }
}

resource "aws_subnet" "test" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.10.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-test"
    }
}

resource "aws_subnet" "test2" {
    vpc_id = "${aws_vpc.test.id}"
    cidr_block = "172.16.11.0/24"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-instance-add-security-group-test"
    }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test"
  name = "test"
}

resource "aws_security_group" "test2" {
  vpc_id = "${aws_vpc.test.id}"
  description = "test2"
  name = "test2"
}

resource "aws_instance" "test" {
    ami = "ami-c5eabbf5"
    instance_type = "t2.micro"
    subnet_id = "${aws_subnet.test.id}"
    associate_public_ip_address = false
    vpc_security_group_ids = [
      "${aws_security_group.test.id}",
      "${aws_security_group.test2.id}"
    ]
  tags = {
        Name = "test-instance-sg-add-test"
    }
}

resource "aws_network_interface" "test" {
    subnet_id = "${aws_subnet.test.id}"
    private_ips = ["172.16.10.100"]
    security_groups = ["${aws_security_group.test.id}"]
    attachment {
        instance = "${aws_instance.test.id}"
        device_index = 1
    }
  tags = {
        Name = "test_interface"
    }
}
`

func testAccInstanceConfig_associatePublic_defaultPrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-associate-public-default-private"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.16.20.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = false

  tags = {
    Name = "tf-acc-instance-associate-public-default-private"
  }
}

resource "aws_instance" "test" {
  ami           = "ami-22b9a343"                   # us-west-2
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.public_subnet.id}"

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt)
}

func testAccInstanceConfig_associatePublic_defaultPublic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-associate-public-default-public"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.16.20.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-instance-associate-public-default-public"
  }
}

resource "aws_instance" "test" {
  ami           = "ami-22b9a343"                   # us-west-2
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.public_subnet.id}"

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt)
}

func testAccInstanceConfig_associatePublic_explicitPublic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-associate-public-explicit-public"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.16.20.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-instance-associate-public-explicit-public"
  }
}

resource "aws_instance" "test" {
  ami                         = "ami-22b9a343"                   # us-west-2
  instance_type               = "t2.micro"
  subnet_id                   = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = true

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt)
}

func testAccInstanceConfig_associatePublic_explicitPrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-associate-public-explicit-private"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.16.20.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = false

  tags = {
    Name = "tf-acc-instance-associate-public-explicit-private"
  }
}

resource "aws_instance" "test" {
  ami                         = "ami-22b9a343"                   # us-west-2
  instance_type               = "t2.micro"
  subnet_id                   = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = false

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt)
}

func testAccInstanceConfig_associatePublic_overridePublic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-associate-public-override-public"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.16.20.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = false

  tags = {
    Name = "tf-acc-instance-associate-public-override-public"
  }
}

resource "aws_instance" "test" {
  ami                         = "ami-22b9a343"                   # us-west-2
  instance_type               = "t2.micro"
  subnet_id                   = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = true

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt)
}

func testAccInstanceConfig_associatePublic_overridePrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-instance-associate-public-override-private"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.16.20.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-instance-associate-public-override-private"
  }
}

resource "aws_instance" "test" {
  ami                         = "ami-22b9a343"                   # us-west-2
  instance_type               = "t2.micro"
  subnet_id                   = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = false

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt)
}

func testAccInstanceConfig_getPasswordData(val bool, rInt int) string {
	return fmt.Sprintf(`
# Find latest Microsoft Windows Server 2016 Core image (Amazon deletes old ones)
data "aws_ami" "win2016core" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["Windows_Server-2016-English-Core-Base-*"]
  }
}

resource "aws_key_pair" "test" {
  key_name   = "tf-acctest-%d"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAq6U3HQYC4g8WzU147gZZ7CKQH8TgYn3chZGRPxaGmHW1RUwsyEs0nmombmIhwxudhJ4ehjqXsDLoQpd6+c7BuLgTMvbv8LgE9LX53vnljFe1dsObsr/fYLvpU9LTlo8HgHAqO5ibNdrAUvV31ronzCZhms/Gyfdaue88Fd0/YnsZVGeOZPayRkdOHSpqme2CBrpa8myBeL1CWl0LkDG4+YCURjbaelfyZlIApLYKy3FcCan9XQFKaL32MJZwCgzfOvWIMtYcU8QtXMgnA3/I3gXk8YDUJv5P4lj0s/PJXuTM8DygVAUtebNwPuinS7wwonm5FXcWMuVGsVpG5K7FGQ== tf-acc-winpasswordtest"
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.win2016core.id}"
  instance_type = "t2.medium"
  key_name      = "${aws_key_pair.test.key_name}"

  get_password_data = %t
}
`, rInt, val)
}

func testAccInstanceConfig_CreditSpecification_Empty_NonBurstable(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}


resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "m5.large"
  subnet_id     = "${aws_subnet.test.id}"

  credit_specification {}
}
`, rName)
}

func testAccInstanceConfig_CreditSpecification_Unspecified_NonBurstable(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "m5.large"
  subnet_id     = "${aws_subnet.test.id}"
}
`, rName)
}

func testAccInstanceConfig_creditSpecification_unspecified(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-22b9a343"               # us-west-2
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.my_subnet.id}"
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unspecified_t3(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-51537029"               # us-west-2
  instance_type = "t3.micro"
  subnet_id     = "${aws_subnet.my_subnet.id}"

}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_standardCpuCredits(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-22b9a343"               # us-west-2
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.my_subnet.id}"

  credit_specification {
    cpu_credits = "standard"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_standardCpuCredits_t3(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-51537029"               # us-west-2
  instance_type = "t3.micro"
  subnet_id     = "${aws_subnet.my_subnet.id}"

  credit_specification {
    cpu_credits = "standard"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unlimitedCpuCredits(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-22b9a343"               # us-west-2
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.my_subnet.id}"

  credit_specification {
    cpu_credits = "unlimited"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unlimitedCpuCredits_t3(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-51537029"               # us-west-2
  instance_type = "t3.micro"
  subnet_id     = "${aws_subnet.my_subnet.id}"

  credit_specification {
    cpu_credits = "unlimited"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_isNotAppliedToNonBurstable(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-22b9a343"               # us-west-2
  instance_type = "m1.small"
  subnet_id     = "${aws_subnet.my_subnet.id}"

  credit_specification {
    cpu_credits = "standard"
  }
}
`, rInt)
}

func testAccInstanceConfig_creditSpecification_unknownCpuCredits(rInt int, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.20.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_instance" "test" {
  ami           = "ami-51537029"               # us-west-2
  instance_type = %q
  subnet_id     = "${aws_subnet.my_subnet.id}"

  credit_specification {
  }
}
`, rInt, instanceType)
}

func testAccInstanceConfig_UserData_Base(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-acctest-%d"
  }
}

resource "aws_subnet" "test" {
  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "172.16.0.0/24"

  tags = {
    Name = "tf-acctest-%d"
  }
}
`, rInt, rInt)
}

func testAccInstanceConfig_UserData_Unspecified(rInt int) string {
	return testAccInstanceConfig_UserData_Base(rInt) + `
resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.test.id}"
}
`
}

func testAccInstanceConfig_UserData_EmptyString(rInt int) string {
	return testAccInstanceConfig_UserData_Base(rInt) + `
resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.test.id}"
  user_data     = ""
}
`
}
