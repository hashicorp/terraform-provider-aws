package ssmcontacts

import (
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setContactResourceData(d *schema.ResourceData, getContactOutput *ssmcontacts.GetContactOutput) error {
	if err := d.Set("arn", getContactOutput.ContactArn); err != nil {
		return err
	}
	if err := d.Set("alias", getContactOutput.Alias); err != nil {
		return err
	}
	if err := d.Set("type", getContactOutput.Type); err != nil {
		return err
	}
	if err := d.Set("display_name", getContactOutput.DisplayName); err != nil {
		return err
	}
	return nil
}

func setContactChannelResourceData(d *schema.ResourceData, out *ssmcontacts.GetContactChannelOutput) error {
	d.Set("activation_status", out.ActivationStatus)
	d.Set("arn", out.ContactChannelArn)
	d.Set("contact_id", out.ContactArn)
	d.Set("name", out.Name)
	d.Set("type", out.Type)
	if err := d.Set("delivery_address", flattenContactChannelAddress(out.DeliveryAddress)); err != nil {
		return err
	}
	return nil
}

func setPlanResourceData(d *schema.ResourceData, getContactOutput *ssmcontacts.GetContactOutput) error {
	if err := d.Set("contact_id", getContactOutput.ContactArn); err != nil {
		return err
	}
	if err := d.Set("stage", flattenStages(getContactOutput.Plan.Stages)); err != nil {
		return err
	}
	return nil
}
