// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package launchwizard_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/launchwizard"
	awstypes "github.com/aws/aws-sdk-go-v2/service/launchwizard/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflaunchwizard "github.com/hashicorp/terraform-provider-aws/internal/service/launchwizard"
	"github.com/hashicorp/terraform-provider-aws/names"
)

/**
 Most test cases run without any precondition. However, in order to test the full SAP installation (test case basicFullInstallation), the following preconditions must be met:

 Before running this test, the following ENV variable must be set:

 LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI     - URI of S3 Partition containing SAP Installation Media (without trailing slash)

 Expected structure:

	Inside the S3 Partition, the content is is expected as described here. We only test with HANA deployments which need to be placed in a folder named "HANA_DB_Software" in the bucket. (https://docs.aws.amazon.com/launchwizard/latest/userguide/launch-wizard-sap-software-install-details.html)
**/

func TestAccLaunchWizardDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}
	testSpec := TestSpecSapHanaSingleInfraOnly

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegionAvailable(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LaunchWizard),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"], "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_pattern", testSpec["deployment_pattern"]),
					resource.TestCheckResourceAttr(resourceName, "workload_name", testSpec["workload_name"]),
					resource.TestCheckResourceAttr(resourceName, "specifications.DatabasePrimaryHostname", "hana-primary"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     rName,
				ImportStateVerifyIgnore: []string{
					"specifications.DatabasePassword",
				},
			},
		},
	})
}

func TestAccLaunchWizardDeployment_basicFullInstallation(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	if os.Getenv("LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI") == "" {
		t.Skip("skipping optional acceptance testing: env variable LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI must be set for acceptance test that installs SAP software")
	}

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}
	testSpec := TestSpecSapHanaSingleFullInstallation

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegionAvailable(ctx, t)
			testAccPreCheckEnvironmentVariableSetSapInstMedia(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LaunchWizard),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"], "", os.Getenv("LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_pattern", testSpec["deployment_pattern"]),
					resource.TestCheckResourceAttr(resourceName, "workload_name", testSpec["workload_name"]),
					resource.TestCheckResourceAttr(resourceName, "specifications.DatabasePrimaryHostname", "hana-primary"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     rName,
				ImportStateVerifyIgnore: []string{
					"specifications.DatabasePassword",
				},
			},
		},
	})
}

func TestAccLaunchWizardDeployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	testSpec := TestSpecSapHanaSingleInfraOnly

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegionAvailable(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LaunchWizard),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"], "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflaunchwizard.ResourceDeployment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLaunchWizardDeployment_SkipDestroyAfterFailure(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	testSpec := TestSpecSapHanaSingleInfraOnly
	testSpecFailure := TestSpecSapHanaSingleFailure

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckRegionAvailable(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LaunchWizard),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDeploymentConfig_basic(rName, testSpecFailure["specification_template"], testSpecFailure["deployment_pattern"], testSpecFailure["workload_name"], "true", ""),
				ExpectError: regexache.MustCompile(".*Deployment will be replaced on next apply to allow troubleshooting.*"),
			},
			{
				Config:      testAccDeploymentConfig_basic(rName, testSpecFailure["specification_template"], testSpecFailure["deployment_pattern"], testSpecFailure["workload_name"], "false", ""),
				ExpectError: regexache.MustCompile(".*Deployment will be deleted*"),
			},
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"], "true", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "status", "COMPLETED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     rName,
				ImportStateVerifyIgnore: []string{
					"specifications.DatabasePassword",
					"skip_destroy_after_failure",
				},
			},
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"], "false", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy_after_failure", "false"),
				),
			},
		},
	})
}

