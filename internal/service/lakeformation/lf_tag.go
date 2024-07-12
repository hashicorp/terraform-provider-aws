// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// This value is defined by AWS API
const lfTagsValuesMaxBatchSize = 50

// @SDKResource("aws_lakeformation_lf_tag")
func ResourceLFTag() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLFTagCreate,
		ReadWithoutTimeout:   resourceLFTagRead,
		UpdateWithoutTimeout: resourceLFTagUpdate,
		DeleteWithoutTimeout: resourceLFTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			names.AttrKey: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrValues: {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				// Soft limit stated in AWS Doc
				// https://docs.aws.amazon.com/lake-formation/latest/dg/TBAC-notes.html
				MaxItems: 1000,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateLFTagValues(),
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceLFTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	tagKey := d.Get(names.AttrKey).(string)
	tagValues := d.Get(names.AttrValues).(*schema.Set)

	var catalogID string
	if v, ok := d.GetOk(names.AttrCatalogID); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}

	tagValueChunks := slices.Chunks(tagValues.List(), lfTagsValuesMaxBatchSize)

	input := &lakeformation.CreateLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
		TagValues: flex.ExpandStringValueList(tagValueChunks[0]),
	}

	_, err := conn.CreateLFTag(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lake Formation LF-Tag: %s", err)
	}

	if len(tagValueChunks) > 1 {
		tagValueChunks = tagValueChunks[1:]

		for _, v := range tagValueChunks {
			in := &lakeformation.UpdateLFTagInput{
				CatalogId:      aws.String(catalogID),
				TagKey:         aws.String(tagKey),
				TagValuesToAdd: flex.ExpandStringValueList(v),
			}

			_, err := conn.UpdateLFTag(ctx, in)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Lake Formation LF-Tag: %s", err)
			}
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, tagKey))

	return append(diags, resourceLFTagRead(ctx, d, meta)...)
}

func resourceLFTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	input := &lakeformation.GetLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	output, err := conn.GetLFTag(ctx, input)
	if !d.IsNewResource() {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			log.Printf("[WARN] Lake Formation LF-Tag (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrKey, output.TagKey)
	d.Set(names.AttrValues, flex.FlattenStringValueSet(output.TagValues))
	d.Set(names.AttrCatalogID, output.CatalogId)

	return diags
}

func resourceLFTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	o, n := d.GetChange(names.AttrValues)
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	toAdd := ns.Difference(os)
	toDelete := os.Difference(ns)

	var toAddChunks, toDeleteChunks [][]interface{}
	if len(toAdd.List()) > 0 {
		toAddChunks = slices.Chunks(toAdd.List(), lfTagsValuesMaxBatchSize)
	}

	if len(toDelete.List()) > 0 {
		toDeleteChunks = slices.Chunks(toDelete.List(), lfTagsValuesMaxBatchSize)
	}

	for {
		if len(toAddChunks) == 0 && len(toDeleteChunks) == 0 {
			break
		}

		input := &lakeformation.UpdateLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		toAddEnd, toDeleteEnd := len(toAddChunks), len(toDeleteChunks)
		var indexAdd, indexDelete int
		if indexAdd < toAddEnd {
			input.TagValuesToAdd = flex.ExpandStringValueList(toAddChunks[0])
			indexAdd++
		}
		if indexDelete < toDeleteEnd {
			input.TagValuesToDelete = flex.ExpandStringValueList(toDeleteChunks[0])
			indexDelete++
		}

		toAddChunks = toAddChunks[indexAdd:]
		toDeleteChunks = toDeleteChunks[indexDelete:]

		_, err = conn.UpdateLFTag(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLFTagRead(ctx, d, meta)...)
}

func resourceLFTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationClient(ctx)

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	input := &lakeformation.DeleteLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	_, err = conn.DeleteLFTag(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	return diags
}

func ReadLFTagID(id string) (string, string, error) {
	catalogID, tagKey, found := strings.Cut(id, ":")

	if !found {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected CATALOG-ID:TAG-KEY", id)
	}

	return catalogID, tagKey, nil
}

func validateLFTagValues() schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(regexache.MustCompile(`^([\p{L}\p{Z}\p{N}_.:\*\/=+\-@%]*)$`), ""),
	)
}
