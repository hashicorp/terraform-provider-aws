// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_record", name="Record")
func resourceRecord() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecordCreate,
		ReadWithoutTimeout:   resourceRecordRead,
		UpdateWithoutTimeout: resourceRecordUpdate,
		DeleteWithoutTimeout: resourceRecordDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := recordParseResourceID(d.Id())
				// We check that we have parsed the id into the correct number of segments.
				// We need at least 3 segments!
				// However, parts[1] can be the empty string if it is the root domain of the zone,
				// and isn't using a FQDN. See https://github.com/hashicorp/terraform-provider-aws/issues/4792
				if parts[0] == "" || parts[2] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected ZONEID_RECORDNAME_TYPE_SET-IDENTIFIER (e.g. Z4KAPRWWNC7JR_dev.example.com_NS_dev), where SET-IDENTIFIER is optional", d.Id())
				}

				d.Set("zone_id", parts[0])
				d.Set(names.AttrName, parts[1])
				d.Set(names.AttrType, parts[2])
				if parts[3] != "" {
					d.Set("set_identifier", parts[3])
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaVersion: 2,
		MigrateState:  recordMigrateState,

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"evaluate_target_health": {
							Type:     schema.TypeBool,
							Required: true,
						},
						names.AttrName: {
							Type:             schema.TypeString,
							Required:         true,
							StateFunc:        normalizeAliasName,
							DiffSuppressFunc: sdkv2.SuppressEquivalentStringCaseInsensitive,
							ValidateFunc:     validation.StringLenBetween(1, 1024),
						},
						"zone_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
						},
					},
				},
				ExactlyOneOf:  []string{names.AttrAlias, "records"},
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
					"geoproximity_routing_policy",
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
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceRecordSetFailover](),
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"geolocation_routing_policy",
					"geoproximity_routing_policy",
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
					"geoproximity_routing_policy",
					"latency_routing_policy",
					"multivalue_answer_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			"geoproximity_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bias": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(-99, 99),
						},
						"coordinates": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"latitude": {
										Type:     schema.TypeString,
										Required: true,
									},
									"longitude": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Optional: true,
						},
						"local_zone_group": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"geolocation_routing_policy",
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
						names.AttrRegion: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceRecordSetRegion](),
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"geolocation_routing_policy",
					"geoproximity_routing_policy",
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
					"geoproximity_routing_policy",
					"latency_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},
			names.AttrName: {
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
				ExactlyOneOf: []string{names.AttrAlias, "records"},
			},
			"set_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ttl": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{names.AttrAlias},
				RequiredWith:  []string{"records", "ttl"},
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RRType](),
			},
			"weighted_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrWeight: {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{
					"cidr_routing_policy",
					"failover_routing_policy",
					"geolocation_routing_policy",
					"geoproximity_routing_policy",
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
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	zoneID := cleanZoneID(d.Get("zone_id").(string))
	zoneRecord, err := findHostedZoneByID(ctx, conn, zoneID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone (%s): %s", zoneID, err)
	}

	// Protect existing DNS records which might be managed in another way.
	// Use UPSERT only if the overwrite flag is true or if the current action is an update
	// Else CREATE is used and fail if the same record exists.
	var action awstypes.ChangeAction
	if d.Get("allow_overwrite").(bool) || !d.IsNewResource() {
		action = awstypes.ChangeActionUpsert
	} else {
		action = awstypes.ChangeActionCreate
	}
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &awstypes.ChangeBatch{
			Changes: []awstypes.Change{
				{
					Action:            action,
					ResourceRecordSet: expandResourceRecordSet(d, aws.ToString(zoneRecord.HostedZone.Name)),
				},
			},
			Comment: aws.String("Managed by Terraform"),
		},
		HostedZoneId: aws.String(cleanZoneID(aws.ToString(zoneRecord.HostedZone.Id))),
	}

	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.NoSuchHostedZone](ctx, 1*time.Minute, func() (interface{}, error) {
		return conn.ChangeResourceRecordSets(ctx, input)
	})

	if v, ok := errs.As[*awstypes.InvalidChangeBatch](err); ok && len(v.Messages) > 0 {
		err = fmt.Errorf("%s: %w", v.ErrorCode(), errors.Join(tfslices.ApplyToAll(v.Messages, errors.New)...))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Record: %s", err)
	}

	vars := []string{
		zoneID,
		strings.ToLower(d.Get(names.AttrName).(string)),
		d.Get(names.AttrType).(string),
	}
	if v, ok := d.GetOk("set_identifier"); ok {
		vars = append(vars, v.(string))
	}
	d.SetId(strings.Join(vars, "_"))

	if output := outputRaw.(*route53.ChangeResourceRecordSetsOutput); output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Record (%s) synchronize: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRecordRead(ctx, d, meta)...)
}

func resourceRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	record, fqdn, err := findResourceRecordSetByFourPartKey(ctx, conn, cleanZoneID(d.Get("zone_id").(string)), d.Get(names.AttrName).(string), d.Get(names.AttrType).(string), d.Get("set_identifier").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Record (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Record (%s): %s", d.Id(), err)
	}

	if alias := record.AliasTarget; alias != nil {
		tfList := []interface{}{map[string]interface{}{
			"evaluate_target_health": alias.EvaluateTargetHealth,
			names.AttrName:           normalizeAliasName(aws.ToString(alias.DNSName)),
			"zone_id":                aws.ToString(alias.HostedZoneId),
		}}

		if err := d.Set(names.AttrAlias, tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting alias: %s", err)
		}
	}
	if cidrRoutingConfig := record.CidrRoutingConfig; cidrRoutingConfig != nil {
		tfList := []interface{}{map[string]interface{}{
			"collection_id": aws.ToString(cidrRoutingConfig.CollectionId),
			"location_name": aws.ToString(cidrRoutingConfig.LocationName),
		}}

		if err := d.Set("cidr_routing_policy", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting cidr_routing_policy: %s", err)
		}
	}
	if failover := record.Failover; failover != "" {
		tfList := []interface{}{map[string]interface{}{
			names.AttrType: failover,
		}}

		if err := d.Set("failover_routing_policy", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting failover_routing_policy: %s", err)
		}
	}
	d.Set("fqdn", fqdn)
	if geoLocation := record.GeoLocation; geoLocation != nil {
		tfList := []interface{}{map[string]interface{}{
			"continent":   aws.ToString(geoLocation.ContinentCode),
			"country":     aws.ToString(geoLocation.CountryCode),
			"subdivision": aws.ToString(geoLocation.SubdivisionCode),
		}}

		if err := d.Set("geolocation_routing_policy", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting geolocation_routing_policy: %s", err)
		}
	}
	if geoProximityLocation := record.GeoProximityLocation; geoProximityLocation != nil {
		tfList := []interface{}{map[string]interface{}{
			"aws_region":       aws.ToString(record.GeoProximityLocation.AWSRegion),
			"bias":             aws.ToInt32((record.GeoProximityLocation.Bias)),
			"coordinates":      flattenCoordinate(record.GeoProximityLocation.Coordinates),
			"local_zone_group": aws.ToString(record.GeoProximityLocation.LocalZoneGroup),
		}}

		if err := d.Set("geoproximity_routing_policy", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting geoproximity_routing_policy: %s", err)
		}
	}
	d.Set("health_check_id", record.HealthCheckId)
	if region := record.Region; region != "" {
		tfList := []interface{}{map[string]interface{}{
			names.AttrRegion: region,
		}}

		if err := d.Set("latency_routing_policy", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting latency_routing_policy: %s", err)
		}
	}
	d.Set("multivalue_answer_routing_policy", record.MultiValueAnswer)
	if err := d.Set("records", flattenResourceRecords(record.ResourceRecords, record.Type)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting records: %s", err)
	}

	d.Set("set_identifier", record.SetIdentifier)
	d.Set("ttl", record.TTL)
	if weight := record.Weight; weight != nil {
		tfList := []interface{}{map[string]interface{}{
			names.AttrWeight: aws.ToInt64((record.Weight)),
		}}

		if err := d.Set("weighted_routing_policy", tfList); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting weighted_routing_policy: %s", err)
		}
	}

	return diags
}

func resourceRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	// Route 53 supports CREATE, DELETE, and UPSERT actions. We use UPSERT, and
	// AWS dynamically determines if a record should be created or updated.
	// Amazon Route 53 can update an existing resource record set only when all
	// of the following values match: Name, Type and SetIdentifier
	// See http://docs.aws.amazon.com/Route53/latest/APIReference/API_ChangeResourceRecordSets.html.

	if !d.HasChange(names.AttrType) && !d.HasChange("set_identifier") {
		// If neither type nor set_identifier changed we use UPSERT,
		// for resource update here we simply fall through to
		// our resource create function.
		return append(diags, resourceRecordCreate(ctx, d, meta)...)
	}

	// Otherwise, we delete the existing record and create a new record within
	// a transactional change.
	zoneID := cleanZoneID(d.Get("zone_id").(string))
	zoneRecord, err := findHostedZoneByID(ctx, conn, zoneID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone (%s): %s", zoneID, err)
	}

	// Build the to be deleted record
	en := expandRecordName(d.Get(names.AttrName).(string), aws.ToString(zoneRecord.HostedZone.Name))
	oldRRType, _ := d.GetChange(names.AttrType)

	oldRec := &awstypes.ResourceRecordSet{
		Name: aws.String(en),
		Type: awstypes.RRType(oldRRType.(string)),
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
					oldRec.CidrRoutingConfig = &awstypes.CidrRoutingConfig{
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
					oldRec.Failover = awstypes.ResourceRecordSetFailover(v[names.AttrType].(string))
				}
			}
		}
	}

	if v, _ := d.GetChange("geolocation_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.GeoLocation = &awstypes.GeoLocation{
						ContinentCode:   nilString(v["continent"].(string)),
						CountryCode:     nilString(v["country"].(string)),
						SubdivisionCode: nilString(v["subdivision"].(string)),
					}
				}
			}
		}
	}

	if v, _ := d.GetChange("geoproximity_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.GeoProximityLocation = &awstypes.GeoProximityLocation{
						AWSRegion:      nilString(v["aws_region"].(string)),
						Bias:           aws.Int32(int32(v["bias"].(int))),
						Coordinates:    expandCoordinates(v["coordinates"].(*schema.Set).List()),
						LocalZoneGroup: nilString(v["local_zone_group"].(string)),
					}
				}
			}
		}
	}

	if v, _ := d.GetChange("latency_routing_policy"); v != nil {
		if o, ok := v.([]interface{}); ok {
			if len(o) == 1 {
				if v, ok := o[0].(map[string]interface{}); ok {
					oldRec.Region = awstypes.ResourceRecordSetRegion(v[names.AttrRegion].(string))
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
					oldRec.Weight = aws.Int64(int64(v[names.AttrWeight].(int)))
				}
			}
		}
	}

	if v, _ := d.GetChange("ttl"); v.(int) != 0 {
		oldRec.TTL = aws.Int64(int64(v.(int)))
	}

	// Resource records
	if v, _ := d.GetChange("records"); v != nil {
		if v.(*schema.Set).Len() > 0 {
			oldRec.ResourceRecords = expandResourceRecords(flex.ExpandStringValueSet(v.(*schema.Set)), awstypes.RRType(oldRRType.(string)))
		}
	}

	// Alias record
	if v, _ := d.GetChange(names.AttrAlias); v != nil {
		aliases := v.([]interface{})
		if len(aliases) == 1 {
			alias := aliases[0].(map[string]interface{})
			oldRec.AliasTarget = &awstypes.AliasTarget{
				DNSName:              aws.String(alias[names.AttrName].(string)),
				EvaluateTargetHealth: alias["evaluate_target_health"].(bool),
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

	// Delete the old and create the new records in a single batch.
	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &awstypes.ChangeBatch{
			Changes: []awstypes.Change{
				{
					Action:            awstypes.ChangeActionDelete,
					ResourceRecordSet: oldRec,
				},
				{
					Action:            awstypes.ChangeActionCreate,
					ResourceRecordSet: expandResourceRecordSet(d, aws.ToString(zoneRecord.HostedZone.Name)),
				},
			},
			Comment: aws.String("Managed by Terraform"),
		},
		HostedZoneId: aws.String(cleanZoneID(aws.ToString(zoneRecord.HostedZone.Id))),
	}

	output, err := conn.ChangeResourceRecordSets(ctx, input)

	if v, ok := errs.As[*awstypes.InvalidChangeBatch](err); ok && len(v.Messages) > 0 {
		err = fmt.Errorf("%s: %w", v.ErrorCode(), errors.Join(tfslices.ApplyToAll(v.Messages, errors.New)...))
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Record (%s): %s", d.Id(), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Record (%s) synchronize: %s", d.Id(), err)
		}
	}

	// Generate a new ID.
	vars := []string{
		zoneID,
		strings.ToLower(d.Get(names.AttrName).(string)),
		d.Get(names.AttrType).(string),
	}
	if v, ok := d.GetOk("set_identifier"); ok {
		vars = append(vars, v.(string))
	}
	d.SetId(strings.Join(vars, "_"))

	return append(diags, resourceRecordRead(ctx, d, meta)...)
}

func resourceRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	zoneID := cleanZoneID(d.Get("zone_id").(string))
	var name string
	// If we're dealing with a change of record name, but we're operating on the old, rather than
	// the new, resource, then we need to use the old name to find it (in order to delete it).
	if !d.IsNewResource() && d.HasChange(names.AttrName) {
		oldName, _ := d.GetChange(names.AttrName)
		name = oldName.(string)
	} else {
		name = d.Get(names.AttrName).(string)
	}
	rec, _, err := findResourceRecordSetByFourPartKey(ctx, conn, zoneID, name, d.Get(names.AttrType).(string), d.Get("set_identifier").(string))

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route 53 Record (%s): %s", d.Id(), err)
	}

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &awstypes.ChangeBatch{
			Changes: []awstypes.Change{
				{
					Action:            awstypes.ChangeActionDelete,
					ResourceRecordSet: rec,
				},
			},
			Comment: aws.String("Deleted by Terraform"),
		},
		HostedZoneId: aws.String(zoneID),
	}

	output, err := conn.ChangeResourceRecordSets(ctx, input)

	// Pre-AWS SDK for Go v2 migration compatibility.
	// https://github.com/hashicorp/terraform-provider-aws/issues/37806.
	if errs.IsA[*awstypes.InvalidChangeBatch](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Record (%s): %s", d.Id(), err)
	}

	if output.ChangeInfo != nil {
		if _, err := waitChangeInsync(ctx, conn, aws.ToString(output.ChangeInfo.Id)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Record (%s) synchronize: %s", d.Id(), err)
		}
	}

	return diags
}

