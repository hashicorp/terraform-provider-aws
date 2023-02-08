package eks

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceClusters() *schema.Resource {
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

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSConn()

	var clusters []*string

	err := conn.ListClustersPagesWithContext(ctx, &eks.ListClustersInput{}, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		clusters = append(clusters, page.Clusters...)

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EKS Clusters: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	d.Set("names", aws.StringValueSlice(clusters))

	return diags
}
