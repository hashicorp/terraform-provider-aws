package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLexBotAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexBotAliasRead,

		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateLexName,
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateLexName,
			},
		},
	}
}

func dataSourceAwsLexBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(d.Get("bot_name").(string)),
		Name:    aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("error getting Lex bot alias: %s", err)
	}

	d.SetId(aws.StringValue(resp.Name))

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.CreatedDate.UTC().String())
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("name", resp.Name)

	return nil
}
