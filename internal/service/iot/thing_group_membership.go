package iot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceThingGroupMembership() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).IoTConn()

	thingGroupName := d.Get("thing_group_name").(string)
	thingName := d.Get("thing_name").(string)
	input := &iot.AddThingToThingGroupInput{
		ThingGroupName: aws.String(thingGroupName),
		ThingName:      aws.String(thingName),
	}

	if v, ok := d.GetOk("override_dynamic_group"); ok {
		input.OverrideDynamicGroups = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating IoT Thing Group Membership: %s", input)
	_, err := conn.AddThingToThingGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding IoT Thing (%s) to IoT Thing Group (%s): %s", thingName, thingGroupName, err)
	}

	d.SetId(ThingGroupMembershipCreateResourceID(thingGroupName, thingName))

	return append(diags, resourceThingGroupMembershipRead(ctx, d, meta)...)
}

func resourceThingGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	thingGroupName, thingName, err := ThingGroupMembershipParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Group Membership (%s): %s", d.Id(), err)
	}

	err = FindThingGroupMembership(ctx, conn, thingGroupName, thingName)

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
	conn := meta.(*conns.AWSClient).IoTConn()

	thingGroupName, thingName, err := ThingGroupMembershipParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Group Membership (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting IoT Thing Group Membership: %s", d.Id())
	_, err = conn.RemoveThingFromThingGroupWithContext(ctx, &iot.RemoveThingFromThingGroupInput{
		ThingGroupName: aws.String(thingGroupName),
		ThingName:      aws.String(thingName),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Group Membership (%s): %s", d.Id(), err)
	}

	return diags
}

const thingGroupMembershipResourceIDSeparator = "/"

func ThingGroupMembershipCreateResourceID(thingGroupName, thingName string) string {
	parts := []string{thingGroupName, thingName}
	id := strings.Join(parts, thingGroupMembershipResourceIDSeparator)

	return id
}

func ThingGroupMembershipParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, thingGroupMembershipResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected thing-group-name%[2]sthing-name", id, thingGroupMembershipResourceIDSeparator)
}
