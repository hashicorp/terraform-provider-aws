// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconnect_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfmediaconnect "github.com/hashicorp/terraform-provider-aws/internal/service/mediaconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaConnectFlow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow types.Flow
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"
	sourceName := sdkacctest.RandomWithPrefix("tf-acc-test-source-name")
	sourceDescription := fmt.Sprintf("Terraform Acceptance Test %s", strings.Title(sdkacctest.RandString(20)))
	sourceProtocol := "rtp"
	sourceWhitelistCidr := "0.0.0.0/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName, sourceName, sourceDescription, sourceProtocol, sourceWhitelistCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttrSet(resourceName, "source.0.protocol"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "source.*", map[string]string{
						"name":           sourceName,
						"description":    sourceDescription,
						"protocol":       sourceProtocol,
						"whitelist_cidr": sourceWhitelistCidr,
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mediaconnect", regexache.MustCompile(`flow:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_flow"},
			},
		},
	})
}

func TestAccMediaConnectFlow_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow types.Flow
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"
	maintenanceDay := "Sunday"
	maintenanceDayUpdated := "Wednesday"
	maintenanceStartHour := "13:00"
	maintenanceStartHourUpdated := "16:00"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_update(rName, maintenanceDay, maintenanceStartHour),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "maintenance.*", map[string]string{
						"day":        maintenanceDay,
						"start_hour": maintenanceStartHour,
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mediaconnect", regexache.MustCompile(`flow:+.`)),
				),
			},
			{
				Config: testAccFlowConfig_update(rName, maintenanceDayUpdated, maintenanceStartHourUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "maintenance.*", map[string]string{
						"day":        maintenanceDayUpdated,
						"start_hour": maintenanceStartHourUpdated,
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mediaconnect", regexache.MustCompile(`flow:+.`)),
				),
			},
		},
	})
}

func TestAccMediaConnectFlow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow types.Flow
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"
	sourceName := sdkacctest.RandomWithPrefix("tf-acc-test-source-name")
	sourceDescription := fmt.Sprintf("Terraform Acceptance Test %s", strings.Title(sdkacctest.RandString(20)))
	sourceProtocol := "rtp"
	sourceWhitelistCidr := "0.0.0.0/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName, sourceName, sourceDescription, sourceProtocol, sourceWhitelistCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfmediaconnect.ResourceFlow, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFlowDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mediaconnect_flow" {
				continue
			}

			_, err := tfmediaconnect.FindFlowByARN(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.NotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.MediaConnect, create.ErrActionCheckingDestroyed, tfmediaconnect.ResNameFlow, rs.Primary.ID, err)
			}

			return create.Error(names.MediaConnect, create.ErrActionCheckingDestroyed, tfmediaconnect.ResNameFlow, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFlowExists(ctx context.Context, name string, flow *types.Flow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MediaConnect, create.ErrActionCheckingExistence, tfmediaconnect.ResNameFlow, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MediaConnect, create.ErrActionCheckingExistence, tfmediaconnect.ResNameFlow, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConnectClient(ctx)
		resp, err := conn.DescribeFlow(ctx, &mediaconnect.DescribeFlowInput{
			FlowArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.MediaConnect, create.ErrActionCheckingExistence, tfmediaconnect.ResNameFlow, rs.Primary.ID, err)
		}

		*flow = *resp.Flow

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaConnectClient(ctx)

	input := &mediaconnect.ListFlowsInput{}
	_, err := conn.ListFlows(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFlowConfig_basic(rName, sourceName, sourceDescription, sourceProtocol, sourceWhitelistCidr string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name           = %[2]q
    description    = %[3]q
    protocol       = %[4]q
    whitelist_cidr = %[5]q
  }
}
`, rName, sourceName, sourceDescription, sourceProtocol, sourceWhitelistCidr)
}

func testAccFlowConfig_update(rName, maintenanceDay, maintenanceStartHour string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name           = "testacc-source-name"
    description    = "testacc-source-description"
    protocol       = "rtp"
    whitelist_cidr = "0.0.0.0/0"
  }

  maintenance {
    day        = %[2]q
    start_hour = %[3]q
  }
}
`, rName, maintenanceDay, maintenanceStartHour)
}
