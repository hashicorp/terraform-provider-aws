package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfsgw "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/waiter"
)

func resourceAwsStorageGatewayFsxAssociateFileSystem() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsStorageGatewayFsxAssociateFileSystemCreate,
		Read:          resourceAwsStorageGatewayFsxAssociateFileSystemRead,
		Update:        resourceAwsStorageGatewayFsxAssociateFileSystemUpdate,
		Delete:        resourceAwsStorageGatewayFsxAssociateFileSystemDelete,
		CustomizeDiff: customdiff.Sequence(SetTagsDiff),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
				Default:      "",
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{0}),
								validation.IntBetween(300, 2592000),
							),
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^\w[\w\.\- ]*$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
		},
	}
}

func resourceAwsStorageGatewayFsxAssociateFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &storagegateway.AssociateFileSystemInput{
		ClientToken: aws.String(resource.UniqueId()),
		GatewayARN:  aws.String(d.Get("gateway_arn").(string)),
		LocationARN: aws.String(d.Get("location_arn").(string)),
		Password:    aws.String(d.Get("password").(string)),
		Tags:        tags.IgnoreAws().StoragegatewayTags(),
		UserName:    aws.String(d.Get("username").(string)),
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandStorageGatewayFsxAssociateFileSystemCacheAttributes(v.([]interface{}))
	}

	log.Printf("[DEBUG] Associating File System to Storage Gateway: %s", input)
	output, err := conn.AssociateFileSystem(input)
	if err != nil {
		if tfsgw.InvalidGatewayRequestErrCodeEquals(err, tfsgw.FileSystemAssociationNotFound) {
			log.Printf("[WARN] FSX File System %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error associating file system to storage gateway (%s): %w", d.Get("gateway_arn").(string), err)
	}

	d.SetId(aws.StringValue(output.FileSystemAssociationARN))
	log.Printf("[INFO] Storage Gateway FSx File System Association ID: %s", d.Id())

	if _, err = waiter.FsxFileSystemAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway FSx File System Association (%q) to be Available: %w", d.Id(), err)
	}

	return resourceAwsStorageGatewayFsxAssociateFileSystemRead(d, meta)
}

func resourceAwsStorageGatewayFsxAssociateFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filesystem, err := finder.FileSystemAssociationByARN(conn, d.Id())

	if err != nil {
		return err
	}

	if !d.IsNewResource() && filesystem == nil {
		log.Printf("[WARN] Storage Gateway FSx File System %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := filesystem.FileSystemAssociationARN
	d.Set("arn", arn)

	d.Set("audit_destination_arn", filesystem.AuditDestinationARN)
	d.Set("gateway_arn", filesystem.GatewayARN)
	d.Set("location_arn", filesystem.LocationARN)

	if err := d.Set("cache_attributes", flattenStorageGatewayFsxAssociateFileSystemCacheAttributes(filesystem.CacheAttributes)); err != nil {
		return fmt.Errorf("error setting cache_attributes: %w", err)
	}

	tags := keyvaluetags.StoragegatewayKeyValueTags(filesystem.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsStorageGatewayFsxAssociateFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	if d.HasChangesExcept("tags_all") {

		input := &storagegateway.UpdateFileSystemAssociationInput{
			AuditDestinationARN:      aws.String(d.Get("audit_destination_arn").(string)),
			Password:                 aws.String(d.Get("password").(string)),
			UserName:                 aws.String(d.Get("username").(string)),
			FileSystemAssociationARN: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandStorageGatewayFsxAssociateFileSystemCacheAttributes(v.([]interface{}))
		}

		log.Printf("[DEBUG] Updating Storage Gateway FSx File System Association: %s", input)
		_, err := conn.UpdateFileSystemAssociation(input)
		if err != nil {
			return fmt.Errorf("error updating Storage Gateway FSx File System Association: %w", err)
		}

		if _, err = waiter.FsxFileSystemAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway FSx File System Association (%q) to be Available: %w", d.Id(), err)
		}
	}

	return resourceAwsStorageGatewayFsxAssociateFileSystemRead(d, meta)
}

func resourceAwsStorageGatewayFsxAssociateFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DisassociateFileSystemInput{
		FileSystemAssociationARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway File System Association: %s", input)
	_, err := conn.DisassociateFileSystem(input)
	if err != nil {
		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file system association") {
			log.Printf("[WARN] Storage Gateway FSx File System Association %q not found, removing from state", d.Id())
			return nil
		}
	}
	if _, err = waiter.FsxFileSystemDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if isResourceNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("error waiting for Storage Gateway FSx File System Association (%q) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandStorageGatewayFsxAssociateFileSystemCacheAttributes(l []interface{}) *storagegateway.CacheAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ca := &storagegateway.CacheAttributes{
		CacheStaleTimeoutInSeconds: aws.Int64(int64(m["cache_stale_timeout_in_seconds"].(int))),
	}

	return ca
}

func flattenStorageGatewayFsxAssociateFileSystemCacheAttributes(ca *storagegateway.CacheAttributes) []interface{} {
	if ca == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cache_stale_timeout_in_seconds": aws.Int64Value(ca.CacheStaleTimeoutInSeconds),
	}

	return []interface{}{m}
}
