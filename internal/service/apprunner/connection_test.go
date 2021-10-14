package apprunner_test

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_apprunner_connection", &resource.Sweeper{
		Name:         "aws_apprunner_connection",
		F:            sweepConnections,
		Dependencies: []string{"aws_apprunner_service"},
	})
}

func sweepConnections(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AppRunnerConn
	sweepResources := make([]*sweep.SweepResource, 0)
	ctx := context.Background()

	var errs *multierror.Error

	input := &apprunner.ListConnectionsInput{}

	err = conn.ListConnectionsPagesWithContext(ctx, input, func(page *apprunner.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.ConnectionSummaryList {
			if c == nil {
				continue
			}

			name := aws.StringValue(c.ConnectionName)

			log.Printf("[INFO] Deleting App Runner Connection: %s", name)

			r := tfapprunner.ResourceConnection()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("arn", c.ConnectionArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Connections: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Connections for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Connections sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func TestAccAppRunnerConnection_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apprunner.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerConnection_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "apprunner", regexp.MustCompile(fmt.Sprintf(`connection/%s/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "connection_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provider_type", apprunner.ProviderTypeGithub),
					resource.TestCheckResourceAttr(resourceName, "status", apprunner.ConnectionStatusPendingHandshake),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAppRunnerConnection_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apprunner.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerConnection_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapprunner.ResourceConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerConnection_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAppRunner(t) },
		ErrorCheck:   acctest.ErrorCheck(t, apprunner.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppRunnerConnectionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppRunnerConnectionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccAppRunnerConnectionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apprunner_connection" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		connection, err := tfapprunner.FindConnectionSummaryByName(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if connection != nil {
			return fmt.Errorf("App Runner Connection (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Connection ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerConn

		connection, err := tfapprunner.FindConnectionSummaryByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if connection == nil {
			return fmt.Errorf("App Runner Connection (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAppRunnerConnection_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %q
  provider_type   = "GITHUB"
}
`, rName)
}

func testAccAppRunnerConnectionConfigTags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %[1]q
  provider_type   = "GITHUB"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppRunnerConnectionConfigTags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_connection" "test" {
  connection_name = %[1]q
  provider_type   = "GITHUB"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
