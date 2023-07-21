// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	recordSetSyncMinDelay = 10
	recordSetSyncMaxDelay = 30
)

// @SDKResource("aws_route53_record")
func ResourceRecord() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecordCreate,
		ReadWithoutTimeout:   resourceRecordRead,
		UpdateWithoutTimeout: resourceRecordUpdate,
		DeleteWithoutTimeout: resourceRecordDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := ParseRecordID(d.Id())
				// We check that we have parsed the id into the correct number of segments.
				// We need at least 3 segments!
				if parts[0] == "" || parts[1] == "" || parts[2] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected ZONEID_RECORDNAME_TYPE_SET-IDENTIFIER (e.g. Z4KAPRWWNC7JR_dev.example.com_NS_dev), where SET-IDENTIFIER is optional", d.Id())
				}

				d.Set("zone_id", parts[0])
				d.Set("name", parts[1])
				d.Set("type", parts[2])
				if parts[3] != "" {
					d.Set("set_identifier", parts[3])
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 2,
		MigrateState:  RecordMigrateState,

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"evaluate_target_health": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"name": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: NormalizeAliasName,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"zone_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
						},
					},
				},
				ExactlyOneOf:  []string{"alias", "records"},
				ConflictsWith: []string{"ttl"},
			},
			"allow_overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"cidr_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"collection_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"location_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{
					"failover_routing_policy",
					"geolocation_routing_policy",
					"latency_routing_policy",
					"multivalue_answer_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"failover_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(route53.ResourceRecordSetFailover_Values(), false),
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"geolocation_routing_policy",
					"latency_routing_policy",
					"multivalue_answer_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"geolocation_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"continent": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"country": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subdivision": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"latency_routing_policy",
					"multivalue_answer_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"health_check_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"latency_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"geolocation_routing_policy",
					"multivalue_answer_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"multivalue_answer_routing_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"geolocation_routing_policy",
					"latency_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					// AWS Provider aws_acm_certification.domain_validation_options.resource_record_name
					// references (and perhaps others) contain a trailing period, requiring a custom StateFunc
					// to trim the string to prevent Route53 API error.
					value := strings.TrimSuffix(v.(string), ".")
					return strings.ToLower(value)
				},
			},
			"records": {
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				ExactlyOneOf: []string{"alias", "records"},
			},
			"set_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ttl": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"alias"},
				RequiredWith:  []string{"records", "ttl"},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(route53.RRType_Values(), false),
			},
			"weighted_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"weight": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"geolocation_routing_policy",
					"latency_routing_policy",
					"multivalue_answer_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"zone_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	zoneID := CleanZoneID(d.Get("zone_id").(string))
	zoneRecord, err := FindHostedZoneByID(ctx, conn, zoneID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone (%s): %s", zoneID, err)
	}

	// Build the record
	rec := expandResourceRecordSet(d, aws.StringValue(zoneRecord.HostedZone.Name))

	// Protect existing DNS records which might be managed in another way.
	// Use UPSERT only if the overwrite flag is true or if the current action is an update
	// Else CREATE is used and fail if the same record exists
	var action string
	if d.Get("allow_overwrite").(bool) || !d.IsNewResource() {
		action = route53.ChangeActionUpsert
	} else {
		action = route53.ChangeActionCreate
	}

	// Create the new records. We abuse StateChangeConf for this to
	// retry for us since Route53 sometimes returns errors about another
	// operation happening at the same time.
	changeBatch := &route53.ChangeBatch{
		Comment: aws.String("Managed by Terraform"),
		Changes: []*route53.Change{
			{
				Action:            aws.String(action),
				ResourceRecordSet: rec,
			},
		},
	}

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  changeBatch,
		HostedZoneId: aws.String(CleanZoneID(aws.StringValue(zoneRecord.HostedZone.Id))),
	}

	changeInfo, err := ChangeResourceRecordSets(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Record: %s", err)
	}

	vars := []string{
		zoneID,
		strings.ToLower(d.Get("name").(string)),
		d.Get("type").(string),
	}
	if v, ok := d.GetOk("set_identifier"); ok {
		vars = append(vars, v.(string))
	}
	d.SetId(strings.Join(vars, "_"))

	if err := WaitForRecordSetToSync(ctx, conn, CleanChangeID(aws.StringValue(changeInfo.Id))); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Record (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRecordRead(ctx, d, meta)...)
}

func resourceRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	record, fqdn, err := FindResourceRecordSetByFourPartKey(ctx, conn, CleanZoneID(d.Get("zone_id").(string)), d.Get("name").(string), d.Get("type").(string), d.Get("set_identifier").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Record (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Record (%s): %s", d.Id(), err)
	}

	d.Set("fqdn", fqdn)

	err = d.Set("records", FlattenResourceRecords(record.ResourceRecords, aws.StringValue(record.Type)))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting records: %s", err)
	}

	if alias := record.AliasTarget; alias != nil {
		name := NormalizeAliasName(aws.StringValue(alias.DNSName))
		v := []map[string]interface{}{{
			"zone_id":                aws.StringValue(alias.HostedZoneId),
			"name":                   name,
			"evaluate_target_health": aws.BoolValue(alias.EvaluateTargetHealth),
		}}
		if err := d.Set("alias", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting alias: %s", err)
		}
	}

	d.Set("ttl", record.TTL)

	if record.CidrRoutingConfig != nil {
		v := []map[string]interface{}{{
			"collection_id": aws.StringValue(record.CidrRoutingConfig.CollectionId),
			"location_name": aws.StringValue(record.CidrRoutingConfig.LocationName),
		}}
		if err := d.Set("cidr_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cidr_routing_policy: %s", err)
		}
	}

	if record.Failover != nil {
		v := []map[string]interface{}{{
			"type": aws.StringValue(record.Failover),
		}}
		if err := d.Set("failover_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting failover_routing_policy: %s", err)
		}
	}

	if record.GeoLocation != nil {
		v := []map[string]interface{}{{
			"continent":   aws.StringValue(record.GeoLocation.ContinentCode),
			"country":     aws.StringValue(record.GeoLocation.CountryCode),
			"subdivision": aws.StringValue(record.GeoLocation.SubdivisionCode),
		}}
		if err := d.Set("geolocation_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting geolocation_routing_policy: %s", err)
		}
	}

	if record.Region != nil {
		v := []map[string]interface{}{{
			"region": aws.StringValue(record.Region),
		}}
		if err := d.Set("latency_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting latency_routing_policy: %s", err)
		}
	}

	d.Set("multivalue_answer_routing_policy", record.MultiValueAnswer)

	if record.Weight != nil {
		v := []map[string]interface{}{{
			"weight": aws.Int64Value((record.Weight)),
		}}
		if err := d.Set("weighted_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting weighted_routing_policy: %s", err)
		}
	}

	d.Set("set_identifier", record.SetIdentifier)
	d.Set("health_check_id", record.HealthCheckId)

	return diags
}

func resourceRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	// Route 53 supports CREATE, DELETE, and UPSERT actions. We use UPSERT, and
	// AWS dynamically determines if a record should be created or updated.
	// Amazon Route 53 can update an existing resource record set only when all
	// of the following values match: Name, Type and SetIdentifier
	// See http://docs.aws.amazon.com/Route53/latest/APIReference/API_ChangeResourceRecordSets.html

	if !d.HasChange("type") && !d.HasChange("set_identifier") {
		// If neither type nor set_identifier changed we use UPSERT,
		// for resource update here we simply fall through to
		// our resource create function.
		return append(diags, resourceRecordCreate(ctx, d, meta)...)
	}

	// Otherwise, we delete the existing record and create a new record within
	// a transactional change.
	zone := CleanZoneID(d.Get("zone_id").(string))

	zoneRecord, err := FindHostedZoneByID(ctx, conn, zone)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route 53 Record (%s): getting Hosted Zone (%s): %s", d.Id(), zone, err)
	}

	// Build the to be deleted record
	en := ExpandRecordName(d.Get("name").(string), aws.StringValue(zoneRecord.HostedZone.Name))
	typeo, _ := d.GetChange("type")

	oldRec := &route53.ResourceRecordSet{
		Name: aws.String(en),
		Type: aws.String(typeo.(string)),
	}

	// If the old record has any of the following, we need to pass that in
	// here because otherwise the API will give us an error:
	// - cidr_routing_policy
	// - failover_routing_policy
	// - geolocation_routing_policy
	// - latency_routing_policy
	// - multivalue_answer_routing_policy
	// - weighted_routing_policy

	if v, _ := d.GetChange("cidr_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.CidrRoutingConfig = &route53.CidrRoutingConfig{
						CollectionId: nilString(v["collection_id"].(string)),
						LocationName: nilString(v["location_name"].(string)),
					}
				}
			}
		}
	}

	if v, _ := d.GetChange("failover_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.Failover = aws.String(v["type"].(string))
				}
			}
		}
	}

	if v, _ := d.GetChange("geolocation_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.GeoLocation = &route53.GeoLocation{
						ContinentCode:   nilString(v["continent"].(string)),
						CountryCode:     nilString(v["country"].(string)),
						SubdivisionCode: nilString(v["subdivision"].(string)),
					}
				}
			}
		}
	}

	if v, _ := d.GetChange("latency_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.Region = aws.String(v["region"].(string))
				}
			}
		}
	}

	if v, _ := d.GetChange("multivalue_answer_routing_policy"); v != nil && v.(bool) {
		oldRec.MultiValueAnswer = aws.Bool(v.(bool))
	}

	if v, _ := d.GetChange("weighted_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.Weight = aws.Int64(int64(v["weight"].(int)))
				}
			}
		}
	}

	if v, _ := d.GetChange("ttl"); v.(int) != 0 {
		oldRec.TTL = aws.Int64(int64(v.(int)))
	}

	// Resource records
	if v, _ := d.GetChange("records"); v != nil {
		recs := v.(*schema.Set).List()
		if len(recs) > 0 {
			oldRec.ResourceRecords = expandResourceRecords(recs, typeo.(string))
		}
	}

	// Alias record
	if v, _ := d.GetChange("alias"); v != nil {
		aliases := v.([]interface{})
		if len(aliases) == 1 {
			alias := aliases[0].(map[string]interface{})
			oldRec.AliasTarget = &route53.AliasTarget{
				DNSName:              aws.String(alias["name"].(string)),
				EvaluateTargetHealth: aws.Bool(alias["evaluate_target_health"].(bool)),
				HostedZoneId:         aws.String(alias["zone_id"].(string)),
			}
		}
	}

	// If health check id is present send that to AWS
	if v, _ := d.GetChange("health_check_id"); v.(string) != "" {
		oldRec.HealthCheckId = aws.String(v.(string))
	}

	if v, _ := d.GetChange("set_identifier"); v.(string) != "" {
		oldRec.SetIdentifier = aws.String(v.(string))
	}

	// Build the to be created record
	rec := expandResourceRecordSet(d, aws.StringValue(zoneRecord.HostedZone.Name))

	// Delete the old and create the new records in a single batch. We abuse
	// StateChangeConf for this to retry for us since Route53 sometimes returns
	// errors about another operation happening at the same time.
	changeBatch := &route53.ChangeBatch{
		Comment: aws.String("Managed by Terraform"),
		Changes: []*route53.Change{
			{
				Action:            aws.String(route53.ChangeActionDelete),
				ResourceRecordSet: oldRec,
			},
			{
				Action:            aws.String(route53.ChangeActionCreate),
				ResourceRecordSet: rec,
			},
		},
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(CleanZoneID(aws.StringValue(zoneRecord.HostedZone.Id))),
		ChangeBatch:  changeBatch,
	}

	log.Printf("[DEBUG] Updating resource records for zone: %s, name: %s", zone, aws.StringValue(rec.Name))

	changeInfo, err := ChangeResourceRecordSets(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route 53 resource record sets: %s", err)
	}

	// Generate an ID
	vars := []string{
		zone,
		strings.ToLower(d.Get("name").(string)),
		d.Get("type").(string),
	}
	if v, ok := d.GetOk("set_identifier"); ok {
		vars = append(vars, v.(string))
	}

	d.SetId(strings.Join(vars, "_"))

	if err := WaitForRecordSetToSync(ctx, conn, CleanChangeID(aws.StringValue(changeInfo.Id))); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Record (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceRecordRead(ctx, d, meta)...)
}

func resourceRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	zoneID := CleanZoneID(d.Get("zone_id").(string))
	var name string
	// If we're dealing with a change of record name, but we're operating on the old, rather than
	// the new, resource, then we need to use the old name to find it (in order to delete it).
	if !d.IsNewResource() && d.HasChange("name") {
		oldName, _ := d.GetChange("name")
		name = oldName.(string)
	} else {
		name = d.Get("name").(string)
	}
	rec, _, err := FindResourceRecordSetByFourPartKey(ctx, conn, zoneID, name, d.Get("type").(string), d.Get("set_identifier").(string))

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Record (%s): %s", d.Id(), err)
	}

	// Change batch for deleting
	changeBatch := &route53.ChangeBatch{
		Comment: aws.String("Deleted by Terraform"),
		Changes: []*route53.Change{
			{
				Action:            aws.String(route53.ChangeActionDelete),
				ResourceRecordSet: rec,
			},
		},
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch:  changeBatch,
		HostedZoneId: aws.String(zoneID),
	}

	respRaw, err := DeleteRecordSet(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route 53 Record (%s): %s", d.Id(), err)
	}

	changeInfo := respRaw.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo
	if changeInfo == nil {
		log.Printf("[INFO] No ChangeInfo Found. Waiting for Sync not required")
		return diags
	}

	if err := WaitForRecordSetToSync(ctx, conn, CleanChangeID(aws.StringValue(changeInfo.Id))); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Record (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindResourceRecordSetByFourPartKey(ctx context.Context, conn *route53.Route53, zoneID, recordName, recordType, recordSetIdentifier string) (*route53.ResourceRecordSet, string, error) {
	zone, err := FindHostedZoneByID(ctx, conn, zoneID)

	if err != nil {
		return nil, "", err
	}

	fqdn := ExpandRecordName(recordName, aws.StringValue(zone.HostedZone.Name))
	recordName = FQDN(strings.ToLower(fqdn))
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneID),
		StartRecordName: aws.String(recordName),
		StartRecordType: aws.String(recordType),
	}
	if recordSetIdentifier == "" {
		input.MaxItems = aws.String("1")
	} else {
		input.MaxItems = aws.String("100")
	}
	var output *route53.ResourceRecordSet

	err = conn.ListResourceRecordSetsPagesWithContext(ctx, input, func(page *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourceRecordSets {
			if recordName != strings.ToLower(CleanRecordName(aws.StringValue(v.Name))) {
				continue
			}
			if recordType != strings.ToUpper(aws.StringValue(v.Type)) {
				continue
			}
			if recordSetIdentifier != aws.StringValue(v.SetIdentifier) {
				continue
			}

			output = v

			return false
		}

		if strings.ToLower(CleanRecordName(aws.StringValue(page.NextRecordName))) != recordName {
			return false
		}

		if strings.ToUpper(aws.StringValue(page.NextRecordType)) != recordType {
			return false
		}

		return !lastPage
	})

	if err != nil {
		return nil, "", err
	}

	if output == nil {
		return nil, "", &retry.NotFoundError{}
	}

	return output, fqdn, nil
}

func ChangeResourceRecordSets(ctx context.Context, conn *route53.Route53, input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeInfo, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 1*time.Minute, func() (interface{}, error) {
		return conn.ChangeResourceRecordSetsWithContext(ctx, input)
	}, route53.ErrCodeNoSuchHostedZone)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo, nil
}

func WaitForRecordSetToSync(ctx context.Context, conn *route53.Route53, requestId string) error {
	wait := retry.StateChangeConf{
		Pending:      []string{route53.ChangeStatusPending},
		Target:       []string{route53.ChangeStatusInsync},
		Delay:        time.Duration(rand.Int63n(recordSetSyncMaxDelay-recordSetSyncMinDelay)+recordSetSyncMinDelay) * time.Second,
		MinTimeout:   5 * time.Second,
		PollInterval: 20 * time.Second,
		Timeout:      30 * time.Minute,
		Refresh: func() (result interface{}, state string, err error) {
			changeRequest := &route53.GetChangeInput{
				Id: aws.String(requestId),
			}
			return resourceGoWait(ctx, conn, changeRequest)
		},
	}
	_, err := wait.WaitForStateContext(ctx)
	return err
}

func DeleteRecordSet(ctx context.Context, conn *route53.Route53, input *route53.ChangeResourceRecordSetsInput) (interface{}, error) {
	out, err := conn.ChangeResourceRecordSetsWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, route53.ErrCodeInvalidChangeBatch) {
		return out, nil
	}

	return out, err
}

