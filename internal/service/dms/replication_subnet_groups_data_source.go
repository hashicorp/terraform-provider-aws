package dms

import (
	"context"

	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	dmstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceReplicationSubnetGroups() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationSubnetGroupsRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceReplicationSubnetGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient()
	input := &dms.DescribeReplicationSubnetGroupsInput{}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		input.Filters = append(input.Filters,
			BuildFiltersDataSourceV2(filters.(*schema.Set))...)
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindReplicationSubnetGroups(ctx, conn, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Subnet Groups: %s", err)
	}

	var subnetGroupIDs []string
	for _, v := range output {
		subnetGroupIDs = append(subnetGroupIDs, *v.ReplicationSubnetGroupIdentifier)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", subnetGroupIDs)

	return nil
}

func FindReplicationSubnetGroups(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationSubnetGroupsInput) ([]dmstypes.ReplicationSubnetGroup, error) {
	var output []dmstypes.ReplicationSubnetGroup

	paginator := dms.NewDescribeReplicationSubnetGroupsPaginator(conn, input, func(o *dms.DescribeReplicationSubnetGroupsPaginatorOptions) {
		o.Limit = 100
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ReplicationSubnetGroups...)
	}

	return output, nil
}
