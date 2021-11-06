package connect

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceBotAssociation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBotAssociationRead,
		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 50),
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lex_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	instanceID := d.Get("instance_id")
	name := d.Get("bot_name")

	lexBot, err := FindBotAssociationV1ByNameWithContext(ctx, conn, instanceID.(string), name.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Bot V1 Association by name (%s): %w", name, err))
	}

	if lexBot == nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Bot V1 Association by name (%s): not found", name))
	}

	d.Set("bot_name", lexBot.Name)
	d.Set("instance_id", instanceID)
	d.Set("lex_region", lexBot.LexRegion)
	d.SetId(fmt.Sprintf("%s:%s:%s", instanceID, d.Get("bot_name").(string), d.Get("lex_region").(string)))

	return nil
}
