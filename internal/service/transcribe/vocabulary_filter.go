package transcribe

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceVocabularyFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVocabularyFilterCreate,
		ReadWithoutTimeout:   resourceVocabularyFilterRead,
		UpdateWithoutTimeout: resourceVocabularyFilterUpdate,
		DeleteWithoutTimeout: resourceVocabularyFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				ValidateFunc: validation.StringInSlice(validateLanguageCodes(types.LanguageCode("").Values()), false),
			},
			"words": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     256,
				ExactlyOneOf: []string{"words", "vocabulary_filter_file_uri"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vocabulary_filter_file_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"words", "vocabulary_filter_file_uri"},
				ValidateFunc: validation.StringLenBetween(1, 2000),
			},
			"vocabulary_filter_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIfChange("words", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) > 0 && len(new.([]interface{})) == 0
			}),
			customdiff.ForceNewIfChange("vocabulary_filter_file_uri", func(_ context.Context, old, new, meta interface{}) bool {
				return new.(string) == ""
			}),
		),
	}
}

const (
	ResNameVocabularyFilter = "Vocabulary Filter"
)

func resourceVocabularyFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	in := &transcribe.CreateVocabularyFilterInput{
		VocabularyFilterName: aws.String(d.Get("vocabulary_filter_name").(string)),
		LanguageCode:         types.LanguageCode(d.Get("language_code").(string)),
	}

	if v, ok := d.GetOk("vocabulary_filter_file_uri"); ok {
		in.VocabularyFilterFileUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("words"); ok {
		in.Words = flex.ExpandStringValueList(v.([]interface{}))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateVocabularyFilter(ctx, in)
	if err != nil {
		return create.DiagError(names.Transcribe, create.ErrActionCreating, ResNameVocabularyFilter, d.Get("vocabulary_filter_name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.Transcribe, create.ErrActionCreating, ResNameVocabularyFilter, d.Get("vocabulary_filter_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.VocabularyFilterName))

	return resourceVocabularyFilterRead(ctx, d, meta)
}

func resourceVocabularyFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	out, err := FindVocabularyFilterByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transcribe VocabularyFilter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Transcribe, create.ErrActionReading, ResNameVocabularyFilter, d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "transcribe",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("vocabulary-filter/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("vocabulary_filter_name", out.VocabularyFilterName)
	d.Set("language_code", out.LanguageCode)

	// GovCloud does not set a download URI
	downloadUri := aws.ToString(out.DownloadUri)
	if downloadUri == "" {
		downloadUri = "NONE"
	}
	d.Set("download_uri", downloadUri)

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return create.DiagError(names.Transcribe, create.ErrActionReading, ResNameVocabularyFilter, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Transcribe, create.ErrActionSetting, ResNameVocabularyFilter, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Transcribe, create.ErrActionSetting, ResNameVocabularyFilter, d.Id(), err)
	}

	return nil
}

func resourceVocabularyFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	if d.HasChangesExcept("tags", "tags_all") {
		in := &transcribe.UpdateVocabularyFilterInput{
			VocabularyFilterName: aws.String(d.Id()),
		}

		if d.HasChanges("vocabulary_filter_file_uri", "words") {
			if d.Get("vocabulary_filter_file_uri").(string) != "" {
				in.VocabularyFilterFileUri = aws.String(d.Get("vocabulary_filter_file_uri").(string))
			} else {
				in.Words = flex.ExpandStringValueList(d.Get("words").([]interface{}))
			}
		}

		log.Printf("[DEBUG] Updating Transcribe VocabularyFilter (%s): %#v", d.Id(), in)
		_, err := conn.UpdateVocabularyFilter(ctx, in)
		if err != nil {
			return create.DiagError(names.Transcribe, create.ErrActionUpdating, ResNameVocabularyFilter, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating Transcribe VocabularyFilter (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceVocabularyFilterRead(ctx, d, meta)
}

func resourceVocabularyFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeClient()

	log.Printf("[INFO] Deleting Transcribe VocabularyFilter %s", d.Id())

	_, err := conn.DeleteVocabularyFilter(ctx, &transcribe.DeleteVocabularyFilterInput{
		VocabularyFilterName: aws.String(d.Id()),
	})

	if err != nil {
		var bre *types.BadRequestException
		if errors.As(err, &bre) {
			return nil
		}

		return create.DiagError(names.Transcribe, create.ErrActionDeleting, ResNameVocabularyFilter, d.Id(), err)
	}

	return nil
}

func FindVocabularyFilterByName(ctx context.Context, conn *transcribe.Client, id string) (*transcribe.GetVocabularyFilterOutput, error) {
	in := &transcribe.GetVocabularyFilterInput{
		VocabularyFilterName: aws.String(id),
	}
	out, err := conn.GetVocabularyFilter(ctx, in)
	if err != nil {
		var bre *types.BadRequestException
		if errors.As(err, &bre) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
