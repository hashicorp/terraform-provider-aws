// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsguru_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdevopsguru "github.com/hashicorp/terraform-provider-aws/internal/service/devopsguru"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEventSourcesConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cfg devopsguru.DescribeEventSourcesConfigOutput
	resourceName := "aws_devopsguru_event_sources_config.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourcesConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourcesConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourcesConfigExists(ctx, t, resourceName, &cfg),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "event_sources.0.amazon_code_guru_profiler.*", map[string]string{
						names.AttrStatus: "ENABLED",
					}),
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

func testAccEventSourcesConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var eventsourcesconfig devopsguru.DescribeEventSourcesConfigOutput
	resourceName := "aws_devopsguru_event_sources_config.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DevOpsGuruEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsGuruServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventSourcesConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventSourcesConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEventSourcesConfigExists(ctx, t, resourceName, &eventsourcesconfig),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdevopsguru.ResourceEventSourcesConfig, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEventSourcesConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DevOpsGuruClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_devopsguru_event_sources_config" {
				continue
			}

			out, err := tfdevopsguru.FindEventSourcesConfig(ctx, conn)
			if out.EventSources == nil || out.EventSources.AmazonCodeGuruProfiler == nil {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameEventSourcesConfig, rs.Primary.ID, errors.New("empty output"))
			}
			if out.EventSources.AmazonCodeGuruProfiler.Status == awstypes.EventSourceOptInStatusDisabled {
				return nil
			}

			if err != nil {
				return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameEventSourcesConfig, rs.Primary.ID, err)
			}
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingDestroyed, tfdevopsguru.ResNameEventSourcesConfig, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEventSourcesConfigExists(ctx context.Context, t *testing.T, name string, cfg *devopsguru.DescribeEventSourcesConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameEventSourcesConfig, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameEventSourcesConfig, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DevOpsGuruClient(ctx)

		out, err := tfdevopsguru.FindEventSourcesConfig(ctx, conn)
		if err != nil {
			return create.Error(names.DevOpsGuru, create.ErrActionCheckingExistence, tfdevopsguru.ResNameEventSourcesConfig, rs.Primary.ID, err)
		}

		*cfg = *out

		return nil
	}
}

func testAccEventSourcesConfigConfig_basic() string {
	return `
resource "aws_devopsguru_event_sources_config" "test" {
  event_sources {
    amazon_code_guru_profiler {
      status = "ENABLED"
    }
  }
}
`
}
