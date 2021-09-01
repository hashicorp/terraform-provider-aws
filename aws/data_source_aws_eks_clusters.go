package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsEksClusters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksClustersRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEksClustersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	var clusters []*string

	log.Printf("[DEBUG] Listing EKS Clusters")
	err := conn.ListClustersPages(&eks.ListClustersInput{},
		func(page *eks.ListClustersOutput, lastPage bool) bool {
			clusters = append(clusters, page.Clusters...)
			return true
		},
	)

	d.SetId(meta.(*AWSClient).region)

	if err != nil {
		log.Printf("[DEBUG] There was an error while listing EKS Clusters: %v", err)
	}
	d.Set("names", clusters)

	return nil
}
