package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(lexSlotTypeDeleteRetryTimeoutMinutes * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexBotNameMinLength, lexBotNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
			"bot_version": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(lexDescriptionMinLength, lexDescriptionMaxLength),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexNameMinLength, lexNameMaxLength),
					validation.StringMatch(regexp.MustCompile(lexNameRegex), ""),
				),
			},
		},
	}
}

func resourceAwsLexBotAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:    aws.String(d.Get("bot_name").(string)),
		BotVersion: aws.String(d.Get("bot_version").(string)),
		Name:       aws.String(name),
	}

	// optional attributes

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if _, err := conn.PutBotAlias(input); err != nil {
		return fmt.Errorf("error creating bot alias %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexBotAliasRead(d, meta)
}

func resourceAwsLexBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(d.Get("bot_name").(string)),
		Name:    aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, "NotFoundException", "") {
			log.Printf("[WARN] Bot alias (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting bot alias %s: %s", d.Id(), err)
	}

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)

	return nil
}

func resourceAwsLexBotAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:    aws.String(d.Get("bot_name").(string)),
		BotVersion: aws.String(d.Get("bot_version").(string)),
		Checksum:   aws.String(d.Get("checksum").(string)),
		Name:       aws.String(d.Id()),
	}

	// optional attributes

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.PutBotAlias(input)
	})
	if err != nil {
		return fmt.Errorf("error updating bot alias %s: %s", d.Id(), err)
	}

	return resourceAwsLexBotAliasRead(d, meta)
}

func resourceAwsLexBotAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	botName := d.Get("bot_name").(string)
	name := d.Get("name").(string)

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.DeleteBotAlias(&lexmodelbuildingservice.DeleteBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(name),
		})
	})
	if err != nil {
		return fmt.Errorf("error deleteing bot alias %s: %s", d.Id(), err)
	}

	// Ensure the bot alias is actually deleted before moving on. This avoids issues with deleting bots that have
	// associated bot aliases.

	return resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(name),
		})
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
}

func resourceAwsLexBotAliasImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Lex Bot Alias resource id, expected BOT_NAME.BOT_ALIAS_NAME")
	}

	d.SetId(parts[1])
	d.Set("bot_name", parts[0])
	d.Set("name", parts[1])

	return []*schema.ResourceData{d}, nil
}
