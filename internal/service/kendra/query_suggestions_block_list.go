// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kendra_query_suggestions_block_list", name="Query Suggestions Block List")
// @Tags(identifierAttribute="arn")
func ResourceQuerySuggestionsBlockList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceQuerySuggestionsBlockListCreate,
		ReadWithoutTimeout:   resourceQuerySuggestionsBlockListRead,
		UpdateWithoutTimeout: resourceQuerySuggestionsBlockListUpdate,
		DeleteWithoutTimeout: resourceQuerySuggestionsBlockListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_suggestions_block_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source_s3_path": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceQuerySuggestionsBlockListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	in := &kendra.CreateQuerySuggestionsBlockListInput{
		ClientToken:  aws.String(id.UniqueId()),
		IndexId:      aws.String(d.Get("index_id").(string)),
		Name:         aws.String(d.Get("name").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		SourceS3Path: expandSourceS3Path(d.Get("source_s3_path").([]interface{})),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateQuerySuggestionsBlockList(ctx, in)
		},
		func(err error) (bool, error) {
			var validationException *types.ValidationException

			if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return diag.Errorf("creating Amazon Kendra QuerySuggestionsBlockList (%s): %s", d.Get("name").(string), err)
	}

	out, ok := outputRaw.(*kendra.CreateQuerySuggestionsBlockListOutput)
	if !ok || out == nil {
		return diag.Errorf("creating Amazon Kendra QuerySuggestionsBlockList (%s): empty output", d.Get("name").(string))
	}

	id := aws.ToString(out.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitQuerySuggestionsBlockListCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Kendra QuerySuggestionsBlockList (%s) create: %s", d.Id(), err)
	}

	return resourceQuerySuggestionsBlockListRead(ctx, d, meta)
}

func resourceQuerySuggestionsBlockListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id, indexId, err := QuerySuggestionsBlockListParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	out, err := FindQuerySuggestionsBlockListByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra QuerySuggestionsBlockList (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "kendra",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/query-suggestions-block-list/%s", indexId, id),
	}.String()

	d.Set("arn", arn)
	d.Set("description", out.Description)
	d.Set("index_id", out.IndexId)
	d.Set("name", out.Name)
	d.Set("query_suggestions_block_list_id", id)
	d.Set("role_arn", out.RoleArn)
	d.Set("status", out.Status)

	if err := d.Set("source_s3_path", flattenSourceS3Path(out.SourceS3Path)); err != nil {
		return diag.Errorf("setting complex argument: %s", err)
	}

	return nil
}

func resourceQuerySuggestionsBlockListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		id, indexId, err := QuerySuggestionsBlockListParseResourceID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		input := &kendra.UpdateQuerySuggestionsBlockListInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("source_s3_path") {
			input.SourceS3Path = expandSourceS3Path(d.Get("source_s3_path").([]interface{}))
		}

		log.Printf("[DEBUG] Updating Kendra QuerySuggestionsBlockList (%s): %#v", d.Id(), input)

		_, err = tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateQuerySuggestionsBlockList(ctx, input)
			},
			func(err error) (bool, error) {
				var validationException *types.ValidationException

				if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return diag.Errorf("updating Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
		}

		if _, err := waitQuerySuggestionsBlockListUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Kendra QuerySuggestionsBlockList (%s) update: %s", d.Id(), err)
		}
	}

	return resourceQuerySuggestionsBlockListRead(ctx, d, meta)
}

func resourceQuerySuggestionsBlockListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	log.Printf("[INFO] Deleting Kendra QuerySuggestionsBlockList %s", d.Id())

	id, indexId, err := QuerySuggestionsBlockListParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteQuerySuggestionsBlockList(ctx, &kendra.DeleteQuerySuggestionsBlockListInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var notFound *types.ResourceNotFoundException

	if errors.As(err, &notFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Kendra QuerySuggestionsBlockList (%s): %s", d.Id(), err)
	}

	if _, err := waitQuerySuggestionsBlockListDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Kendra QuerySuggestionsBlockList (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func statusQuerySuggestionsBlockList(ctx context.Context, conn *kendra.Client, id, indexId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindQuerySuggestionsBlockListByID(ctx, conn, id, indexId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func waitQuerySuggestionsBlockListCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeQuerySuggestionsBlockListOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.QuerySuggestionsBlockListStatusCreating),
		Target:                    enum.Slice(types.QuerySuggestionsBlockListStatusActive),
		Refresh:                   statusQuerySuggestionsBlockList(ctx, conn, id, indexId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeQuerySuggestionsBlockListOutput); ok {
		if out.Status == types.QuerySuggestionsBlockListStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
		return out, err
	}

	return nil, err
}

func waitQuerySuggestionsBlockListUpdated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeQuerySuggestionsBlockListOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.QuerySuggestionsBlockListStatusUpdating),
		Target:                    enum.Slice(types.QuerySuggestionsBlockListStatusActive),
		Refresh:                   statusQuerySuggestionsBlockList(ctx, conn, id, indexId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeQuerySuggestionsBlockListOutput); ok {
		if out.Status == types.QuerySuggestionsBlockListStatusActiveButUpdateFailed || out.Status == types.QuerySuggestionsBlockListStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
		return out, err
	}

	return nil, err
}

func waitQuerySuggestionsBlockListDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeQuerySuggestionsBlockListOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.QuerySuggestionsBlockListStatusDeleting),
		Target:  []string{},
		Refresh: statusQuerySuggestionsBlockList(ctx, conn, id, indexId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeQuerySuggestionsBlockListOutput); ok {
		if out.Status == types.QuerySuggestionsBlockListStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
		return out, err
	}

	return nil, err
}
