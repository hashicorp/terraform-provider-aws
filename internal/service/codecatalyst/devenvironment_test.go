package codecatalyst_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfcodecatalyst "github.com/hashicorp/terraform-provider-aws/internal/service/codecatalyst"
)

func TestAccCodecatalystDevenvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var devEnvironment codecatalyst.GetDevEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecatalyst_devenvironment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCatalyst),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevenvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevenvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDevenvironmentExists(ctx, resourceName, &devEnvironment),
					// resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
					// resource.TestCheckResourceAttrPair(resourceName, "target.0.id", instanceResourceName, "id"),
					// resource.TestCheckResourceAttr(resourceName, "target.0.port", "80"),
				),
			},
		},
	})
}
func TestAccCodecatalystDevenvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var devEnvironment codecatalyst.GetDevEnvironmentOutput
	resourceName := "aws_codecatalyst_devenvironment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDevenvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDevenvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevenvironmentExists(ctx, resourceName, &devEnvironment),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodecatalyst.ResourceDevenvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDevenvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecatalyst_devenvironment" {
				continue
			}
			spaceName := rs.Primary.Attributes["space_name"]
			projectName := rs.Primary.Attributes["project_name"]

			_, err := conn.GetDevEnvironment(ctx, &codecatalyst.GetDevEnvironmentInput{
				Id:          aws.String(rs.Primary.ID),
				SpaceName:   aws.String(spaceName),
				ProjectName: aws.String(projectName),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return nil
			}

			return create.Error(names.CodeCatalyst, create.ErrActionCheckingDestroyed, tfcodecatalyst.ResNameDevenvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDevenvironmentExists(ctx context.Context, name string, devenvironment *codecatalyst.GetDevEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevenvironment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevenvironment, name, errors.New("not set"))
		}
		spaceName := rs.Primary.Attributes["space_name"]
		projectName := rs.Primary.Attributes["project_name"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)
		resp, err := conn.GetDevEnvironment(ctx, &codecatalyst.GetDevEnvironmentInput{
			Id:          aws.String(rs.Primary.ID),
			SpaceName:   aws.String(spaceName),
			ProjectName: aws.String(projectName),
		})

		if err != nil {
			return create.Error(names.CodeCatalyst, create.ErrActionCheckingExistence, tfcodecatalyst.ResNameDevenvironment, rs.Primary.ID, err)
		}

		*devenvironment = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCatalystClient(ctx)

	spaceName := "personal-926562225508"
	projectName := "terraform-contribution"

	input := &codecatalyst.ListDevEnvironmentsInput{
		SpaceName:   aws.String(spaceName),
		ProjectName: aws.String(projectName),
	}
	_, err := conn.ListDevEnvironments(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDevenvironmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecatalyst_devenvironment" "test" {
  alias = %[1]q
  space_name = "personal-926562225508"
  project_name = "terraform-contribution"
  instance_type = "dev.standard1.small"
  persistent_storage  {
	size = 16
  }
  ides {
	name = "VSCode"
  }

  
}
`, rName)
}
