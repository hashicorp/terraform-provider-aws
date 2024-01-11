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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/launchwizard"
	"github.com/aws/aws-sdk-go-v2/service/launchwizard/types"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	awstypes "github.com/aws/aws-sdk-go-v2/service/launchwizard/types"
	tflaunchwizard "github.com/hashicorp/terraform-provider-aws/internal/service/launchwizard"
)

/**
 Before running this test, the following two ENV variables must be set:

 LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI     - URI of S3 Partition containing SAP Installation Media (without trailing slash)

 Expected structure:

	Inside the S3 Partition, the following structure is expected with content matching the NW version defined below (https://docs.aws.amazon.com/launchwizard/latest/userguide/launch-wizard-sap-software-install-details.html)

	/HANA_DB_Software
	/SWPM
	/SAPCAR
	/Exports
	/Kernel
	/HANA_Client_Software
**/

// The following are example test specifications for the different SAP deployment patterns. We currently only use the HANA Single Node deployment pattern for testing as it should suffice to verify the functionality of the provider.
// The other patterns are not tested as they are very long running and error prone while not adding any additional value to the test coverage.

var (
	TestSpecSapHanaSingle = map[string]string{
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
			"DatabaseInstallationMediaS3Uri" :"%[1]s\/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :%[2]q,
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
	TestSpecSapHanaMulti = map[string]string{
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
			"DatabaseInstallationMediaS3Uri" :"%[1]s\/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :%[2]q,
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
	TestSpecSapHanaHA = map[string]string{
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
			"DatabaseInstallationMediaS3Uri" :"%[1]s\/HANA_DB_Software",
			"DatabaseOperatingSystem": "SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :%[2]q,
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
	TestSpecSapNWOnHanaHA = map[string]string{
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
			"AscsAmiId" :%[2]q,
			"AscsInstanceType" :"r5.2xlarge",
			"AscsHostname" :"hallp123",
			"AscsInstanceNumber" :"10",
			"ErsOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"ErsAmiId" :%[2]q,
			"ErsInstanceType" :"r5.2xlarge",
			"ErsHostname" :"hallq123",
			"ErsInstanceNumber" :"11",
			"PasOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"PasAmiId" :%[2]q,
			"PasInstanceType" :"r5.2xlarge",
			"PasHostname" :"hallv123",
			"InstallAas" :"Yes",
			"AasHostCount" :"2",
			"AasInstanceType" :"r5.2xlarge",
			"AasHostnames" :"halgaa123,halhaa124",
			"AasAutomaticRecovery" :"Yes",
			"DatabaseOperatingSystem" :"SuSE-Linux-15-SP4-For-SAP-HVM",
			"DatabaseAmiId" :%[2]q,
			"DatabaseInstanceType" :"r5.8xlarge",
			"DatabasePrimaryHostname" :"haljdbpri",
			"DatabaseSecondaryHostname" :"hakldbsec",
			"SapPassword" :"Password123@",
			"InstallSap":"Yes",
			"DatabaseVirtualIpAddress" :"10.0.0.10",
			"DatabasePrimarySiteName" :"dbsitfp",
			"DatabaseSecondarySiteName" :"dbsitshc",
			"DatabasePacemakerTag" :"hallpag",
			"SetupTransportDomainController" :"No",
			"SapInstallationSpecifications": "{\"parameters\":{\"PRODUCT_ID\":\"saps4hana-2020\",\"HDB_SCHEMA_NAME\":\"SAPABAP1\",\"CI_INSTANCE_NR\":\"50\",\"ASCS_VIRTUAL_HOSTNAME\":\"hallvirascs\",\"ERS_VIRTUAL_HOSTNAME\":\"hallvirers\",\"ASCS_OVERLAY_IP\":\"10.29.26.79\",\"ERS_OVERLAY_IP\":\"10.96.87.80\",\"DB_VIRTUAL_HOSTNAME\":\"halldbvr1\",\"SAP_PACEMAKER_TAG\":\"hllpacetag\",\"SAPINST_CD_SAPCAR\":\"%[1]s/SAPCAR\",\"SAPINST_CD_SWPM\":\"%[1]s/SWPM\",\"SAPINST_CD_KERNEL\":\"%[1]s/Kernel\",\"SAPINST_CD_LOAD\":\"%[1]s/Exports\",\"SAPINST_CD_RDBMS\":\"%[1]s/HANA_DB_Software\",\"SAPINST_CD_RDBMS_CLIENT\":\"%[1]s/HANA_Client_Software\"}, \"onFailureBehaviour\": \"CONTINUE\"}",
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

// func TestAccLaunchWizardDeployment_serial(t *testing.T) {
// 	t.Parallel()

// 	setEnvVars_TODO_REPLACE()

// 	//all test cases require a NAT gateway by design. running serial due to 5 EIP quota per region
// 	testCases := map[string]map[string]func(t *testing.T){
// 		"Deployment": {
// 			"basic_SapHanaSingle":      testAccLaunchWizardDeployment_basicSapHanaSingle,
// 			"disappears_SapHanaSingle": testAccLaunchWizardDeployment_disappearsSapHanaSingle,
// 			"basic_SapHanaHA":          testAccLaunchWizardDeployment_basicSapHanaHA,
// 		},
// 	}

// 	acctest.RunSerialTests2Levels(t, testCases, 0)
// }

func TestAccLaunchWizardDeployment_basicSapHanaSingle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	setEnvVars_TODO_REPLACE()

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	testSpec := TestSpecSapHanaSingle

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.LaunchWizardEndpointID)
			testAccPreCheckRegionAvailable(ctx, t)
			testAccPreCheckEnvironmentVariablesSet(t, testSpec["deployment_pattern"])
		},
		ErrorCheck:               acctest.ErrorCheck(t, "launchwizard"), //TODO: names.LaunchWizardEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"]),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_pattern", testSpec["deployment_pattern"]),
					resource.TestCheckResourceAttr(resourceName, "workload_name", testSpec["workload_name"]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"specifications.SapPassword",
				},
			},
		},
	})
}

