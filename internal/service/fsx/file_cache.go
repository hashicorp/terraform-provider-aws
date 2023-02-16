package fsx

import (
	"context"
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
			"arn": {
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
						"tags": tftags.TagsSchemaComputed(),
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
				Type:     schema.TypeSet,
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
							Type:     schema.TypeSet,
							Computed: true,
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
			"security_group_ids": {
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
	conn := meta.(*conns.AWSClient).FSxConn()
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
	if v, ok := d.GetOk("data_repository_association"); ok && len(v.(*schema.Set).List()) > 0 {
		input.DataRepositoryAssociations = expandDataRepositoryAssociations(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("lustre_configuration"); ok && len(v.(*schema.Set).List()) > 0 {
		input.LustreConfiguration = expandCreateFileCacheLustreConfiguration(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
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
	conn := meta.(*conns.AWSClient).FSxConn()
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

	d.Set("arn", filecache.ResourceARN)
	d.Set("dns_name", filecache.DNSName)
	d.Set("file_cache_id", filecache.FileCacheId)
	d.Set("file_cache_type", filecache.FileCacheType)
	d.Set("file_cache_type_version", filecache.FileCacheTypeVersion)
	d.Set("kms_key_id", filecache.KmsKeyId)
	d.Set("owner_id", filecache.OwnerId)
	d.Set("storage_capacity", filecache.StorageCapacity)
	d.Set("subnet_ids", aws.StringValueSlice(filecache.SubnetIds))
	d.Set("vpc_id", filecache.VpcId)

	if err := d.Set("data_repository_association_ids", filecache.DataRepositoryAssociationIds); err != nil {
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

	dataRepositoryAssociations, err := findDataRepositoryAssociationsByIDs(ctx, conn, filecache.DataRepositoryAssociationIds)

	if err := d.Set("data_repository_association", flattenDataRepositoryAssociations(dataRepositoryAssociations, defaultTagsConfig, ignoreTagsConfig)); err != nil {
		return create.DiagError(names.FSx, create.ErrActionSetting, ResNameFileCache, d.Id(), err)
	}

	//Cache tags do not get returned with describe call so need to make a separate list tags call
	tags, tagserr := ListTags(ctx, conn, *filecache.ResourceARN)

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
	conn := meta.(*conns.AWSClient).FSxConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
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
	conn := meta.(*conns.AWSClient).FSxConn()
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

func flattenDataRepositoryAssociations(dataRepositoryAssociations []*fsx.DataRepositoryAssociation, defaultTagsConfig *tftags.DefaultConfig, ignoreTagsConfig *tftags.IgnoreConfig) []interface{} {
	if len(dataRepositoryAssociations) == 0 {
		return nil
	}

	var flattenedDataRepositoryAssociations []interface{}

	for _, dataRepositoryAssociation := range dataRepositoryAssociations {
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
			"tags":                           tags.RemoveDefaultConfig(defaultTagsConfig).Map(),
		}
		flattenedDataRepositoryAssociations = append(flattenedDataRepositoryAssociations, values)
	}
	return flattenedDataRepositoryAssociations
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
	if len(l) == 0 {
		return nil
	}

	var dataRepositoryAssociations []*fsx.FileCacheDataRepositoryAssociation

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}
		req := &fsx.FileCacheDataRepositoryAssociation{}

		if v, ok := tfMap["data_repository_path"].(string); ok {
			req.DataRepositoryPath = aws.String(v)
		}
		if v, ok := tfMap["data_repository_subdirectories"]; ok {
			req.DataRepositorySubdirectories = flex.ExpandStringSet(v.(*schema.Set))
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

func expandFileCacheNFSConfiguration(l []interface{}) *fsx.FileCacheNFSConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	data := l[0].(map[string]interface{})

	req := &fsx.FileCacheNFSConfiguration{}
	if v, ok := data["dns_ips"]; ok {
		req.DnsIps = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := data["version"].(string); ok {
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
	if v, ok := data["metadata_configuration"]; ok && len(v.(*schema.Set).List()) > 0 {
		req.MetadataConfiguration = expandFileCacheLustreMetadataConfiguration(v.(*schema.Set).List())
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
