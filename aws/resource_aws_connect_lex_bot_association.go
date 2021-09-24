package aws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// ConnectLexBotAssociationCreateTimeout Timeout for connect flow creation
	ConnectLexBotAssociationCreateTimeout = 1 * time.Minute
	// ConnectLexBotAssociationDeleteTimeout Timeout for connect flow deletion
	ConnectLexBotAssociationDeleteTimeout = 1 * time.Minute
)

func resourceAwsConnectLexBotAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectLexBotAssociationCreate,
		ReadContext:   resourceAwsConnectLexBotAssociationRead,
		DeleteContext: resourceAwsConnectLexBotAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				instanceID, name, region, err := resourceAwsConnectLexBotAssociationParseID(d.Id())

				if err != nil {
					return nil, err
				}

				d.Set("instance_id", instanceID)
				d.Set("name", name)
				d.Set("region", name)
				d.SetId(fmt.Sprintf("%s:%s:%s", instanceID, name, region))

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ConnectLexBotAssociationCreateTimeout),
			Delete: schema.DefaultTimeout(ConnectLexBotAssociationDeleteTimeout),
		},
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

func resourceAwsConnectLexBotAssociationParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:name:region", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func resourceAwsConnectLexBotAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	botAssociation := &connect.LexBot{
		Name:      aws.String(d.Get("name").(string)),
		LexRegion: aws.String(d.Get("region").(string)),
	}

	input := &connect.AssociateLexBotInput{
		InstanceId: aws.String(d.Get("instance_id").(string)),
		LexBot:     botAssociation,
	}

	log.Printf("[DEBUG] Creating Connect Instance %s", input)

	_, err := conn.AssociateLexBotWithContext(ctx, input)

	d.SetId(fmt.Sprintf("%s:%s:%s", d.Get("instance_id").(string), d.Get("name").(string), d.Get("region").(string)))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Lex Bot Association (%s): %s", d.Id(), err))
	}

	return resourceAwsConnectLexBotAssociationRead(ctx, d, meta)
}

func resourceAwsConnectLexBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	var matchedLexBot *connect.LexBot

	instanceID, name, _, err := resourceAwsConnectLexBotAssociationParseID(d.Id())

	lexBots, err := resourceAwsConnectGetAllLexBotAssociations(ctx, conn, instanceID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing Connect Lex Bots: %s", err))
	}

	for _, lexBot := range lexBots {
		log.Printf("[DEBUG] Connect Lex Bot Association: %s", lexBot)
		if aws.StringValue(lexBot.Name) == name {
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

func resourceAwsConnectLexBotAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, name, region, err := resourceAwsConnectLexBotAssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.DisassociateLexBotInput{
		InstanceId: aws.String(instanceID),
		BotName:    aws.String(name),
		LexRegion:  aws.String(region),
	}

	log.Printf("[DEBUG] Deleting Connect Lex Bot Association %s", d.Id())

	_, dissErr := conn.DisassociateLexBot(input)

	if dissErr != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Lex Bot Association (%s): %s", d.Id(), err))
	}
	return nil
}

func resourceAwsConnectGetAllLexBotAssociations(ctx context.Context, conn *connect.Connect, instanceID string) ([]*connect.LexBot, error) {
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
