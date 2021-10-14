package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	tftransfer "github.com/hashicorp/terraform-provider-aws/aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/transfer/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func resourceAwsTransferUser() *schema.Resource {

	return &schema.Resource{
		Create: resourceAwsTransferUserCreate,
		Read:   resourceAwsTransferUserRead,
		Update: resourceAwsTransferUserUpdate,
		Delete: resourceAwsTransferUserDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"home_directory": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},

			"home_directory_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entry": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
					},
				},
			},

			"home_directory_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      transfer.HomeDirectoryTypePath,
				ValidateFunc: validation.StringInSlice([]string{transfer.HomeDirectoryTypePath, transfer.HomeDirectoryTypeLogical}, false),
			},

			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"posix_profile": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
						},
					},
				},
			},

			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferServerID,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),

			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateTransferUserName,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsTransferUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	serverID := d.Get("server_id").(string)
	userName := d.Get("user_name").(string)
	id := tftransfer.UserCreateResourceID(serverID, userName)
	input := &transfer.CreateUserInput{
		Role:     aws.String(d.Get("role").(string)),
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	if v, ok := d.GetOk("home_directory"); ok {
		input.HomeDirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("home_directory_mappings"); ok {
		input.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(v.([]interface{}))
	}

	if v, ok := d.GetOk("home_directory_type"); ok {
		input.HomeDirectoryType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("posix_profile"); ok {
		input.PosixProfile = expandTransferUserPosixUser(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().TransferTags()
	}

	log.Printf("[DEBUG] Creating Transfer User: %s", input)
	_, err := conn.CreateUser(input)

	if err != nil {
		return fmt.Errorf("error creating Transfer User (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceAwsTransferUserRead(d, meta)
}

func resourceAwsTransferUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	serverID, userName, err := tftransfer.UserParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %w", err)
	}

	user, err := finder.UserByServerIDAndUserName(conn, serverID, userName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Transfer User (%s): %w", d.Id(), err)
	}

	d.Set("arn", user.Arn)
	d.Set("home_directory", user.HomeDirectory)
	if err := d.Set("home_directory_mappings", flattenAwsTransferHomeDirectoryMappings(user.HomeDirectoryMappings)); err != nil {
		return fmt.Errorf("error setting home_directory_mappings: %w", err)
	}
	d.Set("home_directory_type", user.HomeDirectoryType)
	d.Set("policy", user.Policy)
	if err := d.Set("posix_profile", flattenTransferUserPosixUser(user.PosixProfile)); err != nil {
		return fmt.Errorf("error setting posix_profile: %w", err)
	}
	d.Set("role", user.Role)
	d.Set("server_id", serverID)
	d.Set("user_name", user.UserName)

	tags := keyvaluetags.TransferKeyValueTags(user.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}
	return nil
}

func resourceAwsTransferUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn

	if d.HasChangesExcept("tags", "tags_all") {
		serverID, userName, err := tftransfer.UserParseResourceID(d.Id())

		if err != nil {
			return fmt.Errorf("error parsing Transfer User ID: %w", err)
		}

		input := &transfer.UpdateUserInput{
			ServerId: aws.String(serverID),
			UserName: aws.String(userName),
		}

		if d.HasChange("home_directory") {
			input.HomeDirectory = aws.String(d.Get("home_directory").(string))
		}

		if d.HasChange("home_directory_mappings") {
			input.HomeDirectoryMappings = expandAwsTransferHomeDirectoryMappings(d.Get("home_directory_mappings").([]interface{}))
		}

		if d.HasChange("home_directory_type") {
			input.HomeDirectoryType = aws.String(d.Get("home_directory_type").(string))
		}

		if d.HasChange("policy") {
			input.Policy = aws.String(d.Get("policy").(string))
		}

		if d.HasChange("posix_profile") {
			input.PosixProfile = expandTransferUserPosixUser(d.Get("posix_profile").([]interface{}))
		}

		if d.HasChange("role") {
			input.Role = aws.String(d.Get("role").(string))
		}

		log.Printf("[DEBUG] Updating Transfer User: %s", input)
		_, err = conn.UpdateUser(input)

		if err != nil {
			return fmt.Errorf("error updating Transfer User (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.TransferUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsTransferUserRead(d, meta)
}

func resourceAwsTransferUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).TransferConn

	serverID, userName, err := tftransfer.UserParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Transfer User ID: %w", err)
	}

	return transferUserDelete(conn, serverID, userName)
}

// transferUserDelete attempts to delete a transfer user.
func transferUserDelete(conn *transfer.Transfer, serverID, userName string) error {
	id := tftransfer.UserCreateResourceID(serverID, userName)
	input := &transfer.DeleteUserInput{
		ServerId: aws.String(serverID),
		UserName: aws.String(userName),
	}

	log.Printf("[INFO] Deleting Transfer User: %s", id)
	_, err := conn.DeleteUser(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Transfer User (%s): %w", id, err)
	}

	_, err = waiter.UserDeleted(conn, serverID, userName)

	if err != nil {
		return fmt.Errorf("error waiting for Transfer User (%s) delete: %w", id, err)
	}

	return nil
}

func expandAwsTransferHomeDirectoryMappings(in []interface{}) []*transfer.HomeDirectoryMapEntry {
	mappings := make([]*transfer.HomeDirectoryMapEntry, 0)

	for _, tConfig := range in {
		config := tConfig.(map[string]interface{})

		m := &transfer.HomeDirectoryMapEntry{
			Entry:  aws.String(config["entry"].(string)),
			Target: aws.String(config["target"].(string)),
		}

		mappings = append(mappings, m)
	}

	return mappings
}

func flattenAwsTransferHomeDirectoryMappings(mappings []*transfer.HomeDirectoryMapEntry) []interface{} {
	l := make([]interface{}, len(mappings))
	for i, m := range mappings {
		l[i] = map[string]interface{}{
			"entry":  aws.StringValue(m.Entry),
			"target": aws.StringValue(m.Target),
		}
	}
	return l
}

func expandTransferUserPosixUser(pUser []interface{}) *transfer.PosixProfile {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	posixUser := &transfer.PosixProfile{
		Gid: aws.Int64(int64(m["gid"].(int))),
		Uid: aws.Int64(int64(m["uid"].(int))),
	}

	if v, ok := m["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		posixUser.SecondaryGids = flex.ExpandInt64Set(v)
	}

	return posixUser
}

func flattenTransferUserPosixUser(posixUser *transfer.PosixProfile) []interface{} {
	if posixUser == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"gid":            aws.Int64Value(posixUser.Gid),
		"uid":            aws.Int64Value(posixUser.Uid),
		"secondary_gids": aws.Int64ValueSlice(posixUser.SecondaryGids),
	}

	return []interface{}{m}
}
