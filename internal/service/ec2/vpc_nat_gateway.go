package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNATGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNATGatewayCreate,
		ReadWithoutTimeout:   resourceNATGatewayRead,
		UpdateWithoutTimeout: resourceNATGatewayUpdate,
		DeleteWithoutTimeout: resourceNATGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"connectivity_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.ConnectivityTypePublic,
				ValidateFunc: validation.StringInSlice(ec2.ConnectivityType_Values(), false),
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNATGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateNatGatewayInput{
		ClientToken:       aws.String(resource.UniqueId()),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNatgateway),
	}

	if v, ok := d.GetOk("allocation_id"); ok {
		input.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connectivity_type"); ok {
		input.ConnectivityType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("private_ip"); ok {
		input.PrivateIpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	output, err := conn.CreateNatGatewayWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating EC2 NAT Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.NatGateway.NatGatewayId))

	if _, err := WaitNATGatewayCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for EC2 NAT Gateway (%s) create: %s", d.Id(), err)
	}

	return resourceNATGatewayRead(ctx, d, meta)
}

func resourceNATGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ng, err := FindNATGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 NAT Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	address := ng.NatGatewayAddresses[0]
	d.Set("allocation_id", address.AllocationId)
	d.Set("connectivity_type", ng.ConnectivityType)
	d.Set("network_interface_id", address.NetworkInterfaceId)
	d.Set("private_ip", address.PrivateIp)
	d.Set("public_ip", address.PublicIp)
	d.Set("subnet_id", ng.SubnetId)

	tags := KeyValueTags(ng.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceNATGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating EC2 NAT Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceNATGatewayRead(ctx, d, meta)
}

func resourceNATGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 NAT Gateway: %s", d.Id())
	_, err := conn.DeleteNatGatewayWithContext(ctx, &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNatGatewayNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	if _, err := WaitNATGatewayDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for EC2 NAT Gateway (%s) delete: %s", d.Id(), err)
	}

	return nil
}
