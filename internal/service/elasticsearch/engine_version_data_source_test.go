package elasticsearch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccElasticSearchEngineVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elasticsearch_engine_version.test"
	version := "7.10"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticsearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_basic(version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", version),
				),
			},
		},
	})
}

func TestAccElasticSearchEngineVersionDataSource_preferred_version(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elasticsearch_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccEngineVersionPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticsearchservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEngineVersionDataSourceConfig_preferred(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "version", "1.5"),
				),
			},
		},
	})
}

func testAccEngineVersionDataSourceConfig_basic(version string) string {
	return fmt.Sprintf(`
data "aws_elasticsearch_engine_version" "test" {
  version                = %[1]q
}
`, version)
}

func testAccEngineVersionDataSourceConfig_preferred() string {
	return fmt.Sprintf(`
data "aws_elasticsearch_engine_version" "test" {
  preferred_versions = ["1.5", "7.10"]
}`)
}

func testAccEngineVersionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticsearchConn()

	input := &elasticsearchservice.ListElasticsearchVersionsInput{}

	_, err := conn.ListElasticsearchVersionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
