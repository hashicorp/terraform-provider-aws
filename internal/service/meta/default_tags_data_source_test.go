package meta_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaDefaultTagsDataSource_basic(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("first", "value"),
					testAccDefaultTagsDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.first", "value"),
				),
			},
		},
	})
}

func TestAccMetaDefaultTagsDataSource_empty(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags0(),
					testAccDefaultTagsDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccMetaDefaultTagsDataSource_multiple(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("nuera", "hijo", "escalofrios", "calambres"),
					testAccDefaultTagsDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.nuera", "hijo"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.escalofrios", "calambres"),
				),
			},
		},
	})
}

func TestAccMetaDefaultTagsDataSource_ignore(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("Tabac", "Louis Chiron"),
					testAccDefaultTagsDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Tabac", "Louis Chiron"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("Tabac", "Louis Chiron"),
					testAccDefaultTagsDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDefaultTagsDataSourceConfig_basic() string {
	return `data "aws_default_tags" "test" {}`
}