func recordParseResourceID(id string) [4]string {
	var recZone, recType, recName, recSet string

	parts := strings.Split(id, "_")
	if len(parts) > 1 {
		recZone = parts[0]
	}
	if len(parts) >= 3 {
		recTypeIndex := -1
		for i, maybeRecType := range parts[1:] {
			if slices.Contains(enum.Values[awstypes.RRType](), maybeRecType) {
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

func findResourceRecordSetByFourPartKey(ctx context.Context, conn *route53.Client, zoneID, recordName, recordType, recordSetID string) (*awstypes.ResourceRecordSet, *string, error) {
	zone, err := findHostedZoneByID(ctx, conn, zoneID)

	if err != nil {
		return nil, nil, err
	}

	name := expandRecordName(recordName, aws.ToString(zone.HostedZone.Name))
	recordName = fqdn(strings.ToLower(name))
	rrType := awstypes.RRType(strings.ToUpper(recordType))
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(zoneID),
		StartRecordName: aws.String(recordName),
		StartRecordType: rrType,
	}
	if recordSetID == "" {
		input.MaxItems = aws.Int32(1)
	} else {
		input.MaxItems = aws.Int32(100)
	}
	output, err := findResourceRecordSet(ctx, conn, input, resourceRecordsFor(recordName, rrType), func(v *awstypes.ResourceRecordSet) bool {
		if recordName != strings.ToLower(cleanRecordName(aws.ToString(v.Name))) {
			return false
		}
		if recordType != strings.ToUpper(string(v.Type)) {
			return false
		}
		if recordSetID != aws.ToString(v.SetIdentifier) {
			return false
		}

		return true
	})

	if err != nil {
		return nil, nil, err
	}

	return output, &name, nil
}

func findResourceRecordSet(ctx context.Context, conn *route53.Client, input *route53.ListResourceRecordSetsInput, morePages tfslices.Predicate[*route53.ListResourceRecordSetsOutput], filter tfslices.Predicate[*awstypes.ResourceRecordSet]) (*awstypes.ResourceRecordSet, error) {
	output, err := findResourceRecordSets(ctx, conn, input, morePages, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourceRecordSets(ctx context.Context, conn *route53.Client, input *route53.ListResourceRecordSetsInput, morePages tfslices.Predicate[*route53.ListResourceRecordSetsOutput], filter tfslices.Predicate[*awstypes.ResourceRecordSet]) ([]awstypes.ResourceRecordSet, error) {
	var output []awstypes.ResourceRecordSet

	pages := route53.NewListResourceRecordSetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchHostedZone](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourceRecordSets {
			if filter(&v) {
				output = append(output, v)
			}
		}

		if !morePages(page) {
			break
		}
	}

	return output, nil
}

func resourceRecordsFor(recordName string, recordType awstypes.RRType) tfslices.Predicate[*route53.ListResourceRecordSetsOutput] {
	return func(page *route53.ListResourceRecordSetsOutput) bool {
		if page.IsTruncated {
			if strings.ToLower(cleanRecordName(aws.ToString(page.NextRecordName))) != recordName {
				return false
			}

			if page.NextRecordType != recordType {
				return false
			}
		}

		return true
	}
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

func expandResourceRecordSet(d *schema.ResourceData, zoneName string) *awstypes.ResourceRecordSet {
	// get expanded name
	en := expandRecordName(d.Get(names.AttrName).(string), zoneName)

	// Create the RecordSet request with the fully expanded name, e.g.
	// sub.domain.com. Route 53 requires a fully qualified domain name, but does
	// not require the trailing ".", which it will itself, so we don't call FQDN
	// here.
	rrType := awstypes.RRType(d.Get(names.AttrType).(string))
	apiObject := &awstypes.ResourceRecordSet{
		Name: aws.String(en),
		Type: rrType,
	}

	if v, ok := d.GetOk("ttl"); ok {
		apiObject.TTL = aws.Int64(int64(v.(int)))
	}

	// Resource records
	if v, ok := d.GetOk("records"); ok {
		apiObject.ResourceRecords = expandResourceRecords(flex.ExpandStringValueSet(v.(*schema.Set)), rrType)
	}

	// Alias record
	if v, ok := d.GetOk(names.AttrAlias); ok {
		aliases := v.([]interface{})
		alias := aliases[0].(map[string]interface{})
		apiObject.AliasTarget = &awstypes.AliasTarget{
			DNSName:              aws.String(alias[names.AttrName].(string)),
			EvaluateTargetHealth: alias["evaluate_target_health"].(bool),
			HostedZoneId:         aws.String(alias["zone_id"].(string)),
		}
	}

	if v, ok := d.GetOk("cidr_routing_policy"); ok {
		records := v.([]interface{})
		cidr := records[0].(map[string]interface{})

		apiObject.CidrRoutingConfig = &awstypes.CidrRoutingConfig{
			CollectionId: aws.String(cidr["collection_id"].(string)),
			LocationName: aws.String(cidr["location_name"].(string)),
		}
	}

	if v, ok := d.GetOk("failover_routing_policy"); ok {
		records := v.([]interface{})
		failover := records[0].(map[string]interface{})

		apiObject.Failover = awstypes.ResourceRecordSetFailover(failover[names.AttrType].(string))
	}

	if v, ok := d.GetOk("geolocation_routing_policy"); ok {
		geolocations := v.([]interface{})
		geolocation := geolocations[0].(map[string]interface{})

		apiObject.GeoLocation = &awstypes.GeoLocation{
			ContinentCode:   nilString(geolocation["continent"].(string)),
			CountryCode:     nilString(geolocation["country"].(string)),
			SubdivisionCode: nilString(geolocation["subdivision"].(string)),
		}
	}

	if v, ok := d.GetOk("geoproximity_routing_policy"); ok {
		geoproximityvalues := v.([]interface{})
		geoproximity := geoproximityvalues[0].(map[string]interface{})

		apiObject.GeoProximityLocation = &awstypes.GeoProximityLocation{
			AWSRegion:      nilString(geoproximity["aws_region"].(string)),
			Bias:           aws.Int32(int32(geoproximity["bias"].(int))),
			Coordinates:    expandCoordinates(geoproximity["coordinates"].(*schema.Set).List()),
			LocalZoneGroup: nilString(geoproximity["local_zone_group"].(string)),
		}
	}

	if v, ok := d.GetOk("health_check_id"); ok {
		apiObject.HealthCheckId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("latency_routing_policy"); ok {
		records := v.([]interface{})
		latency := records[0].(map[string]interface{})

		apiObject.Region = awstypes.ResourceRecordSetRegion(latency[names.AttrRegion].(string))
	}

	if v, ok := d.GetOk("multivalue_answer_routing_policy"); ok {
		apiObject.MultiValueAnswer = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("set_identifier"); ok {
		apiObject.SetIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weighted_routing_policy"); ok {
		records := v.([]interface{})
		weight := records[0].(map[string]interface{})

		apiObject.Weight = aws.Int64(int64(weight[names.AttrWeight].(int)))
	}

	return apiObject
}

// Check if the current record name contains the zone suffix.
// If it does not, add the zone name to form a fully qualified name
// and keep AWS happy.
func expandRecordName(name, zone string) string {
	rn := normalizeZoneName(name)
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

func expandCoordinates(tfList []interface{}) *awstypes.Coordinates {
	if len(tfList) == 0 {
		return nil
	}

	apiObject := &awstypes.Coordinates{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject.Latitude = aws.String(tfMap["latitude"].(string))
		apiObject.Longitude = aws.String(tfMap["longitude"].(string))
	}

	return apiObject
}

func flattenCoordinate(apiObject *awstypes.Coordinates) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}
	tfMap := map[string]interface{}{}

	if v := apiObject.Latitude; v != nil {
		tfMap["latitude"] = aws.ToString(v)
	}

	if v := apiObject.Longitude; v != nil {
		tfMap["longitude"] = aws.ToString(v)
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func expandResourceRecords(rrs []string, rrType awstypes.RRType) []awstypes.ResourceRecord {
	apiObjects := make([]awstypes.ResourceRecord, 0, len(rrs))

	for _, rr := range rrs {
		if rrType == awstypes.RRTypeTxt || rrType == awstypes.RRTypeSpf {
			rr = flattenTxtEntry(rr)
		}
		apiObjects = append(apiObjects, awstypes.ResourceRecord{Value: aws.String(rr)})
	}

	return apiObjects
}

func flattenResourceRecords(apiObjects []awstypes.ResourceRecord, rrType awstypes.RRType) []string {
	rrs := []string{}

	for _, apiObject := range apiObjects {
		if apiObject.Value != nil {
			rr := aws.ToString(apiObject.Value)
			if rrType == awstypes.RRTypeTxt || rrType == awstypes.RRTypeSpf {
				rr = expandTxtEntry(rr)
			}
			rrs = append(rrs, rr)
		}
	}

	return rrs
}

// How 'flattenTxtEntry' and 'expandTxtEntry' work.
//
// In the Route 53, TXT entries are written using quoted strings, one per line.
// Example:
//
//	"x=foo"
//	"bar=12"
//
// In Terraform, there are two differences:
// - We use a list of strings instead of separating strings with newlines.
// - Within each string, we dont' include the surrounding quotes.
// Example:
//
//	records = ["x=foo", "bar=12"]    # Instead of ["\"x=foo\", \"bar=12\""]
//
// When we pull from Route 53, `expandTxtEntry` removes the surrounding quotes;
// when we push to Route 53, `flattenTxtEntry` adds them back.
//
// One complication is that a single TXT entry can have multiple quoted strings.
// For example, here are two TXT entries, one with two quoted strings and the
// other with three.
//
//	"x=" "foo"
//	"ba" "r" "=12"
//
// DNS clients are expected to merge the quoted strings before interpreting the
// value.  Since `expandTxtEntry` only removes the quotes at the end we can still
// (hackily) represent the above configuration in Terraform:
//
//	records = ["x=\" \"foo", "ba\" \"r\" \"=12"]
//
// The primary reason to use multiple strings for an entry is that DNS (and Route
// 53) doesn't allow a quoted string to be more than 255 characters long.  If you
// want a longer TXT entry, you must use multiple quoted strings.
//
// It would be nice if this Terraform automatically split strings longer than 255
// characters.  For example, imagine "xxx..xxx" has 256 "x" characters.
//
//	records = ["xxx..xxx"]
//
// When pushing to Route 53, this could be converted to:
//
//	"xxx..xx" "x"
//
// This could also work when the user is already using multiple quoted strings:
//
//	records = ["xxx.xxx\" \"yyy..yyy"]
//
// When pushing to Route 53, this could be converted to:
//
//	"xxx..xx" "xyyy...y" "yy"
//
// If you want to add this feature, make sure to follow all the quoting rules in
// <https://tools.ietf.org/html/rfc1464#section-2>.  If you make a mistake, people
// might end up relying on that mistake so fixing it would be a breaking change.
func expandTxtEntry(s string) string {
	last := len(s) - 1
	if last != 0 && s[0] == '"' && s[last] == '"' {
		s = s[1:last]
	}
	return s
}

func flattenTxtEntry(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}
