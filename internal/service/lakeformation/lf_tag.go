// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"values": {
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
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)

	tagKey := d.Get("key").(string)
	tagValues := d.Get("values").(*schema.Set)

	var catalogID string
	if v, ok := d.GetOk("catalog_id"); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}

	tagValueChunks := splitLFTagValues(tagValues.List(), lfTagsValuesMaxBatchSize)

	input := &lakeformation.CreateLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
		TagValues: flex.ExpandStringList(tagValueChunks[0]),
	}

	_, err := conn.CreateLFTagWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lake Formation LF-Tag: %s", err)
	}

	if len(tagValueChunks) > 1 {
		tagValueChunks = tagValueChunks[1:]

		for _, v := range tagValueChunks {
			in := &lakeformation.UpdateLFTagInput{
				CatalogId:      aws.String(catalogID),
				TagKey:         aws.String(tagKey),
				TagValuesToAdd: flex.ExpandStringList(v),
			}

			_, err := conn.UpdateLFTagWithContext(ctx, in)
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
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	input := &lakeformation.GetLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	output, err := conn.GetLFTagWithContext(ctx, input)
	if !d.IsNewResource() {
		if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Lake Formation LF-Tag (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	d.Set("key", output.TagKey)
	d.Set("values", flex.FlattenStringSet(output.TagValues))
	d.Set("catalog_id", output.CatalogId)

	return diags
}

func resourceLFTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	o, n := d.GetChange("values")
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	toAdd := ns.Difference(os)
	toDelete := os.Difference(ns)

	var toAddChunks, toDeleteChunks [][]interface{}
	if len(toAdd.List()) > 0 {
		toAddChunks = splitLFTagValues(toAdd.List(), lfTagsValuesMaxBatchSize)
	}

	if len(toDelete.List()) > 0 {
		toDeleteChunks = splitLFTagValues(toDelete.List(), lfTagsValuesMaxBatchSize)
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
			input.TagValuesToAdd = flex.ExpandStringList(toAddChunks[indexAdd])
			indexAdd++
		}
		if indexDelete < toDeleteEnd {
			input.TagValuesToDelete = flex.ExpandStringList(toDeleteChunks[indexDelete])
			indexDelete++
		}

		toAddChunks = toAddChunks[indexAdd:]
		toDeleteChunks = toDeleteChunks[indexDelete:]

		_, err = conn.UpdateLFTagWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLFTagRead(ctx, d, meta)...)
}

func resourceLFTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn(ctx)

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	input := &lakeformation.DeleteLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	_, err = conn.DeleteLFTagWithContext(ctx, input)
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
		validation.StringMatch(regexp.MustCompile(`^([\p{L}\p{Z}\p{N}_.:\*\/=+\-@%]*)$`), ""),
	)
}

func splitLFTagValues(in []interface{}, size int) [][]interface{} {
	var out [][]interface{}

	for {
		if len(in) == 0 {
			break
		}

		if len(in) < size {
			size = len(in)
		}

		out = append(out, in[0:size])
		in = in[size:]
	}

	return out
}
