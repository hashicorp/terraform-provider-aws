package securitylake_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityLakeLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_securitylake_log_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLogSourceConfig_basic(rName),
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

func TestAccSecurityLakeLogSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_log_source.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var logSource types.AwsLogSourceConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLogSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogSourceExists(ctx, resourceName, &logSource),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceLogSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLogSourceDestroy(ctx context.Context) resource.TestCheckFunc {
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

			_, err = tfsecuritylake.FindLogSourceById(ctx, conn, regions, rs.Primary.ID)
			// No Datalake
			// "The request failed because Security Lake isn't enabled for your account in any Regions. Enable Security Lake for your account and then try again."
			if tfresource.NotFound(err) || errs.IsAErrorMessageContains[*types.ResourceNotFoundException](err, "Enable Security Lake for your account and then try again") {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameLogSource, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLogSourceExists(ctx context.Context, name string, logSource *types.AwsLogSourceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameLogSource, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameLogSource, name, errors.New("not set"))
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

		logSources, err := tfsecuritylake.FindLogSourceById(ctx, conn, regions, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameLogSource, rs.Primary.ID, err)
		}

		var resp *types.AwsLogSourceConfiguration
		if len(logSources.Sources) > 0 {
			resp, err = tfsecuritylake.ExtractLogSourceConfiguration(logSources)
			if err != nil {
				return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameLogSource, rs.Primary.ID, err)
			}
		}

		*logSource = *resp

		return nil
	}
}

func testAccLogSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(rName), fmt.Sprintf(`


resource "aws_securitylake_log_source" "test" {
  sources {
    regions        = [%[2]q]
    source_name    = "ROUTE53"
    source_version = "1.0"
  }
  depends_on = [aws_securitylake_data_lake.test]
}
`, rName, acctest.Region()))
}
