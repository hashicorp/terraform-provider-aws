package appsync_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/appsync"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var typ appsync.Type
	resourceName := "aws_appsync_type.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeResourceExists(ctx, resourceName, &typ),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile("apis/.+/types/.+")),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "format", "SDL"),
					resource.TestCheckResourceAttr(resourceName, "name", "Mutation"),
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

func testAccType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var typ appsync.Type
	resourceName := "aws_appsync_type.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTypeResourceExists(ctx, resourceName, &typ),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTypeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_type" {
				continue
			}

			apiID, format, name, err := tfappsync.DecodeTypeID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfappsync.FindTypeByID(ctx, conn, apiID, format, name)
			if err == nil {
				if tfresource.NotFound(err) {
					return nil
				}
				return err
			}

			return nil
		}
		return nil
	}
}

func testAccCheckTypeResourceExists(ctx context.Context, resourceName string, typ *appsync.Type) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Appsync Type Not found in state: %s", resourceName)
		}

		apiID, format, name, err := tfappsync.DecodeTypeID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn()
		out, err := tfappsync.FindTypeByID(ctx, conn, apiID, format, name)
		if err != nil {
			return err
		}

		*typ = *out

		return nil
	}
}

func testAccTypeConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_type" "test" {
  api_id     = aws_appsync_graphql_api.test.id
  format     = "SDL"
  definition = <<EOF
type Mutation

{
putPost(id: ID!,title: String! ): Post

}
EOF  
}
`, rName)
}
