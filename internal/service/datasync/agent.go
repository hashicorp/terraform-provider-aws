// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_agent", name="Agent")
// @Tags(identifierAttribute="id")
func ResourceAgent() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAgentCreate,
		ReadWithoutTimeout:   resourceAgentRead,
		UpdateWithoutTimeout: resourceAgentUpdate,
		DeleteWithoutTimeout: resourceAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"activation_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"private_link_endpoint", names.AttrIPAddress},
			},
			names.AttrIPAddress: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"activation_key"},
			},
			"private_link_endpoint": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"activation_key"},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_group_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAgentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	activationKey := d.Get("activation_key").(string)
	agentIpAddress := d.Get(names.AttrIPAddress).(string)

	// Perform one time fetch of activation key from gateway IP address.
	if activationKey == "" {
		if agentIpAddress == "" {
			return sdkdiag.AppendErrorf(diags, "one of activation_key or ip_address is required")
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 10,
		}
		region := meta.(*conns.AWSClient).Region

		var requestURL string
		if v, ok := d.GetOk("private_link_endpoint"); ok {
			requestURL = fmt.Sprintf("http://%s/?gatewayType=SYNC&activationRegion=%s&endpointType=PRIVATE_LINK&privateLinkEndpoint=%s", agentIpAddress, region, v.(string))
		} else {
			requestURL = fmt.Sprintf("http://%s/?gatewayType=SYNC&activationRegion=%s", agentIpAddress, region)
		}

		request, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating HTTP request: %s", err)
		}

		var response *http.Response
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			response, err = client.Do(request)

			if errs.IsA[net.Error](err) {
				return retry.RetryableError(fmt.Errorf("making HTTP request: %w", err))
			}

			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("making HTTP request: %w", err))
			}

			if response == nil {
				return retry.NonRetryableError(fmt.Errorf("no response for activation key request"))
			}

			log.Printf("[DEBUG] Received HTTP response: %#v", response)
			if expected := http.StatusFound; expected != response.StatusCode {
				return retry.NonRetryableError(fmt.Errorf("expected HTTP status code %d, received: %d", expected, response.StatusCode))
			}

			redirectURL, err := response.Location()
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("extracting HTTP Location header: %w", err))
			}

			if errorType := redirectURL.Query().Get("errorType"); errorType == "PRIVATE_LINK_ENDPOINT_UNREACHABLE" {
				errMessage := fmt.Errorf("during activation: %s", errorType)
				return retry.RetryableError(errMessage)
			}

			activationKey = redirectURL.Query().Get("activationKey")

			return nil
		})

		if tfresource.TimedOut(err) {
			return sdkdiag.AppendErrorf(diags, "timeout retrieving activation key from IP Address (%s): %s", agentIpAddress, err)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "retrieving activation key from IP Address (%s): %s", agentIpAddress, err)
		}

		if activationKey == "" {
			return sdkdiag.AppendErrorf(diags, "empty activationKey received from IP Address: %s", agentIpAddress)
		}
	}

	input := &datasync.CreateAgentInput{
		ActivationKey: aws.String(activationKey),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.AgentName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_arns"); ok {
		input.SecurityGroupArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_arns"); ok {
		input.SubnetArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrVPCEndpointID); ok {
		input.VpcEndpointId = aws.String(v.(string))
	}

	output, err := conn.CreateAgent(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Agent: %s", err)
	}

	d.SetId(aws.ToString(output.AgentArn))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return FindAgentByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DataSync Agent (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAgentRead(ctx, d, meta)...)
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := FindAgentByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Agent (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Agent (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.AgentArn)
	d.Set(names.AttrName, output.Name)
	if plc := output.PrivateLinkConfig; plc != nil {
		d.Set("private_link_endpoint", plc.PrivateLinkEndpoint)
		d.Set("security_group_arns", flex.FlattenStringValueList(plc.SecurityGroupArns))
		d.Set("subnet_arns", flex.FlattenStringValueList(plc.SubnetArns))
		d.Set(names.AttrVPCEndpointID, plc.VpcEndpointId)
	} else {
		d.Set("private_link_endpoint", "")
		d.Set("security_group_arns", nil)
		d.Set("subnet_arns", nil)
		d.Set(names.AttrVPCEndpointID, "")
	}

	return diags
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChange(names.AttrName) {
		input := &datasync.UpdateAgentInput{
			AgentArn: aws.String(d.Id()),
			Name:     aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateAgent(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Agent (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAgentRead(ctx, d, meta)...)
}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Agent: %s", d.Id())
	_, err := conn.DeleteAgent(ctx, &datasync.DeleteAgentInput{
		AgentArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Agent (%s): %s", d.Id(), err)
	}

	return diags
}

func FindAgentByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeAgentOutput, error) {
	input := &datasync.DescribeAgentInput{
		AgentArn: aws.String(arn),
	}

	output, err := conn.DescribeAgent(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
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
