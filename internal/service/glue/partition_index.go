// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_partition_index", name="Partition Index")
func resourcePartitionIndex() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePartitionIndexCreate,
		ReadWithoutTimeout:   resourcePartitionIndexRead,
		DeleteWithoutTimeout: resourcePartitionIndexDelete,

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
			names.AttrDatabaseName: {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"partition_index": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keys": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrTableName: {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourcePartitionIndexCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	catalogID, dbName, tableName := cmp.Or(d.Get(names.AttrCatalogID).(string), c.AccountID(ctx)), d.Get(names.AttrDatabaseName).(string), d.Get(names.AttrTableName).(string)
	input := &glue.CreatePartitionIndexInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
	}
	if v, ok := d.GetOk("partition_index"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.PartitionIndex = expandPartitionIndex(v.([]any)[0].(map[string]any))
	}

	_, err := conn.CreatePartitionIndex(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Partition Index: %s", err)
	}

	indexName := aws.ToString(input.PartitionIndex.IndexName)
	d.SetId(partitionIndexCreateResourceID(catalogID, dbName, tableName, indexName))

	if _, err := waitPartitionIndexCreated(ctx, conn, catalogID, dbName, tableName, indexName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Partition Index (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePartitionIndexRead(ctx, d, meta)...)
}

func resourcePartitionIndexRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, tableName, indexName, err := partitionIndexParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	partition, err := findPartitionIndexByFourPartKey(ctx, conn, catalogID, dbName, tableName, indexName)
	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Partition Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Partition Index (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	if err := d.Set("partition_index", []map[string]any{flattenPartitionIndexDescriptor(partition)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting partition_index: %s", err)
	}
	d.Set(names.AttrTableName, tableName)

	return diags
}

func resourcePartitionIndexDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, tableName, indexName, err := partitionIndexParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Glue Partition Index: %s", d.Id())
	input := glue.DeletePartitionIndexInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		IndexName:    aws.String(indexName),
		TableName:    aws.String(tableName),
	}
	_, err = conn.DeletePartitionIndex(ctx, &input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Partition Index (%s): %s", d.Id(), err)
	}

	if _, err := waitPartitionIndexDeleted(ctx, conn, catalogID, dbName, tableName, indexName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Glue Partition Index (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const partitionIndexResourceIDSeparator = ":"

func partitionIndexCreateResourceID(catalogID, dbName, tableName, indexName string) string {
	parts := []string{catalogID, dbName, tableName, indexName}
	id := strings.Join(parts, partitionIndexResourceIDSeparator)

	return id
}

func partitionIndexParseResourceID(id string) (string, string, string, string, error) {
	parts := strings.SplitN(id, partitionIndexResourceIDSeparator, 4)

	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		return "", "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected catalog-id%[2]sdatabase-name%[2]stable-name%[2]sindex-name", id, partitionIndexResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], parts[3], nil
}

func findPartitionIndexByFourPartKey(ctx context.Context, conn *glue.Client, catalogID, dbName, tableName, indexName string) (*awstypes.PartitionIndexDescriptor, error) {
	input := glue.GetPartitionIndexesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
	}

	return findPartitionIndex(ctx, conn, &input, func(v awstypes.PartitionIndexDescriptor) bool {
		return aws.ToString(v.IndexName) == indexName
	})
}

func findPartitionIndex(ctx context.Context, conn *glue.Client, input *glue.GetPartitionIndexesInput, filter tfslices.Predicate[awstypes.PartitionIndexDescriptor]) (*awstypes.PartitionIndexDescriptor, error) {
	output, err := findPartitionIndexes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output, filter))
}

func findPartitionIndexes(ctx context.Context, conn *glue.Client, input *glue.GetPartitionIndexesInput) ([]awstypes.PartitionIndexDescriptor, error) {
	var output []awstypes.PartitionIndexDescriptor

	pages := glue.NewGetPartitionIndexesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PartitionIndexDescriptorList...)
	}

	return output, nil
}

func statusPartitionIndex(conn *glue.Client, catalogID, dbName, tableName, indexName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPartitionIndexByFourPartKey(ctx, conn, catalogID, dbName, tableName, indexName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.IndexStatus), nil
	}
}

func waitPartitionIndexCreated(ctx context.Context, conn *glue.Client, catalogID, dbName, tableName, indexName string, timeout time.Duration) (*awstypes.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PartitionIndexStatusCreating),
		Target:  enum.Slice(awstypes.PartitionIndexStatusActive),
		Refresh: statusPartitionIndex(conn, catalogID, dbName, tableName, indexName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func waitPartitionIndexDeleted(ctx context.Context, conn *glue.Client, catalogID, dbName, tableName, indexName string, timeout time.Duration) (*awstypes.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PartitionIndexStatusDeleting),
		Target:  []string{},
		Refresh: statusPartitionIndex(conn, catalogID, dbName, tableName, indexName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func expandPartitionIndex(tfMap map[string]any) *awstypes.PartitionIndex {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PartitionIndex{}

	if v, ok := tfMap["index_name"].(string); ok && v != "" {
		apiObject.IndexName = aws.String(v)
	}

	if v, ok := tfMap["keys"].([]any); ok && len(v) > 0 {
		apiObject.Keys = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func flattenPartitionIndexDescriptor(apiObject *awstypes.PartitionIndexDescriptor) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]any)

	if v := aws.ToString(apiObject.IndexName); v != "" {
		tfMap["index_name"] = v
	}

	if v := apiObject.IndexStatus; v != "" {
		tfMap["index_status"] = v
	}

	if apiObject.Keys != nil {
		tfMap["keys"] = tfslices.ApplyToAll(apiObject.Keys, func(v awstypes.KeySchemaElement) string {
			return aws.ToString(v.Name)
		})
	}

	return tfMap
}
