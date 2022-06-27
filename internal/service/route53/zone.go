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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	zoneChangeSyncMinDelay        = 10
	zoneChangeSyncMaxDelay        = 30
	zoneChangeSyncMinPollInterval = 15
	zoneChangeSyncMaxPollInterval = 30
)

func ResourceZone() *schema.Resource {
	return &schema.Resource{
		Create: resourceZoneCreate,
		Read:   resourceZoneRead,
		Update: resourceZoneUpdate,
		Delete: resourceZoneDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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

			"comment": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Managed by Terraform",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},

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

			"delegation_set_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vpc"},
				ValidateFunc:  validation.StringLenBetween(0, 32),
			},

			"name_servers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceZoneCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	region := meta.(*conns.AWSClient).Region

	input := &route53.CreateHostedZoneInput{
		CallerReference: aws.String(resource.UniqueId()),
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

	var vpcs []*route53.VPC = expandVPCs(d.Get("vpc").(*schema.Set).List(), region)

	if len(vpcs) > 0 {
		input.VPC = vpcs[0]
	}

	log.Printf("[DEBUG] Creating Route53 hosted zone: %s", input)
	output, err := conn.CreateHostedZone(input)

	if err != nil {
		return fmt.Errorf("error creating Route53 Hosted Zone: %s", err)
	}

	d.SetId(CleanZoneID(aws.StringValue(output.HostedZone.Id)))

	if output.ChangeInfo != nil {
		if err := waitForChangeSynchronization(conn, CleanChangeID(aws.StringValue(output.ChangeInfo.Id))); err != nil {
			return fmt.Errorf("error waiting for Route53 Hosted Zone (%s) creation: %s", d.Id(), err)
		}
	}

	if err := UpdateTags(conn, d.Id(), route53.TagResourceTypeHostedzone, nil, tags); err != nil {
		return fmt.Errorf("error setting Route53 Zone (%s) tags: %s", d.Id(), err)
	}

	// Associate additional VPCs beyond the first
	if len(vpcs) > 1 {
		for _, vpc := range vpcs[1:] {
			err := hostedZoneVPCAssociate(conn, d.Id(), vpc)

			if err != nil {
				return err
			}
		}
	}

	return resourceZoneRead(d, meta)
}

func resourceZoneRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &route53.GetHostedZoneInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Getting Route53 Hosted Zone: %s", input)
	output, err := conn.GetHostedZone(input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		log.Printf("[WARN] Route53 Hosted Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	if output == nil || output.HostedZone == nil {
		log.Printf("[WARN] Route53 Hosted Zone (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

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
			nameServers, err = getNameServers(d.Id(), d.Get("name").(string), conn)

			if err != nil {
				return fmt.Errorf("error getting Route53 Hosted Zone (%s) name servers: %s", d.Id(), err)
			}
		}
	}

	sort.Strings(nameServers)
	if err := d.Set("name_servers", nameServers); err != nil {
		return fmt.Errorf("error setting name_servers: %s", err)
	}

	if err := d.Set("vpc", flattenVPCs(output.VPCs)); err != nil {
		return fmt.Errorf("error setting vpc: %s", err)
	}

	tags, err := ListTags(conn, d.Id(), route53.TagResourceTypeHostedzone)

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("hostedzone/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceZoneUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("comment") {
		input := route53.UpdateHostedZoneCommentInput{
			Id:      aws.String(d.Id()),
			Comment: aws.String(d.Get("comment").(string)),
		}

		_, err := conn.UpdateHostedZoneComment(&input)

		if err != nil {
			return fmt.Errorf("error updating Route53 Hosted Zone (%s) comment: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), route53.TagResourceTypeHostedzone, o, n); err != nil {
			return fmt.Errorf("error updating Route53 Zone (%s) tags: %s", d.Id(), err)
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
			err := hostedZoneVPCAssociate(conn, d.Id(), vpc)

			if err != nil {
				return err
			}
		}

		for _, vpcRaw := range oldVPCs.Difference(newVPCs).List() {
			if vpcRaw == nil {
				continue
			}

			vpc := expandVPC(vpcRaw.(map[string]interface{}), region)
			err := hostedZoneVPCDisassociate(conn, d.Id(), vpc)

			if err != nil {
				return err
			}
		}
	}

	return resourceZoneRead(d, meta)
}

func resourceZoneDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53Conn

	if d.Get("force_destroy").(bool) {
		if err := deleteAllRecordsInHostedZoneId(d.Id(), d.Get("name").(string), conn); err != nil {
			return fmt.Errorf("error while force deleting Route53 Hosted Zone (%s), deleting records: %w", d.Id(), err)
		}

		if err := disableDNSSECForZone(conn, d.Id()); err != nil {
			return fmt.Errorf("error while force deleting Route53 Hosted Zone (%s), disabling DNSSEC: %w", d.Id(), err)
		}
	}

	input := &route53.DeleteHostedZoneInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Route53 Hosted Zone: %s", input)
	_, err := conn.DeleteHostedZone(input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 Hosted Zone (%s): %s", d.Id(), err)
	}

	return nil
}

