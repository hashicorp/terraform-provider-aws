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
	conn := meta.(*conns.AWSClient).LakeFormationConn()

	tagKey := d.Get("key").(string)
	tagValues := d.Get("values").(*schema.Set)
	tagValuesLen := tagValues.Len()

	var catalogID string
	if v, ok := d.GetOk("catalog_id"); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}

	end := lfTagsValuesMaxBatchSize
	if end > tagValuesLen {
		end = tagValuesLen
	}

	valuesSubset := schema.NewSet(tagValues.F, tagValues.List()[0:end])
	input := &lakeformation.CreateLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
		TagValues: flex.ExpandStringSet(valuesSubset),
	}

	_, err := conn.CreateLFTagWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lake Formation LF-Tag: %s", err)
	}

	// If there are more than 50 values, create them in batches of 50 using UpdateLFTag API
	for i := lfTagsValuesMaxBatchSize; i < tagValuesLen; i += lfTagsValuesMaxBatchSize {
		end := i + lfTagsValuesMaxBatchSize

		if end > tagValuesLen {
			end = tagValuesLen
		}

		subset := schema.NewSet(tagValues.F, tagValues.List()[i:end])

		input := &lakeformation.UpdateLFTagInput{
			CatalogId:      aws.String(catalogID),
			TagKey:         aws.String(tagKey),
			TagValuesToAdd: flex.ExpandStringSet(subset),
		}

		_, err := conn.UpdateLFTag(input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "error creating Lake Formation LF-Tag (batch: %d to %d): %w", i, end, err)
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, tagKey))

	return append(diags, resourceLFTagRead(ctx, d, meta)...)
}

func resourceLFTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn()

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
	conn := meta.(*conns.AWSClient).LakeFormationConn()

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s): %s", d.Id(), err)
	}

	o, n := d.GetChange("values")
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	toAdd := ns.Difference(os)
	toDelete := os.Difference(ns)
	toAddLen := toAdd.Len()
	toDeleteLen := toDelete.Len()

	for i := 0; i < Max(toAddLen, toDeleteLen); i += lfTagsValuesMaxBatchSize {
		input := &lakeformation.UpdateLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		if i < toAddLen {
			end := i + lfTagsValuesMaxBatchSize
			if end > toAddLen {
				end = toAddLen
			}

			toAddSubset := schema.NewSet(toAdd.F, toAdd.List()[i:end])
			input.TagValuesToAdd = flex.ExpandStringSet(toAddSubset)
		}

		if i < toDeleteLen {
			end := i + lfTagsValuesMaxBatchSize
			if end > toDeleteLen {
				end = toDeleteLen
			}

			toDeleteSubset := schema.NewSet(toAdd.F, toDelete.List()[i:end])
			input.TagValuesToDelete = flex.ExpandStringSet(toDeleteSubset)
		}

		_, err := conn.UpdateLFTag(input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s) (batch %d): %w", d.Id(), i, err)
		}

		_, err = conn.UpdateLFTagWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lake Formation LF-Tag (%s): %s", d.Id(), err)

		}
	}

	return append(diags, resourceLFTagRead(ctx, d, meta)...)
}

func resourceLFTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LakeFormationConn()

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

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
