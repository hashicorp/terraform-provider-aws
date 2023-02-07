package networkfirewall_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkfirewall"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkFirewallFirewall_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_policy_arn", policyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_mapping.*", map[string]string{
						"ip_address_type": networkfirewall.IPAddressTypeIpv4,
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
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

func TestAccNetworkFirewallFirewall_dualstackSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test.0"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_dualstackSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_policy_arn", policyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "subnet_mapping.*", map[string]string{
						"ip_address_type": networkfirewall.IPAddressTypeDualstack,
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
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

func TestAccNetworkFirewallFirewall_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccFirewallConfig_description(rName, "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "updated"),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccNetworkFirewallFirewall_deleteProtection(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{

			{
				Config: testAccFirewallConfig_deleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
				),
			},
			{
				Config: testAccFirewallConfig_deleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "true"),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
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

func TestAccNetworkFirewallFirewall_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_encryptionConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "0"),
				),
			},
			{
				Config: testAccFirewallConfig_encryptionConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_SubnetMappings_updateSubnet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	subnetResourceName := "aws_subnet.test.0"
	updateSubnetResourceName := "aws_subnet.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{

			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
				),
			},
			{
				Config: testAccFirewallConfig_updateSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", updateSubnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", updateSubnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", updateSubnetResourceName, "id"),
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

func TestAccNetworkFirewallFirewall_SubnetMappings_updateMultipleSubnets(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	subnetResourceName := "aws_subnet.test.0"
	updateSubnetResourceName := "aws_subnet.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
				),
			},
			{
				Config: testAccFirewallConfig_updateMultipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", updateSubnetResourceName, "availability_zone"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", updateSubnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", updateSubnetResourceName, "id"),
				),
			},
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
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

func TestAccNetworkFirewallFirewall_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFirewallConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewall_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceFirewall(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFirewallDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_firewall" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn()

			_, err := tfnetworkfirewall.FindFirewallByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Firewall %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFirewallExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Firewall ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn()

		_, err := tfnetworkfirewall.FindFirewallByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn()

	input := &networkfirewall.ListFirewallsInput{}

	_, err := conn.ListFirewallsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFirewallConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName))
}

func testAccFirewallConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccFirewallConfig_deleteProtection(rName string, deleteProtection bool) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  delete_protection   = %[1]t
  name                = %[2]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, deleteProtection, rName))
}

func testAccFirewallConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccFirewallConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccFirewallConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  description         = %[2]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName, description))
}

func testAccFirewallConfig_updateSubnet(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }
}
`, rName))
}

func testAccFirewallConfig_updateMultipleSubnets(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }
}
`, rName))
}

func testAccFirewallConfig_encryptionConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccFirewallConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type   = "CUSTOMER_KMS"
  }

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccFirewallConfig_dualstackSubnet(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id       = aws_subnet.test[0].id
    ip_address_type = "DUALSTACK"
  }
}
`, rName))
}
