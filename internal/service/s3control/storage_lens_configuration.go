package s3control

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStorageLensConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStorageLensConfigurationCreate,
		ReadWithoutTimeout:   resourceStorageLensConfigurationRead,
		UpdateWithoutTimeout: resourceStorageLensConfigurationUpdate,
		DeleteWithoutTimeout: resourceStorageLensConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStorageLensConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	configID := d.Get("config_id").(string)
	id := StorageLensConfigurationCreateResourceID(accountID, configID)

	input := &s3control.PutStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	}

	if len(tags) > 0 {
		input.Tags = StorageLensTags(tags.IgnoreAWS())
	}

	_, err := conn.PutStorageLensConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Storage Lens Configuration (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceStorageLensConfigurationRead(ctx, d, meta)
}

func resourceStorageLensConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindStorageLensConfigurationByAccountIDAndConfigID(ctx, conn, accountID, configID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Storage Lens Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("config_id", configID)

	tags, err := storageLensConfigurationListTags(ctx, conn, accountID, configID)

	if err != nil {
		return diag.Errorf("listing tags for S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceStorageLensConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := storageLensConfigurationUpdateTags(ctx, conn, accountID, configID, o, n); err != nil {
			return diag.Errorf("updating S3 Storage Lens Configuration (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceStorageLensConfigurationRead(ctx, d, meta)
}

func resourceStorageLensConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting S3 Storage Lens Configuration: %s", d.Id())
	_, err = conn.DeleteStorageLensConfigurationWithContext(ctx, &s3control.DeleteStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchConfiguration) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

const storageLensConfigurationResourceIDSeparator = ":"

func StorageLensConfigurationCreateResourceID(accountID, configID string) string {
	parts := []string{accountID, configID}
	id := strings.Join(parts, storageLensConfigurationResourceIDSeparator)

	return id
}

func StorageLensConfigurationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, storageLensConfigurationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account-id%[2]sconfig-id", id, storageLensConfigurationResourceIDSeparator)
}
