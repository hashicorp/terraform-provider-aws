package amplify

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDomainAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainAssociationCreate,
		ReadWithoutTimeout:   resourceDomainAssociationRead,
		UpdateWithoutTimeout: resourceDomainAssociationUpdate,
		DeleteWithoutTimeout: resourceDomainAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate_verification_dns_record": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"sub_domain": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"dns_record": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"prefix": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"verified": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"wait_for_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceDomainAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyConn()

	appID := d.Get("app_id").(string)
	domainName := d.Get("domain_name").(string)
	id := DomainAssociationCreateResourceID(appID, domainName)

	input := &amplify.CreateDomainAssociationInput{
		AppId:             aws.String(appID),
		DomainName:        aws.String(domainName),
		SubDomainSettings: expandSubDomainSettings(d.Get("sub_domain").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Creating Amplify Domain Association: %s", input)
	_, err := conn.CreateDomainAssociationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amplify Domain Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDomainAssociationCreated(ctx, conn, appID, domainName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Amplify Domain Association (%s) to create: %s", d.Id(), err)
	}

	if d.Get("wait_for_verification").(bool) {
		if _, err := waitDomainAssociationVerified(ctx, conn, appID, domainName); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Amplify Domain Association (%s) to verify: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainAssociationRead(ctx, d, meta)...)
}

func resourceDomainAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyConn()

	appID, domainName, err := DomainAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Amplify Domain Association ID: %s", err)
	}

	domainAssociation, err := FindDomainAssociationByAppIDAndDomainName(ctx, conn, appID, domainName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify Domain Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Amplify Domain Association (%s): %s", d.Id(), err)
	}

	d.Set("app_id", appID)
	d.Set("arn", domainAssociation.DomainAssociationArn)
	d.Set("certificate_verification_dns_record", domainAssociation.CertificateVerificationDNSRecord)
	d.Set("domain_name", domainAssociation.DomainName)
	if err := d.Set("sub_domain", flattenSubDomains(domainAssociation.SubDomains)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sub_domain: %s", err)
	}

	return diags
}

func resourceDomainAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyConn()

	appID, domainName, err := DomainAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Amplify Domain Association ID: %s", err)
	}

	if d.HasChange("sub_domain") {
		input := &amplify.UpdateDomainAssociationInput{
			AppId:             aws.String(appID),
			DomainName:        aws.String(domainName),
			SubDomainSettings: expandSubDomainSettings(d.Get("sub_domain").(*schema.Set).List()),
		}

		log.Printf("[DEBUG] Creating Amplify Domain Association: %s", input)
		_, err := conn.UpdateDomainAssociationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Amplify Domain Association (%s): %s", d.Id(), err)
		}
	}

	if d.Get("wait_for_verification").(bool) {
		if _, err := waitDomainAssociationVerified(ctx, conn, appID, domainName); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Amplify Domain Association (%s) to verify: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainAssociationRead(ctx, d, meta)...)
}

func resourceDomainAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyConn()

	appID, domainName, err := DomainAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Amplify Domain Association ID: %s", err)
	}

	log.Printf("[DEBUG] Deleting Amplify Domain Association: %s", d.Id())
	_, err = conn.DeleteDomainAssociationWithContext(ctx, &amplify.DeleteDomainAssociationInput{
		AppId:      aws.String(appID),
		DomainName: aws.String(domainName),
	})

	if tfawserr.ErrCodeEquals(err, amplify.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Amplify Domain Association (%s): %s", d.Id(), err)
	}

	return diags
}

func expandSubDomainSetting(tfMap map[string]interface{}) *amplify.SubDomainSetting {
	if tfMap == nil {
		return nil
	}

	apiObject := &amplify.SubDomainSetting{}

	if v, ok := tfMap["branch_name"].(string); ok && v != "" {
		apiObject.BranchName = aws.String(v)
	}

	// Empty prefix is allowed.
	if v, ok := tfMap["prefix"].(string); ok {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandSubDomainSettings(tfList []interface{}) []*amplify.SubDomainSetting {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*amplify.SubDomainSetting

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandSubDomainSetting(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenSubDomain(apiObject *amplify.SubDomain) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DnsRecord; v != nil {
		tfMap["dns_record"] = aws.StringValue(v)
	}

	if v := apiObject.SubDomainSetting; v != nil {
		apiObject := v

		if v := apiObject.BranchName; v != nil {
			tfMap["branch_name"] = aws.StringValue(v)
		}

		if v := apiObject.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}
	}

	if v := apiObject.Verified; v != nil {
		tfMap["verified"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenSubDomains(apiObjects []*amplify.SubDomain) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenSubDomain(apiObject))
	}

	return tfList
}
