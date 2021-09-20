package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	tfnet "github.com/hashicorp/terraform-provider-aws/aws/internal/net"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPSetCreate,
		Read:   resourceIPSetRead,
		Update: resourceIPSetUpdate,
		Delete: resourceIPSetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10000,
				Elem:     &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					o, n := d.GetChange("addresses")
					oldAddresses := o.(*schema.Set).List()
					newAddresses := n.(*schema.Set).List()
					if len(oldAddresses) == len(newAddresses) {
						for _, ov := range oldAddresses {
							hasAddress := false
							for _, nv := range newAddresses {
								if verify.CIDRBlocksEqual(ov.(string), nv.(string)) {
									hasAddress = true
									break
								}
							}
							if !hasAddress {
								return false
							}
						}
						return true
					}
					return false
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"ip_address_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.IPAddressVersionIpv4,
					wafv2.IPAddressVersionIpv6,
				}, false),
			},
			"lock_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIPSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	params := &wafv2.CreateIPSetInput{
		Addresses:        aws.StringSlice([]string{}),
		IPAddressVersion: aws.String(d.Get("ip_address_version").(string)),
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		params.Addresses = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.Tags = tags.IgnoreAws().Wafv2Tags()
	}

	resp, err := conn.CreateIPSet(params)

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 IPSet: %s", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 IPSet")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceIPSetRead(d, meta)
}

func resourceIPSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &wafv2.GetIPSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetIPSet(params)
	if err != nil {
		if tfawserr.ErrMessageContains(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.IPSet == nil {
		return fmt.Errorf("Error reading WAFv2 IPSet")
	}

	d.Set("name", resp.IPSet.Name)
	d.Set("description", resp.IPSet.Description)
	d.Set("ip_address_version", resp.IPSet.IPAddressVersion)
	d.Set("arn", resp.IPSet.ARN)
	d.Set("lock_token", resp.LockToken)

	if err := d.Set("addresses", flex.FlattenStringSet(resp.IPSet.Addresses)); err != nil {
		return fmt.Errorf("Error setting addresses: %s", err)
	}

	arn := aws.StringValue(resp.IPSet.ARN)
	tags, err := tftags.Wafv2ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 IpSet (%s): %s", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceIPSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	log.Printf("[INFO] Updating WAFv2 IPSet %s", d.Id())

	params := &wafv2.UpdateIPSetInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		Addresses: aws.StringSlice([]string{}),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		params.Addresses = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateIPSet(params)

	if err != nil {
		return fmt.Errorf("Error updating WAFv2 IPSet: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating tags: %s", err)
		}
	}

	return resourceIPSetRead(d, meta)
}

func resourceIPSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFV2Conn

	log.Printf("[INFO] Deleting WAFv2 IPSet %s", d.Id())

	params := &wafv2.DeleteIPSetInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteIPSet(params)
		if err != nil {
			if tfawserr.ErrMessageContains(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteIPSet(params)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 IPSet: %s", err)
	}

	return nil
}
