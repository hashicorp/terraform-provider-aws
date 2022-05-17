package athena

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceNamedQuery() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNamedQueryRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"workgroup": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "primary",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"database": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"querystring": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNamedQueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	input := &athena.ListNamedQueriesInput{WorkGroup: aws.String(d.Get("workgroup").(string))}
	conn := meta.(*conns.AWSClient).AthenaConn
	var ids []*string

	err := conn.ListNamedQueriesPagesWithContext(ctx, input, func(page *athena.ListNamedQueriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		ids = append(ids, page.NamedQueryIds...)
		if aws.StringValue(page.NextToken) == "" {
			return false
		}
		return !lastPage
	})

	if err != nil {
		return names.DiagError(names.Athena, names.ErrActionReading, ResListNamedQueries, d.Id(), err)
	}

	target := d.Get("name").(string)
	batchInput := &athena.BatchGetNamedQueryInput{NamedQueryIds: ids}
	resp, err := conn.BatchGetNamedQueryWithContext(ctx, batchInput)
	var output *athena.NamedQuery
	if err != nil {
		return names.DiagError(names.Athena, names.ErrActionReading, ResBatchGetNamedQuery, d.Id(), err)
	}
	for _, query := range resp.NamedQueries {
		if aws.StringValue(query.Name) == target {
			output = query
			break
		}
	}

	if output == nil {
		return names.DiagError(names.Athena, names.ErrActionReading, ResFindQueryByName, d.Id(), err)
	}
	d.SetId(aws.StringValue(output.NamedQueryId))
	if err := d.Set("database", output.Database); err != nil {
		return names.DiagError(names.Athena, "setting database", ResFindQueryByName, d.Id(), err)
	}
	if err := d.Set("description", output.Description); err != nil {
		return names.DiagError(names.Athena, "setting description", ResFindQueryByName, d.Id(), err)
	}
	if err := d.Set("name", output.Name); err != nil {
		return names.DiagError(names.Athena, "setting name", ResFindQueryByName, d.Id(), err)
	}
	if err := d.Set("id", output.NamedQueryId); err != nil {
		return names.DiagError(names.Athena, "setting id", ResFindQueryByName, d.Id(), err)
	}
	if err := d.Set("workgroup", output.WorkGroup); err != nil {
		return names.DiagError(names.Athena, "setting workgroup", ResFindQueryByName, d.Id(), err)
	}
	if err := d.Set("querystring", output.QueryString); err != nil {
		return names.DiagError(names.Athena, "setting querystring", ResFindQueryByName, d.Id(), err)
	}
	return nil
}
