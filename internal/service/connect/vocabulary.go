// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_vocabulary", name="Vocabulary")
// @Tags(identifierAttribute="arn")
func resourceVocabulary() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVocabularyCreate,
		ReadWithoutTimeout:   resourceVocabularyRead,
		UpdateWithoutTimeout: resourceVocabularyUpdate,
		DeleteWithoutTimeout: resourceVocabularyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			// It takes about 90 minutes for Amazon Connect to delete a vocabulary.
			// https://docs.aws.amazon.com/connect/latest/adminguide/add-custom-vocabulary.html
			Delete: schema.DefaultTimeout(100 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContent: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 60000),
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrLanguageCode: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VocabularyLanguageCode](),
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 140),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), "must contain only alphanumeric, period, underscore, and hyphen characters"),
				),
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vocabulary_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceVocabularyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	vocabularyName := d.Get(names.AttrName).(string)
	input := &connect.CreateVocabularyInput{
		ClientToken:    aws.String(id.UniqueId()),
		InstanceId:     aws.String(instanceID),
		Content:        aws.String(d.Get(names.AttrContent).(string)),
		LanguageCode:   awstypes.VocabularyLanguageCode(d.Get(names.AttrLanguageCode).(string)),
		Tags:           getTagsIn(ctx),
		VocabularyName: aws.String(vocabularyName),
	}

	output, err := conn.CreateVocabulary(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Vocabulary (%s): %s", vocabularyName, err)
	}

	vocabularyID := aws.ToString(output.VocabularyId)
	id := vocabularyCreateResourceID(instanceID, vocabularyID)
	d.SetId(id)

	if _, err := waitVocabularyCreated(ctx, conn, instanceID, vocabularyID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Connect Vocabulary (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVocabularyRead(ctx, d, meta)...)
}

func resourceVocabularyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, vocabularyID, err := vocabularyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	vocabulary, err := findVocabularyByTwoPartKey(ctx, conn, instanceID, vocabularyID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Vocabulary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Vocabulary (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, vocabulary.Arn)
	d.Set(names.AttrContent, vocabulary.Content)
	d.Set("failure_reason", vocabulary.FailureReason)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrLanguageCode, vocabulary.LanguageCode)
	d.Set("last_modified_time", vocabulary.LastModifiedTime.Format(time.RFC3339))
	d.Set(names.AttrName, vocabulary.Name)
	d.Set(names.AttrState, vocabulary.State)
	d.Set("vocabulary_id", vocabulary.Id)

	setTagsOut(ctx, vocabulary.Tags)

	return diags
}

func resourceVocabularyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceVocabularyRead(ctx, d, meta)
}

func resourceVocabularyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, vocabularyID, err := vocabularyParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect Vocabulary: %s", d.Id())
	input := connect.DeleteVocabularyInput{
		InstanceId:   aws.String(instanceID),
		VocabularyId: aws.String(vocabularyID),
	}
	_, err = conn.DeleteVocabulary(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Vocabulary (%s): %s", d.Id(), err)
	}

	if _, err := waitVocabularyDeleted(ctx, conn, instanceID, vocabularyID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Connect Vocabulary (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const vocabularyResourceIDSeparator = ":"

func vocabularyCreateResourceID(instanceID, vocabularyID string) string {
	parts := []string{instanceID, vocabularyID}
	id := strings.Join(parts, vocabularyResourceIDSeparator)

	return id
}

func vocabularyParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, vocabularyResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]svocabularyID", id, vocabularyResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findVocabularyByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, vocabularyID string) (*awstypes.Vocabulary, error) {
	input := &connect.DescribeVocabularyInput{
		InstanceId:   aws.String(instanceID),
		VocabularyId: aws.String(vocabularyID),
	}

	return findVocabulary(ctx, conn, input)
}

func findVocabulary(ctx context.Context, conn *connect.Client, input *connect.DescribeVocabularyInput) (*awstypes.Vocabulary, error) {
	output, err := conn.DescribeVocabulary(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Vocabulary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Vocabulary, nil
}

func statusVocabulary(ctx context.Context, conn *connect.Client, instanceID, vocabularyID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findVocabularyByTwoPartKey(ctx, conn, instanceID, vocabularyID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitVocabularyCreated(ctx context.Context, conn *connect.Client, instanceID, vocabularyID string, timeout time.Duration) (*awstypes.Vocabulary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VocabularyStateCreationInProgress),
		Target:  enum.Slice(awstypes.VocabularyStateActive, awstypes.VocabularyStateCreationFailed),
		Refresh: statusVocabulary(ctx, conn, instanceID, vocabularyID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vocabulary); ok {
		if state := output.State; state == awstypes.VocabularyStateCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitVocabularyDeleted(ctx context.Context, conn *connect.Client, instanceID, vocabularyID string, timeout time.Duration) (*awstypes.Vocabulary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VocabularyStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusVocabulary(ctx, conn, instanceID, vocabularyID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vocabulary); ok {
		return output, err
	}

	return nil, err
}
