package transcribe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMedicalVocabulary() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMedicalVocabularyCreate,
		ReadWithoutTimeout:   resourceMedicalVocabularyRead,
		UpdateWithoutTimeout: resourceMedicalVocabularyUpdate,
		DeleteWithoutTimeout: resourceMedicalVocabularyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"download_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"language_code": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"en-US"}, false), // en-US is the only supported language for this service
			},
			"vocabulary_file_uri": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 2000),
			},
			"vocabulary_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMedicalVocabularyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	vocabularyName := d.Get("vocabulary_name").(string)
	in := &transcribe.CreateMedicalVocabularyInput{
		VocabularyName:    aws.String(vocabularyName),
		VocabularyFileUri: aws.String(d.Get("vocabulary_file_uri").(string)),
		LanguageCode:      types.LanguageCode(d.Get("language_code").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateMedicalVocabulary(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Transcribe MedicalVocabulary (%s): %s", d.Get("vocabulary_name").(string), err)
	}

	if out == nil {
		return diag.Errorf("creating Amazon Transcribe MedicalVocabulary (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.ToString(out.VocabularyName))

	if _, err := waitMedicalVocabularyCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Transcribe MedicalVocabulary (%s) create: %s", d.Id(), err)
	}

	return resourceMedicalVocabularyRead(ctx, d, meta)
}

func resourceMedicalVocabularyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	out, err := FindMedicalVocabularyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transcribe MedicalVocabulary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "transcribe",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("medical-vocabulary/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("download_uri", out.DownloadUri)
	d.Set("vocabulary_name", out.VocabularyName)
	d.Set("language_code", out.LanguageCode)

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return diag.Errorf("listing tags for Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceMedicalVocabularyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	if d.HasChangesExcept("tags", "tags_all") {
		in := &transcribe.UpdateMedicalVocabularyInput{
			VocabularyName: aws.String(d.Id()),
			LanguageCode:   types.LanguageCode(d.Get("language_code").(string)),
		}

		if d.HasChanges("vocabulary_file_uri") {
			in.VocabularyFileUri = aws.String(d.Get("vocabulary_file_uri").(string))
		}

		log.Printf("[DEBUG] Updating Transcribe MedicalVocabulary (%s): %#v", d.Id(), in)
		_, err := conn.UpdateMedicalVocabulary(ctx, in)
		if err != nil {
			return diag.Errorf("updating Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
		}

		if _, err := waitMedicalVocabularyUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Transcribe MedicalVocabulary (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating Transcribe MedicalVocabulary (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceMedicalVocabularyRead(ctx, d, meta)
}

func resourceMedicalVocabularyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	log.Printf("[INFO] Deleting Transcribe MedicalVocabulary %s", d.Id())

	_, err := conn.DeleteMedicalVocabulary(ctx, &transcribe.DeleteMedicalVocabularyInput{
		VocabularyName: aws.String(d.Id()),
	})

	var badRequestException *types.BadRequestException
	if errors.As(err, &badRequestException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
	}

	if _, err := waitMedicalVocabularyDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Transcribe MedicalVocabulary (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func waitMedicalVocabularyCreated(ctx context.Context, conn *transcribe.Client, id string, timeout time.Duration) (*transcribe.GetMedicalVocabularyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   medicalVocabularyStatus(types.VocabularyStatePending),
		Target:                    medicalVocabularyStatus(types.VocabularyStateReady),
		Refresh:                   statusMedicalVocabulary(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*transcribe.GetMedicalVocabularyOutput); ok {
		return out, err
	}

	return nil, err
}

func waitMedicalVocabularyUpdated(ctx context.Context, conn *transcribe.Client, id string, timeout time.Duration) (*transcribe.GetMedicalVocabularyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   medicalVocabularyStatus(types.VocabularyStatePending),
		Target:                    medicalVocabularyStatus(types.VocabularyStateReady),
		Refresh:                   statusMedicalVocabulary(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*transcribe.GetMedicalVocabularyOutput); ok {
		return out, err
	}

	return nil, err
}

func waitMedicalVocabularyDeleted(ctx context.Context, conn *transcribe.Client, id string, timeout time.Duration) (*transcribe.GetMedicalVocabularyOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: medicalVocabularyStatus(types.VocabularyStatePending),
		Target:  []string{},
		Refresh: statusMedicalVocabulary(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*transcribe.GetMedicalVocabularyOutput); ok {
		return out, err
	}

	return nil, err
}

func statusMedicalVocabulary(ctx context.Context, conn *transcribe.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindMedicalVocabularyByName(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.VocabularyState), nil
	}
}

func FindMedicalVocabularyByName(ctx context.Context, conn *transcribe.Client, id string) (*transcribe.GetMedicalVocabularyOutput, error) {
	in := &transcribe.GetMedicalVocabularyInput{
		VocabularyName: aws.String(id),
	}

	out, err := conn.GetMedicalVocabulary(ctx, in)

	var badRequestException *types.BadRequestException
	if errors.As(err, &badRequestException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func medicalVocabularyStatus(in ...types.VocabularyState) []string {
	var s []string

	for _, v := range in {
		s = append(s, string(v))
	}

	return s
}