func testAccDeploymentConfig_basic(rName string, specTemplate string, deploymentPattern string, workloadName string, skipDestroyAfterFailure string, instMediaURI string) string {
	var specification_template_computed string
	if instMediaURI != "" {
		specification_template_computed = fmt.Sprintf(specTemplate, instMediaURI)
	} else {
		specification_template_computed = specTemplate
	}

	if skipDestroyAfterFailure == "" {
		return acctest.ConfigCompose(
			testAccDeploymentConfigBase(rName),
			fmt.Sprintf(`
resource "aws_launchwizard_deployment" "test" {
  name               = %[1]q
  deployment_pattern = %[3]q
  workload_name      = %[4]q
  specifications     = %[2]s
}
`, rName, specification_template_computed, deploymentPattern, workloadName))
	} else {
		return acctest.ConfigCompose(
			testAccDeploymentConfigBase(rName),
			fmt.Sprintf(`
resource "aws_launchwizard_deployment" "test" {
  name                       = %[1]q
  deployment_pattern         = %[3]q
  workload_name              = %[4]q
  specifications             = %[2]s
  skip_destroy_after_failure = %[5]q
}
`, rName, specification_template_computed, deploymentPattern, workloadName, skipDestroyAfterFailure))
	}
}

func testAccDeploymentConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    Name = %[1]q
  }
}


resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
  tags = {
    Name = %[1]q
  }
}


resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
  tags = {
    Name                        = %[1]q
    LaunchWizardApplicationType = "SAP"
  }
}


resource "aws_main_route_table_association" "test" {
  route_table_id = aws_route_table.public.id
  vpc_id         = aws_vpc.test.id
}


data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}


resource "aws_subnet" "private" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 2)
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = %[1]q
  }
}


resource "aws_eip" "public" {
  count      = 2
  depends_on = [aws_internet_gateway.test]
  domain     = "vpc"
  tags = {
    Name = %[1]q
  }
}


resource "aws_nat_gateway" "public" {
  count         = 2
  allocation_id = aws_eip.public[count.index].id
  subnet_id     = aws_subnet.public[count.index].id
  tags = {
    Name = %[1]q
  }
}


resource "aws_route_table" "private" {
  count  = 2
  vpc_id = aws_vpc.test.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.public[count.index].id
  }
  tags = {
    Name = %[1]q
  }
}


resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}


resource "aws_subnet" "public" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = %[1]q
  }
}


resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


resource "tls_private_key" "key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}


resource "aws_key_pair" "key" {
  public_key = tls_private_key.key.public_key_openssh
}


