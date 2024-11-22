// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_iot_thing_group_membership", name="Thing Group Membership")
func resourceThingGroupMembership() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingGroupMembershipCreate,
		ReadWithoutTimeout:   resourceThingGroupMembershipRead,
		DeleteWithoutTimeout: resourceThingGroupMembershipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"override_dynamic_group": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"thing_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"thing_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceThingGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	thingGroupName := d.Get("thing_group_name").(string)
	thingName := d.Get("thing_name").(string)
	input := &iot.AddThingToThingGroupInput{
		ThingGroupName: aws.String(thingGroupName),
		ThingName:      aws.String(thingName),
	}

	if v, ok := d.GetOk("override_dynamic_group"); ok {
		input.OverrideDynamicGroups = v.(bool)
	}

	_, err := conn.AddThingToThingGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding IoT Thing (%s) to IoT Thing Group (%s): %s", thingName, thingGroupName, err)
	}

	d.SetId(thingGroupMembershipCreateResourceID(thingGroupName, thingName))

	return append(diags, resourceThingGroupMembershipRead(ctx, d, meta)...)
}

func resourceThingGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	thingGroupName, thingName, err := thingGroupMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findThingGroupMembershipByTwoPartKey(ctx, conn, thingGroupName, thingName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing Group Membership (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Group Membership (%s): %s", d.Id(), err)
	}

	d.Set("thing_group_name", thingGroupName)
	d.Set("thing_name", thingName)

	return diags
}

func resourceThingGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	thingGroupName, thingName, err := thingGroupMembershipParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IoT Thing Group Membership: %s", d.Id())
	_, err = conn.RemoveThingFromThingGroup(ctx, &iot.RemoveThingFromThingGroupInput{
		ThingGroupName: aws.String(thingGroupName),
		ThingName:      aws.String(thingName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Group Membership (%s): %s", d.Id(), err)
	}

	return diags
}

func findThingGroupMembershipByTwoPartKey(ctx context.Context, conn *iot.Client, thingGroupName, thingName string) (*awstypes.GroupNameAndArn, error) {
	input := &iot.ListThingGroupsForThingInput{
		ThingName: aws.String(thingName),
	}

	return findThingGroup(ctx, conn, input, func(v *awstypes.GroupNameAndArn) bool {
		return aws.ToString(v.GroupName) == thingGroupName
	})
}

func findThingGroup(ctx context.Context, conn *iot.Client, input *iot.ListThingGroupsForThingInput, filter tfslices.Predicate[*awstypes.GroupNameAndArn]) (*awstypes.GroupNameAndArn, error) {
	output, err := findThingGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findThingGroups(ctx context.Context, conn *iot.Client, input *iot.ListThingGroupsForThingInput, filter tfslices.Predicate[*awstypes.GroupNameAndArn]) ([]awstypes.GroupNameAndArn, error) {
	var output []awstypes.GroupNameAndArn

	pages := iot.NewListThingGroupsForThingPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ThingGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

const thingGroupMembershipResourceIDSeparator = "/"

func thingGroupMembershipCreateResourceID(thingGroupName, thingName string) string {
	parts := []string{thingGroupName, thingName}
	id := strings.Join(parts, thingGroupMembershipResourceIDSeparator)

	return id
}

func thingGroupMembershipParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, thingGroupMembershipResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected thing-group-name%[2]sthing-name", id, thingGroupMembershipResourceIDSeparator)
}
