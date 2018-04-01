package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsLexBotAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexBotAliasCreate,
		Read:   resourceAwsLexBotAliasRead,
		Update: resourceAwsLexBotAliasUpdate,
		Delete: resourceAwsLexBotAliasDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsLexBotAliasImport,
		},

		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexName,
			},
			"bot_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateLexVersion,
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validateMaxLength(lexDescriptionMaxLength),
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexName,
			},
		},
	}
}

func resourceAwsLexBotAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	_, err := conn.PutBotAlias(&lexmodelbuildingservice.PutBotAliasInput{
		BotName:     aws.String(d.Get("bot_name").(string)),
		BotVersion:  aws.String(d.Get("bot_version").(string)),
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("error creating Lex bot alias %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexBotAliasRead(d, meta)
}

func resourceAwsLexBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(d.Get("bot_name").(string)),
		Name:    aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("error getting Lex bot alias: %s", err)
	}

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.UTC().String())
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.UTC().String())
	d.Set("name", resp.Name)

	return nil
}

func resourceAwsLexBotAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	hasChanges := false

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:    aws.String(d.Get("bot_name").(string)),
		BotVersion: aws.String(d.Get("bot_version").(string)),
		Checksum:   aws.String(d.Get("checksum").(string)),
		Name:       aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
		hasChanges = true
	}

	if hasChanges {
		_, err := conn.PutBotAlias(input)
		if err != nil {
			return fmt.Errorf("error updating Lex bot alias %s: %s", d.Id(), err)
		}
	}

	return resourceAwsLexBotAliasRead(d, meta)
}

func resourceAwsLexBotAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	botName := d.Get("bot_name").(string)
	name := d.Get("name").(string)

	_, err := conn.DeleteBotAlias(&lexmodelbuildingservice.DeleteBotAliasInput{
		BotName: aws.String(botName),
		Name:    aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("error deleteing Lex bot alias %s: %s", d.Id(), err)
	}

	// Ensure the bot aliases is actually deleted before moving on. This avoids issues with
	// deleting bots that have associated bot aliases.
	for {
		_, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(name),
		})
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return nil
			}

			return fmt.Errorf("could not get Lex bot alias, %s %s", botName, name)
		}
	}

	return nil
}

func resourceAwsLexBotAliasImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid bot alias resource id, expected BOT_NAME.BOT_ALIAS_NAME")
	}

	d.SetId(parts[1])
	d.Set("bot_name", parts[0])
	d.Set("name", parts[1])

	return []*schema.ResourceData{d}, nil
}
