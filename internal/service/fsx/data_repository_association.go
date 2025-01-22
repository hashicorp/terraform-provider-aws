// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_data_repository_association", name="Data Repository Association")
// @Tags(identifierAttribute="arn")
func resourceDataRepositoryAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataRepositoryAssociationCreate,
		ReadWithoutTimeout:   resourceDataRepositoryAssociationRead,
		UpdateWithoutTimeout: resourceDataRepositoryAssociationUpdate,
		DeleteWithoutTimeout: resourceDataRepositoryAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"batch_import_meta_data_on_create": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"data_repository_path": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 900),
					validation.StringMatch(regexache.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"delete_data_in_filesystem": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(11, 21),
					validation.StringMatch(regexache.MustCompile(`^fs-[0-9a-f]*`), "must begin with fs-"),
				),
			},
			"file_system_path": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 4096),
					validation.StringMatch(regexache.MustCompile(`^/.*`), "path must begin with /"),
				),
			},
			"imported_file_chunk_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 512000),
			},
			"s3": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_export_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"events": {
										Type:     schema.TypeList,
										MaxItems: 3,
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.EventType](),
										},
									},
								},
							},
						},
						"auto_import_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"events": {
										Type:     schema.TypeList,
										MaxItems: 3,
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.EventType](),
										},
									},
								},
							},
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceDataRepositoryAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.CreateDataRepositoryAssociationInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		DataRepositoryPath: aws.String(d.Get("data_repository_path").(string)),
		FileSystemId:       aws.String(d.Get(names.AttrFileSystemID).(string)),
		FileSystemPath:     aws.String(d.Get("file_system_path").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("batch_import_meta_data_on_create"); ok {
		input.BatchImportMetaDataOnCreate = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("imported_file_chunk_size"); ok {
		input.ImportedFileChunkSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("s3"); ok {
		input.S3 = expandDataRepositoryAssociationS3(v.([]interface{}))
	}

	output, err := conn.CreateDataRepositoryAssociation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre Data Repository Association: %s", err)
	}

	d.SetId(aws.ToString(output.Association.AssociationId))

	if _, err := waitDataRepositoryAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre Data Repository Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDataRepositoryAssociationRead(ctx, d, meta)...)
}

func resourceDataRepositoryAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	association, err := findDataRepositoryAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for Lustre Data Repository Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for Lustre Data Repository Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, association.ResourceARN)
	d.Set("batch_import_meta_data_on_create", association.BatchImportMetaDataOnCreate)
	d.Set("data_repository_path", association.DataRepositoryPath)
	d.Set(names.AttrFileSystemID, association.FileSystemId)
	d.Set("file_system_path", association.FileSystemPath)
	d.Set("imported_file_chunk_size", association.ImportedFileChunkSize)
	if err := d.Set("s3", flattenDataRepositoryAssociationS3(association.S3)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3: %s", err)
	}

	setTagsOut(ctx, association.Tags)

	return diags
}

func resourceDataRepositoryAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateDataRepositoryAssociationInput{
			AssociationId:      aws.String(d.Id()),
			ClientRequestToken: aws.String(id.UniqueId()),
		}

		if d.HasChange("imported_file_chunk_size") {
			input.ImportedFileChunkSize = aws.Int32(int32(d.Get("imported_file_chunk_size").(int)))
		}

		if d.HasChange("s3") {
			input.S3 = expandDataRepositoryAssociationS3(d.Get("s3").([]interface{}))
		}

		_, err := conn.UpdateDataRepositoryAssociation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for Lustre Data Repository Association (%s): %s", d.Id(), err)
		}

		if _, err := waitDataRepositoryAssociationUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre Data Repository Association (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataRepositoryAssociationRead(ctx, d, meta)...)
}

func resourceDataRepositoryAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	request := &fsx.DeleteDataRepositoryAssociationInput{
		AssociationId:          aws.String(d.Id()),
		ClientRequestToken:     aws.String(id.UniqueId()),
		DeleteDataInFileSystem: aws.Bool(d.Get("delete_data_in_filesystem").(bool)),
	}

	log.Printf("[DEBUG] Deleting FSx for Lustre Data Repository Association: %s", d.Id())
	_, err := conn.DeleteDataRepositoryAssociation(ctx, request)

	if errs.IsA[*awstypes.DataRepositoryAssociationNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for Lustre Data Repository Association (%s): %s", d.Id(), err)
	}

	if _, err := waitDataRepositoryAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre Data Repository Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDataRepositoryAssociationByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.DataRepositoryAssociation, error) {
	input := &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: []string{id},
	}

	return findDataRepositoryAssociation(ctx, conn, input, tfslices.PredicateTrue[*awstypes.DataRepositoryAssociation]())
}

