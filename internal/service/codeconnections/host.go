// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeconnections

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeconnections_host", name="Host")
func resourceHost() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostCreate,
		ReadWithoutTimeout:   resourceHostRead,
		UpdateWithoutTimeout: resourceHostUpdate,
		DeleteWithoutTimeout: resourceHostDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"provider_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ProviderType](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tls_certificate": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeConnectionsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &codeconnections.CreateHostInput{
		Name:             aws.String(name),
		ProviderEndpoint: aws.String(d.Get("provider_endpoint").(string)),
		ProviderType:     types.ProviderType(d.Get("provider_type").(string)),
		VpcConfiguration: expandHostVPCConfiguration(d.Get(names.AttrVPCConfiguration).([]interface{})),
	}

	output, err := conn.CreateHost(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeConnections Host (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.HostArn))

	if _, err := waitHostPendingOrAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CodeConnections Host (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceHostRead(ctx, d, meta)...)
}

func resourceHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeConnectionsClient(ctx)

	output, err := findHostByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeConnections Host (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeConnections Host (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set(names.AttrName, output.Name)
	d.Set("provider_endpoint", output.ProviderEndpoint)
	d.Set("provider_type", output.ProviderType)
	d.Set(names.AttrStatus, output.Status)
	if err := d.Set(names.AttrVPCConfiguration, flattenHostVPCConfiguration(output.VpcConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_configuration: %s", err)
	}

	return diags
}

func resourceHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeConnectionsClient(ctx)

	if d.HasChanges("provider_endpoint", names.AttrVPCConfiguration) {
		input := &codeconnections.UpdateHostInput{
			HostArn:          aws.String(d.Id()),
			ProviderEndpoint: aws.String(d.Get("provider_endpoint").(string)),
			VpcConfiguration: expandHostVPCConfiguration(d.Get(names.AttrVPCConfiguration).([]interface{})),
		}

		_, err := conn.UpdateHost(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeConnections Host (%s): %s", d.Id(), err)
		}

		if _, err := waitHostPendingOrAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CodeConnections Host (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceHostRead(ctx, d, meta)...)
}

func resourceHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeConnectionsClient(ctx)

	log.Printf("[DEBUG] Deleting CodeConnections Host: %s", d.Id())
	_, err := conn.DeleteHost(ctx, &codeconnections.DeleteHostInput{
		HostArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeConnections Host (%s): %s", d.Id(), err)
	}

	return diags
}

func expandHostVPCConfiguration(l []interface{}) *types.VpcConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	vc := &types.VpcConfiguration{
		SecurityGroupIds: flex.ExpandStringValueSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
		SubnetIds:        flex.ExpandStringValueSet(m[names.AttrSubnetIDs].(*schema.Set)),
		VpcId:            aws.String(m[names.AttrVPCID].(string)),
	}

	if v, ok := m["tls_certificate"].(string); ok && v != "" {
		vc.TlsCertificate = aws.String(v)
	}

	return vc
}

func flattenHostVPCConfiguration(vpcConfig *types.VpcConfiguration) []interface{} {
	if vpcConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrSecurityGroupIDs: vpcConfig.SecurityGroupIds,
		names.AttrSubnetIDs:        vpcConfig.SubnetIds,
		names.AttrVPCID:            aws.ToString(vpcConfig.VpcId),
	}

	if vpcConfig.TlsCertificate != nil {
		m["tls_certificate"] = aws.ToString(vpcConfig.TlsCertificate)
	}

	return []interface{}{m}
}

func findHostByARN(ctx context.Context, conn *codeconnections.Client, arn string) (*codeconnections.GetHostOutput, error) {
	input := &codeconnections.GetHostInput{
		HostArn: aws.String(arn),
	}

	output, err := conn.GetHost(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusHost(ctx context.Context, conn *codeconnections.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findHostByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

const (
	hostStatusAvailable = "AVAILABLE"
	hostStatusPending   = "PENDING"
	// hostStatusVPCConfigDeleting             = "VPC_CONFIG_DELETING"
	// hostStatusVPCConfigFailedInitialization = "VPC_CONFIG_FAILED_INITIALIZATION"
	hostStatusVPCConfigInitializing = "VPC_CONFIG_INITIALIZING"
)

func waitHostPendingOrAvailable(ctx context.Context, conn *codeconnections.Client, arn string, timeout time.Duration) (*codeconnections.GetHostOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{hostStatusVPCConfigInitializing},
		Target:  []string{hostStatusAvailable, hostStatusPending},
		Refresh: statusHost(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*codeconnections.GetHostOutput); ok {
		return output, err
	}

	return nil, err
}
