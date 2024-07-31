// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_vocabulary", name="Vocabulary")
// @Tags(identifierAttribute="arn")
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(connect.VocabularyLanguageCode_Values(), false),
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

func resourceVocabularyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	vocabularyName := d.Get(names.AttrName).(string)
	input := &connect.CreateVocabularyInput{
		ClientToken:    aws.String(id.UniqueId()),
		InstanceId:     aws.String(instanceID),
		Content:        aws.String(d.Get(names.AttrContent).(string)),
		LanguageCode:   aws.String(d.Get(names.AttrLanguageCode).(string)),
		Tags:           getTagsIn(ctx),
		VocabularyName: aws.String(vocabularyName),
	}

	log.Printf("[DEBUG] Creating Connect Vocabulary %s", input)
	output, err := conn.CreateVocabularyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Vocabulary (%s): %s", vocabularyName, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Vocabulary (%s): empty output", vocabularyName)
	}

	vocabularyID := aws.StringValue(output.VocabularyId)

	d.SetId(fmt.Sprintf("%s:%s", instanceID, vocabularyID))

	// waiter since the status changes from CREATION_IN_PROGRESS to either ACTIVE or CREATION_FAILED
	if _, err := waitVocabularyCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), instanceID, vocabularyID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Vocabulary (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceVocabularyRead(ctx, d, meta)...)
}

func resourceVocabularyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, vocabularyID, err := VocabularyParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeVocabularyWithContext(ctx, &connect.DescribeVocabularyInput{
		InstanceId:   aws.String(instanceID),
		VocabularyId: aws.String(vocabularyID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect Vocabulary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Vocabulary (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.Vocabulary == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Vocabulary (%s): empty response", d.Id())
	}

	vocabulary := resp.Vocabulary

	d.Set(names.AttrARN, vocabulary.Arn)
	d.Set(names.AttrContent, vocabulary.Content)
	d.Set("failure_reason", vocabulary.FailureReason)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrLanguageCode, vocabulary.LanguageCode)
	d.Set("last_modified_time", vocabulary.LastModifiedTime.Format(time.RFC3339))
	d.Set(names.AttrName, vocabulary.Name)
	d.Set(names.AttrState, vocabulary.State)
	d.Set("vocabulary_id", vocabulary.Id)

	setTagsOut(ctx, resp.Vocabulary.Tags)

	return diags
}

func resourceVocabularyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceVocabularyRead(ctx, d, meta)
}

func resourceVocabularyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, vocabularyID, err := VocabularyParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteVocabularyWithContext(ctx, &connect.DeleteVocabularyInput{
		InstanceId:   aws.String(instanceID),
		VocabularyId: aws.String(vocabularyID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Vocabulary (%s): %s", d.Id(), err)
	}

	if _, err := waitVocabularyDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete), instanceID, vocabularyID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Vocabulary (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func VocabularyParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:vocabularyID", id)
	}

	return parts[0], parts[1], nil
}
