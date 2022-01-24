package fsx

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDataRepositoryAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataRepositoryAssociationCreate,
		Read:   resourceDataRepositoryAssociationRead,
		Update: resourceDataRepositoryAssociationUpdate,
		Delete: resourceDataRepositoryAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
			},
			"file_system_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(11, 21),
					validation.StringMatch(regexp.MustCompile(`^fs-[0-9a-f]*`), "must begin with fs-"),
				),
			},
			"file_system_path": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 4096),
					validation.StringMatch(regexp.MustCompile(`^/.*`), "path must begin with /"),
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
			"delete_data_in_filesystem": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

func resourceDataRepositoryAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateDataRepositoryAssociationInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		DataRepositoryPath: aws.String(d.Get("data_repository_path").(string)),
		FileSystemId:       aws.String(d.Get("file_system_id").(string)),
		FileSystemPath:     aws.String(d.Get("file_system_path").(string)),
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

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating FSx Lustre Data Repository Association: %s", input)
	result, err := conn.CreateDataRepositoryAssociation(input)

	if err != nil {
		return fmt.Errorf("error creating FSx Lustre Data Repository Association: %w", err)
	}

	d.SetId(aws.StringValue(result.Association.AssociationId))

	if _, err := waitDataRepositoryAssociationCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx Lustre Data Repository Association (%s) create: %w", d.Id(), err)
	}

	return resourceDataRepositoryAssociationRead(d, meta)
}

func resourceDataRepositoryAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Lustre Data Repository Association (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateDataRepositoryAssociationInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			AssociationId:      aws.String(d.Id()),
		}

		if d.HasChange("imported_file_chunk_size") {
			input.ImportedFileChunkSize = aws.Int64(int64(d.Get("imported_file_chunk_size").(int)))
		}

		if d.HasChange("s3") {
			input.S3 = expandDataRepositoryAssociationS3(d.Get("s3").([]interface{}))
		}

		_, err := conn.UpdateDataRepositoryAssociation(input)
		if err != nil {
			return fmt.Errorf("error updating FSX Lustre Data Repository Association (%s): %w", d.Id(), err)
		}

		if _, err := waitDataRepositoryAssociationUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx Lustre Data Repository Association (%s) update: %w", d.Id(), err)
		}
	}

	return resourceDataRepositoryAssociationRead(d, meta)
}

func resourceDataRepositoryAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	association, err := FindDataRepositoryAssociationByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Lustre Data Repository Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Lustre Data Repository Association (%s): %w", d.Id(), err)
	}

	d.Set("arn", association.ResourceARN)
	d.Set("batch_import_meta_data_on_create", association.BatchImportMetaDataOnCreate)
	d.Set("data_repository_path", association.DataRepositoryPath)
	d.Set("file_system_id", association.FileSystemId)
	d.Set("file_system_path", association.FileSystemPath)
	d.Set("imported_file_chunk_size", association.ImportedFileChunkSize)
	if err := d.Set("s3", flattenDataRepositoryAssociationS3(association.S3)); err != nil {
		return fmt.Errorf("error setting s3 data repository configuration: %s", err)
	}

	tags := KeyValueTags(association.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDataRepositoryAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	request := &fsx.DeleteDataRepositoryAssociationInput{
		ClientRequestToken:     aws.String(resource.UniqueId()),
		AssociationId:          aws.String(d.Id()),
		DeleteDataInFileSystem: aws.Bool(d.Get("delete_data_in_filesystem").(bool)),
	}

	log.Printf("[DEBUG] Deleting FSx Lustre Data Repository Association: %s", d.Id())
	_, err := conn.DeleteDataRepositoryAssociation(request)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeDataRepositoryAssociationNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx Lustre Data Repository Association (%s): %w", d.Id(), err)
	}

	if _, err := waitDataRepositoryAssociationDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx Lustre Data Repository Association (%s) to deleted: %w", d.Id(), err)
	}

	return nil
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
