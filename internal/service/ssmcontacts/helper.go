// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setContactResourceData(d *schema.ResourceData, getContactOutput *ssmcontacts.GetContactOutput) error { //nolint:unparam
	d.Set("arn", getContactOutput.ContactArn)
	d.Set("alias", getContactOutput.Alias)
	d.Set("type", getContactOutput.Type)
	d.Set("display_name", getContactOutput.DisplayName)

	return nil
}

func setContactChannelResourceData(d *schema.ResourceData, out *ssmcontacts.GetContactChannelOutput) error {
	d.Set("activation_status", out.ActivationStatus)
	d.Set("arn", out.ContactChannelArn)
	d.Set("contact_id", out.ContactArn)
	if err := d.Set("delivery_address", flattenContactChannelAddress(out.DeliveryAddress)); err != nil {
		return fmt.Errorf("setting delivery_address: %w", err)
	}
	d.Set("name", out.Name)
	d.Set("type", out.Type)

	return nil
}

func setPlanResourceData(d *schema.ResourceData, getContactOutput *ssmcontacts.GetContactOutput) error {
	d.Set("contact_id", getContactOutput.ContactArn)
	if err := d.Set("stage", flattenStages(getContactOutput.Plan.Stages)); err != nil {
		return fmt.Errorf("setting stage: %w", err)
	}

	return nil
}
