package glue

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Define AttrIDs as a string constant
const AttrIDs = "ids"

// @SDKDataSource("aws_glue_catalog_tables")
func DataSourceCatalogTables() *schema.Resource {

	return &schema.Resource{
		ReadWithoutTimeout: DataSourceCatalogTablesRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func DataSourceCatalogTablesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	dbName := d.Get(names.AttrDatabaseName).(string)
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID)

	input := &glue.GetTablesInput{
		DatabaseName: aws.String(dbName),
		CatalogId:    aws.String(catalogID),
	}

	output, err := conn.GetTables(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Glue tables: %s", err)
	}

	var tableIDs []string
	for _, table := range output.TableList {
		tableIDs = append(tableIDs, *table.Name)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, dbName))
	d.Set(AttrIDs, tableIDs)
	d.Set("catalog_id", catalogID)

	return diags
}
