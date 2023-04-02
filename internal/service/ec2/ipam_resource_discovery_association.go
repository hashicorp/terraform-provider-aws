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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_vpc_ipam_resource_discovery_association")
func ResourceIPAMResourceDiscoveryAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMResourceDiscoveryAssociationCreate,
		ReadWithoutTimeout:   resourceIPAMResourceDiscoveryAssociationRead,
		UpdateWithoutTimeout: resourceIPAMResourceDiscoveryAssociationUpdate,
		DeleteWithoutTimeout: resourceIPAMResourceDiscoveryAssociationDelete,

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

func resourceIPAMResourceDiscoveryAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	ipamID := d.Get("ipam_id").(string)
	ipamResourceDiscoveryID := d.Get("ipam_resource_discovery_id").(string)
	input := &ec2.AssociateIpamResourceDiscoveryInput{
		ClientToken:             aws.String(id.UniqueId()),
		IpamId:                  aws.String(ipamID),
		IpamResourceDiscoveryId: aws.String(ipamResourceDiscoveryID),
		TagSpecifications:       tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeIpamResourceDiscoveryAssociation),
	}

	output, err := conn.AssociateIpamResourceDiscoveryWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM (%s) Resource Discovery (%s) Association: %s", ipamID, ipamResourceDiscoveryID, err)
	}

	d.SetId(aws.StringValue(output.IpamResourceDiscoveryAssociation.IpamResourceDiscoveryAssociationId))

	if _, err := WaitIPAMResourceDiscoveryAssociationAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMResourceDiscoveryAssociationRead(ctx, d, meta)...)
}

func resourceIPAMResourceDiscoveryAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rda, err := FindIPAMResourceDiscoveryAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Resource Discovery Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Resource Discovery Association (%s): %s", d.Id(), err)
	}

	d.Set("arn", rda.IpamResourceDiscoveryAssociationArn)
	d.Set("ipam_arn", rda.IpamArn)
	d.Set("ipam_id", rda.IpamId)
	d.Set("ipam_region", rda.IpamRegion)
	d.Set("ipam_resource_discovery_id", rda.IpamResourceDiscoveryId)
	d.Set("is_default", rda.IsDefault)
	d.Set("owner_id", rda.OwnerId)
	d.Set("state", rda.State)

	tags := KeyValueTags(ctx, rda.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return nil
}

func resourceIPAMResourceDiscoveryAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM Resource Discovery Association (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceIPAMResourceDiscoveryAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[DEBUG] Deleting IPAM Resource Discovery Association: %s", d.Id())
	_, err := conn.DisassociateIpamResourceDiscoveryWithContext(ctx, &ec2.DisassociateIpamResourceDiscoveryInput{
		IpamResourceDiscoveryAssociationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryAssociationIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Resource Discovery Association (%s): %s", d.Id(), err)
	}

	if _, err := WaitIPAMResourceDiscoveryAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Resource Discovery Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}
