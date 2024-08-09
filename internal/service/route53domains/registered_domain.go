// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53domains_registered_domain", name="Registered Domain")
// @Tags(identifierAttribute="id")
func resourceRegisteredDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegisteredDomainCreate,
		ReadWithoutTimeout:   resourceRegisteredDomainRead,
		UpdateWithoutTimeout: resourceRegisteredDomainUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			contactSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"address_line_1": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"address_line_2": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"city": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"contact_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.ContactType](),
							},
							"country_code": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.CountryCode](),
							},
							names.AttrEmail: {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 254),
							},
							"extra_params": {
								Type:     schema.TypeMap,
								Optional: true,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"fax": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 30),
							},
							"first_name": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"last_name": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"organization_name": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"phone_number": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 30),
							},
							names.AttrState: {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
							"zip_code": {
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringLenBetween(0, 255),
							},
						},
					},
				}
			}

			return map[string]*schema.Schema{
				"abuse_contact_email": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"abuse_contact_phone": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"admin_contact": contactSchema(),
				"admin_privacy": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"auto_renew": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"billing_contact": contactSchema(),
				"billing_privacy": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				names.AttrCreationDate: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDomainName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"expiration_date": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name_server": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					MaxItems: 6,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"glue_ips": {
								Type:     schema.TypeSet,
								Optional: true,
								MaxItems: 2,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.IsIPAddress,
								},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 255),
									validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_.-]*`), "can contain only alphabetical characters (A-Z or a-z), numeric characters (0-9), underscore (_), the minus sign (-), and the period (.)"),
								),
							},
						},
					},
				},
				"registrant_contact": contactSchema(),
				"registrant_privacy": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"registrar_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"registrar_url": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"reseller": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"status_list": {
					Type:     schema.TypeList,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"tech_contact":    contactSchema(),
				"tech_privacy": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"transfer_lock": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"updated_date": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"whois_server": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegisteredDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.semgrep.tags.calling-UpdateTags-in-resource-create
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53DomainsClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	domainDetail, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Domains Domain (%s): %s", domainName, err)
	}

	d.SetId(aws.ToString(domainDetail.DomainName))

	var adminContact, billingContact, registrantContact, techContact *types.ContactDetail

	if v, ok := d.GetOk("admin_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.AdminContact) {
			adminContact = v
		}
	}

	if v, ok := d.GetOk("billing_contact"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v := expandContactDetail(v.([]interface{})[0].(map[string]interface{})); !reflect.DeepEqual(v, domainDetail.BillingContact) {
			billingContact = v
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

	if adminContact != nil || billingContact != nil || registrantContact != nil || techContact != nil {
		if err := modifyDomainContact(ctx, conn, d.Id(), adminContact, billingContact, registrantContact, techContact, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if adminPrivacy, billingPrivacy, registrantPrivacy, techPrivacy := d.Get("admin_privacy").(bool), d.Get("billing_privacy").(bool), d.Get("registrant_privacy").(bool), d.Get("tech_privacy").(bool); adminPrivacy != aws.ToBool(domainDetail.AdminPrivacy) || billingPrivacy != aws.ToBool(domainDetail.BillingPrivacy) || registrantPrivacy != aws.ToBool(domainDetail.RegistrantPrivacy) || techPrivacy != aws.ToBool(domainDetail.TechPrivacy) {
		if err := modifyDomainContactPrivacy(ctx, conn, d.Id(), adminPrivacy, billingPrivacy, registrantPrivacy, techPrivacy, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if v := d.Get("auto_renew").(bool); v != aws.ToBool(domainDetail.AutoRenew) {
		if err := modifyDomainAutoRenew(ctx, conn, d.Id(), v); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if v, ok := d.GetOk("name_server"); ok && len(v.([]interface{})) > 0 {
		nameservers := expandNameservers(v.([]interface{}))

		if !reflect.DeepEqual(nameservers, domainDetail.Nameservers) {
			if err := modifyDomainNameservers(ctx, conn, d.Id(), nameservers, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if v := d.Get("transfer_lock").(bool); v != hasDomainTransferLock(domainDetail.StatusList) {
		if err := modifyDomainTransferLock(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := KeyValueTags(ctx, getTagsIn(ctx))
	oldTags := tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := updateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route 53 Domains Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegisteredDomainRead(ctx, d, meta)...)
}

func resourceRegisteredDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53DomainsClient(ctx)

	domainDetail, err := findDomainDetailByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Domains Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Domains Domain (%s): %s", d.Id(), err)
	}

	d.Set("abuse_contact_email", domainDetail.AbuseContactEmail)
	d.Set("abuse_contact_phone", domainDetail.AbuseContactPhone)
	if domainDetail.AdminContact != nil {
		if err := d.Set("admin_contact", []interface{}{flattenContactDetail(domainDetail.AdminContact)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting admin_contact: %s", err)
		}
	} else {
		d.Set("admin_contact", nil)
	}
	d.Set("admin_privacy", domainDetail.AdminPrivacy)
	d.Set("auto_renew", domainDetail.AutoRenew)
	if domainDetail.CreationDate != nil {
		d.Set(names.AttrCreationDate, aws.ToTime(domainDetail.CreationDate).Format(time.RFC3339))
	} else {
		d.Set(names.AttrCreationDate, nil)
	}
	if domainDetail.BillingContact != nil {
		if err := d.Set("billing_contact", []interface{}{flattenContactDetail(domainDetail.BillingContact)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting billing_contact: %s", err)
		}
	} else {
		d.Set("billing_contact", nil)
	}
	d.Set("billing_privacy", domainDetail.BillingPrivacy)
	d.Set(names.AttrDomainName, domainDetail.DomainName)
	if domainDetail.ExpirationDate != nil {
		d.Set("expiration_date", aws.ToTime(domainDetail.ExpirationDate).Format(time.RFC3339))
	} else {
		d.Set("expiration_date", nil)
	}
	if err := d.Set("name_server", flattenNameservers(domainDetail.Nameservers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name_servers: %s", err)
	}
	if domainDetail.RegistrantContact != nil {
		if err := d.Set("registrant_contact", []interface{}{flattenContactDetail(domainDetail.RegistrantContact)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting registrant_contact: %s", err)
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
			return sdkdiag.AppendErrorf(diags, "setting tech_contact: %s", err)
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

	return diags
}

func resourceRegisteredDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53DomainsClient(ctx)

	if d.HasChanges("admin_contact", "billing_contact", "registrant_contact", "tech_contact") {
		var adminContact, billingContact, registrantContact, techContact *types.ContactDetail

		if key := "admin_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				adminContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if key := "billing_contact"; d.HasChange(key) {
			if v, ok := d.GetOk(key); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				billingContact = expandContactDetail(v.([]interface{})[0].(map[string]interface{}))
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

		if err := modifyDomainContact(ctx, conn, d.Id(), adminContact, billingContact, registrantContact, techContact, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("admin_privacy", "billing_contact", "registrant_privacy", "tech_privacy") {
		if err := modifyDomainContactPrivacy(ctx, conn, d.Id(), d.Get("admin_privacy").(bool), d.Get("billing_privacy").(bool), d.Get("registrant_privacy").(bool), d.Get("tech_privacy").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("auto_renew") {
		if err := modifyDomainAutoRenew(ctx, conn, d.Id(), d.Get("auto_renew").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("name_server") {
		if v, ok := d.GetOk("name_server"); ok && len(v.([]interface{})) > 0 {
			if err := modifyDomainNameservers(ctx, conn, d.Id(), expandNameservers(v.([]interface{})), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("transfer_lock") {
		if err := modifyDomainTransferLock(ctx, conn, d.Id(), d.Get("transfer_lock").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRegisteredDomainRead(ctx, d, meta)...)
}

func hasDomainTransferLock(statusList []string) bool {
	const (
		eppStatusClientTransferProhibited = "clientTransferProhibited"
	)
	return slices.Contains(statusList, eppStatusClientTransferProhibited)
}

func modifyDomainAutoRenew(ctx context.Context, conn *route53domains.Client, domainName string, autoRenew bool) error {
	if autoRenew {
		input := &route53domains.EnableDomainAutoRenewInput{
			DomainName: aws.String(domainName),
		}

		_, err := conn.EnableDomainAutoRenew(ctx, input)

		if err != nil {
			return fmt.Errorf("enabling Route 53 Domains Domain (%s) auto-renew: %w", domainName, err)
		}
	} else {
		input := &route53domains.DisableDomainAutoRenewInput{
			DomainName: aws.String(domainName),
		}

		_, err := conn.DisableDomainAutoRenew(ctx, input)

		if err != nil {
			return fmt.Errorf("disabling Route 53 Domains Domain (%s) auto-renew: %w", domainName, err)
		}
	}

	return nil
}

func modifyDomainContact(ctx context.Context, conn *route53domains.Client, domainName string, adminContact, billingContact, registrantContact, techContact *types.ContactDetail, timeout time.Duration) error {
	input := &route53domains.UpdateDomainContactInput{
		AdminContact:      adminContact,
		BillingContact:    billingContact,
		DomainName:        aws.String(domainName),
		RegistrantContact: registrantContact,
		TechContact:       techContact,
	}

	output, err := conn.UpdateDomainContact(ctx, input)

	if err != nil {
		return fmt.Errorf("updating Route 53 Domains Domain (%s) contacts: %w", domainName, err)
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), timeout); err != nil {
		return fmt.Errorf("waiting for Route 53 Domains Domain (%s) contacts update: %w", domainName, err)
	}

	return nil
}

func modifyDomainContactPrivacy(ctx context.Context, conn *route53domains.Client, domainName string, adminPrivacy, billingPrivacy, registrantPrivacy, techPrivacy bool, timeout time.Duration) error {
	input := &route53domains.UpdateDomainContactPrivacyInput{
		AdminPrivacy:      aws.Bool(adminPrivacy),
		BillingPrivacy:    aws.Bool(billingPrivacy),
		DomainName:        aws.String(domainName),
		RegistrantPrivacy: aws.Bool(registrantPrivacy),
		TechPrivacy:       aws.Bool(techPrivacy),
	}

	output, err := conn.UpdateDomainContactPrivacy(ctx, input)

	if err != nil {
		return fmt.Errorf("enabling Route 53 Domains Domain (%s) contact privacy: %w", domainName, err)
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), timeout); err != nil {
		return fmt.Errorf("waiting for Route 53 Domains Domain (%s) contact privacy update: %w", domainName, err)
	}

	return nil
}

func modifyDomainNameservers(ctx context.Context, conn *route53domains.Client, domainName string, nameservers []types.Nameserver, timeout time.Duration) error {
	input := &route53domains.UpdateDomainNameserversInput{
		DomainName:  aws.String(domainName),
		Nameservers: nameservers,
	}

	output, err := conn.UpdateDomainNameservers(ctx, input)

	if err != nil {
		return fmt.Errorf("updating Route 53 Domains Domain (%s) name servers: %w", domainName, err)
	}

	if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), timeout); err != nil {
		return fmt.Errorf("waiting for Route 53 Domains Domain (%s) name servers update: %w", domainName, err)
	}

	return nil
}

func modifyDomainTransferLock(ctx context.Context, conn *route53domains.Client, domainName string, transferLock bool, timeout time.Duration) error {
	if transferLock {
		input := &route53domains.EnableDomainTransferLockInput{
			DomainName: aws.String(domainName),
		}

		output, err := conn.EnableDomainTransferLock(ctx, input)

		if err != nil {
			return fmt.Errorf("enabling Route 53 Domains Domain (%s) transfer lock: %w", domainName, err)
		}

		if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), timeout); err != nil {
			return fmt.Errorf("waiting for Route 53 Domains Domain (%s) transfer lock enable: %w", domainName, err)
		}
	} else {
		input := &route53domains.DisableDomainTransferLockInput{
			DomainName: aws.String(domainName),
		}

		output, err := conn.DisableDomainTransferLock(ctx, input)

		if err != nil {
			return fmt.Errorf("disabling Route 53 Domains Domain (%s) transfer lock: %w", domainName, err)
		}

		if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId), timeout); err != nil {
			return fmt.Errorf("waiting for Route 53 Domains Domain (%s) transfer lock disable: %w", domainName, err)
		}
	}

	return nil
}

func flattenContactDetail(apiObject *types.ContactDetail) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddressLine1; v != nil {
		tfMap["address_line_1"] = aws.ToString(v)
	}

	if v := apiObject.AddressLine2; v != nil {
		tfMap["address_line_2"] = aws.ToString(v)
	}

	if v := apiObject.City; v != nil {
		tfMap["city"] = aws.ToString(v)
	}

	tfMap["contact_type"] = apiObject.ContactType
	tfMap["country_code"] = apiObject.CountryCode

	if v := apiObject.Email; v != nil {
		tfMap[names.AttrEmail] = aws.ToString(v)
	}

	if v := apiObject.ExtraParams; v != nil {
		tfMap["extra_params"] = flattenExtraParams(v)
	}

	if v := apiObject.Fax; v != nil {
		tfMap["fax"] = aws.ToString(v)
	}

	if v := apiObject.FirstName; v != nil {
		tfMap["first_name"] = aws.ToString(v)
	}

	if v := apiObject.LastName; v != nil {
		tfMap["last_name"] = aws.ToString(v)
	}

	if v := apiObject.OrganizationName; v != nil {
		tfMap["organization_name"] = aws.ToString(v)
	}

	if v := apiObject.PhoneNumber; v != nil {
		tfMap["phone_number"] = aws.ToString(v)
	}

	if v := apiObject.State; v != nil {
		tfMap[names.AttrState] = aws.ToString(v)
	}

	if v := apiObject.ZipCode; v != nil {
		tfMap["zip_code"] = aws.ToString(v)
	}

	return tfMap
}

func flattenExtraParams(apiObjects []types.ExtraParam) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := make(map[string]interface{}, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap[string(apiObject.Name)] = aws.ToString(apiObject.Value)
	}

	return tfMap
}

func expandContactDetail(tfMap map[string]interface{}) *types.ContactDetail {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ContactDetail{}

	if v, ok := tfMap["address_line_1"].(string); ok {
		apiObject.AddressLine1 = aws.String(v)
	}

	if v, ok := tfMap["address_line_2"].(string); ok {
		apiObject.AddressLine2 = aws.String(v)
	}

	if v, ok := tfMap["city"].(string); ok {
		apiObject.City = aws.String(v)
	}

	if v, ok := tfMap["contact_type"].(string); ok {
		apiObject.ContactType = types.ContactType(v)
	}

	if v, ok := tfMap["country_code"].(string); ok {
		apiObject.CountryCode = types.CountryCode(v)
	}

	if v, ok := tfMap[names.AttrEmail].(string); ok {
		apiObject.Email = aws.String(v)
	}

	if v, ok := tfMap["extra_params"].(map[string]interface{}); ok {
		apiObject.ExtraParams = expandExtraParams(v)
	}

	if v, ok := tfMap["fax"].(string); ok {
		apiObject.Fax = aws.String(v)
	}

	if v, ok := tfMap["first_name"].(string); ok {
		apiObject.FirstName = aws.String(v)
	}

	if v, ok := tfMap["last_name"].(string); ok {
		apiObject.LastName = aws.String(v)
	}

	if v, ok := tfMap["organization_name"].(string); ok {
		apiObject.OrganizationName = aws.String(v)
	}

	if v, ok := tfMap["phone_number"].(string); ok {
		apiObject.PhoneNumber = aws.String(v)
	}

	if v, ok := tfMap[names.AttrState].(string); ok {
		apiObject.State = aws.String(v)
	}

	if v, ok := tfMap["zip_code"].(string); ok {
		apiObject.ZipCode = aws.String(v)
	}

	return apiObject
}

func expandExtraParams(tfMap map[string]interface{}) []types.ExtraParam {
	if len(tfMap) == 0 {
		return nil
	}

	var apiObjects []types.ExtraParam

	for k, vRaw := range tfMap {
		v, ok := vRaw.(string)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, types.ExtraParam{
			Name:  types.ExtraParamName(k),
			Value: aws.String(v),
		})
	}

	return apiObjects
}

func flattenNameserver(apiObject *types.Nameserver) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GlueIps; v != nil {
		tfMap["glue_ips"] = v
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func expandNameserver(tfMap map[string]interface{}) *types.Nameserver {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Nameserver{}

	if v, ok := tfMap["glue_ips"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.GlueIps = aws.ToStringSlice(flex.ExpandStringSet(v))
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandNameservers(tfList []interface{}) []types.Nameserver {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Nameserver

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandNameserver(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenNameservers(apiObjects []types.Nameserver) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNameserver(&apiObject))
	}

	return tfList
}
