// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dataexchange_revision", name="Revision")
// @Tags(identifierAttribute="arn")
func ResourceRevision() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRevisionCreate,
		ReadWithoutTimeout:   resourceRevisionRead,
		UpdateWithoutTimeout: resourceRevisionUpdate,
		DeleteWithoutTimeout: resourceRevisionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16348),
			},
			"data_set_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRevisionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	input := &dataexchange.CreateRevisionInput{
		DataSetId: aws.String(d.Get("data_set_id").(string)),
		Comment:   aws.String(d.Get(names.AttrComment).(string)),
		Tags:      getTagsIn(ctx),
	}

	out, err := conn.CreateRevision(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataExchange Revision: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.ToString(out.DataSetId), aws.ToString(out.Id)))

	return append(diags, resourceRevisionRead(ctx, d, meta)...)
}

func resourceRevisionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	dataSetId, revisionId, err := RevisionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataExchange Revision (%s): %s", d.Id(), err)
	}

	revision, err := FindRevisionById(ctx, conn, dataSetId, revisionId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataExchange Revision (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataExchange Revision (%s): %s", d.Id(), err)
	}

	d.Set("data_set_id", revision.DataSetId)
	d.Set(names.AttrComment, revision.Comment)
	d.Set(names.AttrARN, revision.Arn)
	d.Set("revision_id", revision.Id)

	setTagsOut(ctx, revision.Tags)

	return diags
}

func resourceRevisionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &dataexchange.UpdateRevisionInput{
			RevisionId: aws.String(d.Get("revision_id").(string)),
			DataSetId:  aws.String(d.Get("data_set_id").(string)),
		}

		if d.HasChange(names.AttrComment) {
			input.Comment = aws.String(d.Get(names.AttrComment).(string))
		}

		log.Printf("[DEBUG] Updating DataExchange Revision: %s", d.Id())
		_, err := conn.UpdateRevision(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataExchange Revision (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRevisionRead(ctx, d, meta)...)
}

func resourceRevisionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	input := &dataexchange.DeleteRevisionInput{
		RevisionId: aws.String(d.Get("revision_id").(string)),
		DataSetId:  aws.String(d.Get("data_set_id").(string)),
	}

	log.Printf("[DEBUG] Deleting DataExchange Revision: %s", d.Id())
	_, err := conn.DeleteRevision(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DataExchange Revision: %s", err)
	}

	return diags
}

func RevisionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%s), expected DATA-SET_ID:REVISION-ID", id)
}
