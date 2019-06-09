package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsLexIntent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexIntentCreate,
		Read:   resourceAwsLexIntentRead,
		Update: resourceAwsLexIntentUpdate,
		Delete: resourceAwsLexIntentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conclusion_statement": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexStatementResource,
			},
			"confirmation_prompt": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexPromptResource,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      lexDescriptionDefault,
				ValidateFunc: validation.StringLenBetween(lexDescriptionMinLength, lexDescriptionMaxLength),
			},
			"dialog_code_hook": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexCodeHookResource,
			},
			"follow_up_prompt": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexFollowUpPromptResource,
			},
			// Must be required because required by updates even though optional for creates
			"fulfillment_activity": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexFulfilmentActivityResource,
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
			"parent_intent_signature": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rejection_statement": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     lexStatementResource,
			},
			"sample_utterances": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: lexUtterancesMin,
				MaxItems: lexUtterancesMax,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(lexUtteranceMinLength, lexUtteranceMaxLength),
				},
			},
			"slot": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: lexSlotsMin,
				MaxItems: lexSlotsMax,
				Elem:     lexSlotResource,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  lexVersionDefault,
				ValidateFunc: validation.All(
					validation.StringLenBetween(lexVersionMinLength, lexVersionMaxLength),
					validation.StringMatch(regexp.MustCompile(lexVersionRegex), ""),
				),
			},
		},
	}
}

func resourceAwsLexIntentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutIntentInput{
		Name: aws.String(name),
	}

	// optional attributes

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandLexPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandLexCodeHook(expandLexObject(v))
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandLexFollowUpPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandLexFulfilmentActivity(expandLexObject(v))
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandLexSlots(expandLexSet(v.(*schema.Set)))
	}

	if _, err := conn.PutIntent(input); err != nil {
		return fmt.Errorf("error creating Lex Intent %s: %s", name, err)
	}

	d.SetId(name)

	return resourceAwsLexIntentRead(d, meta)
}

func resourceAwsLexIntentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	version := "$LATEST"
	if v, ok := d.GetOk("version"); ok {
		version = v.(string)
	}

	resp, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(version),
	})
	if err != nil {
		if isAWSErr(err, "NotFoundException", "") {
			log.Printf("[WARN] Intent (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting intent %s: %s", d.Id(), err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("name", resp.Name)
	d.Set("version", resp.Version)

	// optional attributes

	if resp.ConclusionStatement != nil {
		d.Set("conclusion_statement", flattenLexObject(flattenLexStatement(resp.ConclusionStatement)))
	}

	if resp.ConfirmationPrompt != nil {
		d.Set("confirmation_prompt", flattenLexObject(flattenLexPrompt(resp.ConfirmationPrompt)))
	}

	if resp.Description != nil {
		d.Set("description", resp.Description)
	}

	if resp.DialogCodeHook != nil {
		d.Set("dialog_code_hook", flattenLexObject(flattenLexCodeHook(resp.DialogCodeHook)))
	}

	if resp.FollowUpPrompt != nil {
		d.Set("follow_up_prompt", flattenLexObject(flattenLexFollowUpPrompt(resp.FollowUpPrompt)))
	}

	if resp.FulfillmentActivity != nil {
		d.Set("fulfillment_activity", flattenLexObject(flattenLexFulfilmentActivity(resp.FulfillmentActivity)))
	}

	if resp.ParentIntentSignature != nil {
		d.Set("parent_intent_signature", resp.ParentIntentSignature)
	}

	if resp.RejectionStatement != nil {
		d.Set("rejection_statement", flattenLexObject(flattenLexStatement(resp.RejectionStatement)))
	}

	if resp.SampleUtterances != nil {
		d.Set("sample_utterances", resp.SampleUtterances)
	}

	if resp.Slots != nil {
		d.Set("slot", flattenLexSlots(resp.Slots))
	}

	return nil
}

func resourceAwsLexIntentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutIntentInput{
		Checksum: aws.String(d.Get("checksum").(string)),
		Name:     aws.String(d.Id()),
	}

	// optional attributes

	if v, ok := d.GetOk("conclusion_statement"); ok {
		input.ConclusionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("confirmation_prompt"); ok {
		input.ConfirmationPrompt = expandLexPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dialog_code_hook"); ok {
		input.DialogCodeHook = expandLexCodeHook(expandLexObject(v))
	}

	if v, ok := d.GetOk("follow_up_prompt"); ok {
		input.FollowUpPrompt = expandLexFollowUpPrompt(expandLexObject(v))
	}

	if v, ok := d.GetOk("fulfillment_activity"); ok {
		input.FulfillmentActivity = expandLexFulfilmentActivity(expandLexObject(v))
	}

	if v, ok := d.GetOk("parent_intent_signature"); ok {
		input.ParentIntentSignature = aws.String(v.(string))
	}

	if v, ok := d.GetOk("rejection_statement"); ok {
		input.RejectionStatement = expandLexStatement(expandLexObject(v))
	}

	if v, ok := d.GetOk("sample_utterances"); ok {
		input.SampleUtterances = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("slot"); ok {
		input.Slots = expandLexSlots(expandLexSet(v.(*schema.Set)))
	}

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.PutIntent(input)
	})
	if err != nil {
		return fmt.Errorf("error updating intent %s: %s", d.Id(), err)
	}

	return resourceAwsLexIntentRead(d, meta)
}

func resourceAwsLexIntentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	_, err := RetryOnAwsCodes([]string{"ConflictException"}, func() (interface{}, error) {
		return conn.DeleteIntent(&lexmodelbuildingservice.DeleteIntentInput{
			Name: aws.String(d.Id()),
		})
	})

	if err != nil {
		return fmt.Errorf("error deleteing intent %s: %s", d.Id(), err)
	}

	return nil
}
