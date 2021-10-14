package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceClusters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClustersRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceClustersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn

	var clusters []*string

	err := conn.ListClustersPages(&eks.ListClustersInput{}, func(page *eks.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		clusters = append(clusters, page.Clusters...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing EKS Clusters: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	d.Set("names", aws.StringValueSlice(clusters))

	return nil
}
