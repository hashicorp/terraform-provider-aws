// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_zone", name="Hosted Zone")
// @Tags(identifierAttribute="zone_id", resourceType="hostedzone")
func resourceZone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceZoneCreate,
		ReadWithoutTimeout:   resourceZoneRead,
		UpdateWithoutTimeout: resourceZoneUpdate,
		DeleteWithoutTimeout: resourceZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Managed by Terraform",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"delegation_set_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vpc"},
				ValidateFunc:  validation.StringLenBetween(0, 32),
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				// AWS Provider 3.0.0 - trailing period removed from name
				// returned from API, no longer requiring custom DiffSuppressFunc;
				// instead a StateFunc allows input to be provided
				// with or without the trailing period
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				StateFunc:    normalizeZoneName,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"primary_name_server": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc": {
				Type:          schema.TypeSet,
				Optional:      true,
				MinItems:      1,
				ConflictsWith: []string{"delegation_set_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrVPCID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"vpc_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Set: sdkv2.SimpleSchemaSetFunc(names.AttrVPCID),
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(id.UniqueId()),
		Name:            aws.String(name),
		HostedZoneConfig: &awstypes.HostedZoneConfig{
			Comment: aws.String(d.Get(names.AttrComment).(string)),
		},
	}

	if v, ok := d.GetOk("delegation_set_id"); ok {
		input.DelegationSetId = aws.String(v.(string))
	}

	// Private Route53 Hosted Zones can only be created with their first VPC association,
	// however we need to associate the remaining after creation.
	vpcs := expandVPCs(d.Get("vpc").(*schema.Set).List(), meta.(*conns.AWSClient).Region)
	if len(vpcs) > 0 {
		input.VPC = vpcs[0]
	}

	output, err := conn.CreateHostedZone(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Hosted Zone (%s): %s", name, err)
	}

	d.SetId(cleanZoneID(aws.ToString(output.HostedZone.Id)))

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone (%s) synchronize: %s", d.Id(), err)
		}
	}

	if err := createTags(ctx, conn, d.Id(), string(awstypes.TagResourceTypeHostedzone), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Zone (%s) tags: %s", d.Id(), err)
	}

	// Associate additional VPCs beyond the first.
	if len(vpcs) > 1 {
		for _, v := range vpcs[1:] {
			if err := hostedZoneAssociateVPC(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceZoneRead(ctx, d, meta)...)
}

func resourceZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	output, err := findHostedZoneByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Hosted Zone %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	zoneID := cleanZoneID(aws.ToString(output.HostedZone.Id))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  "hostedzone/" + zoneID,
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrComment, "")
	d.Set("delegation_set_id", "")
	// To be consistent with other AWS services (e.g. ACM) that do not accept a trailing period,
	// we remove the suffix from the Hosted Zone Name returned from the API.
	d.Set(names.AttrName, normalizeZoneName(aws.ToString(output.HostedZone.Name)))
	d.Set("zone_id", zoneID)

	var nameServers []string

	if output.DelegationSet != nil {
		d.Set("delegation_set_id", cleanDelegationSetID(aws.ToString(output.DelegationSet.Id)))

		nameServers = output.DelegationSet.NameServers
	}

	if output.HostedZone.Config != nil {
		d.Set(names.AttrComment, output.HostedZone.Config.Comment)

		if output.HostedZone.Config.PrivateZone {
			nameServers, err = findNameServersByZone(ctx, conn, d.Id(), d.Get(names.AttrName).(string))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Route53 Hosted Zone (%s) name servers: %s", d.Id(), err)
			}
		}
	}

	d.Set("primary_name_server", nameServers[0])
	sort.Strings(nameServers)
	d.Set("name_servers", nameServers)
	if err := d.Set("vpc", flattenVPCs(output.VPCs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc: %s", err)
	}

	return diags
}

func resourceZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	if d.HasChange(names.AttrComment) {
		input := route53.UpdateHostedZoneCommentInput{
			Comment: aws.String(d.Get(names.AttrComment).(string)),
			Id:      aws.String(d.Id()),
		}

		_, err := conn.UpdateHostedZoneComment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Hosted Zone (%s) comment: %s", d.Id(), err)
		}
	}

	if d.HasChange("vpc") {
		region := meta.(*conns.AWSClient).Region
		o, n := d.GetChange("vpc")
		os, ns := o.(*schema.Set), n.(*schema.Set)

		// VPCs cannot be empty, so add first and then remove.
		for _, tfMapRaw := range ns.Difference(os).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			if err := hostedZoneAssociateVPC(ctx, conn, d.Id(), expandVPC(tfMap, region)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		for _, tfMapRaw := range os.Difference(ns).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			if err := hostedZoneDisassociateVPC(ctx, conn, d.Id(), expandVPC(tfMap, region)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceZoneRead(ctx, d, meta)...)
}

func resourceZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	if d.Get(names.AttrForceDestroy).(bool) {
		if err := deleteAllResourceRecordsFromHostedZone(ctx, conn, d.Id(), d.Get(names.AttrName).(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		hostedZoneDNSSEC, err := findHostedZoneDNSSECByZoneID(ctx, conn, d.Id())

		switch {
		case tfresource.NotFound(err),
			errs.IsAErrorMessageContains[*awstypes.InvalidArgument](err, "Operation is unsupported for private"),
			tfawserr.ErrMessageContains(err, errCodeAccessDenied, "The operation GetDNSSEC is not available for the current AWS account"):
		case err != nil:
			return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone DNSSEC (%s): %s", d.Id(), err)
		case aws.ToString(hostedZoneDNSSEC.Status.ServeSignature) == serveSignatureSigning:
			err := hostedZoneDNSSECDisable(ctx, conn, d.Id())

			switch {
			case errs.IsA[*awstypes.DNSSECNotFound](err), errs.IsA[*awstypes.NoSuchHostedZone](err), errs.IsAErrorMessageContains[*awstypes.InvalidArgument](err, "Operation is unsupported for private"):
			case err != nil:
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting Route53 Hosted Zone: %s", d.Id())
	output, err := conn.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchHostedZone](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Hosted Zone (%s) synchronize: %s", d.Id(), err)
		}
	}

	return diags
}

func findHostedZoneByID(ctx context.Context, conn *route53.Client, id string) (*route53.GetHostedZoneOutput, error) {
	input := &route53.GetHostedZoneInput{
		Id: aws.String(id),
	}

	output, err := conn.GetHostedZone(ctx, input)

	if errs.IsA[*awstypes.NoSuchHostedZone](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.HostedZone == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func deleteAllResourceRecordsFromHostedZone(ctx context.Context, conn *route53.Client, hostedZoneID, hostedZoneName string) error {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	resourceRecordSets, err := findResourceRecordSets(ctx, conn, input, tfslices.PredicateTrue[*route53.ListResourceRecordSetsOutput](), func(v *awstypes.ResourceRecordSet) bool {
		// Zone NS & SOA records cannot be deleted.
		if normalizeZoneName(v.Name) == normalizeZoneName(hostedZoneName) && (v.Type == awstypes.RRTypeNs || v.Type == awstypes.RRTypeSoa) {
			return false
		}
		return true
	})

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Route53 Hosted Zone (%s) resource record sets: %w", hostedZoneID, err)
	}

	const (
		chunkSize = 100
	)
	chunks := tfslices.Chunks(resourceRecordSets, chunkSize)
	for _, chunk := range chunks {
		changes := tfslices.ApplyToAll(chunk, func(v awstypes.ResourceRecordSet) awstypes.Change {
			return awstypes.Change{
				Action:            awstypes.ChangeActionDelete,
				ResourceRecordSet: &v,
			}
		})
		input := &route53.ChangeResourceRecordSetsInput{
			ChangeBatch: &awstypes.ChangeBatch{
				Changes: changes,
				Comment: aws.String("Deleted by Terraform"),
			},
			HostedZoneId: aws.String(hostedZoneID),
		}

		output, err := conn.ChangeResourceRecordSets(ctx, input)

		if v, ok := errs.As[*awstypes.InvalidChangeBatch](err); ok && len(v.Messages) > 0 {
			err = fmt.Errorf("%s: %w", v.ErrorCode(), errors.Join(tfslices.ApplyToAll(v.Messages, errors.New)...))
		}

		if err != nil {
			return fmt.Errorf("deleting Route53 Hosted Zone (%s) resource record sets: %w", hostedZoneID, err)
		}

		if output.ChangeInfo != nil {
			if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
				return fmt.Errorf("waiting for Route 53 Hosted Zone (%s) synchronize: %w", hostedZoneID, err)
			}
		}
	}

	return nil
}

func findNameServersByZone(ctx context.Context, conn *route53.Client, zoneID, zoneName string) ([]string, error) {
	rrType := awstypes.RRTypeNs
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneID),
		StartRecordName: aws.String(zoneName),
		StartRecordType: rrType,
	}
	output, err := findResourceRecordSets(ctx, conn, input, resourceRecordsFor(zoneName, rrType), tfslices.PredicateTrue[*awstypes.ResourceRecordSet]())

	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}
	records := output[0].ResourceRecords

	ns := tfslices.ApplyToAll(records, func(v awstypes.ResourceRecord) string {
		return aws.ToString(v.Value)
	})
	slices.Sort(ns)

	return ns, nil
}

func hostedZoneAssociateVPC(ctx context.Context, conn *route53.Client, zoneID string, vpc *awstypes.VPC) error {
	input := &route53.AssociateVPCWithHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC:          vpc,
	}

	output, err := conn.AssociateVPCWithHostedZone(ctx, input)

	if err != nil {
		return fmt.Errorf("associating Route53 Hosted Zone (%s) with VPC (%s): %w", zoneID, aws.ToString(vpc.VPCId), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for Route 53 Hosted Zone (%s) synchronize: %w", zoneID, err)
		}
	}

	return nil
}

func hostedZoneDisassociateVPC(ctx context.Context, conn *route53.Client, zoneID string, vpc *awstypes.VPC) error {
	input := &route53.DisassociateVPCFromHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC:          vpc,
	}

	output, err := conn.DisassociateVPCFromHostedZone(ctx, input)

	if err != nil {
		return fmt.Errorf("disassociating Route53 Hosted Zone (%s) from VPC (%s): %w", zoneID, aws.ToString(vpc.VPCId), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for Route 53 Hosted Zone (%s) synchronize: %w", zoneID, err)
		}
	}

	return nil
}

func expandVPCs(tfList []interface{}, currentRegion string) []*awstypes.VPC {
	apiObjects := []*awstypes.VPC{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandVPC(tfMap, currentRegion))
	}

	return apiObjects
}

func expandVPC(tfMap map[string]interface{}, currentRegion string) *awstypes.VPC {
	apiObject := &awstypes.VPC{
		VPCId:     aws.String(tfMap[names.AttrVPCID].(string)),
		VPCRegion: awstypes.VPCRegion(currentRegion),
	}

	if v, ok := tfMap["vpc_region"]; ok && v.(string) != "" {
		apiObject.VPCRegion = awstypes.VPCRegion(v.(string))
	}

	return apiObject
}

func flattenVPCs(apiObjects []awstypes.VPC) []interface{} {
	tfList := []interface{}{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrVPCID: aws.ToString(apiObject.VPCId),
			"vpc_region":    apiObject.VPCRegion,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
