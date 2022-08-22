package lakeformation

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// This value is defined by AWS API
const lfTagsValuesMaxBatchSize = 50

func ResourceLFTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceLFTagCreate,
		Read:   resourceLFTagRead,
		Update: resourceLFTagUpdate,
		Delete: resourceLFTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				MaxItems: 500,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateLFTagValues(),
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceLFTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

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

	_, err := conn.CreateLFTag(input)
	if err != nil {
		return fmt.Errorf("error creating Lake Formation LF-Tag: %w", err)
	}

	// If there are more than 50 values, create them in batches of 50 using UpdateLFTag API
	for i := 50; i < tagValuesLen; i += lfTagsValuesMaxBatchSize {
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
			return fmt.Errorf("error creating Lake Formation LF-Tag (batch: %d to %d): %w", i, end, err)
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, tagKey))

	return resourceLFTagRead(d, meta)
}

func resourceLFTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return err
	}

	input := &lakeformation.GetLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	output, err := conn.GetLFTag(input)
	if !d.IsNewResource() {
		if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Lake Formation LF-Tag (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("error reading Lake Formation LF-Tag: %s", err.Error())
	}

	d.Set("key", output.TagKey)
	d.Set("values", flex.FlattenStringSet(output.TagValues))
	d.Set("catalog_id", output.CatalogId)

	return nil
}

func resourceLFTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return err
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
			return fmt.Errorf("error updating Lake Formation LF-Tag (%s) (batch %d): %w", d.Id(), i, err)
		}
	}

	return resourceLFTagRead(d, meta)
}

func resourceLFTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	catalogID, tagKey, err := ReadLFTagID(d.Id())
	if err != nil {
		return err
	}

	input := &lakeformation.DeleteLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	_, err = conn.DeleteLFTag(input)
	if err != nil {
		return fmt.Errorf("error deleting Lake Formation LF-Tag (%s): %w", d.Id(), err)
	}

	return nil
}

func ReadLFTagID(id string) (catalogID string, tagKey string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected CATALOG-ID:TAG-KEY", id)
	}
	return idParts[0], idParts[1], nil
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
