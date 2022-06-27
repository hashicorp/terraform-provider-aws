package glacier

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVault() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultCreate,
		Read:   resourceVaultRead,
		Update: resourceVaultUpdate,
		Delete: resourceVaultDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
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

func resourceVaultCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlacierConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &glacier.CreateVaultInput{
		VaultName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.CreateVault(input)
	if err != nil {
		return fmt.Errorf("Error creating Glacier Vault: %w", err)
	}

	d.SetId(d.Get("name").(string))

	if len(tags) > 0 {
		if err := UpdateTags(conn, d.Id(), nil, tags.Map()); err != nil {
			return fmt.Errorf("error updating Glacier Vault (%s) tags: %w", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("access_policy"); ok {
		if err := resourceVaultPolicyUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating Glacier Vault (%s) access policy: %w", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("notification"); ok {
		if err := resourceVaultNotificationUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating Glacier Vault (%s) notification: %w", d.Id(), err)
		}
	}

	return resourceVaultRead(d, meta)
}

func resourceVaultUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlacierConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Glacier Vault (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("access_policy") {
		if err := resourceVaultPolicyUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating Glacier Vault (%s) access policy: %w", d.Id(), err)
		}
	}

	if d.HasChange("notification") {
		if err := resourceVaultNotificationUpdate(conn, d); err != nil {
			return fmt.Errorf("error updating Glacier Vault (%s) notification: %w", d.Id(), err)
		}
	}

	return resourceVaultRead(d, meta)
}

func resourceVaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlacierConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glacier.DescribeVaultInput{
		VaultName: aws.String(d.Id()),
	}

	out, err := conn.DescribeVault(input)
	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Glaier Vault (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error reading Glacier Vault: %w", err)
	}

	awsClient := meta.(*conns.AWSClient)
	d.Set("name", out.VaultName)
	d.Set("arn", out.VaultARN)

	location, err := buildVaultLocation(awsClient.AccountID, d.Id())
	if err != nil {
		return err
	}
	d.Set("location", location)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Glacier Vault (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	log.Printf("[DEBUG] Getting the access_policy for Vault %s", d.Id())
	pol, err := conn.GetVaultAccessPolicy(&glacier.GetVaultAccessPolicyInput{
		VaultName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		d.Set("access_policy", "")
	} else if err != nil {
		return fmt.Errorf("error getting access policy for Glacier Vault (%s): %w", d.Id(), err)
	} else if pol != nil && pol.Policy != nil {
		policy, err := verify.PolicyToSet(d.Get("access_policy").(string), aws.StringValue(pol.Policy.Policy))

		if err != nil {
			return err
		}

		d.Set("access_policy", policy)
	}

	notifications, err := getVaultNotification(conn, d.Id())
	if tfawserr.ErrCodeEquals(err, glacier.ErrCodeResourceNotFoundException) {
		d.Set("notification", []map[string]interface{}{})
	} else if pol != nil {
		d.Set("notification", notifications)
	} else {
		return fmt.Errorf("error setting notification: %w", err)
	}

	return nil
}

func resourceVaultDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlacierConn

	log.Printf("[DEBUG] Glacier Delete Vault: %s", d.Id())
	_, err := conn.DeleteVault(&glacier.DeleteVaultInput{
		VaultName: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting Glacier Vault: %w", err)
	}
	return nil
}

func resourceVaultNotificationUpdate(conn *glacier.Glacier, d *schema.ResourceData) error {

	if v, ok := d.GetOk("notification"); ok {
		settings := v.([]interface{})

		s := settings[0].(map[string]interface{})

		_, err := conn.SetVaultNotifications(&glacier.SetVaultNotificationsInput{
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
		_, err := conn.DeleteVaultNotifications(&glacier.DeleteVaultNotificationsInput{
			VaultName: aws.String(d.Id()),
		})

		if err != nil {
			return fmt.Errorf("Error Removing Glacier Vault Notifications: %w", err)
		}

	}

	return nil
}

func resourceVaultPolicyUpdate(conn *glacier.Glacier, d *schema.ResourceData) error {
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

		_, err := conn.SetVaultAccessPolicy(&glacier.SetVaultAccessPolicyInput{
			VaultName: aws.String(d.Id()),
			Policy:    policy,
		})

		if err != nil {
			return fmt.Errorf("Error putting Glacier Vault policy: %w", err)
		}
	} else {
		log.Printf("[DEBUG] Glacier Vault: %s, delete policy: %s", vaultName, policy)
		_, err := conn.DeleteVaultAccessPolicy(&glacier.DeleteVaultAccessPolicyInput{
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

func getVaultNotification(conn *glacier.Glacier, vaultName string) ([]map[string]interface{}, error) {
	request := &glacier.GetVaultNotificationsInput{
		VaultName: aws.String(vaultName),
	}

	response, err := conn.GetVaultNotifications(request)
	if err != nil {
		return nil, fmt.Errorf("Error reading Glacier Vault Notifications: %w", err)
	}

	notifications := make(map[string]interface{})

	log.Print("[DEBUG] Flattening Glacier Vault Notifications")

	notifications["events"] = aws.StringValueSlice(response.VaultNotificationConfig.Events)
	notifications["sns_topic"] = aws.StringValue(response.VaultNotificationConfig.SNSTopic)

	return []map[string]interface{}{notifications}, nil
}
