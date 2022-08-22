package imagebuilder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
)

func DataSourceContainerRecipes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceContainerRecipesRead,
		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": namevaluesfilters.Schema(),
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Self", "Shared", "Amazon"}, false),
			},
		},
	}
}

func dataSourceContainerRecipesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn

	input := &imagebuilder.ListContainerRecipesInput{}

	if v, ok := d.GetOk("owner"); ok {
		input.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).ImagebuilderFilters()
	}

	var results []*imagebuilder.ContainerRecipeSummary

	err := conn.ListContainerRecipesPages(input, func(page *imagebuilder.ListContainerRecipesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, containerRecipeSummary := range page.ContainerRecipeSummaryList {
			if containerRecipeSummary == nil {
				continue
			}

			results = append(results, containerRecipeSummary)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading Image Builder Container Recipes: %w", err)
	}

	var arns, names []string

	for _, r := range results {
		arns = append(arns, aws.StringValue(r.Arn))
		names = append(names, aws.StringValue(r.Name))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("names", names)

	return nil
}
