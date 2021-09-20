package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSDefaultTagsDataSource_basic(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("first", "value"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.first", "value"),
				),
			},
		},
	})
}

func TestAccAWSDefaultTagsDataSource_empty(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags0(),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDefaultTagsDataSource_multiple(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags2("nuera", "hijo", "escalofrios", "calambres"),
					testAccAWSDefaultTagsDataSource(),
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

func TestAccAWSDefaultTagsDataSource_ignore(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultTags_Tags1("Tabac", "Louis Chiron"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Tabac", "Louis Chiron"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigDefaultAndIgnoreTagsKeys1("Tabac", "Louis Chiron"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccAWSDefaultTagsDataSource() string {
	return `data "aws_default_tags" "test" {}`
}