func expandResourceRecordSet(d *schema.ResourceData, zoneName string) *route53.ResourceRecordSet {
	// get expanded name
	en := ExpandRecordName(d.Get("name").(string), zoneName)

	// Create the RecordSet request with the fully expanded name, e.g.
	// sub.domain.com. Route 53 requires a fully qualified domain name, but does
	// not require the trailing ".", which it will itself, so we don't call FQDN
	// here.
	rec := &route53.ResourceRecordSet{
		Name: aws.String(en),
		Type: aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("ttl"); ok {
		rec.TTL = aws.Int64(int64(v.(int)))
	}

	// Resource records
	if v, ok := d.GetOk("records"); ok {
		recs := v.(*schema.Set).List()
		rec.ResourceRecords = expandResourceRecords(recs, d.Get("type").(string))
	}

	// Alias record
	if v, ok := d.GetOk("alias"); ok {
		aliases := v.([]interface{})
		alias := aliases[0].(map[string]interface{})
		rec.AliasTarget = &route53.AliasTarget{
			DNSName:              aws.String(alias["name"].(string)),
			EvaluateTargetHealth: aws.Bool(alias["evaluate_target_health"].(bool)),
			HostedZoneId:         aws.String(alias["zone_id"].(string)),
		}
	}

	if v, ok := d.GetOk("cidr_routing_policy"); ok {
		records := v.([]interface{})
		cidr := records[0].(map[string]interface{})

		rec.CidrRoutingConfig = &route53.CidrRoutingConfig{
			CollectionId: aws.String(cidr["collection_id"].(string)),
			LocationName: aws.String(cidr["location_name"].(string)),
		}
	}

	if v, ok := d.GetOk("failover_routing_policy"); ok {
		records := v.([]interface{})
		failover := records[0].(map[string]interface{})

		rec.Failover = aws.String(failover["type"].(string))
	}

	if v, ok := d.GetOk("geolocation_routing_policy"); ok {
		geolocations := v.([]interface{})
		geolocation := geolocations[0].(map[string]interface{})

		rec.GeoLocation = &route53.GeoLocation{
			ContinentCode:   nilString(geolocation["continent"].(string)),
			CountryCode:     nilString(geolocation["country"].(string)),
			SubdivisionCode: nilString(geolocation["subdivision"].(string)),
		}
	}

	if v, ok := d.GetOk("health_check_id"); ok {
		rec.HealthCheckId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("latency_routing_policy"); ok {
		records := v.([]interface{})
		latency := records[0].(map[string]interface{})

		rec.Region = aws.String(latency["region"].(string))
	}

	if v, ok := d.GetOk("multivalue_answer_routing_policy"); ok {
		rec.MultiValueAnswer = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("set_identifier"); ok {
		rec.SetIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weighted_routing_policy"); ok {
		records := v.([]interface{})
		weight := records[0].(map[string]interface{})

		rec.Weight = aws.Int64(int64(weight["weight"].(int)))
	}

	return rec
}

func FQDN(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	} else {
		return name + "."
	}
}

// Route 53 stores certain characters with the octal equivalent in ASCII format.
// This function converts all of these characters back into the original character.
// E.g. "*" is stored as "\\052" and "@" as "\\100"
func CleanRecordName(name string) string {
	str := name
	s, err := strconv.Unquote(`"` + str + `"`)
	if err != nil {
		return str
	}
	return s
}

// Check if the current record name contains the zone suffix.
// If it does not, add the zone name to form a fully qualified name
// and keep AWS happy.
func ExpandRecordName(name, zone string) string {
	rn := strings.ToLower(strings.TrimSuffix(name, "."))
	zone = strings.TrimSuffix(zone, ".")
	if !strings.HasSuffix(rn, zone) {
		if len(name) == 0 {
			rn = zone
		} else {
			rn = strings.Join([]string{rn, zone}, ".")
		}
	}
	return rn
}

// nilString takes a string as an argument and returns a string
// pointer. The returned pointer is nil if the string argument is
// empty. Otherwise, it is a pointer to a copy of the string.
func nilString(s string) *string {
	if s == "" {
		return nil
	}
	return aws.String(s)
}

func NormalizeAliasName(alias interface{}) string {
	output := strings.ToLower(alias.(string))
	return strings.TrimSuffix(output, ".")
}

func ParseRecordID(id string) [4]string {
	var recZone, recType, recName, recSet string
	parts := strings.Split(id, "_")
	if len(parts) > 1 {
		recZone = parts[0]
	}
	if len(parts) >= 3 {
		recTypeIndex := -1
		for i, maybeRecType := range parts[1:] {
			if validRecordType(maybeRecType) {
				recTypeIndex = i + 1
				break
			}
		}
		if recTypeIndex > 1 {
			recName = strings.Join(parts[1:recTypeIndex], "_")
			recName = strings.TrimSuffix(recName, ".")
			recType = parts[recTypeIndex]
			recSet = strings.Join(parts[recTypeIndex+1:], "_")
		}
	}
	return [4]string{recZone, recName, recType, recSet}
}

func validRecordType(s string) bool {
	for _, v := range route53.RRType_Values() {
		if v == s {
			return true
		}
	}
	return false
}
