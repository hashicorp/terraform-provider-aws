package ssoadmin

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"identity_store_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	output, err := findInstanceMetadatas(conn)

	if err != nil {
		return fmt.Errorf("error reading SSO Instances: %w", err)
	}

	var identityStoreIDs, arns []string

	for _, v := range output {
		identityStoreIDs = append(identityStoreIDs, aws.StringValue(v.IdentityStoreId))
		arns = append(arns, aws.StringValue(v.InstanceArn))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("arns", arns)
	d.Set("identity_store_ids", identityStoreIDs)

	return nil
}

func findInstanceMetadatas(conn *ssoadmin.SSOAdmin) ([]*ssoadmin.InstanceMetadata, error) {
	input := &ssoadmin.ListInstancesInput{}
	var output []*ssoadmin.InstanceMetadata

	err := conn.ListInstancesPages(input, func(page *ssoadmin.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Instances {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
