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

// @SDKResource("aws_fsx_file_cache", name="File Cache")
// @Tags(identifierAttribute="arn")
func resourceFileCache() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFileCacheCreate,
		ReadWithoutTimeout:   resourceFileCacheRead,
		UpdateWithoutTimeout: resourceFileCacheUpdate,
		DeleteWithoutTimeout: resourceFileCacheDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"copy_tags_to_data_repository_associations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"data_repository_association": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAssociationID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"data_repository_path": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(3, 4357),
							),
						},
						"data_repository_subdirectories": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 500,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 4096),
								),
							},
						},
						"file_cache_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"file_cache_path": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 4096),
							),
						},
						names.AttrFileSystemID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"file_system_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"imported_file_chunk_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"nfs": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_ips": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringLenBetween(7, 15),
												validation.StringMatch(regexache.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`), "invalid pattern"),
											),
										},
									},
									names.AttrVersion: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.NfsVersion](),
									},
								},
							},
						},
						names.AttrResourceARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
					},
				},
			},
			"data_repository_association_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_cache_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_cache_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FileCacheType](),
			},
			"file_cache_type_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexache.MustCompile(`^[0-9](.[0-9]*)*$`), "invalid pattern"),
				),
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"lustre_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"deployment_type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FileCacheLustreDeploymentType](),
						},
						"log_configuration": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDestination: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"level": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"metadata_configuration": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MaxItems: 8,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"storage_capacity": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.IntBetween(0, 2147483647),
										),
									},
								},
							},
						},
						"mount_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"per_unit_storage_throughput": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.IntBetween(12, 1000),
							),
						},
						"weekly_maintenance_start_time": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(7, 7),
								validation.StringMatch(regexache.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "invalid pattern"),
							),
						},
					},
				},
			},
			"network_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_capacity": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.IntBetween(0, 2147483647),
				),
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFileCacheCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.CreateFileCacheInput{
		ClientRequestToken:   aws.String(id.UniqueId()),
		FileCacheType:        awstypes.FileCacheType(d.Get("file_cache_type").(string)),
		FileCacheTypeVersion: aws.String(d.Get("file_cache_type_version").(string)),
		StorageCapacity:      aws.Int32(int32(d.Get("storage_capacity").(int))),
		SubnetIds:            flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("copy_tags_to_data_repository_associations"); ok {
		input.CopyTagsToDataRepositoryAssociations = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("data_repository_association"); ok && len(v.(*schema.Set).List()) > 0 {
		input.DataRepositoryAssociations = expandDataRepositoryAssociations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("lustre_configuration"); ok && len(v.(*schema.Set).List()) > 0 {
		input.LustreConfiguration = expandCreateFileCacheLustreConfiguration(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateFileCache(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for Lustre File Cache: %s", err)
	}

	d.SetId(aws.ToString(output.FileCache.FileCacheId))

	if _, err := waitFileCacheCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File Cache (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFileCacheRead(ctx, d, meta)...)
}

func resourceFileCacheRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	filecache, err := findFileCacheByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx FileCache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for Lustre File Cache (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, filecache.ResourceARN)
	dataRepositoryAssociationIDs := filecache.DataRepositoryAssociationIds
	d.Set("data_repository_association_ids", dataRepositoryAssociationIDs)
	d.Set(names.AttrDNSName, filecache.DNSName)
	d.Set("file_cache_id", filecache.FileCacheId)
	d.Set("file_cache_type", filecache.FileCacheType)
	d.Set("file_cache_type_version", filecache.FileCacheTypeVersion)
	d.Set(names.AttrKMSKeyID, filecache.KmsKeyId)
	if err := d.Set("lustre_configuration", flattenFileCacheLustreConfiguration(filecache.LustreConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lustre_configuration: %s", err)
	}
	d.Set("network_interface_ids", filecache.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, filecache.OwnerId)
	d.Set("storage_capacity", filecache.StorageCapacity)
	d.Set(names.AttrSubnetIDs, filecache.SubnetIds)
	d.Set(names.AttrVPCID, filecache.VpcId)

	dataRepositoryAssociations, err := findDataRepositoryAssociationsByIDs(ctx, conn, dataRepositoryAssociationIDs)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for Lustre  Data Repository Associations: %s", err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	if err := d.Set("data_repository_association", flattenDataRepositoryAssociations(ctx, dataRepositoryAssociations, defaultTagsConfig, ignoreTagsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_repository_association: %s", err)
	}

	return diags
}

func resourceFileCacheUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateFileCacheInput{
			ClientRequestToken:  aws.String(id.UniqueId()),
			FileCacheId:         aws.String(d.Id()),
			LustreConfiguration: &awstypes.UpdateFileCacheLustreConfiguration{},
		}

		if d.HasChanges("lustre_configuration") {
			input.LustreConfiguration = expandUpdateFileCacheLustreConfiguration(d.Get("lustre_configuration").([]interface{}))
		}

		_, err := conn.UpdateFileCache(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for Lustre File Cache (%s): %s", d.Id(), err)
		}

		if _, err := waitFileCacheUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File Cache (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFileCacheRead(ctx, d, meta)...)
}

func resourceFileCacheDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	log.Printf("[INFO] Deleting FSx FileCache: %s", d.Id())
	_, err := conn.DeleteFileCache(ctx, &fsx.DeleteFileCacheInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileCacheId:        aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.FileCacheNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for Lustre File Cache (%s): %s", d.Id(), err)
	}

	if _, err := waitFileCacheDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for Lustre File Cache (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFileCacheByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.FileCache, error) {
	input := &fsx.DescribeFileCachesInput{
		FileCacheIds: []string{id},
	}

	return findFileCache(ctx, conn, input, tfslices.PredicateTrue[*awstypes.FileCache]())
}

func findFileCache(ctx context.Context, conn *fsx.Client, input *fsx.DescribeFileCachesInput, filter tfslices.Predicate[*awstypes.FileCache]) (*awstypes.FileCache, error) {
	output, err := findFileCaches(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFileCaches(ctx context.Context, conn *fsx.Client, input *fsx.DescribeFileCachesInput, filter tfslices.Predicate[*awstypes.FileCache]) ([]awstypes.FileCache, error) {
	var output []awstypes.FileCache

	pages := fsx.NewDescribeFileCachesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.FileCacheNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileCaches {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusFileCache(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileCacheByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitFileCacheCreated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.FileCache, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileCacheLifecycleCreating),
		Target:  enum.Slice(awstypes.FileCacheLifecycleAvailable),
		Refresh: statusFileCache(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileCache); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}
	return nil, err
}

func waitFileCacheUpdated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.FileCache, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileCacheLifecycleUpdating),
		Target:  enum.Slice(awstypes.FileCacheLifecycleAvailable),
		Refresh: statusFileCache(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileCache); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitFileCacheDeleted(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.FileCache, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FileCacheLifecycleAvailable, awstypes.FileCacheLifecycleDeleting),
		Target:  []string{},
		Refresh: statusFileCache(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileCache); ok {
		if status, details := output.Lifecycle, output.FailureDetails; status == awstypes.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func findDataRepositoryAssociationsByIDs(ctx context.Context, conn *fsx.Client, ids []string) ([]awstypes.DataRepositoryAssociation, error) {
	input := &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: ids,
	}

	return findDataRepositoryAssociations(ctx, conn, input, tfslices.PredicateTrue[*awstypes.DataRepositoryAssociation]())
}

func flattenDataRepositoryAssociations(ctx context.Context, dataRepositoryAssociations []awstypes.DataRepositoryAssociation, defaultTagsConfig *tftags.DefaultConfig, ignoreTagsConfig *tftags.IgnoreConfig) []interface{} {
	if len(dataRepositoryAssociations) == 0 {
		return nil
	}

	var flattenedDataRepositoryAssociations []interface{}

	for _, dataRepositoryAssociation := range dataRepositoryAssociations {
		tags := KeyValueTags(ctx, dataRepositoryAssociation.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		values := map[string]interface{}{
			names.AttrAssociationID:          dataRepositoryAssociation.AssociationId,
			"data_repository_path":           dataRepositoryAssociation.DataRepositoryPath,
			"data_repository_subdirectories": dataRepositoryAssociation.DataRepositorySubdirectories,
			"file_cache_id":                  dataRepositoryAssociation.FileCacheId,
			"file_cache_path":                dataRepositoryAssociation.FileCachePath,
			"imported_file_chunk_size":       dataRepositoryAssociation.ImportedFileChunkSize,
			"nfs":                            flattenNFSDataRepositoryConfiguration(dataRepositoryAssociation.NFS),
			names.AttrResourceARN:            dataRepositoryAssociation.ResourceARN,
			names.AttrTags:                   tags.RemoveDefaultConfig(defaultTagsConfig).Map(),
		}
		flattenedDataRepositoryAssociations = append(flattenedDataRepositoryAssociations, values)
	}
	return flattenedDataRepositoryAssociations
}

func flattenNFSDataRepositoryConfiguration(nfsDataRepositoryConfiguration *awstypes.NFSDataRepositoryConfiguration) []map[string]interface{} {
	if nfsDataRepositoryConfiguration == nil {
		return []map[string]interface{}{}
	}

	values := map[string]interface{}{
		"dns_ips":         nfsDataRepositoryConfiguration.DnsIps,
		names.AttrVersion: string(nfsDataRepositoryConfiguration.Version),
	}
	return []map[string]interface{}{values}
}

func flattenFileCacheLustreConfiguration(fileCacheLustreConfiguration *awstypes.FileCacheLustreConfiguration) []interface{} {
	if fileCacheLustreConfiguration == nil {
		return []interface{}{}
	}
	values := make(map[string]interface{})

	values["deployment_type"] = string(fileCacheLustreConfiguration.DeploymentType)

	if fileCacheLustreConfiguration.LogConfiguration != nil {
		values["log_configuration"] = flattenLustreLogConfiguration(fileCacheLustreConfiguration.LogConfiguration)
	}
	if fileCacheLustreConfiguration.MetadataConfiguration != nil {
		values["metadata_configuration"] = flattenFileCacheLustreMetadataConfiguration(fileCacheLustreConfiguration.MetadataConfiguration)
	}
	if fileCacheLustreConfiguration.MountName != nil {
		values["mount_name"] = aws.ToString(fileCacheLustreConfiguration.MountName)
	}
	if fileCacheLustreConfiguration.PerUnitStorageThroughput != nil {
		values["per_unit_storage_throughput"] = aws.ToInt32(fileCacheLustreConfiguration.PerUnitStorageThroughput)
	}
	if fileCacheLustreConfiguration.WeeklyMaintenanceStartTime != nil {
		values["weekly_maintenance_start_time"] = aws.ToString(fileCacheLustreConfiguration.WeeklyMaintenanceStartTime)
	}

	return []interface{}{values}
}

func flattenFileCacheLustreMetadataConfiguration(fileCacheLustreMetadataConfiguration *awstypes.FileCacheLustreMetadataConfiguration) []interface{} {
	values := make(map[string]interface{})
	if fileCacheLustreMetadataConfiguration.StorageCapacity != nil {
		values["storage_capacity"] = aws.ToInt32(fileCacheLustreMetadataConfiguration.StorageCapacity)
	}

	return []interface{}{values}
}

func expandDataRepositoryAssociations(l []interface{}) []awstypes.FileCacheDataRepositoryAssociation {
	if len(l) == 0 {
		return nil
	}

	var dataRepositoryAssociations []awstypes.FileCacheDataRepositoryAssociation

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}
		req := awstypes.FileCacheDataRepositoryAssociation{}

		if v, ok := tfMap["data_repository_path"].(string); ok {
			req.DataRepositoryPath = aws.String(v)
		}
		if v, ok := tfMap["data_repository_subdirectories"]; ok {
			req.DataRepositorySubdirectories = flex.ExpandStringValueSet(v.(*schema.Set))
		}
		if v, ok := tfMap["file_cache_path"].(string); ok {
			req.FileCachePath = aws.String(v)
		}
		if v, ok := tfMap["nfs"]; ok && len(v.(*schema.Set).List()) > 0 {
			req.NFS = expandFileCacheNFSConfiguration(v.(*schema.Set).List())
		}
		dataRepositoryAssociations = append(dataRepositoryAssociations, req)
	}

	return dataRepositoryAssociations
}

func expandFileCacheNFSConfiguration(l []interface{}) *awstypes.FileCacheNFSConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	data := l[0].(map[string]interface{})

	req := &awstypes.FileCacheNFSConfiguration{}
	if v, ok := data["dns_ips"]; ok {
		req.DnsIps = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := data[names.AttrVersion].(string); ok {
		req.Version = awstypes.NfsVersion(v)
	}

	return req
}

func expandUpdateFileCacheLustreConfiguration(l []interface{}) *awstypes.UpdateFileCacheLustreConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &awstypes.UpdateFileCacheLustreConfiguration{}

	if v, ok := data["weekly_maintenance_start_time"].(string); ok {
		req.WeeklyMaintenanceStartTime = aws.String(v)
	}

	return req
}

func expandCreateFileCacheLustreConfiguration(l []interface{}) *awstypes.CreateFileCacheLustreConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	data := l[0].(map[string]interface{})
	req := &awstypes.CreateFileCacheLustreConfiguration{}

	if v, ok := data["deployment_type"].(string); ok {
		req.DeploymentType = awstypes.FileCacheLustreDeploymentType(v)
	}
	if v, ok := data["metadata_configuration"]; ok && len(v.(*schema.Set).List()) > 0 {
		req.MetadataConfiguration = expandFileCacheLustreMetadataConfiguration(v.(*schema.Set).List())
	}
	if v, ok := data["per_unit_storage_throughput"].(int); ok {
		req.PerUnitStorageThroughput = aws.Int32(int32(v))
	}
	if v, ok := data["weekly_maintenance_start_time"].(string); ok {
		req.WeeklyMaintenanceStartTime = aws.String(v)
	}

	return req
}

func expandFileCacheLustreMetadataConfiguration(l []interface{}) *awstypes.FileCacheLustreMetadataConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	data := l[0].(map[string]interface{})
	req := &awstypes.FileCacheLustreMetadataConfiguration{}

	if v, ok := data["storage_capacity"].(int); ok {
		req.StorageCapacity = aws.Int32(int32(v))
	}
	return req
}
