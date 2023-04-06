package vpclattice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service_networkservice_association")
// @Tags(identifierAttribute="arn")
func ResourceServiceNetworkAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceNetworkAssociationCreate,
		ReadWithoutTimeout:   resourceServiceNetworkAssociationRead,
		UpdateWithoutTimeout: resourceServiceNetworkAssociationUpdate,
		DeleteWithoutTimeout: resourceServiceNetworkAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_entry": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"service_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_network_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameServiceNetworkAssociation = "ServiceNetworkAssociation"
)

func resourceServiceNetworkAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	in := &vpclattice.CreateServiceNetworkServiceAssociationInput{
		ClientToken:              aws.String(id.UniqueId()),
		ServiceIdentifier:        aws.String(d.Get("service_identifier").(string)),
		ServiceNetworkIdentifier: aws.String(d.Get("service_network_identifier").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateServiceNetworkServiceAssociation(ctx, in)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkAssociation, "", err)
	}

	if out == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameServiceNetworkAssociation, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitServiceNetworkAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	return resourceServiceNetworkAssociationRead(ctx, d, meta)
}

func resourceServiceNetworkAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	out, err := findServiceNetworkAscByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("created_by", out.CreatedBy)
	d.Set("custom_domain_name", out.CustomDomainName)
	d.Set("status", out.Status)
	d.Set("service_identifier", out.ServiceId)
	d.Set("service_network_identifier", out.ServiceNetworkId)

	if err := d.Set("dns_entry", flattenServiceDNSEntry(out.DnsEntry)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, *out.Arn)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	return nil
}

func resourceServiceNetworkAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	// Tags only.
	return resourceServiceNetworkAssociationRead(ctx, d, meta)
}

func resourceServiceNetworkAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	log.Printf("[INFO] Deleting VPCLattice Service Network Association %s", d.Id())

	_, err := conn.DeleteServiceNetworkServiceAssociation(ctx, &vpclattice.DeleteServiceNetworkServiceAssociationInput{
		ServiceNetworkServiceAssociationIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.VPCLattice, create.ErrActionDeleting, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	if _, err := waitServiceNetworkAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameServiceNetworkAssociation, d.Id(), err)
	}

	return nil
}

func findServiceNetworkAscByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceNetworkServiceAssociationOutput, error) {
	in := &vpclattice.GetServiceNetworkServiceAssociationInput{
		ServiceNetworkServiceAssociationIdentifier: aws.String(id),
	}
	out, err := conn.GetServiceNetworkServiceAssociation(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func waitServiceNetworkAssociationCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkServiceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ServiceNetworkVpcAssociationStatusCreateInProgress),
		Target:                    enum.Slice(types.ServiceNetworkVpcAssociationStatusActive),
		Refresh:                   statusServiceNetworkAssociation(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkServiceAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitServiceNetworkAssociationDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceNetworkServiceAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceNetworkVpcAssociationStatusDeleteInProgress, types.ServiceNetworkVpcAssociationStatusActive),
		Target:  []string{},
		Refresh: statusServiceNetworkAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceNetworkServiceAssociationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusServiceNetworkAssociation(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findServiceNetworkAscByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func flattenServiceDNSEntry(apiObject *types.DnsEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.DomainName; v != nil {
		m["domain_name"] = aws.ToString(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		m["hosted_zone_id"] = aws.ToString(v)
	}

	return m
}
