package storagegateway_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccStorageGatewayGateway_GatewayType_cached(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeCached(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "CACHED"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_security_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_GatewayType_fileFSxSMB(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeFileFSxSMB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "FILE_FSX_SMB"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_GatewayType_fileS3(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeFileS3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "FILE_S3"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_GatewayType_stored(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeStored(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "STORED"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_GatewayType_vtl(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeVtl(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "VTL"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_tags(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGatewayConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccStorageGatewayGateway_gatewayName(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeFileS3(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName1),
				),
			},
			{
				Config: testAccGatewayConfig_typeFileS3(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_cloudWatchLogs(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"
	resourceName2 := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_logGroup(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", resourceName2, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_gatewayTimezone(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_timezone(rName, "GMT-1:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT-1:00"),
				),
			},
			{
				Config: testAccGatewayConfig_timezone(rName, "GMT-2:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT-2:00"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_gatewayVPCEndpoint(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_vpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_vpc_endpoint", vpcEndpointResourceName, "dns_entry.0.dns_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_smbActiveDirectorySettings(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbActiveDirectorySettings(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.username", "Administrator"),
					resource.TestCheckResourceAttrSet(resourceName, "smb_active_directory_settings.0.active_directory_status"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_SMBActiveDirectorySettings_timeout(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbActiveDirectorySettingsTimeout(rName, domainName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.timeout_in_seconds", "50"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_smbMicrosoftActiveDirectorySettings(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"
	domainName := acctest.RandomDomainName()
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbMicrosoftActiveDirectorySettings(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.username", username),
					resource.TestCheckResourceAttrSet(resourceName, "smb_active_directory_settings.0.active_directory_status"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_SMBMicrosoftActiveDirectorySettings_timeout(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbMicrosoftActiveDirectorySettingsTimeout(rName, domainName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.timeout_in_seconds", "50"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_smbGuestPassword(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbGuestPassword(rName, "myguestpassword1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", "myguestpassword1"),
				),
			},
			{
				Config: testAccGatewayConfig_smbGuestPassword(rName, "myguestpassword2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", "myguestpassword2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_guest_password"},
			},
		},
	})
}

func TestAccStorageGatewayGateway_smbSecurityStrategy(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbSecurityStrategy(rName, "ClientSpecified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_security_strategy", "ClientSpecified"),
					resource.TestCheckResourceAttr(resourceName, "smb_file_share_visibility", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_smbSecurityStrategy(rName, "MandatorySigning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_security_strategy", "MandatorySigning"),
				),
			},
		},
	})
}

func TestAccStorageGatewayGateway_smbVisibility(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_smbVisibility(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_file_share_visibility", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_smbVisibility(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_file_share_visibility", "false"),
				),
			},
			{
				Config: testAccGatewayConfig_smbVisibility(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_file_share_visibility", "true"),
				),
			},
		},
	})
}

func TestAccStorageGatewayGateway_disappears(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_typeCached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					acctest.CheckResourceDisappears(acctest.Provider, tfstoragegateway.ResourceGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccStorageGatewayGateway_bandwidthUpload(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_bandwidthUpload(rName, 102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "102400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_bandwidthUpload(rName, 2*102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "204800"),
				),
			},
			{
				Config: testAccGatewayConfig_typeCached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "0"),
				),
			},
		},
	})
}

func TestAccStorageGatewayGateway_bandwidthDownload(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_bandwidthDownload(rName, 102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "102400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_bandwidthDownload(rName, 2*102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "204800"),
				),
			},
			{
				Config: testAccGatewayConfig_typeCached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "0"),
				),
			},
		},
	})
}

func TestAccStorageGatewayGateway_bandwidthAll(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_bandwidthAll(rName, 102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "102400"),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "102400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_bandwidthAll(rName, 2*102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "204800"),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "204800"),
				),
			},
			{
				Config: testAccGatewayConfig_typeCached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "0"),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "0"),
				),
			},
		},
	})
}

