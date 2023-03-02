package swf_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/swf"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfswf "github.com/hashicorp/terraform-provider-aws/internal/service/swf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPreCheckDomainTestingEnabled(t *testing.T) {
	if os.Getenv("SWF_DOMAIN_TESTING_ENABLED") == "" {
		t.Skip(
			"Environment variable SWF_DOMAIN_TESTING_ENABLED is not set. " +
				"SWF limits domains per region and the API does not support " +
				"deletions. Set the environment variable to any value to enable.")
	}
}

func TestAccSWFDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, swf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "swf", regexp.MustCompile(`/domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "workflow_execution_retention_period_in_days", "1"),
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

func TestAccSWFDomain_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, swf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_nameGenerated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
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

func TestAccSWFDomain_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, swf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

func TestAccSWFDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, swf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSWFDomain_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, swf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SWFConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_swf_domain" {
				continue
			}

			// Retrying as Read after Delete is not always consistent.
			err := resource.RetryContext(ctx, 2*time.Minute, func() *resource.RetryError {
				_, err := tfswf.FindDomainByName(ctx, conn, rs.Primary.ID)

				if tfresource.NotFound(err) {
					return nil
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}

				return resource.RetryableError(fmt.Errorf("SWF Domain still exists: %s", rs.Primary.ID))
			})

			return err
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SWF Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SWFConn()

		_, err := tfswf.FindDomainByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1
}
`, rName)
}

func testAccDomainConfig_nameGenerated() string {
	return `
resource "aws_swf_domain" "test" {
  workflow_execution_retention_period_in_days = 1
}
`
}

func testAccDomainConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name_prefix                                 = %[1]q
  workflow_execution_retention_period_in_days = 1
}
`, namePrefix)
}

func testAccDomainConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDomainConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDomainConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  description                                 = %[2]q
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1
}
`, rName, description)
}
