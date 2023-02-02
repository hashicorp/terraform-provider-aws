package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMResourceDiscoveryAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceIPAMResourceDiscoveryAssociationCreate,
		ReadWithoutTimeout:   ResourceIPAMResourceDiscoveryAssociationRead,
		UpdateWithoutTimeout: ResourceIPAMResourceDiscoveryAssociationUpdate,
		DeleteWithoutTimeout: ResourceIPAMResourceDiscoveryAssociationDelete,

		CustomizeDiff: customdiff.Sequence(verify.SetTagsDiff),

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_resource_discovery_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func ResourceIPAMResourceDiscoveryAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.AssociateIpamResourceDiscoveryInput{
		ClientToken:             aws.String(resource.UniqueId()),
		TagSpecifications:       tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeIpamResourceDiscoveryAssociation),
		IpamId:                  aws.String(d.Get("ipam_id").(string)),
		IpamResourceDiscoveryId: aws.String(d.Get("ipam_resource_discovery_id").(string)),
	}

	log.Printf("[DEBUG] Creating IPAM Resource Discovery Association: %s", input)
	output, err := conn.AssociateIpamResourceDiscoveryWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error associating ipam resource discovery: %s", err)
	}
	d.SetId(aws.StringValue(output.IpamResourceDiscoveryAssociation.IpamResourceDiscoveryAssociationId))
	log.Printf("[INFO] IPAM Resource Discovery Association ID: %s", d.Id())

	if _, err = WaitIPAMResourceDiscoveryAssociationAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "error waiting for IPAM Resource Discovery Association (%s) to be Available: %s", d.Id(), err)
	}

	return append(diags, ResourceIPAMResourceDiscoveryAssociationRead(ctx, d, meta)...)
}

func ResourceIPAMResourceDiscoveryAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rda, err := FindIPAMResourceDiscoveryAssociationById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Resource Discovery Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", rda.IpamResourceDiscoveryAssociationArn)
	d.Set("owner_id", rda.OwnerId)
	d.Set("ipam_arn", rda.IpamArn)
	d.Set("ipam_region", rda.IpamRegion)
	d.Set("ipam_id", rda.IpamId)
	d.Set("state", rda.State)
	d.Set("ipam_resource_discovery_id", rda.IpamResourceDiscoveryId)
	d.Set("is_default", rda.IsDefault)

	tags := KeyValueTags(rda.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "error setting tags_all: %s", err)
	}

	return nil
}

func ResourceIPAMResourceDiscoveryAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "error updating IPAM ResourceDiscovery Association (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func ResourceIPAMResourceDiscoveryAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.DisassociateIpamResourceDiscoveryInput{
		IpamResourceDiscoveryAssociationId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Disassociating IPAM Resource Discovery: %s", d.Id())
	_, err := conn.DisassociateIpamResourceDiscoveryWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "error disassociating IPAM Resource Discovery: (%s): %s", d.Id(), err)
	}

	if _, err = WaiterIPAMResourceDiscoveryAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryAssociationIDNotFound) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "error waiting for IPAM Resource Discovery Association (%s) to be dissociated: %s", d.Id(), err)
	}

	return diags
}
