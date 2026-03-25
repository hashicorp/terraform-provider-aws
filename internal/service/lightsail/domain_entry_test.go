// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomainEntry_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainEntryName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "A"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Validate that we can import an existing resource using the legacy separator
			// Validate that the ID is updated to use the new common separator
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDomainEntryStateLegacyIdFunc(resourceName),
				ImportStateVerify: true,
				Check:             resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%s,%s,%s,%s", domainEntryName, domainName, "A", "127.0.0.1")),
			},
		},
	})
}

func testAccDomainEntry_underscore(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := "_" + acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainEntryName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "A"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Validate that we can import an existing resource using the legacy separator
			// Validate that the ID is updated to use the new common separator
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDomainEntryStateLegacyIdFunc(resourceName),
				ImportStateVerify: true,
				Check:             resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%s,%s,%s,%s", domainEntryName, domainName, "A", "127.0.0.1")),
			},
		},
	})
}

func testAccDomainEntry_apex(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := ""

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainEntryName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "A"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Validate that we can import an existing resource using the legacy separator
			// Validate that the ID is updated to use the new common separator
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDomainEntryStateLegacyIdFunc(resourceName),
				ImportStateVerify: true,
				Check:             resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%s,%s,%s,%s", domainEntryName, domainName, "A", "127.0.0.1")),
			},
		},
	})
}

func testAccDomainEntry_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceDomainEntry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainEntry_typeAAAA(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckLightsailSynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, string(types.RegionNameUsEast1))
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_typeAAAA(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainEntryName),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, "::1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AAAA"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Validate that we can import an existing resource using the legacy separator
			// Validate that the ID is updated to use the new common separator
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDomainEntryStateLegacyIdFunc(resourceName),
				ImportStateVerify: true,
				Check:             resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("%s,%s,%s,%s", domainEntryName, domainName, "AAAA", "::1")),
			},
		},
	})
}

func testAccCheckDomainEntryExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain Entry ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		resp, err := tflightsail.FindDomainEntryById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("DomainEntry %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainEntryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_domain_entry" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

			_, err := tflightsail.FindDomainEntryById(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResDomainEntry, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccDomainEntryConfig_basic(domainName string, domainEntryName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = %[1]q
}

resource "aws_lightsail_domain_entry" "test" {
  domain_name = aws_lightsail_domain.test.id
  name        = %[2]q
  type        = "A"
  target      = "127.0.0.1"
}
`, domainName, domainEntryName)
}

func testAccDomainEntryConfig_typeAAAA(domainName string, domainEntryName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = %[1]q
}

resource "aws_lightsail_domain_entry" "test" {
  domain_name = aws_lightsail_domain.test.id
  name        = %[2]q
  type        = "AAAA"
  target      = "::1"
}
`, domainName, domainEntryName)
}

func testAccDomainEntryStateLegacyIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s_%s_%s_%s", rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes[names.AttrType], rs.Primary.Attributes[names.AttrTarget]), nil
	}
}
