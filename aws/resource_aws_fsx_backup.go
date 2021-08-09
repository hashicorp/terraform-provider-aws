package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/waiter"
)

func resourceAwsFsxBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFsxBackupCreate,
		Read:   resourceAwsFsxBackupRead,
		Update: resourceAwsFsxBackupUpdate,
		Delete: resourceAwsFsxBackupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			SetTagsDiff,
		),
	}
}

func resourceAwsFsxBackupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemId:       aws.String(d.Get("file_system_id").(string)),
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().FsxTags()
	}

	result, err := conn.CreateBackup(input)
	if err != nil {
		return fmt.Errorf("Error creating FSx Backup: %w", err)
	}

	d.SetId(aws.StringValue(result.Backup.BackupId))

	log.Println("[DEBUG] Waiting for FSx backup to become available")
	if _, err := waiter.BackupAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for FSx backup (%s) to be available: %w", d.Id(), err)
	}

	return resourceAwsFsxBackupRead(d, meta)
}

func resourceAwsFsxBackupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.FsxUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Lustre File System (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsFsxBackupRead(d, meta)
}

func resourceAwsFsxBackupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	backup, err := finder.BackupByID(conn, d.Id())
	if err != nil {
		return err
	}

	if !d.IsNewResource() && backup == nil {
		log.Printf("[WARN] FSx Backup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", backup.ResourceARN)
	d.Set("type", backup.Type)

	// d.Set("dns_name", filesystem.DNSName)
	// d.Set("export_path", lustreConfig.DataRepositoryConfiguration.ExportPath)
	// d.Set("import_path", lustreConfig.DataRepositoryConfiguration.ImportPath)
	// d.Set("auto_import_policy", lustreConfig.DataRepositoryConfiguration.AutoImportPolicy)
	// d.Set("imported_file_chunk_size", lustreConfig.DataRepositoryConfiguration.ImportedFileChunkSize)
	// d.Set("deployment_type", lustreConfig.DeploymentType)
	// if lustreConfig.PerUnitStorageThroughput != nil {
	// 	d.Set("per_unit_storage_throughput", lustreConfig.PerUnitStorageThroughput)
	// }
	// d.Set("mount_name", lustreConfig.MountName)
	// d.Set("storage_type", filesystem.StorageType)
	// if lustreConfig.DriveCacheType != nil {
	// 	d.Set("drive_cache_type", lustreConfig.DriveCacheType)
	// }

	if backup.KmsKeyId != nil {
		d.Set("kms_key_id", backup.KmsKeyId)
	}

	// if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
	// 	return fmt.Errorf("error setting network_interface_ids: %w", err)
	// }

	d.Set("owner_id", backup.OwnerId)

	tags := keyvaluetags.FsxKeyValueTags(backup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsFsxBackupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.DeleteBackupInput{
		BackupId: aws.String(d.Id()),
	}

	_, err := retryOnAwsCode(fsx.ErrCodeBackupInProgress, func() (interface{}, error) {
		return conn.DeleteBackup(request)
	})
	if isAWSErr(err, fsx.ErrCodeBackupNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting FSx Backup: %w", err)
	}

	log.Println("[DEBUG] Waiting for filesystem to delete")
	if _, err := waiter.BackupDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for FSx backup (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}
