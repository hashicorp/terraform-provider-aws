// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package customerprofiles

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles"
	"github.com/aws/aws-sdk-go-v2/service/customerprofiles/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_customerprofiles_profile")
func ResourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceProfileImport,
		},

		Schema: map[string]*schema.Schema{
			"account_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"additional_information": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrAddress: customerProfileAddressSchema(),
			names.AttrAttributes: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"billing_address": customerProfileAddressSchema(),
			"birth_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"business_email_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"business_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"business_phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"email_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"gender_string": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"home_phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mailing_address": customerProfileAddressSchema(),
			"middle_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mobile_phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"party_type_string": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"personal_email_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"shipping_address": customerProfileAddressSchema(),
		},
	}
}

func customerProfileAddressSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address_1": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"address_2": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"address_3": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"address_4": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"city": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"country": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"county": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"postal_code": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"province": {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrState: {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	input := &customerprofiles.CreateProfileInput{
		DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
	}

	if v, ok := d.GetOk("account_number"); ok {
		input.AccountNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("additional_information"); ok {
		input.AdditionalInformation = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrAddress); ok {
		input.Address = expandAddress(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrAttributes); ok {
		input.Attributes = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("billing_address"); ok {
		input.BillingAddress = expandAddress(v.([]interface{}))
	}

	if v, ok := d.GetOk("birth_date"); ok {
		input.BirthDate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("business_email_address"); ok {
		input.BusinessEmailAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("business_name"); ok {
		input.BusinessName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("business_phone_number"); ok {
		input.BusinessPhoneNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("email_address"); ok {
		input.EmailAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("first_name"); ok {
		input.FirstName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("gender_string"); ok {
		input.GenderString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("home_phone_number"); ok {
		input.HomePhoneNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("last_name"); ok {
		input.LastName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("mailing_address"); ok {
		input.MailingAddress = expandAddress(v.([]interface{}))
	}

	if v, ok := d.GetOk("middle_name"); ok {
		input.MiddleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("mobile_phone_number"); ok {
		input.MobilePhoneNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("party_type_string"); ok {
		input.PartyTypeString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("personal_email_address"); ok {
		input.PersonalEmailAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("phone_number"); ok {
		input.PhoneNumber = aws.String(v.(string))
	}

	if v, ok := d.GetOk("shipping_address"); ok {
		input.ShippingAddress = expandAddress(v.([]interface{}))
	}

	output, err := conn.CreateProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Customer Profiles Profile: %s", err)
	}

	d.SetId(aws.ToString(output.ProfileId))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindProfileByTwoPartKey(ctx, conn, d.Id(), d.Get(names.AttrDomainName).(string))
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Customer Profiles Profile (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	output, err := FindProfileByTwoPartKey(ctx, conn, d.Id(), domainName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Customer Profiles Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Customer Profiles Profile: (%s) %s", d.Id(), err)
	}

	d.Set("account_number", output.AccountNumber)
	d.Set("additional_information", output.AdditionalInformation)
	d.Set(names.AttrAddress, flattenAddress(output.Address))
	d.Set("account_number", output.AccountNumber)
	d.Set(names.AttrAttributes, output.Attributes)
	d.Set("billing_address", flattenAddress(output.BillingAddress))
	d.Set("birth_date", output.BirthDate)
	d.Set("business_email_address", output.BusinessEmailAddress)
	d.Set("business_name", output.BusinessName)
	d.Set("business_phone_number", output.BusinessPhoneNumber)
	d.Set("email_address", output.EmailAddress)
	d.Set("first_name", output.FirstName)
	d.Set("gender_string", output.GenderString)
	d.Set("home_phone_number", output.HomePhoneNumber)
	d.Set("last_name", output.LastName)
	d.Set("mailing_address", flattenAddress(output.MailingAddress))
	d.Set("middle_name", output.MiddleName)
	d.Set("mobile_phone_number", output.MobilePhoneNumber)
	d.Set("party_type_string", output.PartyTypeString)
	d.Set("personal_email_address", output.PersonalEmailAddress)
	d.Set("phone_number", output.PhoneNumber)
	d.Set("shipping_address", flattenAddress(output.ShippingAddress))

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	input := &customerprofiles.UpdateProfileInput{
		DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
		ProfileId:  aws.String(d.Id()),
	}

	if d.HasChange("account_number") {
		input.AccountNumber = aws.String(d.Get("account_number").(string))
	}

	if d.HasChange("additional_information") {
		input.AdditionalInformation = aws.String(d.Get("additional_information").(string))
	}

	if d.HasChange(names.AttrAddress) {
		input.Address = expandUpdateAddress(d.Get(names.AttrAddress).([]interface{}))
	}

	if d.HasChange(names.AttrAttributes) {
		input.Attributes = flex.ExpandStringValueMap(d.Get(names.AttrAttributes).(map[string]interface{}))
	}

	if d.HasChange("billing_address") {
		input.BillingAddress = expandUpdateAddress(d.Get("billing_address").([]interface{}))
	}

	if d.HasChange("additional_information") {
		input.AdditionalInformation = aws.String(d.Get("additional_information").(string))
	}

	if d.HasChange("birth_date") {
		input.BirthDate = aws.String(d.Get("birth_date").(string))
	}

	if d.HasChange("business_email_address") {
		input.BusinessEmailAddress = aws.String(d.Get("business_email_address").(string))
	}

	if d.HasChange("business_name") {
		input.BusinessName = aws.String(d.Get("business_name").(string))
	}

	if d.HasChange("business_phone_number") {
		input.BusinessPhoneNumber = aws.String(d.Get("business_phone_number").(string))
	}

	if d.HasChange("email_address") {
		input.EmailAddress = aws.String(d.Get("email_address").(string))
	}

	if d.HasChange("first_name") {
		input.FirstName = aws.String(d.Get("first_name").(string))
	}

	if d.HasChange("gender_string") {
		input.GenderString = aws.String(d.Get("gender_string").(string))
	}

	if d.HasChange("home_phone_number") {
		input.HomePhoneNumber = aws.String(d.Get("home_phone_number").(string))
	}

	if d.HasChange("last_name") {
		input.LastName = aws.String(d.Get("last_name").(string))
	}

	if d.HasChange("mailing_address") {
		input.MailingAddress = expandUpdateAddress(d.Get("mailing_address").([]interface{}))
	}

	if d.HasChange("middle_name") {
		input.MiddleName = aws.String(d.Get("middle_name").(string))
	}

	if d.HasChange("mobile_phone_number") {
		input.MobilePhoneNumber = aws.String(d.Get("mobile_phone_number").(string))
	}

	if d.HasChange("party_type_string") {
		input.PartyTypeString = aws.String(d.Get("party_type_string").(string))
	}

	if d.HasChange("personal_email_address") {
		input.PersonalEmailAddress = aws.String(d.Get("personal_email_address").(string))
	}

	if d.HasChange("phone_number") {
		input.PhoneNumber = aws.String(d.Get("phone_number").(string))
	}

	if d.HasChange("shipping_address") {
		input.ShippingAddress = expandUpdateAddress(d.Get("shipping_address").([]interface{}))
	}

	_, err := conn.UpdateProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Customer Profiles Profile (%s): %s", d.Id(), err)
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CustomerProfilesClient(ctx)

	log.Printf("[DEBUG] Deleting Customer Profiles Profile: %s", d.Id())
	_, err := conn.DeleteProfile(ctx, &customerprofiles.DeleteProfileInput{
		DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
		ProfileId:  aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Customer Profiles Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceProfileImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return []*schema.ResourceData{}, fmt.Errorf("unexpected format of import ID (%s), use: 'domain-name/profile-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set(names.AttrDomainName, parts[0])

	return []*schema.ResourceData{d}, nil
}

func FindProfileByTwoPartKey(ctx context.Context, conn *customerprofiles.Client, profileId, domainName string) (*types.Profile, error) {
	input := &customerprofiles.SearchProfilesInput{
		DomainName: aws.String(domainName),
		KeyName:    aws.String("_profileId"),
		Values:     []string{profileId},
	}

	output, err := conn.SearchProfiles(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Items) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Items); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.Items[0], nil
}

func flattenAddress(apiObject *types.Address) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Address1; v != nil {
		tfMap["address_1"] = aws.ToString(v)
	}

	if v := apiObject.Address2; v != nil {
		tfMap["address_2"] = aws.ToString(v)
	}

	if v := apiObject.Address3; v != nil {
		tfMap["address_3"] = aws.ToString(v)
	}

	if v := apiObject.Address4; v != nil {
		tfMap["address_4"] = aws.ToString(v)
	}

	if v := apiObject.City; v != nil {
		tfMap["city"] = aws.ToString(v)
	}

	if v := apiObject.Country; v != nil {
		tfMap["country"] = aws.ToString(v)
	}

	if v := apiObject.County; v != nil {
		tfMap["county"] = aws.ToString(v)
	}

	if v := apiObject.PostalCode; v != nil {
		tfMap["postal_code"] = aws.ToString(v)
	}

	if v := apiObject.Province; v != nil {
		tfMap["province"] = aws.ToString(v)
	}

	if v := apiObject.State; v != nil {
		tfMap[names.AttrState] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func expandAddress(tfMap []interface{}) *types.Address {
	if tfMap == nil || tfMap[0] == nil {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.Address{}

	if v, ok := tfList["address_1"]; ok {
		apiObject.Address1 = aws.String(v.(string))
	}

	if v, ok := tfList["address_2"]; ok {
		apiObject.Address2 = aws.String(v.(string))
	}

	if v, ok := tfList["address_3"]; ok {
		apiObject.Address3 = aws.String(v.(string))
	}

	if v, ok := tfList["address_4"]; ok {
		apiObject.Address4 = aws.String(v.(string))
	}

	if v, ok := tfList["city"]; ok {
		apiObject.City = aws.String(v.(string))
	}

	if v, ok := tfList["country"]; ok {
		apiObject.Country = aws.String(v.(string))
	}

	if v, ok := tfList["county"]; ok {
		apiObject.County = aws.String(v.(string))
	}

	if v, ok := tfList["postal_code"]; ok {
		apiObject.PostalCode = aws.String(v.(string))
	}

	if v, ok := tfList["province"]; ok {
		apiObject.Province = aws.String(v.(string))
	}

	if v, ok := tfList[names.AttrState]; ok {
		apiObject.State = aws.String(v.(string))
	}

	return apiObject
}

func expandUpdateAddress(tfMap []interface{}) *types.UpdateAddress {
	if tfMap == nil || tfMap[0] == nil {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.UpdateAddress{}

	if v, ok := tfList["address_1"]; ok {
		apiObject.Address1 = aws.String(v.(string))
	}

	if v, ok := tfList["address_2"]; ok {
		apiObject.Address2 = aws.String(v.(string))
	}

	if v, ok := tfList["address_3"]; ok {
		apiObject.Address3 = aws.String(v.(string))
	}

	if v, ok := tfList["address_4"]; ok {
		apiObject.Address4 = aws.String(v.(string))
	}

	if v, ok := tfList["city"]; ok {
		apiObject.City = aws.String(v.(string))
	}

	if v, ok := tfList["country"]; ok {
		apiObject.Country = aws.String(v.(string))
	}

	if v, ok := tfList["county"]; ok {
		apiObject.County = aws.String(v.(string))
	}

	if v, ok := tfList["postal_code"]; ok {
		apiObject.PostalCode = aws.String(v.(string))
	}

	if v, ok := tfList["province"]; ok {
		apiObject.Province = aws.String(v.(string))
	}

	if v, ok := tfList[names.AttrState]; ok {
		apiObject.State = aws.String(v.(string))
	}

	return apiObject
}
