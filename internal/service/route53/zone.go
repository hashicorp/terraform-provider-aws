// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	zoneChangeSyncMinDelay        = 10
	zoneChangeSyncMaxDelay        = 30
	zoneChangeSyncMinPollInterval = 15
	zoneChangeSyncMaxPollInterval = 30
)

// @SDKResource("aws_route53_zone", name="Hosted Zone")
// @Tags(identifierAttribute="id", resourceType="hostedzone")
func ResourceZone() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceZoneCreate,
		ReadWithoutTimeout:   resourceZoneRead,
		UpdateWithoutTimeout: resourceZoneUpdate,
		DeleteWithoutTimeout: resourceZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
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
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name": {
				// AWS Provider 3.0.0 - trailing period removed from name
				// returned from API, no longer requiring custom DiffSuppressFunc;
				// instead a StateFunc allows input to be provided
				// with or without the trailing period
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				StateFunc:    TrimTrailingPeriod,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name_servers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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
						"vpc_id": {
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
				Set: hostedZoneVPCHash,
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
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	input := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(id.UniqueId()),
		Name:            aws.String(d.Get("name").(string)),
		HostedZoneConfig: &route53.HostedZoneConfig{
			Comment: aws.String(d.Get("comment").(string)),
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

	output, err := conn.CreateHostedZoneWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Hosted Zone: %s", err)
	}

	d.SetId(CleanZoneID(aws.StringValue(output.HostedZone.Id)))

	if output.ChangeInfo != nil {
		if err := waitForChangeSynchronization(ctx, conn, CleanChangeID(aws.StringValue(output.ChangeInfo.Id))); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route53 Hosted Zone (%s) create: %s", d.Id(), err)
		}
	}

	if err := createTags(ctx, conn, d.Id(), route53.TagResourceTypeHostedzone, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Zone (%s) tags: %s", d.Id(), err)
	}

	// Associate additional VPCs beyond the first
	if len(vpcs) > 1 {
		for _, vpc := range vpcs[1:] {
			err := hostedZoneAssociateVPC(ctx, conn, d.Id(), vpc)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceZoneRead(ctx, d, meta)...)
}

func resourceZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	output, err := FindHostedZoneByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Hosted Zone %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("hostedzone/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("comment", "")
	d.Set("delegation_set_id", "")
	// To be consistent with other AWS services (e.g. ACM) that do not accept a trailing period,
	// we remove the suffix from the Hosted Zone Name returned from the API
	d.Set("name", TrimTrailingPeriod(aws.StringValue(output.HostedZone.Name)))
	d.Set("zone_id", CleanZoneID(aws.StringValue(output.HostedZone.Id)))

	var nameServers []string

	if output.DelegationSet != nil {
		d.Set("delegation_set_id", CleanDelegationSetID(aws.StringValue(output.DelegationSet.Id)))

		nameServers = aws.StringValueSlice(output.DelegationSet.NameServers)
	}

	if output.HostedZone.Config != nil {
		d.Set("comment", output.HostedZone.Config.Comment)

		if aws.BoolValue(output.HostedZone.Config.PrivateZone) {
			var err error
			nameServers, err = findNameServers(ctx, conn, d.Id(), d.Get("name").(string))

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
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("comment") {
		input := route53.UpdateHostedZoneCommentInput{
			Comment: aws.String(d.Get("comment").(string)),
			Id:      aws.String(d.Id()),
		}

		_, err := conn.UpdateHostedZoneCommentWithContext(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Hosted Zone (%s) comment: %s", d.Id(), err)
		}
	}

	if d.HasChange("vpc") {
		o, n := d.GetChange("vpc")
		oldVPCs := o.(*schema.Set)
		newVPCs := n.(*schema.Set)

		// VPCs cannot be empty, so add first and then remove
		for _, vpcRaw := range newVPCs.Difference(oldVPCs).List() {
			if vpcRaw == nil {
				continue
			}

			vpc := expandVPC(vpcRaw.(map[string]interface{}), region)
			err := hostedZoneAssociateVPC(ctx, conn, d.Id(), vpc)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		for _, vpcRaw := range oldVPCs.Difference(newVPCs).List() {
			if vpcRaw == nil {
				continue
			}

			vpc := expandVPC(vpcRaw.(map[string]interface{}), region)
			err := hostedZoneDisassociateVPC(ctx, conn, d.Id(), vpc)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceZoneRead(ctx, d, meta)...)
}

func resourceZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	if d.Get("force_destroy").(bool) {
		if err := deleteAllResourceRecordsFromHostedZone(ctx, conn, d.Id(), d.Get("name").(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if err := disableDNSSECForHostedZone(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "while force deleting Route53 Hosted Zone (%s), disabling DNSSEC: %s", d.Id(), err)
		}
	}

	input := &route53.DeleteHostedZoneInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Route53 Hosted Zone: %s", input)
	_, err := conn.DeleteHostedZoneWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	return diags
}

func FindHostedZoneByID(ctx context.Context, conn *route53.Route53, id string) (*route53.GetHostedZoneOutput, error) {
	input := &route53.GetHostedZoneInput{
		Id: aws.String(id),
	}

	output, err := conn.GetHostedZoneWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
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

func deleteAllResourceRecordsFromHostedZone(ctx context.Context, conn *route53.Route53, hostedZoneID, hostedZoneName string) error {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	var lastDeleteErr, lastErrorFromWaiter error
	var pageNum = 0
	err := conn.ListResourceRecordSetsPagesWithContext(ctx, input, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		sets := page.ResourceRecordSets
		pageNum += 1

		changes := make([]*route53.Change, 0)
		// 100 items per page returned by default
		for _, set := range sets {
			if strings.TrimSuffix(aws.StringValue(set.Name), ".") == strings.TrimSuffix(hostedZoneName, ".") && (aws.StringValue(set.Type) == "NS" || aws.StringValue(set.Type) == "SOA") {
				// Zone NS & SOA records cannot be deleted
				continue
			}
			changes = append(changes, &route53.Change{
				Action:            aws.String(route53.ChangeActionDelete),
				ResourceRecordSet: set,
			})
		}

		if len(changes) == 0 {
			return !lastPage
		}

		log.Printf("[DEBUG] Deleting %d records (page %d) from %s", len(changes), pageNum, hostedZoneID)

		req := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(hostedZoneID),
			ChangeBatch: &route53.ChangeBatch{
				Comment: aws.String("Deleted by Terraform"),
				Changes: changes,
			},
		}

		var resp interface{}
		resp, lastDeleteErr = DeleteRecordSet(ctx, conn, req)
		if out, ok := resp.(*route53.ChangeResourceRecordSetsOutput); ok {
			log.Printf("[DEBUG] Waiting for change batch to become INSYNC: %#v", out)
			if out.ChangeInfo != nil && out.ChangeInfo.Id != nil {
				lastErrorFromWaiter = WaitForRecordSetToSync(ctx, conn, CleanChangeID(aws.StringValue(out.ChangeInfo.Id)))
			} else {
				log.Printf("[DEBUG] Change info was empty")
			}
		} else {
			log.Printf("[DEBUG] Unable to wait for change batch because of an error: %s", lastDeleteErr)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Failed listing/deleting record sets: %s\nLast error from deletion: %s\nLast error from waiter: %s",
			err, lastDeleteErr, lastErrorFromWaiter)
	}

	return nil
}

func dnsSECStatus(ctx context.Context, conn *route53.Route53, hostedZoneID string) (string, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	var output *route53.GetDNSSECOutput
	err := tfresource.Retry(ctx, 3*time.Minute, func() *retry.RetryError {
		var err error

		output, err = conn.GetDNSSECWithContext(ctx, input)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[DEBUG] Retrying to get DNS SEC for zone %s: %s", hostedZoneID, err)
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelayRand(1*time.Minute), tfresource.WithPollInterval(30*time.Second))

	if tfresource.TimedOut(err) {
		output, err = conn.GetDNSSECWithContext(ctx, input)
	}

	if tfawserr.ErrMessageContains(err, route53.ErrCodeInvalidArgument, "Operation is unsupported for private") {
		return "NOT_SIGNING", nil
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Status == nil {
		return "", fmt.Errorf("getting DNS SEC for hosted zone (%s): empty response (%v)", hostedZoneID, output)
	}

	return aws.StringValue(output.Status.ServeSignature), nil
}

func disableDNSSECForHostedZone(ctx context.Context, conn *route53.Route53, hostedZoneID string) error {
	// hosted zones cannot be deleted if DNSSEC Key Signing Keys exist
	log.Printf("[DEBUG] Disabling DNS SEC for zone %s", hostedZoneID)

	status, err := dnsSECStatus(ctx, conn, hostedZoneID)

	if err != nil {
		return fmt.Errorf("could not get DNS SEC status for hosted zone (%s): %w", hostedZoneID, err)
	}

	if status != "SIGNING" {
		log.Printf("[DEBUG] Not necessary to disable DNS SEC for hosted zone (%s): %s (status)", hostedZoneID, status)
		return nil
	}

	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	var output *route53.DisableHostedZoneDNSSECOutput
	err = tfresource.Retry(ctx, 5*time.Minute, func() *retry.RetryError {
		var err error

		output, err = conn.DisableHostedZoneDNSSECWithContext(ctx, input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53.ErrCodeKeySigningKeyInParentDSRecord) {
				log.Printf("[DEBUG] Unable to disable DNS SEC for zone %s because key-signing key in parent DS record. Retrying... (%s)", hostedZoneID, err)
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelayRand(1*time.Minute), tfresource.WithPollInterval(20*time.Second))

	if tfresource.TimedOut(err) {
		output, err = conn.DisableHostedZoneDNSSECWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, "InvalidArgument", "Operation is unsupported for private") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("disabling Route 53 Hosted Zone DNSSEC (%s): %w", hostedZoneID, err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(ctx, conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for Route 53 Hosted Zone DNSSEC (%s) disable: %w", hostedZoneID, err)
		}
	}

	return nil
}

func resourceGoWait(ctx context.Context, conn *route53.Route53, input *route53.GetChangeInput) (result interface{}, state string, err error) {
	status, err := conn.GetChangeWithContext(ctx, input)
	if err != nil {
		return nil, "UNKNOWN", err
	}
	return true, aws.StringValue(status.ChangeInfo.Status), nil
}

// CleanChangeID is used to remove the leading /change/
func CleanChangeID(ID string) string {
	return strings.TrimPrefix(ID, "/change/")
}

// CleanZoneID is used to remove the leading /hostedzone/
func CleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}

// TrimTrailingPeriod is used to remove the trailing period
// of "name" or "domain name" attributes often returned from
// the Route53 API or provided as user input.
// The single dot (".") domain name is returned as-is.
func TrimTrailingPeriod(v interface{}) string {
	var str string
	switch value := v.(type) {
	case *string:
		str = aws.StringValue(value)
	case string:
		str = value
	default:
		return ""
	}

	if str == "." {
		return str
	}

	return strings.TrimSuffix(str, ".")
}

func findNameServers(ctx context.Context, conn *route53.Route53, zoneId string, zoneName string) ([]string, error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneId),
		StartRecordName: aws.String(zoneName),
		StartRecordType: aws.String(route53.RRTypeNs),
	}

	output, err := conn.ListResourceRecordSetsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(output.ResourceRecordSets) == 0 {
		return nil, nil
	}
	ns := make([]string, len(output.ResourceRecordSets[0].ResourceRecords))
	for i := range output.ResourceRecordSets[0].ResourceRecords {
		ns[i] = aws.StringValue(output.ResourceRecordSets[0].ResourceRecords[i].Value)
	}
	sort.Strings(ns)
	return ns, nil
}

func expandVPCs(l []interface{}, currentRegion string) []*route53.VPC {
	vpcs := []*route53.VPC{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}

		vpcs = append(vpcs, expandVPC(mRaw.(map[string]interface{}), currentRegion))
	}

	return vpcs
}

func expandVPC(m map[string]interface{}, currentRegion string) *route53.VPC {
	vpc := &route53.VPC{
		VPCId:     aws.String(m["vpc_id"].(string)),
		VPCRegion: aws.String(currentRegion),
	}

	if v, ok := m["vpc_region"]; ok && v.(string) != "" {
		vpc.VPCRegion = aws.String(v.(string))
	}

	return vpc
}

func flattenVPCs(vpcs []*route53.VPC) []interface{} {
	l := []interface{}{}

	for _, vpc := range vpcs {
		if vpc == nil {
			continue
		}

		m := map[string]interface{}{
			"vpc_id":     aws.StringValue(vpc.VPCId),
			"vpc_region": aws.StringValue(vpc.VPCRegion),
		}

		l = append(l, m)
	}

	return l
}

func hostedZoneAssociateVPC(ctx context.Context, conn *route53.Route53, zoneID string, vpc *route53.VPC) error {
	input := &route53.AssociateVPCWithHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC:          vpc,
	}

	output, err := conn.AssociateVPCWithHostedZoneWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("associating Route53 Hosted Zone (%s) to VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	if err := waitForChangeSynchronization(ctx, conn, CleanChangeID(aws.StringValue(output.ChangeInfo.Id))); err != nil {
		return fmt.Errorf("waiting for Route53 Hosted Zone (%s) association to VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	return nil
}

func hostedZoneDisassociateVPC(ctx context.Context, conn *route53.Route53, zoneID string, vpc *route53.VPC) error {
	input := &route53.DisassociateVPCFromHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC:          vpc,
	}

	output, err := conn.DisassociateVPCFromHostedZoneWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("disassociating Route53 Hosted Zone (%s) from VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	if err := waitForChangeSynchronization(ctx, conn, CleanChangeID(aws.StringValue(output.ChangeInfo.Id))); err != nil {
		return fmt.Errorf("waiting for Route53 Hosted Zone (%s) disassociation from VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	return nil
}

func hostedZoneVPCHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["vpc_id"].(string)))

	return create.StringHashcode(buf.String())
}

func waitForChangeSynchronization(ctx context.Context, conn *route53.Route53, changeID string) error {
	conf := retry.StateChangeConf{
		Pending:      []string{route53.ChangeStatusPending},
		Target:       []string{route53.ChangeStatusInsync},
		Delay:        time.Duration(rand.Int63n(zoneChangeSyncMaxDelay-zoneChangeSyncMinDelay)+zoneChangeSyncMinDelay) * time.Second,
		MinTimeout:   5 * time.Second,
		PollInterval: time.Duration(rand.Int63n(zoneChangeSyncMaxPollInterval-zoneChangeSyncMinPollInterval)+zoneChangeSyncMinPollInterval) * time.Second,
		Timeout:      15 * time.Minute,
		Refresh: func() (result interface{}, state string, err error) {
			input := &route53.GetChangeInput{
				Id: aws.String(changeID),
			}

			log.Printf("[DEBUG] Getting Route53 Change status: %s", input)
			output, err := conn.GetChangeWithContext(ctx, input)

			if err != nil {
				return nil, "UNKNOWN", err
			}

			if output == nil || output.ChangeInfo == nil {
				return nil, "UNKNOWN", fmt.Errorf("Route53 GetChange response empty for ID: %s", changeID)
			}

			return true, aws.StringValue(output.ChangeInfo.Status), nil
		},
	}

	_, err := conf.WaitForStateContext(ctx)

	return err
}
