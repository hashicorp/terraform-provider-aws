package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
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

	lexBot, err := dataSourceAwsConnectGetLexBotAssociationByName(ctx, conn, instanceID.(string), name.(string))
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

func dataSourceAwsConnectGetLexBotAssociationByName(ctx context.Context, conn *connect.Connect, instanceID string, name string) (*connect.LexBot, error) {
	var result *connect.LexBot

	input := &connect.ListLexBotsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(tfconnect.ListLexBotsMaxResults),
	}

	err := conn.ListLexBotsPagesWithContext(ctx, input, func(page *connect.ListLexBotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.LexBots {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
