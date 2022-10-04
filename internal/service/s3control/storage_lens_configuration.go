package s3control

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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

	input := &s3control.PutStorageLensConfigurationInput{}

	d.SetId(id)

	return resourceStorageLensConfigurationRead(ctx, d, meta)
}

func resourceStorageLensConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStorageLensConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceStorageLensConfigurationRead(ctx, d, meta)
}

func resourceStorageLensConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
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