func TestAccLaunchWizardDeployment_disappearsSapHanaSingle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	setEnvVars_TODO_REPLACE()

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	testSpec := TestSpecSapHanaSingle

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.LaunchWizardEndpointID)
			testAccPreCheckRegionAvailable(ctx, t)
			testAccPreCheckEnvironmentVariablesSet(t, testSpec["deployment_pattern"])
		},
		ErrorCheck:               acctest.ErrorCheck(t, "launchwizard"), //TODO: names.LaunchWizardEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"]),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_pattern", testSpec["deployment_pattern"]),
					resource.TestCheckResourceAttr(resourceName, "workload_name", testSpec["workload_name"]),
				),
			},
		},
	})
}

func TestAccLaunchWizardDeployment_basicSapHanaHA(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	setEnvVars_TODO_REPLACE()

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	testSpec := TestSpecSapHanaHA

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.LaunchWizardEndpointID)  //TODO: replace once the new Service is registered;
			testAccPreCheckRegionAvailable(ctx, t)
			testAccPreCheckEnvironmentVariablesSet(t, testSpec["deployment_pattern"])
		},
		ErrorCheck:               acctest.ErrorCheck(t, "launchwizard"), //TODO: replace once the new Service is registered; names.LaunchWizardEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"]),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_pattern", testSpec["deployment_pattern"]),
					resource.TestCheckResourceAttr(resourceName, "workload_name", testSpec["workload_name"]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"specifications.DatabasePassword",
				},
			},
		},
	})
}

func TestAccLaunchWizardDeployment_basicSapNWOnHanaHA(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	setEnvVars_TODO_REPLACE()

	var deployment launchwizard.GetDeploymentOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_launchwizard_deployment.test"

	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	testSpec := TestSpecSapNWOnHanaHA

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.LaunchWizardEndpointID)
			testAccPreCheckRegionAvailable(ctx, t)
			testAccPreCheckEnvironmentVariablesSet(t, testSpec["deployment_pattern"])
		},
		ErrorCheck:               acctest.ErrorCheck(t, "launchwizard"), //TODO: names.LaunchWizardEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, testSpec["specification_template"], testSpec["deployment_pattern"], testSpec["workload_name"]),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_pattern", testSpec["deployment_pattern"]),
					resource.TestCheckResourceAttr(resourceName, "workload_name", testSpec["workload_name"]),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"specifications.DatabasePassword",
				},
			},
		},
	})
}

