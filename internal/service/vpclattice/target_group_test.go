package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeTargetGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup vpclattice.GetTargetGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_basic(rName, "INSTANCE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("targetgroup/.+$")),
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
func TestAccVPCLatticeTargetGroup_full(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup vpclattice.GetTargetGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_fullIP(rName, "IP"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "config.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "config.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "config.0.ip_address_type", "IPV4"),
					resource.TestCheckResourceAttr(resourceName, "config.0.protocol_version", "HTTP1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.health_check.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.health_check.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "config.0.health_check.0.timeout", "5"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("targetgroup/.+$")),
				),
			},
			{
				Config: testAccTargetGroupConfig_fulllambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("targetgroup/.+$")),
				),
			},
			{
				Config: testAccTargetGroupConfig_fullInstance(rName, "INSTANCE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "config.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "config.0.protocol", "HTTP"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("targetgroup/.+$")),
				),
			},
			{
				Config: testAccTargetGroupConfig_fullAlb(rName, "ALB"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "config.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "config.0.protocol", "HTTPS"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "vpc-lattice", regexp.MustCompile("targetgroup/.+$")),
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

func TestAccVPCLatticeTargetGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup vpclattice.GetTargetGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupConfig_fulllambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceTargetGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTargetGroupConfig_fulllambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "LAMBDA"
}
`, rName)
}

func testAccTargetGroupConfig_fullIP(rName, rType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
	name     = %[1]q
	type     = %[2]q

	config {
	  port             = 443
	  protocol         = "HTTPS"
	  vpc_identifier   =  aws_vpc.test.id
	  ip_address_type  = "IPV4"
	  protocol_version = "HTTP1"

	  health_check {
		enabled             	  = false
		interval            	  = 30
		timeout             	  = 5
		healthy_threshold   	  = 2
		unhealthy_threshold 	  = 2
		matcher 		 		  = "200-299"
		path             		  = "/"
		port             		  = 80
		protocol         		  = "HTTP"
		protocol_version 		  = "HTTP1"
	  }
	}
  }

resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
 }
`, rName, rType))
}

func testAccTargetGroupConfig_fullInstance(rName, rType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
	name     	 = %[1]q
	type     	 = %[2]q
	client_token = "tstclienttoken"

	config {
	  port             = 80
	  protocol         = "HTTP"
	  vpc_identifier   =  aws_vpc.test.id
	  protocol_version = "GRPC"

	  health_check {
		enabled             = true
		interval            = 20
		timeout             = 10
		healthy_threshold   = 2
		unhealthy_threshold = 2
		matcher 		    = "200-299"
		path             	= "/instance"
		port             	= 80
		protocol         	= "HTTP"
		protocol_version 	= "HTTP1"
	  }
	}
  }

resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
 }
`, rName, rType))
}

func testAccTargetGroupConfig_fullAlb(rName, rType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
	name     = %[1]q
	type     = %[2]q

	config {
		port             = 443
		protocol         = "HTTPS"
		vpc_identifier   = aws_vpc.test.id
		protocol_version = "HTTP1"
	}
  }

resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
 }
`, rName, rType))
}

func testAccTargetGroupConfig_basic(rName, rType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
	name     = %[1]q
	type     = %[2]q

	config {
		port             = 443
		protocol         = "HTTPS"
	  	vpc_identifier   = aws_vpc.test.id
	}
  }

resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
 }
`, rName, rType))
}

func TestAccVPCLatticeTargetGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var targetGroup1, targetGroup2 vpclattice.GetTargetGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_target_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetgroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccTargetgroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetGroupExists(ctx, resourceName, &targetGroup2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccTargetgroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "LAMBDA"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTargetgroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_target_group" "test" {
  name = %[1]q
  type = "INSTANCE"

  config {
	port             = 80
	protocol         = "HTTP"
	vpc_identifier   = aws_vpc.test.id
	protocol_version = "HTTP1"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
resource "aws_vpc" "test" {
	cidr_block = "10.0.0.0/16"
 }
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCheckTargetGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_target_group" {
				continue
			}

			_, err := conn.GetTargetGroup(ctx, &vpclattice.GetTargetGroupInput{
				TargetGroupIdentifier: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameService, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTargetGroupExists(ctx context.Context, name string, targetGroup *vpclattice.GetTargetGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient()
		resp, err := conn.GetTargetGroup(ctx, &vpclattice.GetTargetGroupInput{
			TargetGroupIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameService, rs.Primary.ID, err)
		}

		*targetGroup = *resp

		return nil
	}
}
