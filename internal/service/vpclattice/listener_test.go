// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeListener_defaultActionUpdate(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_fixedResponseHTTPS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.status_code", "404"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
				),
			},
			{
				Config: testAccListenerConfig_forwardTargetGroupHTTPSServiceID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "100"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
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

func TestAccVPCLatticeListener_fixedResponseHTTP(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_fixedResponseHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.status_code", "404"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
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

func TestAccVPCLatticeListener_fixedResponseHTTPS(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_fixedResponseHTTPS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.fixed_response.0.status_code", "404"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.+`)),
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

func TestAccVPCLatticeListener_forwardTLSPassthrough(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardTLSPassthrough(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TLS_PASSTHROUGH"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "80"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
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

func TestAccVPCLatticeListener_forwardHTTPTargetGroup(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardTargetGroupHTTPServiceID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "100"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
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

func TestAccVPCLatticeListener_forwardHTTPTargetGroupCustomPort(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardTargetGroupHTTPServiceIDCustomPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "100"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
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

func TestAccVPCLatticeListener_forwardHTTPSTargetGroupARN(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardTargetGroupHTTPServiceARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttrPair(resourceName, "service_arn", serviceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "100"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
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

func TestAccVPCLatticeListener_forwardHTTPSTargetGroupCustomPort(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardTargetGroupHTTPSServiceIDCustomPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8443"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTPS"),
					resource.TestCheckResourceAttrPair(resourceName, "service_arn", serviceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "100"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
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

func TestAccVPCLatticeListener_forwardHTTPMultipleTargetGroups(t *testing.T) {
	ctx := acctest.Context(t)
	targetGroupName1 := fmt.Sprintf("testtargetgroup-%s", sdkacctest.RandString(10))

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"
	serviceName := "aws_vpclattice_service.test"
	targetGroupResourceName := "aws_vpclattice_target_group.test"
	targetGroup1ResourceName := "aws_vpclattice_target_group.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardMultiTargetGroupHTTP(rName, targetGroupName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", serviceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.0.target_group_identifier", targetGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.0.weight", "80"),
					resource.TestCheckResourceAttrPair(resourceName, "default_action.0.forward.0.target_groups.1.target_group_identifier", targetGroup1ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.forward.0.target_groups.1.weight", "20"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
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

func TestAccVPCLatticeListener_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_forwardTargetGroupHTTPServiceID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceListener(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeListener_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var listener vpclattice.GetListenerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_tags1(rName, "key0", "value0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.key0", "value0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service\/svc-.*\/listener\/listener-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerConfig_tags2(rName, "key0", "value0updated", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.key0", "value0updated"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccListenerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckListenerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_listener" {
				continue
			}

			_, err := tfvpclattice.FindListenerByTwoPartKey(ctx, conn, rs.Primary.Attributes["listener_id"], rs.Primary.Attributes["service_identifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Listener %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckListenerExists(ctx context.Context, name string, listener *vpclattice.GetListenerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListener, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := tfvpclattice.FindListenerByTwoPartKey(ctx, conn, rs.Primary.Attributes["listener_id"], rs.Primary.Attributes["service_identifier"])

		if err != nil {
			return err
		}

		*listener = *resp

		return nil
	}
}

func testAccListenerConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}
`, rName))
}

func testAccListenerConfig_fixedResponseHTTP(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}
`, rName))
}

func testAccListenerConfig_fixedResponseHTTPS(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTPS"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}
`, rName))
}

func testAccListenerConfig_forwardTLSPassthrough(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name               = %[1]q
  auth_type          = "AWS_IAM"
  custom_domain_name = "example.com"
}

resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "TCP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "TLS_PASSTHROUGH"
  port               = 8443
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 80
      }
    }
  }
}
`, rName))
}

func testAccListenerConfig_forwardMultiTargetGroupHTTP(rName string, targetGroupName1 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test1" {
  name = %[2]q
  type = "INSTANCE"

  config {
    port           = 8080
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 80
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test1.id
        weight                  = 20
      }
    }
  }
}
`, rName, targetGroupName1))
}

func testAccListenerConfig_forwardTargetGroupHTTPServiceID(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
}
`, rName))
}

func testAccListenerConfig_forwardTargetGroupHTTPServiceIDCustomPort(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  port               = 8080
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
}
`, rName))
}

func testAccListenerConfig_forwardTargetGroupHTTPServiceARN(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name        = %[1]q
  protocol    = "HTTPS"
  service_arn = aws_vpclattice_service.test.arn
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
}`, rName))
}

func testAccListenerConfig_forwardTargetGroupHTTPSServiceID(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTPS"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
}`, rName))
}

func testAccListenerConfig_forwardTargetGroupHTTPSServiceIDCustomPort(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  port               = 8443
  protocol           = "HTTPS"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
}`, rName))
}

func testAccListenerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccListenerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccListenerConfig_basic(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test.id
        weight                  = 100
      }
    }
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
