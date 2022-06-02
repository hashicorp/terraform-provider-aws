package imagebuilder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
)

func DataSourceInfrastructureConfigurations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInfrastructureConfigurationsRead,
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
		},
	}
}

func dataSourceInfrastructureConfigurationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ImageBuilderConn

	input := &imagebuilder.ListInfrastructureConfigurationsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).ImagebuilderFilters()
	}

	var results []*imagebuilder.InfrastructureConfigurationSummary

	err := conn.ListInfrastructureConfigurationsPages(input, func(page *imagebuilder.ListInfrastructureConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, infrastructureConfigurationSummary := range page.InfrastructureConfigurationSummaryList {
			if infrastructureConfigurationSummary == nil {
				continue
			}

			results = append(results, infrastructureConfigurationSummary)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading Image Builder Infrastructure Configurations: %w", err)
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
