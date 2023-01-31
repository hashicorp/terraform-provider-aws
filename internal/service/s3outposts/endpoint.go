package s3outposts

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointCreate,
		ReadWithoutTimeout:   resourceEndpointRead,
		DeleteWithoutTimeout: resourceEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceEndpointImportState,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interfaces": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_interface_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"outpost_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"security_group_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsConn()

	input := &s3outposts.CreateEndpointInput{
		OutpostId:       aws.String(d.Get("outpost_id").(string)),
		SecurityGroupId: aws.String(d.Get("security_group_id").(string)),
		SubnetId:        aws.String(d.Get("subnet_id").(string)),
	}

	output, err := conn.CreateEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Outposts Endpoint: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Outposts Endpoint: empty response")
	}

	d.SetId(aws.StringValue(output.EndpointArn))

	if _, err := waitEndpointStatusCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Outposts Endpoint (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceEndpointRead(ctx, d, meta)...)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsConn()

	endpoint, err := FindEndpoint(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Outposts Endpoint (%s): %s", d.Id(), err)
	}

	if endpoint == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading S3 Outposts Endpoint (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] S3 Outposts Endpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", endpoint.EndpointArn)
	d.Set("cidr_block", endpoint.CidrBlock)

	if endpoint.CreationTime != nil {
		d.Set("creation_time", aws.TimeValue(endpoint.CreationTime).Format(time.RFC3339))
	}

	if err := d.Set("network_interfaces", flattenNetworkInterfaces(endpoint.NetworkInterfaces)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_interfaces: %s", err)
	}

	d.Set("outpost_id", endpoint.OutpostsId)

	return diags
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3OutpostsConn()

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Outposts Endpoint ARN (%s): %s", d.Id(), err)
	}

	// ARN resource format: outpost/<outpost-id>/endpoint/<endpoint-id>
	arnResourceParts := strings.Split(parsedArn.Resource, "/")

	if parsedArn.AccountID == "" || len(arnResourceParts) != 4 {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Outposts Endpoint ARN (%s): unknown format", d.Id())
	}

	input := &s3outposts.DeleteEndpointInput{
		EndpointId: aws.String(arnResourceParts[3]),
		OutpostId:  aws.String(arnResourceParts[1]),
	}

	_, err = conn.DeleteEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Outposts Endpoint (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceEndpointImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%s), expected ENDPOINT-ARN,SECURITY-GROUP-ID,SUBNET-ID", d.Id())
	}

	endpointArn := idParts[0]
	securityGroupId := idParts[1]
	subnetId := idParts[2]

	d.SetId(endpointArn)
	d.Set("security_group_id", securityGroupId)
	d.Set("subnet_id", subnetId)

	return []*schema.ResourceData{d}, nil
}

func flattenNetworkInterfaces(apiObjects []*s3outposts.NetworkInterface) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterface(apiObject))
	}

	return tfList
}

func flattenNetworkInterface(apiObject *s3outposts.NetworkInterface) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	return tfMap
}
