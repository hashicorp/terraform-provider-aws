package lexmodels

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	botAliasCreateTimeout = 1 * time.Minute
	botAliasUpdateTimeout = 1 * time.Minute
	botAliasDeleteTimeout = 5 * time.Minute
)

func ResourceBotAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceBotAliasCreate,
		Read:   resourceBotAliasRead,
		Update: resourceBotAliasUpdate,
		Delete: resourceBotAliasDelete,
		Importer: &schema.ResourceImporter{
			State: resourceBotAliasImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(botAliasCreateTimeout),
			Update: schema.DefaultTimeout(botAliasUpdateTimeout),
			Delete: schema.DefaultTimeout(botAliasDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validBotName,
			},
			"bot_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validBotVersion,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conversation_logs": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iam_role_arn": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(20, 2048),
								verify.ValidARN,
							),
						},
						// Currently the API docs do not list a min and max for this list.
						// https://docs.aws.amazon.com/lex/latest/dg/API_PutBotAlias.html#lex-PutBotAlias-request-conversationLogs
						"log_settings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     logSettings,
						},
					},
				},
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validBotAliasName,
			},
		},
	}
}

var validBotAliasName = validation.All(
	validation.StringLenBetween(1, 100),
	validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
)

func resourceBotAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get("name").(string)
	id := fmt.Sprintf("%s:%s", botName, botAliasName)

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:     aws.String(botName),
		BotVersion:  aws.String(d.Get("bot_version").(string)),
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(botAliasName),
	}

	if v, ok := d.GetOk("conversation_logs"); ok {
		conversationLogs, err := expandConversationLogs(v)
		if err != nil {
			return err
		}
		input.ConversationLogs = conversationLogs
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err := conn.PutBotAlias(input)

		input.Checksum = output.Checksum
		// IAM eventual consistency
		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeBadRequestException, "Lex can't access your IAM role") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q bot alias still creating, another operation is pending: %w", id, err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) { // nosemgrep: helper-schema-TimeoutError-check-doesnt-return-output
		_, err = conn.PutBotAlias(input)
	}

	if err != nil {
		return fmt.Errorf("error creating bot alias '%s': %w", id, err)
	}

	d.SetId(id)

	return resourceBotAliasRead(d, meta)
}

func resourceBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	resp, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(d.Get("bot_name").(string)),
		Name:    aws.String(d.Get("name").(string)),
	})
	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		log.Printf("[WARN] Bot alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting bot alias '%s': %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "lex",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("bot:%s", d.Id()),
	}
	d.Set("arn", arn.String())

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)

	if resp.ConversationLogs != nil {
		d.Set("conversation_logs", flattenConversationLogs(resp.ConversationLogs))
	}

	return nil
}

func resourceBotAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:    aws.String(d.Get("bot_name").(string)),
		BotVersion: aws.String(d.Get("bot_version").(string)),
		Checksum:   aws.String(d.Get("checksum").(string)),
		Name:       aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("conversation_logs"); ok {
		conversationLogs, err := expandConversationLogs(v)
		if err != nil {
			return err
		}
		input.ConversationLogs = conversationLogs
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutBotAlias(input)

		// IAM eventual consistency
		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeBadRequestException, "Lex can't access your IAM role") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q bot alias still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBotAlias(input)
	}

	if err != nil {
		return fmt.Errorf("error updating bot alias '%s': %w", d.Id(), err)
	}

	return resourceBotAliasRead(d, meta)
}

func resourceBotAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LexModelsConn

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get("name").(string)

	input := &lexmodelbuildingservice.DeleteBotAliasInput{
		BotName: aws.String(botName),
		Name:    aws.String(botAliasName),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteBotAlias(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("'%q': bot alias still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteBotAlias(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting bot alias '%s': %w", d.Id(), err)
	}

	_, err = waitBotAliasDeleted(conn, botAliasName, botName)

	return err
}

func resourceBotAliasImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Lex Bot Alias resource id '%s', expected BOT_NAME:BOT_ALIAS_NAME", d.Id())
	}

	d.Set("bot_name", parts[0])
	d.Set("name", parts[1])

	return []*schema.ResourceData{d}, nil
}

var logSettings = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"destination": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.Destination_Values(), false),
		},
		"kms_key_arn": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(20, 2048),
				verify.ValidARN,
			),
		},
		"log_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.LogType_Values(), false),
		},
		"resource_arn": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 2048),
				verify.ValidARN,
			),
		},
		"resource_prefix": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

func flattenConversationLogs(response *lexmodelbuildingservice.ConversationLogsResponse) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			"iam_role_arn": aws.StringValue(response.IamRoleArn),
			"log_settings": flattenLogSettings(response.LogSettings),
		},
	}
}

func expandConversationLogs(rawObject interface{}) (*lexmodelbuildingservice.ConversationLogsRequest, error) {
	request := rawObject.([]interface{})[0].(map[string]interface{})

	logSettings, err := expandLogSettings(request["log_settings"].(*schema.Set).List())
	if err != nil {
		return nil, err
	}
	return &lexmodelbuildingservice.ConversationLogsRequest{
		IamRoleArn:  aws.String(request["iam_role_arn"].(string)),
		LogSettings: logSettings,
	}, nil
}

func flattenLogSettings(responses []*lexmodelbuildingservice.LogSettingsResponse) (flattened []map[string]interface{}) {
	for _, response := range responses {
		flattened = append(flattened, map[string]interface{}{
			"destination":     response.Destination,
			"kms_key_arn":     response.KmsKeyArn,
			"log_type":        response.LogType,
			"resource_arn":    response.ResourceArn,
			"resource_prefix": response.ResourcePrefix,
		})
	}
	return
}

func expandLogSettings(rawValues []interface{}) ([]*lexmodelbuildingservice.LogSettingsRequest, error) {
	requests := make([]*lexmodelbuildingservice.LogSettingsRequest, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}
		destination := value["destination"].(string)
		request := &lexmodelbuildingservice.LogSettingsRequest{
			Destination: aws.String(destination),
			LogType:     aws.String(value["log_type"].(string)),
			ResourceArn: aws.String(value["resource_arn"].(string)),
		}

		if v, ok := value["kms_key_arn"]; ok && v != "" {
			if destination != lexmodelbuildingservice.DestinationS3 {
				return nil, fmt.Errorf("`kms_key_arn` cannot be specified when `destination` is %q", destination)
			}
			request.KmsKeyArn = aws.String(value["kms_key_arn"].(string))
		}

		requests = append(requests, request)
	}

	return requests, nil
}