func findDataRepositoryAssociation(ctx context.Context, conn *fsx.Client, input *fsx.DescribeDataRepositoryAssociationsInput, filter tfslices.Predicate[*awstypes.DataRepositoryAssociation]) (*awstypes.DataRepositoryAssociation, error) {
	output, err := findDataRepositoryAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDataRepositoryAssociations(ctx context.Context, conn *fsx.Client, input *fsx.DescribeDataRepositoryAssociationsInput, filter tfslices.Predicate[*awstypes.DataRepositoryAssociation]) ([]awstypes.DataRepositoryAssociation, error) {
	var output []awstypes.DataRepositoryAssociation

	pages := fsx.NewDescribeDataRepositoryAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DataRepositoryAssociationNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Associations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDataRepositoryAssociation(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDataRepositoryAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitDataRepositoryAssociationCreated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.DataRepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataRepositoryLifecycleCreating),
		Target:  enum.Slice(awstypes.DataRepositoryLifecycleAvailable),
		Refresh: statusDataRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataRepositoryAssociation); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDataRepositoryAssociationUpdated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.DataRepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataRepositoryLifecycleUpdating),
		Target:  enum.Slice(awstypes.DataRepositoryLifecycleAvailable),
		Refresh: statusDataRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataRepositoryAssociation); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDataRepositoryAssociationDeleted(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.DataRepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataRepositoryLifecycleAvailable, awstypes.DataRepositoryLifecycleDeleting),
		Target:  []string{},
		Refresh: statusDataRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataRepositoryAssociation); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func expandDataRepositoryAssociationS3(cfg []interface{}) *awstypes.S3DataRepositoryConfiguration {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}

	m := cfg[0].(map[string]interface{})

	s3Config := &awstypes.S3DataRepositoryConfiguration{}

	if v, ok := m["auto_export_policy"]; ok {
		policy := v.([]interface{})
		s3Config.AutoExportPolicy = expandDataRepositoryAssociationS3AutoExportPolicy(policy)
	}
	if v, ok := m["auto_import_policy"]; ok {
		policy := v.([]interface{})
		s3Config.AutoImportPolicy = expandDataRepositoryAssociationS3AutoImportPolicy(policy)
	}

	return s3Config
}

func expandDataRepositoryAssociationS3AutoExportPolicy(policy []interface{}) *awstypes.AutoExportPolicy {
	if len(policy) == 0 || policy[0] == nil {
		return nil
	}

	m := policy[0].(map[string]interface{})
	autoExportPolicy := &awstypes.AutoExportPolicy{}

	if v, ok := m["events"]; ok {
		autoExportPolicy.Events = flex.ExpandStringyValueList[awstypes.EventType](v.([]interface{}))
	}

	return autoExportPolicy
}

func expandDataRepositoryAssociationS3AutoImportPolicy(policy []interface{}) *awstypes.AutoImportPolicy {
	if len(policy) == 0 || policy[0] == nil {
		return nil
	}

	m := policy[0].(map[string]interface{})
	autoImportPolicy := &awstypes.AutoImportPolicy{}

	if v, ok := m["events"]; ok {
		autoImportPolicy.Events = flex.ExpandStringyValueList[awstypes.EventType](v.([]interface{}))
	}

	return autoImportPolicy
}

func flattenDataRepositoryAssociationS3(s3Config *awstypes.S3DataRepositoryConfiguration) []map[string]interface{} {
	result := make(map[string]interface{})
	if s3Config == nil {
		return []map[string]interface{}{result}
	}

	if s3Config.AutoExportPolicy != nil {
		result["auto_export_policy"] = flattenS3AutoExportPolicy(s3Config.AutoExportPolicy)
	}
	if s3Config.AutoImportPolicy != nil {
		result["auto_import_policy"] = flattenS3AutoImportPolicy(s3Config.AutoImportPolicy)
	}

	return []map[string]interface{}{result}
}

func flattenS3AutoExportPolicy(policy *awstypes.AutoExportPolicy) []map[string][]interface{} {
	result := make(map[string][]interface{})
	if policy == nil {
		return []map[string][]interface{}{result}
	}
	if policy.Events != nil {
		result["events"] = flex.FlattenStringyValueList(policy.Events)
	}

	return []map[string][]interface{}{result}
}

func flattenS3AutoImportPolicy(policy *awstypes.AutoImportPolicy) []map[string][]interface{} {
	result := make(map[string][]interface{})
	if policy == nil {
		return []map[string][]interface{}{result}
	}
	if policy.Events != nil {
		result["events"] = flex.FlattenStringyValueList(policy.Events)
	}

	return []map[string][]interface{}{result}
}
