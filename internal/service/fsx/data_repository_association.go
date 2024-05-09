// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"association_id": {
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
			"file_system_id": {
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
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(fsx.EventType_Values(), false),
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
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(fsx.EventType_Values(), false),
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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.CreateDataRepositoryAssociationInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		DataRepositoryPath: aws.String(d.Get("data_repository_path").(string)),
		FileSystemId:       aws.String(d.Get("file_system_id").(string)),
		FileSystemPath:     aws.String(d.Get("file_system_path").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("batch_import_meta_data_on_create"); ok {
		input.BatchImportMetaDataOnCreate = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("imported_file_chunk_size"); ok {
		input.ImportedFileChunkSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("s3"); ok {
		input.S3 = expandDataRepositoryAssociationS3(v.([]interface{}))
	}

	output, err := conn.CreateDataRepositoryAssociationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre Data Repository Association: %s", err)
	}

	d.SetId(aws.StringValue(output.Association.AssociationId))

	if _, err := waitDataRepositoryAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre Data Repository Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDataRepositoryAssociationRead(ctx, d, meta)...)
}

func resourceDataRepositoryAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

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
	d.Set("file_system_id", association.FileSystemId)
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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateDataRepositoryAssociationInput{
			AssociationId:      aws.String(d.Id()),
			ClientRequestToken: aws.String(id.UniqueId()),
		}

		if d.HasChange("imported_file_chunk_size") {
			input.ImportedFileChunkSize = aws.Int64(int64(d.Get("imported_file_chunk_size").(int)))
		}

		if d.HasChange("s3") {
			input.S3 = expandDataRepositoryAssociationS3(d.Get("s3").([]interface{}))
		}

		_, err := conn.UpdateDataRepositoryAssociationWithContext(ctx, input)

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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	request := &fsx.DeleteDataRepositoryAssociationInput{
		AssociationId:          aws.String(d.Id()),
		ClientRequestToken:     aws.String(id.UniqueId()),
		DeleteDataInFileSystem: aws.Bool(d.Get("delete_data_in_filesystem").(bool)),
	}

	log.Printf("[DEBUG] Deleting FSx for Lustre Data Repository Association: %s", d.Id())
	_, err := conn.DeleteDataRepositoryAssociationWithContext(ctx, request)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeDataRepositoryAssociationNotFound) {
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

func expandDataRepositoryAssociationS3(cfg []interface{}) *fsx.S3DataRepositoryConfiguration {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}

	m := cfg[0].(map[string]interface{})

	s3Config := &fsx.S3DataRepositoryConfiguration{}

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

func expandDataRepositoryAssociationS3AutoExportPolicy(policy []interface{}) *fsx.AutoExportPolicy {
	if len(policy) == 0 || policy[0] == nil {
		return nil
	}

	m := policy[0].(map[string]interface{})
	autoExportPolicy := &fsx.AutoExportPolicy{}

	if v, ok := m["events"]; ok {
		autoExportPolicy.Events = flex.ExpandStringList(v.([]interface{}))
	}

	return autoExportPolicy
}

func expandDataRepositoryAssociationS3AutoImportPolicy(policy []interface{}) *fsx.AutoImportPolicy {
	if len(policy) == 0 || policy[0] == nil {
		return nil
	}

	m := policy[0].(map[string]interface{})
	autoImportPolicy := &fsx.AutoImportPolicy{}

	if v, ok := m["events"]; ok {
		autoImportPolicy.Events = flex.ExpandStringList(v.([]interface{}))
	}

	return autoImportPolicy
}

func flattenDataRepositoryAssociationS3(s3Config *fsx.S3DataRepositoryConfiguration) []map[string]interface{} {
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

func flattenS3AutoExportPolicy(policy *fsx.AutoExportPolicy) []map[string][]interface{} {
	result := make(map[string][]interface{})
	if policy == nil {
		return []map[string][]interface{}{result}
	}
	if policy.Events != nil {
		result["events"] = flex.FlattenStringList(policy.Events)
	}

	return []map[string][]interface{}{result}
}

func flattenS3AutoImportPolicy(policy *fsx.AutoImportPolicy) []map[string][]interface{} {
	result := make(map[string][]interface{})
	if policy == nil {
		return []map[string][]interface{}{result}
	}
	if policy.Events != nil {
		result["events"] = flex.FlattenStringList(policy.Events)
	}

	return []map[string][]interface{}{result}
}

func findDataRepositoryAssociationByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.DataRepositoryAssociation, error) {
	input := &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: aws.StringSlice([]string{id}),
	}

	return findDataRepositoryAssociation(ctx, conn, input, tfslices.PredicateTrue[*fsx.DataRepositoryAssociation]())
}

func findDataRepositoryAssociation(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeDataRepositoryAssociationsInput, filter tfslices.Predicate[*fsx.DataRepositoryAssociation]) (*fsx.DataRepositoryAssociation, error) {
	output, err := findDataRepositoryAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDataRepositoryAssociations(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeDataRepositoryAssociationsInput, filter tfslices.Predicate[*fsx.DataRepositoryAssociation]) ([]*fsx.DataRepositoryAssociation, error) {
	var output []*fsx.DataRepositoryAssociation

	err := conn.DescribeDataRepositoryAssociationsPagesWithContext(ctx, input, func(page *fsx.DescribeDataRepositoryAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Associations {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeDataRepositoryAssociationNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusDataRepositoryAssociation(ctx context.Context, conn *fsx.FSx, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDataRepositoryAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func waitDataRepositoryAssociationCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.DataRepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.DataRepositoryLifecycleCreating},
		Target:  []string{fsx.DataRepositoryLifecycleAvailable},
		Refresh: statusDataRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.DataRepositoryAssociation); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDataRepositoryAssociationUpdated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.DataRepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.DataRepositoryLifecycleUpdating},
		Target:  []string{fsx.DataRepositoryLifecycleAvailable},
		Refresh: statusDataRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.DataRepositoryAssociation); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDataRepositoryAssociationDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.DataRepositoryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.DataRepositoryLifecycleAvailable, fsx.DataRepositoryLifecycleDeleting},
		Target:  []string{},
		Refresh: statusDataRepositoryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.DataRepositoryAssociation); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}