func TestAccStorageGatewayGateway_maintenanceStartTime(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_maintenanceStartTime(rName, 22, 0, "3", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.hour_of_day", "22"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.minute_of_hour", "0"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.day_of_week", "3"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.day_of_month", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccGatewayConfig_maintenanceStartTime(rName, 21, 10, "", "12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.hour_of_day", "21"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.minute_of_hour", "10"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.day_of_week", ""),
					resource.TestCheckResourceAttr(resourceName, "maintenance_start_time.0.day_of_month", "12"),
				),
			},
		},
	})
}

func testAccCheckGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_gateway" {
			continue
		}

		_, err := tfstoragegateway.FindGatewayByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Storage Gateway Gateway %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckGatewayExists(resourceName string, gateway *storagegateway.DescribeGatewayInformationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

		output, err := tfstoragegateway.FindGatewayByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*gateway = *output

		return nil
	}
}

// testAcc_VPCBase provides a publicly accessible subnet
// and security group, suitable for Storage Gateway EC2 instances of any type
func testAcc_VPCBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

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

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
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

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAcc_FileGatewayBase(rName string) string {
	return acctest.ConfigCompose(
		testAcc_VPCBase(rName),
		// Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "m5.xlarge", "m4.xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-file.html
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAcc_TapeAndVolumeGatewayBase(rName string) string {
	return acctest.ConfigCompose(
		testAcc_VPCBase(rName),
		// Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "m5.xlarge", "m4.xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-common.html
# NOTE: CACHED, STORED, and VTL Gateway Types share the same AMI
data "aws_ssm_parameter" "aws_service_storagegateway_ami_CACHED_latest" {
  name = "/aws/service/storagegateway/ami/CACHED/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_CACHED_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccGatewayConfig_typeCached(rName string) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"
}
`, rName))
}

func testAccGatewayConfig_typeFileFSxSMB(rName string) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_FSX_SMB"
}
`, rName))
}

func testAccGatewayConfig_typeFileS3(rName string) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
}
`, rName))
}

func testAccGatewayConfig_logGroup(rName string) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address       = aws_instance.test.public_ip
  gateway_name             = %[1]q
  gateway_timezone         = "GMT"
  gateway_type             = "FILE_S3"
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
}
`, rName))
}

func testAccGatewayConfig_typeStored(rName string) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "STORED"
}
`, rName))
}

func testAccGatewayConfig_typeVtl(rName string) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "VTL"
}
`, rName))
}

func testAccGatewayConfig_timezone(rName, gatewayTimezone string) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = %[2]q
  gateway_type       = "FILE_S3"
}
`, rName, gatewayTimezone))
}

func testAccGatewayConfig_vpcEndpoint(rName string) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
data "aws_vpc_endpoint_service" "storagegateway" {
  service = "storagegateway"
}

resource "aws_vpc_endpoint" "test" {
  security_group_ids = [aws_security_group.test.id]
  service_name       = data.aws_vpc_endpoint_service.storagegateway.service_name
  subnet_ids         = [aws_subnet.test.id]
  vpc_endpoint_type  = data.aws_vpc_endpoint_service.storagegateway.service_type
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address   = aws_instance.test.public_ip
  gateway_name         = %[1]q
  gateway_timezone     = "GMT"
  gateway_type         = "CACHED"
  gateway_vpc_endpoint = aws_vpc_endpoint.test.dns_entry[0].dns_name
}
`, rName))
}

func testAccGatewayConfig_DirectoryServiceSimpleDirectory(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}

`, rName, domainName)
}

func testAccGatewayConfig_DirectoryServiceMicrosoftAD(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}

`, rName, domainName)
}

func testAccGatewaySMBActiveDirectorySettingsBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		// Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", "m5.xlarge", "m4.xlarge"),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
# Directory Service Directories must be deployed across multiple EC2 Availability Zones
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

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

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options" "test" {
  domain_name         = aws_directory_service_directory.test.name
  domain_name_servers = aws_directory_service_directory.test.dns_ip_addresses

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options_association" "test" {
  dhcp_options_id = aws_vpc_dhcp_options.test.id
  vpc_id          = aws_vpc.test.id
}

# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-file.html
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test, aws_vpc_dhcp_options_association.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccGatewayConfig_smbActiveDirectorySettings(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccGatewaySMBActiveDirectorySettingsBaseConfig(rName),
		testAccGatewayConfig_DirectoryServiceSimpleDirectory(rName, domainName),
		fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Administrator"
  }
}
`, rName))
}

func testAccGatewayConfig_smbActiveDirectorySettingsTimeout(rName, domainName string, timeout int) string {
	return acctest.ConfigCompose(
		testAccGatewaySMBActiveDirectorySettingsBaseConfig(rName),
		testAccGatewayConfig_DirectoryServiceSimpleDirectory(rName, domainName),
		fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name        = aws_directory_service_directory.test.name
    password           = aws_directory_service_directory.test.password
    username           = "Administrator"
    timeout_in_seconds = %[2]d
  }
}
`, rName, timeout))
}

func testAccGatewayConfig_smbMicrosoftActiveDirectorySettings(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccGatewaySMBActiveDirectorySettingsBaseConfig(rName),
		testAccGatewayConfig_DirectoryServiceMicrosoftAD(rName, domainName),
		fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Admin"
  }
}
`, rName))
}

func testAccGatewayConfig_smbMicrosoftActiveDirectorySettingsTimeout(rName, domainName string, timeout int) string {
	return acctest.ConfigCompose(
		testAccGatewaySMBActiveDirectorySettingsBaseConfig(rName),
		testAccGatewayConfig_DirectoryServiceMicrosoftAD(rName, domainName),
		fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name        = aws_directory_service_directory.test.name
    password           = aws_directory_service_directory.test.password
    username           = "Admin"
    timeout_in_seconds = %[2]d
  }
}
`, rName, timeout))
}

func testAccGatewayConfig_smbGuestPassword(rName, smbGuestPassword string) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
  smb_guest_password = %[2]q
}
`, rName, smbGuestPassword))
}

func testAccGatewayConfig_smbSecurityStrategy(rName, strategy string) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address    = aws_instance.test.public_ip
  gateway_name          = %[1]q
  gateway_timezone      = "GMT"
  gateway_type          = "FILE_S3"
  smb_security_strategy = %[2]q
}
`, rName, strategy))
}

func testAccGatewayConfig_smbVisibility(rName string, visible bool) string {
	return acctest.ConfigCompose(testAcc_FileGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address        = aws_instance.test.public_ip
  gateway_name              = %[1]q
  gateway_timezone          = "GMT"
  gateway_type              = "FILE_S3"
  smb_file_share_visibility = %[2]t
}
`, rName, visible))
}

func testAccGatewayConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccGatewayConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccGatewayConfig_bandwidthUpload(rName string, rate int) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address                        = aws_instance.test.public_ip
  gateway_name                              = %[1]q
  gateway_timezone                          = "GMT"
  gateway_type                              = "CACHED"
  average_upload_rate_limit_in_bits_per_sec = %[2]d
}
`, rName, rate))
}

func testAccGatewayConfig_bandwidthDownload(rName string, rate int) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address                          = aws_instance.test.public_ip
  gateway_name                                = %[1]q
  gateway_timezone                            = "GMT"
  gateway_type                                = "CACHED"
  average_download_rate_limit_in_bits_per_sec = %[2]d
}
`, rName, rate))
}

func testAccGatewayConfig_bandwidthAll(rName string, rate int) string {
	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address                          = aws_instance.test.public_ip
  gateway_name                                = %[1]q
  gateway_timezone                            = "GMT"
  gateway_type                                = "CACHED"
  average_upload_rate_limit_in_bits_per_sec   = %[2]d
  average_download_rate_limit_in_bits_per_sec = %[2]d
}
`, rName, rate))
}

func testAccGatewayConfig_maintenanceStartTime(rName string, hourOfDay, minuteOfHour int, dayOfWeek, dayOfMonth string) string {
	if dayOfWeek == "" {
		dayOfWeek = strconv.Quote(dayOfWeek)
	}
	if dayOfMonth == "" {
		dayOfMonth = strconv.Quote(dayOfMonth)
	}

	return acctest.ConfigCompose(testAcc_TapeAndVolumeGatewayBase(rName), fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  maintenance_start_time {
    hour_of_day    = %[2]d
    minute_of_hour = %[3]d
    day_of_week    = %[4]s
    day_of_month   = %[5]s
  }
}
`, rName, hourOfDay, minuteOfHour, dayOfWeek, dayOfMonth))
}
