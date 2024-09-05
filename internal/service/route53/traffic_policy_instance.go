// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_traffic_policy_instance", name="Traffic Policy Instance")
func resourceTrafficPolicyInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficPolicyInstanceCreate,
		ReadWithoutTimeout:   resourceTrafficPolicyInstanceRead,
		UpdateWithoutTimeout: resourceTrafficPolicyInstanceUpdate,
		DeleteWithoutTimeout: resourceTrafficPolicyInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrHostedZoneID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
				StateFunc: func(v interface{}) string {
					value := strings.TrimSuffix(v.(string), ".")
					return strings.ToLower(value)
				},
			},
			"traffic_policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 36),
			},
			"traffic_policy_version": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 1000),
			},
			"ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtMost(2147483647),
			},
		},
	}
}

func resourceTrafficPolicyInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53.CreateTrafficPolicyInstanceInput{
		HostedZoneId:         aws.String(d.Get(names.AttrHostedZoneID).(string)),
		Name:                 aws.String(name),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int32(int32(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.NoSuchTrafficPolicy](ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateTrafficPolicyInstance(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Traffic Policy Instance (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*route53.CreateTrafficPolicyInstanceOutput).TrafficPolicyInstance.Id))

	if _, err = waitTrafficPolicyInstanceStateCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Traffic Policy Instance (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTrafficPolicyInstanceRead(ctx, d, meta)...)
}

func resourceTrafficPolicyInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	trafficPolicyInstance, err := findTrafficPolicyInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Traffic Policy Instance %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Traffic Policy Instance (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrHostedZoneID, trafficPolicyInstance.HostedZoneId)
	d.Set(names.AttrName, strings.TrimSuffix(aws.ToString(trafficPolicyInstance.Name), "."))
	d.Set("traffic_policy_id", trafficPolicyInstance.TrafficPolicyId)
	d.Set("traffic_policy_version", trafficPolicyInstance.TrafficPolicyVersion)
	d.Set("ttl", trafficPolicyInstance.TTL)

	return diags
}

func resourceTrafficPolicyInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	input := &route53.UpdateTrafficPolicyInstanceInput{
		Id:                   aws.String(d.Id()),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int32(int32(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	_, err := conn.UpdateTrafficPolicyInstance(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Traffic Policy Instance (%s): %s", d.Id(), err)
	}

	if _, err = waitTrafficPolicyInstanceStateUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Traffic Policy Instance (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceTrafficPolicyInstanceRead(ctx, d, meta)...)
}

func resourceTrafficPolicyInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	log.Printf("[INFO] Deleting Route53 Traffic Policy Instance: %s", d.Id())
	_, err := conn.DeleteTrafficPolicyInstance(ctx, &route53.DeleteTrafficPolicyInstanceInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchTrafficPolicyInstance](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Traffic Policy Instance (%s): %s", d.Id(), err)
	}

	if _, err = waitTrafficPolicyInstanceStateDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Traffic Policy Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findTrafficPolicyInstanceByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.TrafficPolicyInstance, error) {
	input := &route53.GetTrafficPolicyInstanceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetTrafficPolicyInstance(ctx, input)

	if errs.IsA[*awstypes.NoSuchTrafficPolicyInstance](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TrafficPolicyInstance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TrafficPolicyInstance, nil
}

func statusTrafficPolicyInstanceState(ctx context.Context, conn *route53.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTrafficPolicyInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

const (
	trafficPolicyInstanceStateApplied  = "Applied"
	trafficPolicyInstanceStateCreating = "Creating"
	trafficPolicyInstanceStateDeleting = "Deleting"
	trafficPolicyInstanceStateFailed   = "Failed"
	trafficPolicyInstanceStateUpdating = "Updating"
)

const (
	trafficPolicyInstanceOperationTimeout = 4 * time.Minute
)

func waitTrafficPolicyInstanceStateCreated(ctx context.Context, conn *route53.Client, id string) (*awstypes.TrafficPolicyInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{trafficPolicyInstanceStateCreating},
		Target:  []string{trafficPolicyInstanceStateApplied},
		Refresh: statusTrafficPolicyInstanceState(ctx, conn, id),
		Timeout: trafficPolicyInstanceOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TrafficPolicyInstance); ok {
		if state := aws.ToString(output.State); state == trafficPolicyInstanceStateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTrafficPolicyInstanceStateUpdated(ctx context.Context, conn *route53.Client, id string) (*awstypes.TrafficPolicyInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{trafficPolicyInstanceStateUpdating},
		Target:  []string{trafficPolicyInstanceStateApplied},
		Refresh: statusTrafficPolicyInstanceState(ctx, conn, id),
		Timeout: trafficPolicyInstanceOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TrafficPolicyInstance); ok {
		if state := aws.ToString(output.State); state == trafficPolicyInstanceStateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTrafficPolicyInstanceStateDeleted(ctx context.Context, conn *route53.Client, id string) (*awstypes.TrafficPolicyInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{trafficPolicyInstanceStateDeleting},
		Target:  []string{},
		Refresh: statusTrafficPolicyInstanceState(ctx, conn, id),
		Timeout: trafficPolicyInstanceOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TrafficPolicyInstance); ok {
		if state := aws.ToString(output.State); state == trafficPolicyInstanceStateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Message)))
		}

		return output, err
	}

	return nil, err
}