data "aws_ami" "suse" {
  most_recent = true
  owners      = ["aws-marketplace"]
  filter {
    name   = "name"
    values = ["suse-sles-sap-15-sp4*hvm*"]
  }
}
`, rName)
}

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LaunchWizardClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_launchwizard_deployment" {
				continue
			}
			resp, err := conn.GetDeployment(ctx, &launchwizard.GetDeploymentInput{
				DeploymentId: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.LaunchWizard, create.ErrActionCheckingDestroyed, tflaunchwizard.ResNameDeployment, rs.Primary.ID, err)
			}
			if awstypes.DeploymentStatusDeleted == resp.Deployment.Status {
				return nil
			}
			return create.Error(names.LaunchWizard, create.ErrActionCheckingDestroyed, tflaunchwizard.ResNameDeployment, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckDeploymentExists(ctx context.Context, name string, deployment *launchwizard.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LaunchWizard, create.ErrActionCheckingExistence, tflaunchwizard.ResNameDeployment, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.LaunchWizard, create.ErrActionCheckingExistence, tflaunchwizard.ResNameDeployment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LaunchWizardClient(ctx)
		resp, err := conn.GetDeployment(ctx, &launchwizard.GetDeploymentInput{
			DeploymentId: aws.String(rs.Primary.ID),
		})

		*deployment = *resp

		if err != nil {
			return create.Error(names.LaunchWizard, create.ErrActionCheckingExistence, tflaunchwizard.ResNameDeployment, rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccPreCheckRegionAvailable(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LaunchWizardClient(ctx)
	input := &launchwizard.ListDeploymentsInput{}
	_, err := conn.ListDeployments(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckEnvironmentVariableSetSapInstMedia(t *testing.T) {
	if v := os.Getenv("LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI"); v == "" {
		t.Fatal("env variable LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI must be set for acceptance tests")
	}
}

// The following are example test specifications for the different SAP deployment patterns. We currently only use the HANA Single Node deployment pattern for testing as it should suffice to verify the functionality of the provider.
// The other patterns are not automatically tested as they are very long running and error prone while not adding any additional value to the test coverage. Still, they are included here for reference and for manual testing purposes.

var (
	TestSpecSapHanaSingleFullInstallation = map[string]string{
		"specification_template": `{
			"KeyPairName": "${aws_key_pair.key.key_name}",
			"AvailabilityZone1PrivateSubnet1Id": "${aws_subnet.private.0.id}",
			"VpcId": "${aws_vpc.test.id}",
			"Timezone" :"UTC",
			"EnableEbsVolumeEncryption" :"No",
			"CreateSecurityGroup": "No",
			"DatabaseSecurityGroupId": "${aws_security_group.test.id}",
			"SapSysGroupId" :"5002",
			"DatabaseSystemId" :"HDB",
			"DatabaseInstanceNumber": "00",
			"DatabasePassword" :"Password123@",
			"InstallDatabaseSoftware" :"Yes",
			"DatabaseInstallationMediaS3Uri" :"%[1]s/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :"${data.aws_ami.suse.id}",
			"DatabasePrimaryHostname" :"hana-primary",
			"DatabaseInstanceType" :"r5.2xlarge",
			"InstallAwsBackintAgent" :"No",
			"DisableDeploymentRollback" :"Yes",
			"SaveDeploymentArtifacts" :"No",
			"DatabaseAutomaticRecovery": "No",
			"DatabaseLogVolumeType": "gp3",
			"DatabaseDataVolumeType": "gp3"
		}
		`,
		"deployment_pattern": "SapHanaSingle",
		"workload_name":      "SAP",
	}
	TestSpecSapHanaSingleInfraOnly = map[string]string{
		"specification_template": `{
			"KeyPairName": "${aws_key_pair.key.key_name}",
			"AvailabilityZone1PrivateSubnet1Id": "${aws_subnet.private.0.id}",
			"VpcId": "${aws_vpc.test.id}",
			"Timezone" :"UTC",
			"EnableEbsVolumeEncryption" :"No",
			"CreateSecurityGroup": "No",
			"DatabaseSecurityGroupId": "${aws_security_group.test.id}",
			"SapSysGroupId" :"5002",
			"DatabaseSystemId" :"HDB",
			"DatabaseInstanceNumber": "00",
			"DatabasePassword" :"Password123@",
			"InstallDatabaseSoftware" :"No",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :"${data.aws_ami.suse.id}",
			"DatabasePrimaryHostname" :"hana-primary",
			"DatabaseInstanceType" :"r5.2xlarge",
			"InstallAwsBackintAgent" :"No",
			"DisableDeploymentRollback" :"Yes",
			"SaveDeploymentArtifacts" :"No",
			"DatabaseAutomaticRecovery": "No",
			"DatabaseLogVolumeType": "gp3",
			"DatabaseDataVolumeType": "gp3"
		}
		`,
		"deployment_pattern": "SapHanaSingle",
		"workload_name":      "SAP",
	}
	TestSpecSapHanaSingleFailure = map[string]string{
		"specification_template": `{
			"KeyPairName": "${aws_key_pair.key.key_name}",
			"AvailabilityZone1PrivateSubnet1Id": "${aws_subnet.private.0.id}",
			"VpcId": "${aws_vpc.test.id}",
			"Timezone" :"UTC",
			"EnableEbsVolumeEncryption" :"No",
			"CreateSecurityGroup": "No",
			"DatabaseSecurityGroupId": "${aws_security_group.test.id}",
			"SapSysGroupId" :"5002",
			"DatabaseSystemId" :"HDB",
			"DatabaseInstanceNumber": "00",
			"DatabasePassword" :"Password123@",
			"InstallDatabaseSoftware" :"No",
			"DatabaseInstallationMediaS3Uri" :"s3://invalid_uri/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :"${data.aws_ami.suse.id}",
			"DatabasePrimaryHostname" :"hana-primary",
			"DatabaseInstanceType" :"r5.xlarge",
			"InstallAwsBackintAgent" :"No",
			"DisableDeploymentRollback" :"Yes",
			"SaveDeploymentArtifacts" :"No",
			"DatabaseAutomaticRecovery": "No",
			"DatabaseLogVolumeType": "gp3",
			"DatabaseDataVolumeType": "gp3"
		}
		`,
		"deployment_pattern": "SapHanaSingle",
		"workload_name":      "SAP",
	}
	TestSpecSapHanaMulti = map[string]string{ //unused
		"specification_template": `{
			"KeyPairName": "${aws_key_pair.key.key_name}",
			"AvailabilityZone1PrivateSubnet1Id": "${aws_subnet.private.0.id}",
			"VpcId": "${aws_vpc.test.id}",
			"Timezone" :"UTC",
			"EnableEbsVolumeEncryption" :"No",
			"CreateSecurityGroup": "No",
			"DatabaseSecurityGroupId": "${aws_security_group.test.id}",
			"SapSysGroupId" :"5002",
			"DatabaseSystemId" :"HDB",
			"DatabaseInstanceNumber": "00",
			"DatabasePassword" :"Password123@",
			"InstallDatabaseSoftware" :"Yes",
			"DatabaseInstallationMediaS3Uri" :"%[1]s/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :"${data.aws_ami.suse.id}",
			"DatabasePrimaryHostname" :"hana-primary",
			"DatabaseSubordinateHostnames" :"hanasub1,hanasub2",
			"DatabaseHostCount" :"3",
			"DatabaseInstanceType" :"r5.2xlarge",
			"InstallAwsBackintAgent" :"No",
			"DisableDeploymentRollback" :"Yes",
			"SaveDeploymentArtifacts" :"No",
			"DatabaseAutomaticRecovery": "No",
			"DatabaseLogVolumeType": "gp3",
			"DatabaseDataVolumeType": "gp3"
		}
		`,
		"deployment_pattern": "SapHanaMulti",
		"workload_name":      "SAP",
	}
	TestSpecSapHanaHA = map[string]string{ //unused
		"specification_template": `{
			"KeyPairName": "${aws_key_pair.key.key_name}",
			"AvailabilityZone1PrivateSubnet1Id": "${aws_subnet.private.0.id}",
			"AvailabilityZone2PrivateSubnet1Id": "${aws_subnet.private.1.id}",
			"VpcId": "${aws_vpc.test.id}",
			"Timezone" :"UTC",
			"EnableEbsVolumeEncryption" :"No",
			"CreateSecurityGroup": "No",
			"DatabaseSecurityGroupId": "${aws_security_group.test.id}",
			"SapSysGroupId" :"5002",
			"DatabaseSystemId" :"HDB",
			"DatabaseInstanceNumber": "00",
			"DatabasePassword" :"Password123@",
			"InstallDatabaseSoftware" :"Yes",
			"DatabaseInstallationMediaS3Uri" :"%[1]s/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :"${data.aws_ami.suse.id}",
			"DatabasePrimaryHostname" :"hana-primary",
			"DatabaseSecondaryHostname" :"hana-secondary",
			"DatabaseVirtualIpAddress": "192.168.100.100",
			"DatabasePrimarySiteName": "primsite",
			"DatabaseSecondarySiteName": "secsite",
			"DatabasePacemakerTag": "pacemaker",
			"DatabaseInstanceType" :"r5.2xlarge",
			"InstallAwsBackintAgent" :"No",
			"DisableDeploymentRollback" :"Yes",
			"SaveDeploymentArtifacts" :"No",
			"DatabaseLogVolumeType": "gp3",
			"DatabaseDataVolumeType": "gp3"
		}
		`,
		"deployment_pattern": "SapHanaHA",
		"workload_name":      "SAP",
	}
	TestSpecSapNWOnHanaHA = map[string]string{ //unused
		"specification_template": `{
			"KeyPairName": "${aws_key_pair.key.key_name}",
			"VpcId": "${aws_vpc.test.id}",
			"AvailabilityZone1PrivateSubnet1Id": "${aws_subnet.private.0.id}",
			"AvailabilityZone2PrivateSubnet1Id": "${aws_subnet.private.1.id}",
			"Timezone" :"UTC",
			"EnableEbsVolumeEncryption" :"No",
			"CreateSecurityGroup" :"No",
			"DatabaseSecurityGroupId" :"${aws_security_group.test.id}",
			"ApplicationSecurityGroupId" :"${aws_security_group.test.id}",
			"SidAdmUserId" :"7002",
			"SapSysGroupId" :"5001",
			"DatabaseSystemId" :"HYD",
			"SapSid" :"S4K",
			"DatabaseInstanceNumber" :"30",
			"InstallAwsBackintAgent" :"No",
			"AscsOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"AscsAmiId" :"${data.aws_ami.suse.id}",
			"AscsInstanceType" :"r5.2xlarge",
			"AscsHostname" :"hallp123",
			"AscsInstanceNumber" :"10",
			"ErsOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"ErsAmiId" :"${data.aws_ami.suse.id}",
			"ErsInstanceType" :"r5.2xlarge",
			"ErsHostname" :"hallq123",
			"ErsInstanceNumber" :"11",
			"PasOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"PasAmiId" :"${data.aws_ami.suse.id}",
			"PasInstanceType" :"r5.2xlarge",
			"PasHostname" :"hallv123",
			"InstallAas" :"Yes",
			"AasHostCount" :"2",
			"AasInstanceType" :"r5.2xlarge",
			"AasHostnames" :"halgaa123,halhaa124",
			"AasAutomaticRecovery" :"Yes",
			"DatabaseOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :"${data.aws_ami.suse.id}",
			"DatabaseInstanceType" :"r5.8xlarge",
			"DatabasePrimaryHostname" :"haljdbpri",
			"DatabaseSecondaryHostname" :"hakldbsec",
			"SapPassword" :"Password123@",
			"InstallSap":"Yes",
			"DatabaseVirtualIpAddress" :"192.168.100.100",
			"DatabasePrimarySiteName" :"dbsitfp",
			"DatabaseSecondarySiteName" :"dbsitshc",
			"DatabasePacemakerTag" :"hallpag",
			"SetupTransportDomainController" :"No",
			"SapInstallationSpecifications": "{\"parameters\":{\"PRODUCT_ID\":\"saps4hana-2020\",\"HDB_SCHEMA_NAME\":\"SAPABAP1\",\"CI_INSTANCE_NR\":\"50\",\"ASCS_VIRTUAL_HOSTNAME\":\"hallvirascs\",\"ERS_VIRTUAL_HOSTNAME\":\"hallvirers\",\"ASCS_OVERLAY_IP\":\"192.168.100.110\",\"ERS_OVERLAY_IP\":\"192.168.100.120\",\"DB_VIRTUAL_HOSTNAME\":\"halldbvr1\",\"SAP_PACEMAKER_TAG\":\"hllpacetag\",\"SAPINST_CD_SAPCAR\":\"%[1]s/SAPCAR\",\"SAPINST_CD_SWPM\":\"%[1]s/SWPM\",\"SAPINST_CD_KERNEL\":\"%[1]s/Kernel\",\"SAPINST_CD_LOAD\":\"%[1]s/Exports\",\"SAPINST_CD_RDBMS\":\"%[1]s/HANA_DB_Software\",\"SAPINST_CD_RDBMS_CLIENT\":\"%[1]s/HANA_Client_Software\"}, \"onFailureBehaviour\": \"CONTINUE\"}",
			"SaveDeploymentArtifacts" :"No",
			"DisableDeploymentRollback" :"Yes",
			"ApplicationDataVolumeType": "gp3",    
			"DatabaseLogVolumeType": "gp3", 
			"DatabaseDataVolumeType": "gp3",
			"PasAutomaticRecovery": "Yes"
		  }
		`,
		"deployment_pattern": "SapNWOnHanaHA",
		"workload_name":      "SAP",
	}
)
