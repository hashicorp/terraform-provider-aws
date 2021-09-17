package aws

import (
	"context"
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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/fsx/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsFsxOntapFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFsxOntapFileSystemCreate,
		Read:   resourceAwsFsxOntapFileSystemRead,
		Update: resourceAwsFsxOntapFileSystemUpdate,
		Delete: resourceAwsFsxOntapFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 90),
			},
			"backup_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(5, 5),
					validation.StringMatch(regexp.MustCompile(`^([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format HH:MM"),
				),
			},
			"deployment_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.OntapDeploymentType_Values(), false),
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_ip_address_range": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateIpv4CIDRNetworkAddress,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"network_interface_ids": {
				// As explained in https://docs.aws.amazon.com/fsx/latest/OntapGuide/mounting-on-premises.html, the first
				// network_interface_id is the primary one, so ordering matters. Use TypeList instead of TypeSet to preserve it.
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1024, 192*1024),
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 2,
				MaxItems: 2,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"throughput_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntInSlice([]int{512, 1024, 2048}),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(7, 7),
					validation.StringMatch(regexp.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format d:HH:MM"),
				),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			SetTagsDiff,
			resourceFsxOntapFileSystemSchemaCustomizeDiff,
		),
	}
}

func resourceFsxOntapFileSystemSchemaCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// we want to force a new resource if the new storage capacity is less than the old one
	if d.HasChange("storage_capacity") {
		o, n := d.GetChange("storage_capacity")
		if n.(int) < o.(int) {
			if err := d.ForceNew("storage_capacity"); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceAwsFsxOntapFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeOntap),
		StorageCapacity:    aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          expandStringList(d.Get("subnet_ids").([]interface{})),
		OntapConfiguration: &fsx.CreateFileSystemOntapConfiguration{
			DeploymentType:               aws.String(d.Get("deployment_type").(string)),
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
			ThroughputCapacity:           aws.Int64(int64(d.Get("throughput_capacity").(int))),
			PreferredSubnetId:            aws.String(d.Get("preferred_subnet_id").(string)),
		},
	}

	backupInput := &fsx.CreateFileSystemFromBackupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		StorageType:        aws.String(d.Get("storage_type").(string)),
		SubnetIds:          expandStringList(d.Get("subnet_ids").([]interface{})),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
		backupInput.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("endpoint_ip_address_range"); ok {
		input.OntapConfiguration.EndpointIpAddressRange = aws.String(v.(string))
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		input.OntapConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = expandStringSet(v.(*schema.Set))
		backupInput.SecurityGroupIds = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("route_table_ids"); ok {
		input.OntapConfiguration.RouteTableIds = expandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().FsxTags()
		backupInput.Tags = tags.IgnoreAws().FsxTags()
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		input.OntapConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("backup_id"); ok {
		backupInput.BackupId = aws.String(v.(string))

		log.Printf("[DEBUG] Creating FSx Ontap File System: %s", backupInput)
		result, err := conn.CreateFileSystemFromBackup(backupInput)

		if err != nil {
			return fmt.Errorf("error creating FSx Ontap File System from backup: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	} else {
		log.Printf("[DEBUG] Creating FSx Ontap File System: %s", input)
		result, err := conn.CreateFileSystem(input)

		if err != nil {
			return fmt.Errorf("error creating FSx Ontap File System: %w", err)
		}

		d.SetId(aws.StringValue(result.FileSystem.FileSystemId))
	}

	if _, err := waiter.FileSystemAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx Ontap File System (%s) to be available: %w", d.Id(), err)
	}

	return resourceAwsFsxOntapFileSystemRead(d, meta)
}

func resourceAwsFsxOntapFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.FsxUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx Ontap File System (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			FileSystemId:       aws.String(d.Id()),
			OntapConfiguration: &fsx.UpdateFileSystemOntapConfiguration{},
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.OntapConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.OntapConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.OntapConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int64(int64(d.Get("storage_capacity").(int)))
		}

		_, err := conn.UpdateFileSystem(input)
		if err != nil {
			return fmt.Errorf("error updating FSX Ontap File System (%s): %w", d.Id(), err)
		}

		if _, err := waiter.FileSystemAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for FSx Ontap File System (%s) to be available: %w", d.Id(), err)
		}
	}

	return resourceAwsFsxOntapFileSystemRead(d, meta)
}

func resourceAwsFsxOntapFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filesystem, err := finder.FileSystemByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx Ontap File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx Ontap File System (%s): %w", d.Id(), err)
	}

	ontapConfig := filesystem.OntapConfiguration
	if ontapConfig == nil {
		return fmt.Errorf("error describing FSx Ontap File System (%s): empty Ontap configuration", d.Id())
	}

	d.Set("arn", filesystem.ResourceARN)
	d.Set("dns_name", filesystem.DNSName)
	d.Set("deployment_type", ontapConfig.DeploymentType)
	d.Set("storage_type", filesystem.StorageType)
	d.Set("vpc_id", filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", ontapConfig.WeeklyMaintenanceStartTime)
	d.Set("automatic_backup_retention_days", ontapConfig.AutomaticBackupRetentionDays)
	d.Set("daily_automatic_backup_start_time", ontapConfig.DailyAutomaticBackupStartTime)
	d.Set("throughput_capacity", ontapConfig.ThroughputCapacity)
	d.Set("preferred_subnet_id", ontapConfig.PreferredSubnetId)
	d.Set("endpoint_ip_address_range", ontapConfig.EndpointIpAddressRange)
	d.Set("owner_id", filesystem.OwnerId)
	d.Set("storage_capacity", filesystem.StorageCapacity)

	if filesystem.KmsKeyId != nil {
		d.Set("kms_key_id", filesystem.KmsKeyId)
	}

	if err := d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds)); err != nil {
		return fmt.Errorf("error setting network_interface_ids: %w", err)
	}

	if err := d.Set("subnet_ids", aws.StringValueSlice(filesystem.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	if err := d.Set("route_table_ids", aws.StringValueSlice(ontapConfig.RouteTableIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	tags := keyvaluetags.FsxKeyValueTags(filesystem.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsFsxOntapFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fsxconn

	request := &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting FSx Ontap File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(request)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx Ontap File System (%s): %w", d.Id(), err)
	}

	if _, err := waiter.FileSystemDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx Ontap File System (%s) to deleted: %w", d.Id(), err)
	}

	return nil
}
