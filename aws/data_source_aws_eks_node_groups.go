package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func dataSourceAwsEksNodeGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksNodeGroupsRead,

		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEksNodeGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName := d.Get("cluster_name").(string)

	input := &eks.ListNodegroupsInput{
		ClusterName: aws.String(clusterName),
	}

	var nodegroups []*string

	err := conn.ListNodegroupsPages(input, func(page *eks.ListNodegroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		nodegroups = append(nodegroups, page.Nodegroups...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing EKS Node Groups: %w", err)
	}

	d.SetId(clusterName)

	d.Set("cluster_name", clusterName)
	d.Set("names", aws.StringValueSlice(nodegroups))

	return nil
}
