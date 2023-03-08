package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailDomainEntry_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var domainEntry lightsail.DomainEntry
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckDomain(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, resourceName, &domainEntry),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "name", domainEntryName),
					resource.TestCheckResourceAttr(resourceName, "target", "127.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "type", "A"),
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

func TestAccLightsailDomainEntry_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domainEntry lightsail.DomainEntry
	resourceName := "aws_lightsail_domain_entry.test"
	domainName := acctest.RandomDomainName()
	domainEntryName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	testDestroy := func(*terraform.State) error {
		conn := testAccProviderLightsailDomain.Meta().(*conns.AWSClient).LightsailConn()
		_, err := conn.DeleteDomainEntryWithContext(ctx, &lightsail.DeleteDomainEntryInput{
			DomainName: aws.String(domainName),
			DomainEntry: &lightsail.DomainEntry{
				Name:   aws.String(fmt.Sprintf("%s.%s", domainEntryName, domainName)),
				Type:   aws.String("A"),
				Target: aws.String("127.0.0.1"),
			},
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Domain Entry in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckDomain(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainEntryConfig_basic(domainName, domainEntryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainEntryExists(ctx, resourceName, &domainEntry),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainEntryExists(ctx context.Context, n string, domainEntry *lightsail.DomainEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain Entry ID is set")
		}

		conn := testAccProviderLightsailDomain.Meta().(*conns.AWSClient).LightsailConn()

		resp, err := tflightsail.FindDomainEntryById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("DomainEntry %q does not exist", rs.Primary.ID)
		}

		*domainEntry = *resp

		return nil
	}
}

func testAccCheckDomainEntryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_domain_entry" {
				continue
			}

			conn := testAccProviderLightsailDomain.Meta().(*conns.AWSClient).LightsailConn()

			_, err := tflightsail.FindDomainEntryById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
	return acctest.ConfigCompose(
		testAccDomainRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_lightsail_domain" "test" {
  domain_name = %[1]q
}
resource "aws_lightsail_domain_entry" "test" {
  domain_name = aws_lightsail_domain.test.id
  name        = %[2]q
  type        = "A"
  target      = "127.0.0.1"
}
`, domainName, domainEntryName))
}
