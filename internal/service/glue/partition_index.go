// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
			names.AttrTableName: {
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
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourcePartitionIndexCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)
	catalogID := createCatalogID(d, meta.(*conns.AWSClient).AccountID(ctx))
	dbName := d.Get(names.AttrDatabaseName).(string)
	tableName := d.Get(names.AttrTableName).(string)

	input := &glue.CreatePartitionIndexInput{
		CatalogId:      aws.String(catalogID),
		DatabaseName:   aws.String(dbName),
		TableName:      aws.String(tableName),
		PartitionIndex: expandPartitionIndex(d.Get("partition_index").([]any)),
	}

	log.Printf("[DEBUG] Creating Glue Partition Index: %#v", input)
	_, err := conn.CreatePartitionIndex(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Partition Index: %s", err)
	}

	d.SetId(createPartitionIndexID(catalogID, dbName, tableName, aws.ToString(input.PartitionIndex.IndexName)))

	if _, err := waitPartitionIndexCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "while waiting for Glue Partition Index (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourcePartitionIndexRead(ctx, d, meta)...)
}

func resourcePartitionIndexRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, tableName, _, err := readPartitionIndexID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Partition Index (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Reading Glue Partition Index: %s", d.Id())
	partition, err := findPartitionIndexByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Partition Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Partition Index (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrTableName, tableName)
	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)

	if err := d.Set("partition_index", []map[string]any{flattenPartitionIndex(*partition)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting partition_index: %s", err)
	}

	return diags
}

func resourcePartitionIndexDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, tableName, partIndex, err := readPartitionIndexID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Partition Index: %s", err)
	}

	log.Printf("[DEBUG] Deleting Glue Partition Index: %s", d.Id())
	_, err = conn.DeletePartitionIndex(ctx, &glue.DeletePartitionIndexInput{
		CatalogId:    aws.String(catalogID),
		TableName:    aws.String(tableName),
		DatabaseName: aws.String(dbName),
		IndexName:    aws.String(partIndex),
	})
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Glue Partition Index: %s", err)
	}

	if _, err := waitPartitionIndexDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "while waiting for Glue Partition Index (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func findPartitionIndexByName(ctx context.Context, conn *glue.Client, id string) (*awstypes.PartitionIndexDescriptor, error) {
	catalogID, dbName, tableName, partIndex, err := readPartitionIndexID(id)
	if err != nil {
		return nil, err
	}

	input := &glue.GetPartitionIndexesInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(tableName),
	}

	var result *awstypes.PartitionIndexDescriptor

	output, err := conn.GetPartitionIndexes(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, partInd := range output.PartitionIndexDescriptorList {
		if aws.ToString(partInd.IndexName) == partIndex {
			result = &partInd
			break
		}
	}

	if result == nil {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	return result, nil
}

func statusPartitionIndex(ctx context.Context, conn *glue.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPartitionIndexByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.IndexStatus), nil
	}
}

func waitPartitionIndexCreated(ctx context.Context, conn *glue.Client, id string, timeout time.Duration) (*awstypes.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PartitionIndexStatusCreating),
		Target:  enum.Slice(awstypes.PartitionIndexStatusActive),
		Refresh: statusPartitionIndex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func waitPartitionIndexDeleted(ctx context.Context, conn *glue.Client, id string, timeout time.Duration) (*awstypes.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PartitionIndexStatusDeleting),
		Target:  []string{},
		Refresh: statusPartitionIndex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func expandPartitionIndex(l []any) *awstypes.PartitionIndex {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := l[0].(map[string]any)
	parIndex := &awstypes.PartitionIndex{}

	if v, ok := s["keys"].([]any); ok && len(v) > 0 {
		parIndex.Keys = flex.ExpandStringValueList(v)
	}

	if v, ok := s["index_name"].(string); ok && v != "" {
		parIndex.IndexName = aws.String(v)
	}

	return parIndex
}
