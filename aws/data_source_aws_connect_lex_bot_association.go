package aws

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
)

func dataSourceAwsConnectLexBotAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsConnectLexBotAssociationRead,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsConnectLexBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	instanceID := d.Get("instance_id")
	name := d.Get("name")

	lexBot, err := finder.LexBotAssociationByName(ctx, conn, instanceID.(string), name.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding LexBot Association by name (%s): %w", name, err))
	}

	if lexBot == nil {
		return diag.FromErr(fmt.Errorf("error finding LexBot Association by name (%s): not found", name))
	}

	d.Set("name", lexBot.Name)
	d.Set("instance_id", instanceID)
	d.Set("region", lexBot.LexRegion)
	d.SetId(fmt.Sprintf("%s:%s:%s", instanceID, d.Get("name").(string), d.Get("region").(string)))

	return nil
}
