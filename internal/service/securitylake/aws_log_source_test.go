package securitylake_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityLakeAwsLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_securitylake_aws_log_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAwsLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLogSourceConfig_basic(),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{""},
			},
		},
	})
}

func TestAccSecurityLakeAwsLogSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_aws_log_source.test"
	var awslogSource types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAwsLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLogSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLogSourceExists(ctx, resourceName, &awslogSource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceAwsLogSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsLogSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_aws_log_source" {
				continue
			}

			regionsCount, err := strconv.Atoi(rs.Primary.Attributes["sources.0.regions.#"])
			if err != nil {
				return fmt.Errorf("error parsing regions count: %s", err)
			}

			var regions []string
			for i := 0; i < regionsCount; i++ {
				regions = append(regions, rs.Primary.Attributes[fmt.Sprintf("sources.0.regions.%d", i)])
			}

			out, err := tfsecuritylake.FindAwsLogSourceById(ctx, conn, regions, rs.Primary.ID)

			if tfresource.NotFound(err) || len(out.Sources) == 0 {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameAwsLogSource, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAwsLogSourceExists(ctx context.Context, name string, awsLogSource *awstypes.AwsLogSourceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		regionsCount, err := strconv.Atoi(rs.Primary.Attributes["sources.0.regions.#"])
		if err != nil {
			return fmt.Errorf("error parsing regions count: %s", err)
		}

		var regions []string
		for i := 0; i < regionsCount; i++ {
			regions = append(regions, rs.Primary.Attributes[fmt.Sprintf("sources.0.regions.%d", i)])
		}

		logSources, err := tfsecuritylake.FindAwsLogSourceById(ctx, conn, regions, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, rs.Primary.ID, err)
		}

		var resp *awstypes.AwsLogSourceConfiguration
		if len(logSources.Sources) > 0 {
			resp, err = tfsecuritylake.ExtractAwsLogSourceConfiguration(logSources)
			if err != nil {
				return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, rs.Primary.ID, err)
			}
		}

		*awsLogSource = *resp

		return nil
	}
}

func testAccAwsLogSourceConfig_basic() string {
	return fmt.Sprintf(`

resource "aws_securitylake_aws_log_source" "test" {
	sources {
		regions        = ["eu-west-2","eu-west-1"]
		source_name    = "ROUTE53"
		source_version = "1.0"
	}
}
`)
}
