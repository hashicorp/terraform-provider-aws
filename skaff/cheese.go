package skaff
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: You have opted to include helpful guiding comments. These comments are
// meant to teach and remind. However, they should be removed before submitting
// your work in a PR. Thank you!

import (
	// TIP: This is a common set of imports but not fully customized to your code
	// since your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/skaff"
	"github.com/aws/aws-sdk-go-v2/service/skaff/types" // TIP: Some v2 packages use a separate package for types while some do not.
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCheese() *schema.Resource {
	return &schema.Resource{
		// TIP: These 4 functions handle CRUD responsibilities below.
		CreateWithoutTimeout: resourceCheeseCreate,
		ReadWithoutTimeout:   resourceCheeseRead,
		UpdateWithoutTimeout: resourceCheeseUpdate,
		DeleteWithoutTimeout: resourceCheeseDelete,

		// TIP: Users can configure timeout lengths (if you use the times they
		// provide). These are the defaults if they don't configure timeouts.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		// TIP: In the schema, add each of the arguments and attributes in
		// snake case (e.g., delete_automated_backups).
		// * Alphabetize arguments to make it easier to find them.
		// * Do not add a blank line between arguments/attributes.
		// 
		// Users can configure argument values while attribute values cannot be
		// configured and are read as output. Arguments have either:
		// Required: true,
		// Optional: true,
		// 
		// All attributes will be computed and some arguments. If users will
		// want to read updated information or detect drift for an argument,
		// it should be computed:
		// Computed: true,
		//
		// You will typically find arguments in the input struct 
		// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
		// they are only in the input struct (e.g., ModifyDBInstanceInput) for
		// the modify operation.
		//
		// For more about schema options, visit 
		// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Schema
		Schema: map[string]*schema.Schema{
			"arn": { // TIP: Many, but not all, resources have an `arn` attribute.
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace_with_arguments": { // TIP: Add all your arguments and attributes.
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":         tftags.TagsSchema(), // TIP: Many, but not all, resources have `tags` and `tags_all` attributes.
			"tags_all":     tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCheeseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	domainName := d.Get("domain_name").(string)
	domainDetail, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return diag.Errorf("error reading Route 53 Domains Domain (%s): %s", domainName, err)
	}

	d.SetId(aws.ToString(domainDetail.DomainName))

	var adminContact, registrantContact, techContact *types.ContactDetail

	if v, ok := d.GetOk("admin_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.AdminContact) {
			adminContact = v
		}
	}

	if v, ok := d.GetOk("registrant_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.RegistrantContact) {
			registrantContact = v
		}
	}

	if v, ok := d.GetOk("tech_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.TechContact) {
			techContact = v
		}
	}

	if adminContact != nil || registrantContact != nil || techContact != nil {
		if err := modifyDomainContact(ctx, conn, d.Id(), adminContact, registrantContact, techContact, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if adminPrivacy, registrantPrivacy, techPrivacy := d.Get("admin_privacy").(bool), d.Get("registrant_privacy").(bool), d.Get("tech_privacy").(bool); adminPrivacy != aws.ToBool(domainDetail.AdminPrivacy) || registrantPrivacy != aws.ToBool(domainDetail.RegistrantPrivacy) || techPrivacy != aws.ToBool(domainDetail.TechPrivacy) {
		if err := modifyDomainContactPrivacy(ctx, conn, d.Id(), adminPrivacy, registrantPrivacy, techPrivacy, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if v := d.Get("auto_renew").(bool); v != aws.ToBool(domainDetail.AutoRenew) {
		if err := modifyDomainAutoRenew(ctx, conn, d.Id(), v); err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("name_server"); ok && len(v.([]interface{})) > 0 {
		nameservers := expandNameservers(v.([]interface{}))

		if !reflect.DeepEqual(nameservers, domainDetail.Nameservers) {
			if err := modifyDomainNameservers(ctx, conn, d.Id(), nameservers, d.Timeout(schema.TimeoutCreate)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if v := d.Get("transfer_lock").(bool); v != hasDomainTransferLock(domainDetail.StatusList) {
		if err := modifyDomainTransferLock(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("error listing tags for Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{}))).IgnoreConfig(ignoreTagsConfig)
	oldTags := tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := UpdateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return diag.Errorf("error updating Route 53 Domains Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCheeseRead(ctx, d, meta)
}

func resourceCheeseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53DomainsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	domainDetail, err := findDomainDetailByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Domains Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	d.Set("abuse_contact_email", domainDetail.AbuseContactEmail)
	d.Set("abuse_contact_phone", domainDetail.AbuseContactPhone)
	if domainDetail.AdminContact != nil {
		if err := d.Set("admin_contact", []interface{}{flattenContactDetail(domainDetail.AdminContact)}); err != nil {
			return diag.Errorf("error setting admin_contact: %s", err)
		}
	} else {
		d.Set("admin_contact", nil)
	}
	d.Set("admin_privacy", domainDetail.AdminPrivacy)
	d.Set("auto_renew", domainDetail.AutoRenew)
	if domainDetail.CreationDate != nil {
		d.Set("creation_date", aws.ToTime(domainDetail.CreationDate).Format(time.RFC3339))
	} else {
		d.Set("creation_date", nil)
	}
	d.Set("domain_name", domainDetail.DomainName)
	if domainDetail.ExpirationDate != nil {
		d.Set("expiration_date", aws.ToTime(domainDetail.ExpirationDate).Format(time.RFC3339))
	} else {
		d.Set("expiration_date", nil)
	}
	if err := d.Set("name_server", flattenNameservers(domainDetail.Nameservers)); err != nil {
		return diag.Errorf("error setting name_servers: %s", err)
	}
	if domainDetail.RegistrantContact != nil {
		if err := d.Set("registrant_contact", []interface{}{flattenContactDetail(domainDetail.RegistrantContact)}); err != nil {
			return diag.Errorf("error setting registrant_contact: %s", err)
		}
	} else {
		d.Set("registrant_contact", nil)
	}
	d.Set("registrant_privacy", domainDetail.RegistrantPrivacy)
	d.Set("registrar_name", domainDetail.RegistrarName)
	d.Set("registrar_url", domainDetail.RegistrarUrl)
	d.Set("reseller", domainDetail.Reseller)
	statusList := domainDetail.StatusList
	d.Set("status_list", statusList)
	if domainDetail.TechContact != nil {
		if err := d.Set("tech_contact", []interface{}{flattenContactDetail(domainDetail.TechContact)}); err != nil {
			return diag.Errorf("error setting tech_contact: %s", err)
		}
	} else {
		d.Set("tech_contact", nil)
	}
	d.Set("tech_privacy", domainDetail.TechPrivacy)
	d.Set("transfer_lock", hasDomainTransferLock(statusList))
	if domainDetail.UpdatedDate != nil {
		d.Set("updated_date", aws.ToTime(domainDetail.UpdatedDate).Format(time.RFC3339))
	} else {
		d.Set("updated_date", nil)
	}
	d.Set("whois_server", domainDetail.WhoIsServer)

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("error listing tags for Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceCheeseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	if d.HasChanges("admin_contact", "registrant_contact", "tech_contact") {
		var adminContact, registrantContact, techContact *types.ContactDetail

		if key := "admin_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				adminContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if key := "registrant_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				registrantContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if key := "tech_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				techContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if err := modifyDomainContact(ctx, conn, d.Id(), adminContact, registrantContact, techContact, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges("admin_privacy", "registrant_privacy", "tech_privacy") {
		if err := modifyDomainContactPrivacy(ctx, conn, d.Id(), d.Get("admin_privacy").(bool), d.Get("registrant_privacy").(bool), d.Get("tech_privacy").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("auto_renew") {
		if err := modifyDomainAutoRenew(ctx, conn, d.Id(), d.Get("auto_renew").(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name_server") {
		if v, ok := d.GetOk("name_server"); ok && len(v.([]interface{})) > 0 {
			if err := modifyDomainNameservers(ctx, conn, d.Id(), expandNameservers(v.([]interface{})), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("transfer_lock") {
		if err := modifyDomainTransferLock(ctx, conn, d.Id(), d.Get("transfer_lock").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error updating Route 53 Domains Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCheeseRead(ctx, d, meta)
}
