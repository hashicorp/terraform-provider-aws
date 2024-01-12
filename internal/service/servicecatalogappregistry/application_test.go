package servicecatalogappregistry_test

import (
	"context"
	"fmt"
	"testing"

	servicecatalogappregistry_sdkv2 "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/aws/aws-sdk-go/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistry),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr("aws_servicecatalog_application.test", "name", resourceName),
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

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_application" {
				continue
			}

			_, err := conn.GetApplication(ctx, &servicecatalogappregistry_sdkv2.GetApplicationInput{
				Application: aws.String(rs.Primary.Attributes["name"]),
			})

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckApplication(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry_sdkv2.GetApplicationInput{}

	_, err := conn.GetApplication(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckApplicationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

		_, err := servicecatalogappregistry.FindApplicationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccApplicationConfig_basic(name string) string {
	return fmt.Sprintf(`
	resource "aws_servicecatalog_application" "test" {
		name = "%s"
	}`, name)
}
