package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Optional: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsConnectLexBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	instanceID := d.Get("instance_id")
	name := d.Get("name")

	var matchedLexBot *connect.LexBot

	lexBots, err := dataSourceAwsConnectGetAllLexBotAssociations(ctx, conn, instanceID.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing Connect Lex Bots: %s", err))
	}

	for _, lexBot := range lexBots {
		log.Printf("[DEBUG] Connect Lex Bot Association: %s", lexBot)
		if aws.StringValue(lexBot.Name) == name.(string) {
			matchedLexBot = lexBot
			break
		}
	}
	d.Set("name", matchedLexBot.Name)
	d.Set("instance_id", instanceID)
	d.Set("region", matchedLexBot.LexRegion)
	d.SetId(fmt.Sprintf("%s:%s:%s", instanceID, d.Get("name").(string), d.Get("region").(string)))

	return nil
}

func dataSourceAwsConnectGetAllLexBotAssociations(ctx context.Context, conn *connect.Connect, instanceID string) ([]*connect.LexBot, error) {
	var bots []*connect.LexBot
	var nextToken string

	for {
		input := &connect.ListLexBotsInput{
			InstanceId: aws.String(instanceID),
			// MaxResults Valid Range: Minimum value of 1. Maximum value of 60
			MaxResults: aws.Int64(int64(60)),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		log.Printf("[DEBUG] Listing Connect Lex Bots: %s", input)

		output, err := conn.ListLexBotsWithContext(ctx, input)
		if err != nil {
			return bots, err
		}
		bots = append(bots, output.LexBots...)

		if output.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(output.NextToken)
	}

	return bots, nil
}
