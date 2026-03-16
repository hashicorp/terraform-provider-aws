// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediaconnect_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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

	var flow mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", rName+"-source"),
					resource.TestCheckResourceAttr(resourceName, "source.0.description", "Test source"),
					resource.TestCheckResourceAttr(resourceName, "source.0.protocol", "srt-listener"),
					resource.TestCheckResourceAttrSet(resourceName, "source.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "source.0.ingest_ip"),
					resource.TestCheckResourceAttr(resourceName, "start_flow", "false"),
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

func TestAccMediaConnectFlow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmediaconnect.ResourceFlow, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMediaConnectFlow_sourceUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow1, flow2 mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_sourceWhitelist(rName, "10.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow1),
					resource.TestCheckResourceAttr(resourceName, "source.0.whitelist_cidr", "10.0.0.0/16"),
				),
			},
			{
				Config: testAccFlowConfig_sourceWhitelist(rName, "172.16.0.0/12"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow2),
					testAccCheckFlowNotRecreated(&flow1, &flow2),
					resource.TestCheckResourceAttr(resourceName, "source.0.whitelist_cidr", "172.16.0.0/12"),
				),
			},
		},
	})
}

func TestAccMediaConnectFlow_maintenance(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow1, flow2 mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_maintenance(rName, "Sunday", "13:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow1),
					resource.TestCheckResourceAttr(resourceName, "maintenance.0.day", "Sunday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance.0.start_hour", "13:00"),
				),
			},
			{
				Config: testAccFlowConfig_maintenance(rName, "Wednesday", "16:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow2),
					testAccCheckFlowNotRecreated(&flow1, &flow2),
					resource.TestCheckResourceAttr(resourceName, "maintenance.0.day", "Wednesday"),
					resource.TestCheckResourceAttr(resourceName, "maintenance.0.start_hour", "16:00"),
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

func TestAccMediaConnectFlow_entitlement(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_entitlement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "entitlement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "entitlement.0.name", rName+"-entitlement"),
					resource.TestCheckResourceAttr(resourceName, "entitlement.0.description", "Test entitlement"),
					resource.TestCheckResourceAttrSet(resourceName, "entitlement.0.arn"),
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

func TestAccMediaConnectFlow_output(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_output(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "output.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output.0.name", rName+"-output"),
					resource.TestCheckResourceAttr(resourceName, "output.0.description", "Test output"),
					resource.TestCheckResourceAttr(resourceName, "output.0.protocol", "srt-listener"),
					resource.TestCheckResourceAttrSet(resourceName, "output.0.arn"),
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

func TestAccMediaConnectFlow_startFlow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow1, flow2 mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_startFlow(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow1),
					resource.TestCheckResourceAttr(resourceName, "start_flow", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
				),
			},
			{
				Config: testAccFlowConfig_startFlow(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow2),
					testAccCheckFlowNotRecreated(&flow1, &flow2),
					resource.TestCheckResourceAttr(resourceName, "start_flow", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", "STANDBY"),
				),
			},
		},
	})
}

func TestAccMediaConnectFlow_sourceFailover(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_sourceFailover(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "source.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "source_failover_config.0.failover_mode", "FAILOVER"),
					resource.TestCheckResourceAttr(resourceName, "source_failover_config.0.state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "source_failover_config.0.source_priority.0.primary_source", rName+"-primary"),
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

func TestAccMediaConnectFlow_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var flow mediaconnect.DescribeFlowOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mediaconnect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MediaConnectServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFlowDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccFlowConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFlowConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowExists(ctx, resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
				return err
			}

			return create.Error(names.MediaConnect, create.ErrActionCheckingDestroyed, tfmediaconnect.ResNameFlow, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFlowExists(ctx context.Context, name string, flow *mediaconnect.DescribeFlowOutput) resource.TestCheckFunc {
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

		*flow = *resp

		return nil
	}
}

func testAccCheckFlowNotRecreated(before, after *mediaconnect.DescribeFlowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Flow.FlowArn), aws.ToString(after.Flow.FlowArn); before != after {
			return create.Error(names.MediaConnect, create.ErrActionCheckingNotRecreated, tfmediaconnect.ResNameFlow, before, errors.New("recreated"))
		}

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

func testAccFlowConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name        = "%[1]s-source"
    description = "Test source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }
}
`, rName)
}

func testAccFlowConfig_sourceWhitelist(rName, cidr string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name           = "%[1]s-source"
    description    = "Test source"
    protocol       = "rtp"
    whitelist_cidr = %[2]q
  }
}
`, rName, cidr)
}

func testAccFlowConfig_maintenance(rName, day, startHour string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name           = "%[1]s-source"
    description    = "Test source"
    protocol       = "rtp"
    whitelist_cidr = "0.0.0.0/0"
  }

  maintenance {
    day        = %[2]q
    start_hour = %[3]q
  }
}
`, rName, day, startHour)
}

func testAccFlowConfig_entitlement(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name        = "%[1]s-source"
    description = "Test source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  entitlement {
    name        = "%[1]s-entitlement"
    description = "Test entitlement"
    subscriber  = [data.aws_caller_identity.current.account_id]
  }
}
`, rName)
}

func testAccFlowConfig_output(rName string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name        = "%[1]s-source"
    description = "Test source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  output {
    name        = "%[1]s-output"
    description = "Test output"
    protocol    = "srt-listener"
    port        = 5001
  }
}
`, rName)
}

func testAccFlowConfig_startFlow(rName string, start bool) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name       = %[1]q
  start_flow = %[2]t

  source {
    name        = "%[1]s-source"
    description = "Test source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }
}
`, rName, start)
}

func testAccFlowConfig_sourceFailover(rName string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name        = "%[1]s-primary"
    description = "Primary source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  source {
    name        = "%[1]s-backup"
    description = "Backup source"
    protocol    = "srt-listener"
    ingest_port = 5001
  }

  source_failover_config {
    failover_mode   = "FAILOVER"
    recovery_window = 200
    state           = "ENABLED"

    source_priority {
      primary_source = "%[1]s-primary"
    }
  }
}
`, rName)
}

func testAccFlowConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name        = "%[1]s-source"
    description = "Test source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFlowConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_mediaconnect_flow" "test" {
  name = %[1]q

  source {
    name        = "%[1]s-source"
    description = "Test source"
    protocol    = "srt-listener"
    ingest_port = 5000
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