func testAccDeploymentConfig_basic(rName string, specTemplate string, deploymentPattern string, workloadName string) string {

	s3_uri := os.Getenv("LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI")

	specification_template_computed := fmt.Sprintf(specTemplate, s3_uri, "ami-0b3d1b3396c00b1cd") //TODO: get from data and ENV Variables

	return acctest.ConfigCompose(
		testAccDeploymentConfigBase(rName),
		fmt.Sprintf(`

resource "aws_launchwizard_deployment" "test" {
  name             = %[1]q
  deployment_pattern = %[3]q
  workload_name = %[4]q
  specifications = %[2]s

  

}
`, rName, specification_template_computed, deploymentPattern, workloadName))
}

func testAccDeploymentConfigBase(rName string) string {
	return fmt.Sprintf(`

	resource "aws_vpc" "test" {
		cidr_block           = "10.0.0.0/16"
		enable_dns_hostnames = true
		enable_dns_support   = true
	  
		tags = {
		  Name                          = %[1]q
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
		  Name = %[1]q
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
		count = 2
	  
		availability_zone = data.aws_availability_zones.available.names[count.index]
		cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 2)
		vpc_id            = aws_vpc.test.id
	  
		tags = {
		  Name                          = %[1]q
		}
	  }
	  
	  resource "aws_eip" "public" {
		count      = 2
		depends_on = [aws_internet_gateway.test]
	  
		domain = "vpc"
	  
		tags = {
		  Name = %[1]q
		}
	  }
	  
	  resource "aws_nat_gateway" "public" {
		count = 2
	  
		allocation_id = aws_eip.public[count.index].id
		subnet_id     = aws_subnet.public[count.index].id
	  
		tags = {
		  Name = %[1]q
		}
	  }
	  
	  resource "aws_route_table" "private" {
		count = 2
	  
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
		count = 2
	  
		subnet_id      = aws_subnet.private[count.index].id
		route_table_id = aws_route_table.private[count.index].id
	  }
	  
	  resource "aws_subnet" "public" {
		count = 2
	  
		availability_zone = data.aws_availability_zones.available.names[count.index]
		cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
		vpc_id            = aws_vpc.test.id
	  
		tags = {
		  Name                          = %[1]q
		}
	  }

	  resource "aws_security_group" "test" {
		name        = %[1]q
		vpc_id      = aws_vpc.test.id
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


	`, rName)
}

func setEnvVars_TODO_REPLACE() {
	//TODO: remove...
	os.Setenv("TF_ACC", "1")
	os.Setenv("TF_LOG", "info")
	os.Setenv("GOFLAGS", "-mod=readonly")
	os.Setenv("AWS_PROFILE", "aws_provider_profile")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
	os.Setenv("AWS_ALTERNATE_PROFILE", "aws_alternate_profile")
	os.Setenv("AWS_ALTERNATE_REGION", "us-east-1")
	os.Setenv("AWS_THIRD_REGION", "us-east-2")
	os.Setenv("ACM_CERTIFICATE_ROOT_DOMAIN", "terraform-provider-aws-acctest-acm.com")

	os.Setenv("LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI", `s3://2023-trc-sap-installation-files/launch_wizard_sources2`)
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
			if errs.IsA[*types.ResourceNotFoundException](err) {
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

func testAccPreCheckEnvironmentVariablesSet(t *testing.T, deploymentPattern string) {
	if v := os.Getenv("LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI"); v == "" {
		t.Fatal("env variable LAUNCHWIZARD_SAP_INST_MEDIA_S3_URI must be set for acceptance tests")
	}
}
