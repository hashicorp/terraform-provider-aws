package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func dataSourceAwsLexBotAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexBotAliasRead,

		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexBotNameMinLength, lexBotNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
			"bot_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
		},
	}
}

func dataSourceAwsLexBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	botName := d.Get("bot_name").(string)
	botAliasName := d.Get("name").(string)

	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(botName),
		Name:    aws.String(botAliasName),
	})
	if err != nil {
		return fmt.Errorf("error getting bot alias %s: %s", botAliasName, err)
	}

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("name", resp.Name)

	d.SetId(fmt.Sprintf("%s.%s", botName, botAliasName))

	return nil
}
