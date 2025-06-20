// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func setResponsePlanResourceData(
	d *schema.ResourceData,
	getResponsePlanOutput *ssmincidents.GetResponsePlanOutput,
) (*schema.ResourceData, error) {
	if err := d.Set(names.AttrAction, flattenAction(getResponsePlanOutput.Actions)); err != nil {
		return d, err
	}
	if err := d.Set(names.AttrARN, getResponsePlanOutput.Arn); err != nil {
		return d, err
	}
	if err := d.Set("chat_channel", flattenChatChannel(getResponsePlanOutput.ChatChannel)); err != nil {
		return d, err
	}
	if err := d.Set(names.AttrDisplayName, getResponsePlanOutput.DisplayName); err != nil {
		return d, err
	}
	if err := d.Set("engagements", flex.FlattenStringValueSet(getResponsePlanOutput.Engagements)); err != nil {
		return d, err
	}
	if err := d.Set("incident_template", flattenIncidentTemplate(getResponsePlanOutput.IncidentTemplate)); err != nil {
		return d, err
	}
	if err := d.Set("integration", flattenIntegration(getResponsePlanOutput.Integrations)); err != nil {
		return d, err
	}
	if err := d.Set(names.AttrName, getResponsePlanOutput.Name); err != nil {
		return d, err
	}
	return d, nil
}
