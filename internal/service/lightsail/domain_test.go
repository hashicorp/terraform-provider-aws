// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
)

func testAccDomain_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))
	resourceName := "aws_lightsail_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(lightsailDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccDomain_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	lightsailDomainName := fmt.Sprintf("tf-test-lightsail-%s.com", sdkacctest.RandString(5))
	resourceName := "aws_lightsail_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(lightsailDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		resp, err := conn.GetDomain(ctx, &lightsail.GetDomainInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_domain" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

			resp, err := conn.GetDomain(ctx, &lightsail.GetDomainInput{
				DomainName: aws.String(rs.Primary.ID),
			})

			if tflightsail.IsANotFoundError(err) {
				continue
			}

			if err == nil {
				if resp.Domain != nil {
					return fmt.Errorf("Lightsail Domain %q still exists", rs.Primary.ID)
				}
			}

			return err
		}

		return nil
	}
}

func testAccDomainConfig_basic(lightsailDomainName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = "%s"
}
`, lightsailDomainName)
}
