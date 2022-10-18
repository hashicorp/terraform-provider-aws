package storagegateway

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNFSFileShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceNFSFileShareCreate,
		Read:   resourceNFSFileShareRead,
		Update: resourceNFSFileShareUpdate,
		Delete: resourceNFSFileShareDelete,

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
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"bucket_region": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"vpc_endpoint_dns_name"},
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(300, 2592000),
						},
					},
				},
			},
			"client_list": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.Any(
						verify.ValidIPv4CIDRNetworkAddress,
						validation.IsIPv4Address,
					),
				},
			},
			"default_storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultStorageClassS3Standard,
				ValidateFunc: validation.StringInSlice(defaultStorageClass_Values(), false),
			},
			"file_share_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"fileshare_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"guess_mime_type_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"nfs_file_share_defaults": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "0777",
							ValidateFunc: validLinuxFileMode,
						},
						"file_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "0666",
							ValidateFunc: validLinuxFileMode,
						},
						"group_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "65534",
							ValidateFunc: valid4ByteASN,
						},
						"owner_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "65534",
							ValidateFunc: valid4ByteASN,
						},
					},
				},
			},
			"notification_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "{}",
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^\{[\w\s:\{\}\[\]"]*}$`), ""),
					validation.StringLenBetween(2, 100),
				),
			},
			"object_acl": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      storagegateway.ObjectACLPrivate,
				ValidateFunc: validation.StringInSlice(storagegateway.ObjectACL_Values(), false),
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"requester_pays": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"squash": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      squashRootSquash,
				ValidateFunc: validation.StringInSlice(squash_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_endpoint_dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNFSFileShareCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	fileShareDefaults, err := expandNFSFileShareDefaults(d.Get("nfs_file_share_defaults").([]interface{}))

	if err != nil {
		return err
	}

	input := &storagegateway.CreateNFSFileShareInput{
		ClientList:           flex.ExpandStringSet(d.Get("client_list").(*schema.Set)),
		ClientToken:          aws.String(resource.UniqueId()),
		DefaultStorageClass:  aws.String(d.Get("default_storage_class").(string)),
		GatewayARN:           aws.String(d.Get("gateway_arn").(string)),
		GuessMIMETypeEnabled: aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
		KMSEncrypted:         aws.Bool(d.Get("kms_encrypted").(bool)),
		LocationARN:          aws.String(d.Get("location_arn").(string)),
		NFSFileShareDefaults: fileShareDefaults,
		ObjectACL:            aws.String(d.Get("object_acl").(string)),
		ReadOnly:             aws.Bool(d.Get("read_only").(bool)),
		RequesterPays:        aws.Bool(d.Get("requester_pays").(bool)),
		Role:                 aws.String(d.Get("role_arn").(string)),
		Squash:               aws.String(d.Get("squash").(string)),
		Tags:                 Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bucket_region"); ok {
		input.BucketRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandNFSFileShareCacheAttributes(v.([]interface{}))
	}

	if v, ok := d.GetOk("file_share_name"); ok {
		input.FileShareName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_policy"); ok {
		input.NotificationPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_endpoint_dns_name"); ok {
		input.VPCEndpointDNSName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Storage Gateway NFS File Share: %s", input)
	output, err := conn.CreateNFSFileShare(input)

	if err != nil {
		return fmt.Errorf("error creating Storage Gateway NFS File Share: %w", err)
	}

	d.SetId(aws.StringValue(output.FileShareARN))

	if _, err = waitNFSFileShareCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway NFS File Share (%s) create: %w", d.Id(), err)
	}

	return resourceNFSFileShareRead(d, meta)
}

func resourceNFSFileShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	fileshare, err := FindNFSFileShareByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway NFS File Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Storage Gateway NFS File Share (%s): %w", d.Id(), err)
	}

	d.Set("arn", fileshare.FileShareARN)
	d.Set("audit_destination_arn", fileshare.AuditDestinationARN)
	d.Set("bucket_region", fileshare.BucketRegion)
	if err := d.Set("cache_attributes", flattenNFSFileShareCacheAttributes(fileshare.CacheAttributes)); err != nil {
		return fmt.Errorf("error setting cache_attributes: %w", err)
	}
	if err := d.Set("client_list", flex.FlattenStringSet(fileshare.ClientList)); err != nil {
		return fmt.Errorf("error setting client_list: %w", err)
	}
	d.Set("default_storage_class", fileshare.DefaultStorageClass)
	d.Set("file_share_name", fileshare.FileShareName)
	d.Set("fileshare_id", fileshare.FileShareId)
	d.Set("gateway_arn", fileshare.GatewayARN)
	d.Set("guess_mime_type_enabled", fileshare.GuessMIMETypeEnabled)
	d.Set("kms_encrypted", fileshare.KMSEncrypted)
	d.Set("kms_key_arn", fileshare.KMSKey)
	d.Set("location_arn", fileshare.LocationARN)
	if err := d.Set("nfs_file_share_defaults", flattenNFSFileShareDefaults(fileshare.NFSFileShareDefaults)); err != nil {
		return fmt.Errorf("error setting nfs_file_share_defaults: %w", err)
	}
	d.Set("notification_policy", fileshare.NotificationPolicy)
	d.Set("object_acl", fileshare.ObjectACL)
	d.Set("path", fileshare.Path)
	d.Set("read_only", fileshare.ReadOnly)
	d.Set("requester_pays", fileshare.RequesterPays)
	d.Set("role_arn", fileshare.Role)
	d.Set("squash", fileshare.Squash)
	d.Set("vpc_endpoint_dns_name", fileshare.VPCEndpointDNSName)

	tags := KeyValueTags(fileshare.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceNFSFileShareUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	if d.HasChangesExcept("tags_all", "tags") {
		fileShareDefaults, err := expandNFSFileShareDefaults(d.Get("nfs_file_share_defaults").([]interface{}))

		if err != nil {
			return err
		}

		input := &storagegateway.UpdateNFSFileShareInput{
			ClientList:           flex.ExpandStringSet(d.Get("client_list").(*schema.Set)),
			DefaultStorageClass:  aws.String(d.Get("default_storage_class").(string)),
			FileShareARN:         aws.String(d.Id()),
			GuessMIMETypeEnabled: aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
			KMSEncrypted:         aws.Bool(d.Get("kms_encrypted").(bool)),
			NFSFileShareDefaults: fileShareDefaults,
			ObjectACL:            aws.String(d.Get("object_acl").(string)),
			ReadOnly:             aws.Bool(d.Get("read_only").(bool)),
			RequesterPays:        aws.Bool(d.Get("requester_pays").(bool)),
			Squash:               aws.String(d.Get("squash").(string)),
		}

		if v, ok := d.GetOk("audit_destination_arn"); ok {
			input.AuditDestinationARN = aws.String(v.(string))
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandNFSFileShareCacheAttributes(v.([]interface{}))
		}

		if v, ok := d.GetOk("file_share_name"); ok {
			input.FileShareName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("kms_key_arn"); ok {
			input.KMSKey = aws.String(v.(string))
		}

		if v, ok := d.GetOk("notification_policy"); ok {
			input.NotificationPolicy = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Storage Gateway NFS File Share: %s", input)
		_, err = conn.UpdateNFSFileShare(input)

		if err != nil {
			return fmt.Errorf("error updating Storage Gateway NFS File Share (%s): %w", d.Id(), err)
		}

		if _, err = waitNFSFileShareUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway NFS File Share (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceNFSFileShareRead(d, meta)
}

func resourceNFSFileShareDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).StorageGatewayConn

	log.Printf("[DEBUG] Deleting Storage Gateway NFS File Share: %s", d.Id())
	_, err := conn.DeleteFileShare(&storagegateway.DeleteFileShareInput{
		FileShareARN: aws.String(d.Id()),
	})

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Storage Gateway NFS File Share (%s): %w", d.Id(), err)
	}

	if _, err = waitNFSFileShareDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway NFS File Share (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandNFSFileShareDefaults(l []interface{}) (*storagegateway.NFSFileShareDefaults, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	groupID, err := strconv.ParseInt(m["group_id"].(string), 10, 64)
	if err != nil {
		return nil, err
	}

	ownerID, err := strconv.ParseInt(m["owner_id"].(string), 10, 64)
	if err != nil {
		return nil, err
	}

	nfsFileShareDefaults := &storagegateway.NFSFileShareDefaults{
		DirectoryMode: aws.String(m["directory_mode"].(string)),
		FileMode:      aws.String(m["file_mode"].(string)),
		GroupId:       aws.Int64(groupID),
		OwnerId:       aws.Int64(ownerID),
	}

	return nfsFileShareDefaults, nil
}

func flattenNFSFileShareDefaults(nfsFileShareDefaults *storagegateway.NFSFileShareDefaults) []interface{} {
	if nfsFileShareDefaults == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"directory_mode": aws.StringValue(nfsFileShareDefaults.DirectoryMode),
		"file_mode":      aws.StringValue(nfsFileShareDefaults.FileMode),
		"group_id":       strconv.Itoa(int(aws.Int64Value(nfsFileShareDefaults.GroupId))),
		"owner_id":       strconv.Itoa(int(aws.Int64Value(nfsFileShareDefaults.OwnerId))),
	}

	return []interface{}{m}
}

func expandNFSFileShareCacheAttributes(l []interface{}) *storagegateway.CacheAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ca := &storagegateway.CacheAttributes{
		CacheStaleTimeoutInSeconds: aws.Int64(int64(m["cache_stale_timeout_in_seconds"].(int))),
	}

	return ca
}

func flattenNFSFileShareCacheAttributes(ca *storagegateway.CacheAttributes) []interface{} {
	if ca == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cache_stale_timeout_in_seconds": aws.Int64Value(ca.CacheStaleTimeoutInSeconds),
	}

	return []interface{}{m}
}
