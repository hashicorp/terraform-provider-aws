package connect

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVocabulary() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVocabularyCreate,
		ReadWithoutTimeout:   resourceVocabularyRead,
		UpdateWithoutTimeout: resourceVocabularyUpdate,
		DeleteWithoutTimeout: resourceVocabularyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(vocabularyCreatedTimeout),
			Delete: schema.DefaultTimeout(vocabularyDeletedTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 60000),
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"language_code": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.VocabularyLanguageCode_Values(), false),
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 140),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z._-]+`), "must contain only alphanumeric, period, underscore, and hyphen characters"),
				),
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vocabulary_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVocabularyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	instanceID := d.Get("instance_id").(string)
	vocabularyName := d.Get("name").(string)

	input := &connect.CreateVocabularyInput{
		ClientToken:    aws.String(resource.UniqueId()),
		InstanceId:     aws.String(instanceID),
		Content:        aws.String(d.Get("content").(string)),
		LanguageCode:   aws.String(d.Get("language_code").(string)),
		VocabularyName: aws.String(vocabularyName),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Connect Vocabulary %s", input)
	output, err := conn.CreateVocabularyWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Vocabulary (%s): %w", vocabularyName, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Vocabulary (%s): empty output", vocabularyName))
	}

	vocabularyID := aws.StringValue(output.VocabularyId)

	d.SetId(fmt.Sprintf("%s:%s", instanceID, vocabularyID))

	// waiter since the status changes from CREATION_IN_PROGRESS to either ACTIVE or CREATION_FAILED
	if _, err := waitVocabularyCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), instanceID, vocabularyID); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Vocabulary (%s) creation: %w", d.Id(), err))
	}

	return resourceVocabularyRead(ctx, d, meta)
}

func resourceVocabularyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID, vocabularyID, err := VocabularyParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := conn.DescribeVocabularyWithContext(ctx, &connect.DescribeVocabularyInput{
		InstanceId:   aws.String(instanceID),
		VocabularyId: aws.String(vocabularyID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Vocabulary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Vocabulary (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.Vocabulary == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Vocabulary (%s): empty response", d.Id()))
	}

	vocabulary := resp.Vocabulary

	d.Set("arn", vocabulary.Arn)
	d.Set("content", vocabulary.Content)
	d.Set("failure_reason", vocabulary.FailureReason)
	d.Set("instance_id", instanceID)
	d.Set("language_code", vocabulary.LanguageCode)
	d.Set("last_modified_time", vocabulary.LastModifiedTime.Format(time.RFC3339))
	d.Set("name", vocabulary.Name)
	d.Set("state", vocabulary.State)
	d.Set("vocabulary_id", vocabulary.Id)

	tags := KeyValueTags(vocabulary.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceVocabularyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceVocabularyRead(ctx, d, meta)
}

func resourceVocabularyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn()

	instanceID, vocabularyID, err := VocabularyParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteVocabularyWithContext(ctx, &connect.DeleteVocabularyInput{
		InstanceId:   aws.String(instanceID),
		VocabularyId: aws.String(vocabularyID),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Vocabulary (%s): %w", d.Id(), err))
	}

	if _, err := waitVocabularyDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete), instanceID, vocabularyID); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Vocabulary (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func VocabularyParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:vocabularyID", id)
	}

	return parts[0], parts[1], nil
}
