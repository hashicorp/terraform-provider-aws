package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceBotAssociation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBotAssociationRead,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lex_bot": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lex_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(2, 50),
						},
					},
				},
			},
		},
	}
}

func dataSourceBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn()

	instanceID := d.Get("instance_id").(string)

	var name, region string
	if v, ok := d.GetOk("lex_bot"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		lexBot := expandLexBot(v.([]interface{}))
		name = aws.StringValue(lexBot.Name)
		region = aws.StringValue(lexBot.LexRegion)
	}

	lexBot, err := FindBotAssociationV1ByNameAndRegionWithContext(ctx, conn, instanceID, name, region)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Bot Association (%s,%s): %w", instanceID, name, err))
	}

	if lexBot == nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Bot Association (%s,%s) : not found", instanceID, name))
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	d.Set("instance_id", instanceID)
	if err := d.Set("lex_bot", flattenLexBot(lexBot)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting lex_bot: %w", err))
	}

	return nil
}
