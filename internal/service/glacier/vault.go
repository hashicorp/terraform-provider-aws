package glacier

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVault() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVaultCreate,
		ReadWithoutTimeout:   resourceVaultRead,
		UpdateWithoutTimeout: resourceVaultUpdate,
		DeleteWithoutTimeout: resourceVaultDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[.0-9A-Za-z-_]+$`),
						"only alphanumeric characters, hyphens, underscores, and periods are allowed"),
				),
			},

			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},

			"notification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"events": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"ArchiveRetrievalCompleted",
									"InventoryRetrievalCompleted",
								}, false),
							},
							Set: schema.HashString,
						},
						"sns_topic": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVaultCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &glacier.CreateVaultInput{
		VaultName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.CreateVaultWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glacier Vault: %s", err)
	}

	d.SetId(d.Get("name").(string))

	if len(tags) > 0 {
		if err := UpdateTags(ctx, conn, d.Id(), nil, tags.Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glacier Vault (%s) tags: %s", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("access_policy"); ok {
		if err := resourceVaultPolicyUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glacier Vault (%s) access policy: %s", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("notification"); ok {
		if err := resourceVaultNotificationUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glacier Vault (%s) notification: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVaultRead(ctx, d, meta)...)
}

func resourceVaultUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glacier Vault (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("access_policy") {
		if err := resourceVaultPolicyUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glacier Vault (%s) access policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("notification") {
		if err := resourceVaultNotificationUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glacier Vault (%s) notification: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVaultRead(ctx, d, meta)...)
}

func resourceVaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glacier.DescribeVaultInput{
		VaultName: aws.String(d.Id()),
	}

	out, err := conn.DescribeVaultWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Glaier Vault (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glacier Vault (%s): %s", d.Id(), err)
	}

	awsClient := meta.(*conns.AWSClient)
	d.Set("name", out.VaultName)
	d.Set("arn", out.VaultARN)

	location, err := buildVaultLocation(awsClient.AccountID, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glacier Vault (%s): %s", d.Id(), err)
	}
	d.Set("location", location)

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glacier Vault (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	log.Printf("[DEBUG] Getting the access_policy for Vault %s", d.Id())
	pol, err := conn.GetVaultAccessPolicyWithContext(ctx, &glacier.GetVaultAccessPolicyInput{
		VaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		d.Set("access_policy", "")
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glacier Vault (%s): reading policy: %s", d.Id(), err)
	} else if pol != nil && pol.Policy != nil {
		policy, err := verify.PolicyToSet(d.Get("access_policy").(string), aws.StringValue(pol.Policy.Policy))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Glacier Vault (%s): setting policy: %s", d.Id(), err)
		}

		d.Set("access_policy", policy)
	}

	notifications, err := getVaultNotification(ctx, conn, d.Id())
	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		d.Set("notification", []map[string]interface{}{})
	} else if pol != nil {
		d.Set("notification", notifications)
	} else {
		return sdkdiag.AppendErrorf(diags, "setting notification: %s", err)
	}

	return diags
}

func resourceVaultDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlacierConn()

	log.Printf("[DEBUG] Deleting Glacier Vault: %s", d.Id())
	_, err := conn.DeleteVaultWithContext(ctx, &glacier.DeleteVaultInput{
		VaultName: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glacier Vault: %s", err)
	}
	return diags
}

func resourceVaultNotificationUpdate(ctx context.Context, conn *glacier.Glacier, d *schema.ResourceData) error {
	if v, ok := d.GetOk("notification"); ok {
		settings := v.([]interface{})

		s := settings[0].(map[string]interface{})

		_, err := conn.SetVaultNotificationsWithContext(ctx, &glacier.SetVaultNotificationsInput{
			VaultName: aws.String(d.Id()),
			VaultNotificationConfig: &glacier.VaultNotificationConfig{
				SNSTopic: aws.String(s["sns_topic"].(string)),
				Events:   flex.ExpandStringSet(s["events"].(*schema.Set)),
			},
		})

		if err != nil {
			return fmt.Errorf("Error Updating Glacier Vault Notifications: %w", err)
		}
	} else {
		_, err := conn.DeleteVaultNotificationsWithContext(ctx, &glacier.DeleteVaultNotificationsInput{
			VaultName: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("Error Removing Glacier Vault Notifications: %w", err)
		}
	}

	return nil
}

func resourceVaultPolicyUpdate(ctx context.Context, conn *glacier.Glacier, d *schema.ResourceData) error {
	vaultName := d.Id()
	policyContents, err := structure.NormalizeJsonString(d.Get("access_policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyContents, err)
	}

	policy := &glacier.VaultAccessPolicy{
		Policy: aws.String(policyContents),
	}

	if policyContents != "" {
		log.Printf("[DEBUG] Glacier Vault: %s, put policy", vaultName)

		_, err := conn.SetVaultAccessPolicyWithContext(ctx, &glacier.SetVaultAccessPolicyInput{
			VaultName: aws.String(d.Id()),
			Policy:    policy,
		})

		if err != nil {
			return fmt.Errorf("Error putting Glacier Vault policy: %w", err)
		}
	} else {
		log.Printf("[DEBUG] Glacier Vault: %s, delete policy: %s", vaultName, policy)
		_, err := conn.DeleteVaultAccessPolicyWithContext(ctx, &glacier.DeleteVaultAccessPolicyInput{
			VaultName: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("Error deleting Glacier Vault policy: %w", err)
		}
	}

	return nil
}

func buildVaultLocation(accountId, vaultName string) (string, error) {
	if accountId == "" {
		return "", errors.New("AWS account ID unavailable - failed to construct Vault location")
	}
	return fmt.Sprintf("/" + accountId + "/vaults/" + vaultName), nil
}

func getVaultNotification(ctx context.Context, conn *glacier.Glacier, vaultName string) ([]map[string]interface{}, error) {
	request := &glacier.GetVaultNotificationsInput{
		VaultName: aws.String(vaultName),
	}

	response, err := conn.GetVaultNotificationsWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("Error reading Glacier Vault Notifications: %w", err)
	}

	notifications := make(map[string]interface{})

	log.Print("[DEBUG] Flattening Glacier Vault Notifications")

	notifications["events"] = aws.StringValueSlice(response.VaultNotificationConfig.Events)
	notifications["sns_topic"] = aws.StringValue(response.VaultNotificationConfig.SNSTopic)

	return []map[string]interface{}{notifications}, nil
}
