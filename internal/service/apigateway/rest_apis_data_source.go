package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceRestAPIs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClustersRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSName = "Rest APIs Data Source"
)

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	input := &apigateway.GetRestApisInput{}

	var apiNames []string

	err := conn.GetRestApisPagesWithContext(ctx, input, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			apiNames = append(apiNames, aws.StringValue(item.Name))

		}

		return !lastPage
	})

	if err != nil {
		return create.DiagError(names.APIGateway, create.ErrActionReading, DSName, "", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("names", apiNames)

	return nil
}
