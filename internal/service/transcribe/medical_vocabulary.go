// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transcribe_medical_vocabulary", name="Medical Vocabulary")
// @Tags(identifierAttribute="arn")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"download_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLanguageCode: {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMedicalVocabularyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TranscribeClient(ctx)

	vocabularyName := d.Get("vocabulary_name").(string)
	in := &transcribe.CreateMedicalVocabularyInput{
		VocabularyName:    aws.String(vocabularyName),
		VocabularyFileUri: aws.String(d.Get("vocabulary_file_uri").(string)),
		LanguageCode:      types.LanguageCode(d.Get(names.AttrLanguageCode).(string)),
		Tags:              getTagsIn(ctx),
	}

	out, err := conn.CreateMedicalVocabulary(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amazon Transcribe MedicalVocabulary (%s): %s", d.Get("vocabulary_name").(string), err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "creating Amazon Transcribe MedicalVocabulary (%s): empty output", d.Get(names.AttrName).(string))
	}

	d.SetId(aws.ToString(out.VocabularyName))

	if _, err := waitMedicalVocabularyCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Amazon Transcribe MedicalVocabulary (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMedicalVocabularyRead(ctx, d, meta)...)
}

func resourceMedicalVocabularyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TranscribeClient(ctx)

	out, err := FindMedicalVocabularyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transcribe MedicalVocabulary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "transcribe",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("medical-vocabulary/%s", d.Id()),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set("download_uri", out.DownloadUri)
	d.Set("vocabulary_name", out.VocabularyName)
	d.Set(names.AttrLanguageCode, out.LanguageCode)

	return diags
}

func resourceMedicalVocabularyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TranscribeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &transcribe.UpdateMedicalVocabularyInput{
			VocabularyName: aws.String(d.Id()),
			LanguageCode:   types.LanguageCode(d.Get(names.AttrLanguageCode).(string)),
		}

		if d.HasChanges("vocabulary_file_uri") {
			in.VocabularyFileUri = aws.String(d.Get("vocabulary_file_uri").(string))
		}

		log.Printf("[DEBUG] Updating Transcribe MedicalVocabulary (%s): %#v", d.Id(), in)
		_, err := conn.UpdateMedicalVocabulary(ctx, in)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
		}

		if _, err := waitMedicalVocabularyUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Transcribe MedicalVocabulary (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMedicalVocabularyRead(ctx, d, meta)...)
}

func resourceMedicalVocabularyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TranscribeClient(ctx)

	log.Printf("[INFO] Deleting Transcribe MedicalVocabulary %s", d.Id())

	_, err := conn.DeleteMedicalVocabulary(ctx, &transcribe.DeleteMedicalVocabularyInput{
		VocabularyName: aws.String(d.Id()),
	})

	var badRequestException *types.BadRequestException
	if errors.As(err, &badRequestException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transcribe MedicalVocabulary (%s): %s", d.Id(), err)
	}

	if _, err := waitMedicalVocabularyDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Transcribe MedicalVocabulary (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func waitMedicalVocabularyCreated(ctx context.Context, conn *transcribe.Client, id string, timeout time.Duration) (*transcribe.GetMedicalVocabularyOutput, error) {
	stateConf := &retry.StateChangeConf{
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
	stateConf := &retry.StateChangeConf{
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
	stateConf := &retry.StateChangeConf{
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

func statusMedicalVocabulary(ctx context.Context, conn *transcribe.Client, id string) retry.StateRefreshFunc {
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
		return nil, &retry.NotFoundError{
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
