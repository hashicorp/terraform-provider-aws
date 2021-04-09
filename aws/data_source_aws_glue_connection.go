package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsGlueConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsGlueConnectionRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsGlueConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).glueconn
	catalogID, connectionName, err := decodeGlueConnectionID(d.Id())
	input := &glue.GetConnectionInput{
		CatalogId: aws.String(catalogID),
		Name:      aws.String(connectionName),
	}
	output, err := conn.GetConnection(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return diag.Errorf("error Glue Connection (%s) not found", d.Id())
		}
		return diag.Errorf("error reading Glue Connection (%s): %s", d.Id(), err)
	}
	d.Set("catalog_id", catalogID)
	d.Set("creation_time", aws.TimeValue(output.Connection.CreationTime).Format(time.RFC3339))
	d.Set("connection_type", output.Connection.ConnectionType)
	d.Set("name", connectionName)
	return nil
}
