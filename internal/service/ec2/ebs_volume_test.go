package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

// testAccErrorCheckSkipEBSVolume skips EBS volume tests that have error messages indicating unsupported features
func testAccErrorCheckSkipEBSVolume(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"specified zone does not support multi-attach-enabled volumes",
		"Unsupported volume type",
	)
}

func TestAccEC2EBSVolume_basic(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
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

func TestAccEC2EBSVolume_updateAttachedEBSVolume(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSAttachedVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSAttachedVolumeUpdateSizeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "20"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_updateSize(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeUpdateSizeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_updateType(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeUpdateTypeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "sc1"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_UpdateIops_io1(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeWithIopsIo1Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeWithIopsIo1UpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "200"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_UpdateIops_io2(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeWithIopsIo2Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeWithIopsIo2UpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "200"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_kmsKey(t *testing.T) {
	var v ec2.Volume
	ri := sdkacctest.RandInt()
	config := fmt.Sprintf(testAccEBSVolumeWithKMSKeyConfig, ri)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
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

func TestAccEC2EBSVolume_noIops(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeWithNoIopsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
func TestAccEC2EBSVolume_invalidIopsForType(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccEBSVolumeWithInvalidIopsForTypeConfig,
				ExpectError: regexp.MustCompile(`'iops' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccEC2EBSVolume_invalidThroughputForType(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccEBSVolumeWithInvalidThroughputForTypeConfig,
				ExpectError: regexp.MustCompile(`'throughput' must not be set when 'type' is`),
			},
		},
	})
}

func TestAccEC2EBSVolume_withTags(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeWithTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TerraformTest"),
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

func TestAccEC2EBSVolume_multiAttach(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeMultiAttachConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
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

func TestAccEC2EBSVolume_outpost(t *testing.T) {
	var v ec2.Volume
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeOutpostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, "arn"),
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

func TestAccEC2EBSVolume_GP3_basic(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp3", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "125"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
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

func TestAccEC2EBSVolume_GP3_iops(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp3", "4000", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "4000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "200"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp3", "5000", "200"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "5000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "200"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_GP3_throughput(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp3", "", "400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "400"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp3", "", "600"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "600"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_gp3ToGP2(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp3", "3000", "400"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "400"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, "10", "gp2", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
				),
			},
		},
	})
}

func TestAccEC2EBSVolume_snapshotID(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeSnapshotIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
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

func TestAccEC2EBSVolume_snapshotIDAndSize(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"
	snapshotResourceName := "aws_ebs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeSnapshotIdAndSizeConfig(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "multi_attach_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "size", "20"),
					resource.TestCheckResourceAttrPair(resourceName, "snapshot_id", snapshotResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
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

func TestAccEC2EBSVolume_disappears(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipEBSVolume(t),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEBSVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVolumeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_volume" {
			continue
		}

		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeVolumes(request)

		if tfawserr.ErrMessageContains(err, "InvalidVolume.NotFound", "") {
			continue
		}

		if err == nil {
			for _, volume := range resp.Volumes {
				if aws.StringValue(volume.VolumeId) == rs.Primary.ID {
					return fmt.Errorf("Volume still exists")
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckVolumeExists(n string, v *ec2.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeVolumes(request)
		if err == nil {
			if response.Volumes != nil && len(response.Volumes) > 0 {
				*v = *response.Volumes[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 volume %s", rs.Primary.ID)
	}
}

const testAccEBSVolumeConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 1
}
`

const testAccEBSAttachedVolumeConfig = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
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

  owners = ["amazon"]
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.debian_jessie_latest.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name = "test-terraform"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  depends_on        = [aws_instance.test]
  availability_zone = aws_instance.test.availability_zone
  type              = "gp2"
  size              = "10"
}

resource "aws_volume_attachment" "test" {
  depends_on  = [aws_ebs_volume.test]
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`

const testAccEBSAttachedVolumeUpdateSizeConfig = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
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

  owners = ["amazon"]
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.debian_jessie_latest.id
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name = "test-terraform"
  }
}

resource "aws_ebs_volume" "test" {
  depends_on        = [aws_instance.test]
  availability_zone = aws_instance.test.availability_zone
  type              = "gp2"
  size              = "20"
}

resource "aws_volume_attachment" "test" {
  depends_on  = [aws_ebs_volume.test]
  device_name = "/dev/xvdg"
  volume_id   = aws_ebs_volume.test.id
  instance_id = aws_instance.test.id
}
`

const testAccEBSVolumeUpdateSizeConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 10

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccEBSVolumeUpdateTypeConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "sc1"
  size              = 500

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccEBSVolumeWithIopsIo1Config = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 4
  iops              = 100

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccEBSVolumeWithIopsIo1UpdatedConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 4
  iops              = 200

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccEBSVolumeWithIopsIo2Config = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io2"
  size              = 4
  iops              = 100

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccEBSVolumeWithIopsIo2UpdatedConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io2"
  size              = 4
  iops              = 200

  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccEBSVolumeWithKMSKeyConfig = `
resource "aws_kms_key" "test" {
  description = "Terraform acc test %d"
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

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
  kms_key_id        = aws_kms_key.test.arn
}
`

const testAccEBSVolumeWithTagsConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccEBSVolumeWithNoIopsConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  type              = "gp2"
  iops              = 0

  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccEBSVolumeWithInvalidIopsForTypeConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  iops              = 100

  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccEBSVolumeWithInvalidThroughputForTypeConfig = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 10
  iops              = 100
  throughput        = 500
  type              = "io1"

  tags = {
    Name = "TerraformTest"
  }
}
`

func testAccEBSVolumeOutpostConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  size              = 1
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = "tf-acc-volume-outpost"
  }
}
`
}

func testAccEBSVolumeMultiAttachConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  type                 = "io1"
  multi_attach_enabled = true
  size                 = 4
  iops                 = 100

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccEBSVolumeSizeTypeIopsThroughputConfig(rName, size, volumeType, iops, throughput string) string {
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

func testAccEBSVolumeSnapshotIDConfig(rName string) string {
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

func testAccEBSVolumeSnapshotIdAndSizeConfig(rName string, size int) string {
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
