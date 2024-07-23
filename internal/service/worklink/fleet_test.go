// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package worklink_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfworklink "github.com/hashicorp/terraform-provider-aws/internal/service/worklink"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkLinkFleet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "company_code"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
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

func TestAccWorkLinkFleet_displayName(t *testing.T) {
	ctx := acctest.Context(t)
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_displayName(suffix, "display1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "display1"),
				),
			},
			{
				Config: testAccFleetConfig_displayName(suffix, "display2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "display2"),
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

func TestAccWorkLinkFleet_optimizeForEndUserLocation(t *testing.T) {
	ctx := acctest.Context(t)
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_optimizeForEndUserLocation(suffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", acctest.CtFalse),
				),
			},
			{
				Config: testAccFleetConfig_optimizeForEndUserLocation(suffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", acctest.CtTrue),
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

func TestAccWorkLinkFleet_auditStreamARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_auditStreamARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "audit_stream_arn", "aws_kinesis_stream.test_stream", names.AttrARN),
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

func TestAccWorkLinkFleet_network(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_network(rName, "192.168.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", acctest.Ct1),
				),
			},
			{
				Config: testAccFleetConfig_network(rName, "10.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccFleetConfig_basic(rName),
				ExpectError: regexache.MustCompile(`Company Network Configuration cannot be removed`),
			},
		},
	})
}

func TestAccWorkLinkFleet_deviceCaCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_deviceCaCertificate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "device_ca_certificate", regexache.MustCompile("^-----BEGIN CERTIFICATE-----")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_ca_certificate", ""),
				),
			},
		},
	})
}

func TestAccWorkLinkFleet_identityProvider(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_identityProvider(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_provider.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "identity_provider.0.type", "SAML"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccFleetConfig_basic(rName),
				ExpectError: regexache.MustCompile(`Identity Provider Configuration cannot be removed`),
			},
		},
	})
}

func TestAccWorkLinkFleet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkLinkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName),
					testAccCheckFleetDisappears(ctx, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetDisappears(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn(ctx)

		input := &worklink.DeleteFleetInput{
			FleetArn: aws.String(rs.Primary.ID),
		}

		if _, err := conn.DeleteFleetWithContext(ctx, input); err != nil {
			return err
		}

		stateConf := &retry.StateChangeConf{
			Pending:    []string{"DELETING"},
			Target:     []string{"DELETED"},
			Refresh:    tfworklink.FleetStateRefresh(ctx, conn, rs.Primary.ID),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err := stateConf.WaitForStateContext(ctx)

		return err
	}
}

func testAccCheckFleetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_worklink_fleet" {
				continue
			}

			_, err := conn.DescribeFleetMetadataWithContext(ctx, &worklink.DescribeFleetMetadataInput{
				FleetArn: aws.String(rs.Primary.ID),
			})

			if err != nil {
				// Return nil if the Worklink Fleet is already destroyed
				if tfawserr.ErrCodeEquals(err, worklink.ErrCodeResourceNotFoundException) {
					return nil
				}

				return err
			}
			return fmt.Errorf("Worklink Fleet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFleetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worklink Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn(ctx)
		_, err := conn.DescribeFleetMetadataWithContext(ctx, &worklink.DescribeFleetMetadataInput{
			FleetArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn(ctx)

	input := &worklink.ListFleetsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListFleetsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFleetConfig_basic(r string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"
}
`, r)
}

func testAccFleetConfig_displayName(r, displayName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name         = "tf-worklink-fleet-%s"
  display_name = "%s"
}
`, r, displayName)
}

func testAccFleetConfig_optimizeForEndUserLocation(r string, b bool) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name                           = "tf-worklink-fleet-%s"
  optimize_for_end_user_location = %t
}
`, r, b)
}

func testAccFleetNetworkConfig_Base(rName, cidrBlock string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "%s"

  tags = {
    Name = %q
  }
}

resource "aws_security_group" "test" {
  name        = "tf_test_foo"
  description = "foo"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol  = "icmp"
    from_port = -1
    to_port   = -1
    self      = true
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %q
  }
}
`, cidrBlock, rName, rName)
}

func testAccFleetConfig_network(r, cidrBlock string) string {
	return acctest.ConfigCompose(
		testAccFleetNetworkConfig_Base(r, cidrBlock),
		fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  network {
    vpc_id             = aws_vpc.test.id
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }
}
`, r))
}

func testAccFleetConfig_auditStreamARN(r string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
  name        = "AmazonWorkLink-%[1]s_kinesis_test"
  shard_count = 1
}

resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%[1]s"

  audit_stream_arn = aws_kinesis_stream.test_stream.arn
}
`, r)
}

func testAccFleetConfig_deviceCaCertificate(rName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%[1]s"

  device_ca_certificate = file("./test-fixtures/worklink-device-ca-certificate.pem")
}
`, rName)
}

func testAccFleetConfig_identityProvider(rName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%[1]s"

  identity_provider {
    type          = "SAML"
    saml_metadata = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
  }
}
`, rName, idpEntityId)
}
