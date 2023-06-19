package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ec2_instance_connect_endpoint")
func ResourceInstanceConnectEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceConnectEndpointCreate,
		ReadWithoutTimeout:   resourceInstanceConnectEndpointRead,
		UpdateWithoutTimeout: resourceInstanceConnectEndpointUpdate,
		DeleteWithoutTimeout: resourceInstanceConnectEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"preserve_client_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Default:  true,
			},
		},
	}
}

func resourceInstanceConnectEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateInstanceConnectEndpointInput{}

	output, err := conn.CreateInstanceConnectEndpointWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "importing EC2 Key Pair (%s): %s", keyName, err)
	}

	d.SetId(aws.StringValue(output.KeyName))

	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	InstanceConnectEndpoint, err := FindInstanceConnectEndpointByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Key Pair (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Key Pair (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("key-pair/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("fingerprint", InstanceConnectEndpoint.KeyFingerprint)
	d.Set("key_name", InstanceConnectEndpoint.KeyName)
	d.Set("key_name_prefix", create.NamePrefixFromName(aws.StringValue(InstanceConnectEndpoint.KeyName)))
	d.Set("key_type", InstanceConnectEndpoint.KeyType)
	d.Set("key_pair_id", InstanceConnectEndpoint.InstanceConnectEndpointId)

	setTagsOut(ctx, InstanceConnectEndpoint.Tags)

	return diags
}

func resourceInstanceConnectEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceInstanceConnectEndpointRead(ctx, d, meta)...)
}

func resourceInstanceConnectEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Key Pair: %s", d.Id())
	_, err := conn.DeleteInstanceConnectEndpointWithContext(ctx, &ec2.DeleteInstanceConnectEndpointInput{
		KeyName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Key Pair (%s): %s", d.Id(), err)
	}

	return diags
}