func deleteAllRecordsInHostedZoneId(hostedZoneId, hostedZoneName string, conn *route53.Route53) error {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneId),
	}

	var lastDeleteErr, lastErrorFromWaiter error
	var pageNum = 0
	err := conn.ListResourceRecordSetsPages(input, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
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

		log.Printf("[DEBUG] Deleting %d records (page %d) from %s", len(changes), pageNum, hostedZoneId)

		req := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(hostedZoneId),
			ChangeBatch: &route53.ChangeBatch{
				Comment: aws.String("Deleted by Terraform"),
				Changes: changes,
			},
		}

		var resp interface{}
		resp, lastDeleteErr = DeleteRecordSet(conn, req)
		if out, ok := resp.(*route53.ChangeResourceRecordSetsOutput); ok {
			log.Printf("[DEBUG] Waiting for change batch to become INSYNC: %#v", out)
			if out.ChangeInfo != nil && out.ChangeInfo.Id != nil {
				lastErrorFromWaiter = WaitForRecordSetToSync(conn, CleanChangeID(aws.StringValue(out.ChangeInfo.Id)))
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

func dnsSECStatus(conn *route53.Route53, hostedZoneID string) (string, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	var output *route53.GetDNSSECOutput
	err := tfresource.RetryConfigContext(context.Background(), 0*time.Millisecond, 1*time.Minute, 0*time.Millisecond, 30*time.Second, 3*time.Minute, func() *resource.RetryError {
		var err error

		output, err = conn.GetDNSSEC(input)

		if err != nil {
			if strings.Contains(err.Error(), "Throttling") {
				log.Printf("[DEBUG] Retrying to get DNS SEC for zone %s: %s", hostedZoneID, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetDNSSEC(input)
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

func disableDNSSECForZone(conn *route53.Route53, hostedZoneId string) error {
	// hosted zones cannot be deleted if DNSSEC Key Signing Keys exist
	log.Printf("[DEBUG] Disabling DNS SEC for zone %s", hostedZoneId)

	status, err := dnsSECStatus(conn, hostedZoneId)

	if err != nil {
		return fmt.Errorf("could not get DNS SEC status for hosted zone (%s): %w", hostedZoneId, err)
	}

	if status != "SIGNING" {
		log.Printf("[DEBUG] Not necessary to disable DNS SEC for hosted zone (%s): %s (status)", hostedZoneId, status)
		return nil
	}

	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: aws.String(hostedZoneId),
	}

	var output *route53.DisableHostedZoneDNSSECOutput
	err = tfresource.RetryConfigContext(context.Background(), 0*time.Millisecond, 1*time.Minute, 0*time.Millisecond, 20*time.Second, 5*time.Minute, func() *resource.RetryError {
		var err error

		output, err = conn.DisableHostedZoneDNSSEC(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53.ErrCodeKeySigningKeyInParentDSRecord) {
				log.Printf("[DEBUG] Unable to disable DNS SEC for zone %s because key-signing key in parent DS record. Retrying... (%s)", hostedZoneId, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.DisableHostedZoneDNSSEC(input)
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
		return fmt.Errorf("disabling Route 53 Hosted Zone DNSSEC (%s): %w", hostedZoneId, err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waitChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("waiting for Route 53 Hosted Zone DNSSEC (%s) disable: %w", hostedZoneId, err)
		}
	}

	return nil
}

func resourceGoWait(r53 *route53.Route53, ref *route53.GetChangeInput) (result interface{}, state string, err error) {

	status, err := r53.GetChange(ref)
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

func getNameServers(zoneId string, zoneName string, r53 *route53.Route53) ([]string, error) {
	resp, err := r53.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneId),
		StartRecordName: aws.String(zoneName),
		StartRecordType: aws.String("NS"),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.ResourceRecordSets) == 0 {
		return nil, nil
	}
	ns := make([]string, len(resp.ResourceRecordSets[0].ResourceRecords))
	for i := range resp.ResourceRecordSets[0].ResourceRecords {
		ns[i] = aws.StringValue(resp.ResourceRecordSets[0].ResourceRecords[i].Value)
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

func hostedZoneVPCAssociate(conn *route53.Route53, zoneID string, vpc *route53.VPC) error {
	input := &route53.AssociateVPCWithHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC:          vpc,
	}

	log.Printf("[DEBUG] Associating Route53 Hosted Zone with VPC: %s", input)
	output, err := conn.AssociateVPCWithHostedZone(input)

	if err != nil {
		return fmt.Errorf("error associating Route53 Hosted Zone (%s) to VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	if err := waitForChangeSynchronization(conn, CleanChangeID(aws.StringValue(output.ChangeInfo.Id))); err != nil {
		return fmt.Errorf("error waiting for Route53 Hosted Zone (%s) association to VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	return nil
}

func hostedZoneVPCDisassociate(conn *route53.Route53, zoneID string, vpc *route53.VPC) error {
	input := &route53.DisassociateVPCFromHostedZoneInput{
		HostedZoneId: aws.String(zoneID),
		VPC:          vpc,
	}

	log.Printf("[DEBUG] Disassociating Route53 Hosted Zone with VPC: %s", input)
	output, err := conn.DisassociateVPCFromHostedZone(input)

	if err != nil {
		return fmt.Errorf("error disassociating Route53 Hosted Zone (%s) from VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	if err := waitForChangeSynchronization(conn, CleanChangeID(aws.StringValue(output.ChangeInfo.Id))); err != nil {
		return fmt.Errorf("error waiting for Route53 Hosted Zone (%s) disassociation from VPC (%s): %s", zoneID, aws.StringValue(vpc.VPCId), err)
	}

	return nil
}

func hostedZoneVPCHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["vpc_id"].(string)))

	return create.StringHashcode(buf.String())
}

func waitForChangeSynchronization(conn *route53.Route53, changeID string) error {
	rand.Seed(time.Now().UTC().UnixNano())

	conf := resource.StateChangeConf{
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
			output, err := conn.GetChange(input)

			if err != nil {
				return nil, "UNKNOWN", err
			}

			if output == nil || output.ChangeInfo == nil {
				return nil, "UNKNOWN", fmt.Errorf("Route53 GetChange response empty for ID: %s", changeID)
			}

			return true, aws.StringValue(output.ChangeInfo.Status), nil
		},
	}

	_, err := conf.WaitForState()

	return err
}
