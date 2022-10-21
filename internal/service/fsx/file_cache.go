package fsx

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceFileCache() *schema.Resource {
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
			"copy_tags_to_data_repository_associations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"data_repository_associations": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"association_id": {
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
							Type:     schema.TypeList,
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
						"file_system_id": {
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
							Type:     schema.TypeList,
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
												validation.StringMatch(regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`), "invalid pattern"),
											),
										},
									},
									"version": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringInSlice(fsx.NfsVersion_Values(), false),
										),
									},
								},
							},
						},
						"resource_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": {
							Type:     schema.TypeMap,
							Computed: true,
						},
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
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_cache_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_cache_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringInSlice(fsx.FileCacheType_Values(), false),
				),
			},
			"file_cache_type_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexp.MustCompile(`^[0-9](.[0-9]*)*$`), "invalid pattern"),
				),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"lustre_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"deployment_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringInSlice(fsx.FileCacheLustreDeploymentType_Values(), false),
							),
						},
						"log_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination": {
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
							Type:     schema.TypeList,
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
								validation.StringMatch(regexp.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "invalid pattern"),
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
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeList,
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
			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameFileCache = "File Cache"
)

func resourceFileCacheCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateFileCacheInput{
		ClientRequestToken:   aws.String(resource.UniqueId()),
		FileCacheType:        aws.String(d.Get("file_cache_type").(string)),
		FileCacheTypeVersion: aws.String(d.Get("file_cache_type_version").(string)),
		StorageCapacity:      aws.Int64(int64(d.Get("storage_capacity").(int))),
		SubnetIds:            flex.ExpandStringList(d.Get("subnet_ids").([]interface{})),
	}
	if v, ok := d.GetOk("copy_tags_to_data_repository_associations"); ok {
		input.CopyTagsToDataRepositoryAssociations = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("data_repository_associations"); ok && len(v.([]interface{})) > 0 {
		input.DataRepositoryAssociations = expandDataRepositoryAssociations(v.([]interface{}))
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("lustre_configuration"); ok && len(v.([]interface{})) > 0 {
		input.LustreConfiguration = expandCreateFileCacheLustreConfiguration(v.([]interface{}))
	}
	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = flex.ExpandStringList(v.([]interface{}))
	}
	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	result, err := conn.CreateFileCacheWithContext(ctx, input)

	if err != nil {
		return create.DiagError(names.FSx, create.ErrActionCreating, ResNameFileCache, "", err)
	}

	d.SetId(aws.StringValue(result.FileCache.FileCacheId))

	if _, err := waitFileCacheCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionWaitingForCreation, ResNameFileCache, d.Id(), err)
	}

	return resourceFileCacheRead(ctx, d, meta)
}

func resourceFileCacheRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filecache, err := findFileCacheByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx FileCache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.FSx, create.ErrActionReading, ResNameFileCache, d.Id(), err)
	}

	d.Set("dns_name", filecache.DNSName)
	d.Set("file_cache_type", filecache.FileCacheType)
	d.Set("file_cache_type_version", filecache.FileCacheTypeVersion)
	d.Set("kms_key_id", filecache.KmsKeyId)
	d.Set("owner_id", filecache.OwnerId)
	d.Set("resource_arn", filecache.ResourceARN)
	d.Set("storage_capacity", filecache.StorageCapacity)
	d.Set("subnet_ids", aws.StringValueSlice(filecache.SubnetIds))
	d.Set("vpc_id", filecache.VpcId)

	if err := d.Set("data_repository_association_ids", aws.StringValueSlice(filecache.DataRepositoryAssociationIds)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}
	if err := d.Set("lustre_configuration", flattenFileCacheLustreConfiguration(filecache.LustreConfiguration)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}
	if err := d.Set("network_interface_ids", aws.StringValueSlice(filecache.NetworkInterfaceIds)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}
	if err := d.Set("subnet_ids", aws.StringValueSlice(filecache.SubnetIds)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}

	// Lookup and set Data Repository Associations
	data_repository_associations, err := flattenDataRepositoryAssociations(ctx, conn, meta, filecache.DataRepositoryAssociationIds)

	if err := d.Set("data_repository_associations", data_repository_associations); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}

	//Volume tags do not get returned with describe call so need to make a separate list tags call
	tags, tagserr := ListTags(conn, *filecache.ResourceARN)

	if tagserr != nil {
		return create.DiagError(names.FSx, create.ErrActionReading, ResNameFileCache, d.Id(), err)
	} else {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	}
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}
	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}
	return nil
}

func resourceFileCacheUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("resource_arn").(string), o, n); err != nil {
			return create.DiagError(names.FSx, create.ErrActionUpdating, ResNameFileCache, d.Id(), err)
		}
	}

	if d.HasChangesExcept("tags_all") {
		input := &fsx.UpdateFileCacheInput{
			ClientRequestToken:  aws.String(resource.UniqueId()),
			FileCacheId:         aws.String(d.Id()),
			LustreConfiguration: &fsx.UpdateFileCacheLustreConfiguration{},
		}

		if d.HasChanges("lustre_configuration") {
			input.LustreConfiguration = expandUpdateFileCacheLustreConfiguration(d.Get("lustre_configuration").([]interface{}))
		}

		log.Printf("[DEBUG] Updating FSx FileCache (%s): %#v", d.Id(), input)

		result, err := conn.UpdateFileCacheWithContext(ctx, input)
		if err != nil {
			return create.DiagError(names.FSx, create.ErrActionUpdating, ResNameFileCache, d.Id(), err)
		}
		if _, err := waitFileCacheUpdated(ctx, conn, aws.StringValue(result.FileCache.FileCacheId), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.FSx, create.ErrActionWaitingForUpdate, ResNameFileCache, d.Id(), err)
		}

	}
	return resourceFileCacheRead(ctx, d, meta)
}

func resourceFileCacheDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).FSxConn
	log.Printf("[INFO] Deleting FSx FileCache %s", d.Id())

	_, err := conn.DeleteFileCacheWithContext(ctx, &fsx.DeleteFileCacheInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileCacheId:        aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileCacheNotFound) {
		return nil
	}
	if err != nil {
		return create.DiagError(names.FSx, create.ErrActionDeleting, ResNameFileCache, d.Id(), err)
	}
	if _, err := waitFileCacheDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionWaitingForDeletion, ResNameFileCache, d.Id(), err)
	}

	return nil
}

func waitFileCacheCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileCache, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileCacheLifecycleCreating},
		Target:  []string{fsx.FileCacheLifecycleAvailable},
		Refresh: statusFileCache(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileCache); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}
		return output, err
	}
	return nil, err
}

func waitFileCacheUpdated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileCache, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileCacheLifecycleUpdating},
		Target:  []string{fsx.FileCacheLifecycleAvailable},
		Refresh: statusFileCache(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileCache); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitFileCacheDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileCache, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileCacheLifecycleAvailable, fsx.FileCacheLifecycleDeleting},
		Target:  []string{},
		Refresh: statusFileCache(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileCache); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}
		return output, err
	}

	return nil, err
}

func statusFileCache(ctx context.Context, conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findFileCacheByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.Lifecycle), nil
	}
}

func findFileCacheByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileCache, error) {

	input := &fsx.DescribeFileCachesInput{
		FileCacheIds: []*string{aws.String(id)},
	}
	var fileCaches []*fsx.FileCache

	err := conn.DescribeFileCachesPages(input, func(page *fsx.DescribeFileCachesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		fileCaches = append(fileCaches, page.FileCaches...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileCacheNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}
	if len(fileCaches) == 0 || fileCaches[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}
	if count := len(fileCaches); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
	return fileCaches[0], nil
}

func flattenDataRepositoryAssociations(ctx context.Context, conn *fsx.FSx, meta interface{}, dataRepositoryAssociationIds []*string) ([]interface{}, error) {
	in := &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: dataRepositoryAssociationIds,
	}
	result, err := conn.DescribeDataRepositoryAssociationsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileCacheNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}
	if result == nil || result.Associations == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	dataRepositoryAssociationsList := []interface{}{}

	for _, dataRepositoryAssociation := range result.Associations {
		tags := KeyValueTags(dataRepositoryAssociation.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		values := map[string]interface{}{
			"association_id":                 dataRepositoryAssociation.AssociationId,
			"data_repository_path":           dataRepositoryAssociation.DataRepositoryPath,
			"data_repository_subdirectories": aws.StringValueSlice(dataRepositoryAssociation.DataRepositorySubdirectories),
			"file_cache_id":                  dataRepositoryAssociation.FileCacheId,
			"file_cache_path":                dataRepositoryAssociation.FileCachePath,
			"imported_file_chunk_size":       dataRepositoryAssociation.ImportedFileChunkSize,
			"nfs":                            flattenNFSDataRepositoryConfiguration(dataRepositoryAssociation.NFS),
			"resource_arn":                   dataRepositoryAssociation.ResourceARN,
			"tags":                           tags,
		}

		dataRepositoryAssociationsList = append(dataRepositoryAssociationsList, values)
	}
	return dataRepositoryAssociationsList, nil
}

func flattenDataRepositoryAssociationTags(tags []*fsx.Tag) []map[string]interface{} {

	dataRepositoryAssociationTags := make([]map[string]interface{}, 0)

	for _, tag := range tags {
		values := map[string]interface{}{
			"key":   tag.Key,
			"value": tag.Value,
		}
		dataRepositoryAssociationTags = append(dataRepositoryAssociationTags, values)
	}
	return dataRepositoryAssociationTags
}

func flattenNFSDataRepositoryConfiguration(nfsDataRepositoryConfiguration *fsx.NFSDataRepositoryConfiguration) []map[string]interface{} {
	if nfsDataRepositoryConfiguration == nil {
		return []map[string]interface{}{}
	}

	values := map[string]interface{}{
		"dns_ips": aws.StringValueSlice(nfsDataRepositoryConfiguration.DnsIps),
		"version": aws.StringValue(nfsDataRepositoryConfiguration.Version),
	}
	return []map[string]interface{}{values}
}

func flattenFileCacheLustreConfiguration(fileCacheLustreConfiguration *fsx.FileCacheLustreConfiguration) []interface{} {
	if fileCacheLustreConfiguration == nil {
		return []interface{}{}
	}
	values := make(map[string]interface{})

	if fileCacheLustreConfiguration.DeploymentType != nil {
		values["deployment_type"] = aws.StringValue(fileCacheLustreConfiguration.DeploymentType)
	}
	if fileCacheLustreConfiguration.LogConfiguration != nil {
		values["log_configuration"] = flattenLustreLogConfiguration(fileCacheLustreConfiguration.LogConfiguration)
	}
	if fileCacheLustreConfiguration.MetadataConfiguration != nil {
		values["metadata_configuration"] = flattenFileCacheLustreMetadataConfiguration(fileCacheLustreConfiguration.MetadataConfiguration)
	}
	if fileCacheLustreConfiguration.MountName != nil {
		values["mount_name"] = aws.StringValue(fileCacheLustreConfiguration.MountName)
	}
	if fileCacheLustreConfiguration.PerUnitStorageThroughput != nil {
		values["per_unit_storage_throughput"] = aws.Int64Value(fileCacheLustreConfiguration.PerUnitStorageThroughput)
	}
	if fileCacheLustreConfiguration.WeeklyMaintenanceStartTime != nil {
		values["weekly_maintenance_start_time"] = aws.StringValue(fileCacheLustreConfiguration.WeeklyMaintenanceStartTime)
	}

	return []interface{}{values}
}

func flattenFileCacheLustreMetadataConfiguration(fileCacheLustreMetadataConfiguration *fsx.FileCacheLustreMetadataConfiguration) []interface{} {
	values := make(map[string]interface{})
	if fileCacheLustreMetadataConfiguration.StorageCapacity != nil {
		values["storage_capacity"] = aws.Int64Value(fileCacheLustreMetadataConfiguration.StorageCapacity)
	}

	return []interface{}{values}
}

func expandDataRepositoryAssociations(l []interface{}) []*fsx.FileCacheDataRepositoryAssociation {
	dataRepositoryAssociations := []*fsx.FileCacheDataRepositoryAssociation{}

	for _, dataRepositoryAssociation := range l {
		tfMap := dataRepositoryAssociation.(map[string]interface{})
		req := &fsx.FileCacheDataRepositoryAssociation{}

		if v, ok := tfMap["data_repository_path"].(string); ok {
			req.DataRepositoryPath = aws.String(v)
		}
		if v, ok := tfMap["data_repository_subdirectories"]; ok {
			req.DataRepositorySubdirectories = flex.ExpandStringList(v.([]interface{}))
		}
		if v, ok := tfMap["file_cache_path"].(string); ok {
			req.FileCachePath = aws.String(v)
		}
		if v, ok := tfMap["nfs"]; ok && len(v.([]interface{})) > 0 {
			req.NFS = expandFileCacheNFSConfiguration(v.(map[string]interface{}))
		}
		dataRepositoryAssociations = append(dataRepositoryAssociations, req)
	}

	return dataRepositoryAssociations
}

func expandFileCacheNFSConfiguration(l map[string]interface{}) *fsx.FileCacheNFSConfiguration {
	req := &fsx.FileCacheNFSConfiguration{}
	if v, ok := l["dns_ips"]; ok {
		req.DnsIps = flex.ExpandStringList(v.([]interface{}))
	}
	if v, ok := l["version"].(string); ok {
		req.Version = aws.String(v)
	}

	return req
}

func expandUpdateFileCacheLustreConfiguration(l []interface{}) *fsx.UpdateFileCacheLustreConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]interface{})
	req := &fsx.UpdateFileCacheLustreConfiguration{}

	if v, ok := data["weekly_maintenance_start_time"].(string); ok {
		req.WeeklyMaintenanceStartTime = aws.String(v)
	}

	return req
}

func expandCreateFileCacheLustreConfiguration(l []interface{}) *fsx.CreateFileCacheLustreConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	data := l[0].(map[string]interface{})
	req := &fsx.CreateFileCacheLustreConfiguration{}

	if v, ok := data["deployment_type"].(string); ok {
		req.DeploymentType = aws.String(v)
	}
	if v, ok := data["metadata_configuration"]; ok && len(v.([]interface{})) > 0 {
		req.MetadataConfiguration = expandFileCacheLustreMetadataConfiguration(v.([]interface{}))
	}
	if v, ok := data["per_unit_storage_throughput"].(int); ok {
		req.PerUnitStorageThroughput = aws.Int64(int64(v))
	}
	if v, ok := data["weekly_maintenance_start_time"].(string); ok {
		req.WeeklyMaintenanceStartTime = aws.String(v)
	}

	return req
}

func expandFileCacheLustreMetadataConfiguration(l []interface{}) *fsx.FileCacheLustreMetadataConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	data := l[0].(map[string]interface{})
	req := &fsx.FileCacheLustreMetadataConfiguration{}

	if v, ok := data["storage_capacity"].(int); ok {
		req.StorageCapacity = aws.Int64(int64(v))
	}
	return req
}
