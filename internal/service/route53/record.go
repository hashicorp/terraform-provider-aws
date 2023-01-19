package route53

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

var (
	errNoRecordsFound    = errors.New("No matching records found")
	errNoHostedZoneFound = errors.New("No matching Hosted Zone found")
)

func ResourceRecord() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecordCreate,
		ReadWithoutTimeout:   resourceRecordRead,
		UpdateWithoutTimeout: resourceRecordUpdate,
		DeleteWithoutTimeout: resourceRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 2,
		MigrateState:  RecordMigrateState,
		Schema: map[string]*schema.Schema{
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

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(route53.RRType_Values(), false),
			},

			"zone_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"ttl": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"alias"},
				RequiredWith:  []string{"records", "ttl"},
			},

			"set_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"alias": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ExactlyOneOf:  []string{"alias", "records"},
				ConflictsWith: []string{"ttl"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
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

						"evaluate_target_health": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},

			"failover_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ConflictsWith: []string{
					"geolocation_routing_policy",
					"latency_routing_policy",
					"weighted_routing_policy",
					"multivalue_answer_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
								value := v.(string)
								if value != "PRIMARY" && value != "SECONDARY" {
									es = append(es, fmt.Errorf("Failover policy type must be PRIMARY or SECONDARY"))
								}
								return
							},
						},
					},
				},
			},

			"latency_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ConflictsWith: []string{
					"failover_routing_policy",
					"geolocation_routing_policy",
					"weighted_routing_policy",
					"multivalue_answer_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"geolocation_routing_policy": { // AWS Geolocation
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ConflictsWith: []string{
					"failover_routing_policy",
					"latency_routing_policy",
					"weighted_routing_policy",
					"multivalue_answer_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
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
			},

			"weighted_routing_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ConflictsWith: []string{
					"failover_routing_policy",
					"geolocation_routing_policy",
					"latency_routing_policy",
					"multivalue_answer_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"weight": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},

			"multivalue_answer_routing_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				ConflictsWith: []string{
					"failover_routing_policy",
					"geolocation_routing_policy",
					"latency_routing_policy",
					"weighted_routing_policy",
				},
				RequiredWith: []string{"set_identifier"},
			},

			"health_check_id": { // ID of health check
				Type:     schema.TypeString,
				Optional: true,
			},

			"records": {
				Type:         schema.TypeSet,
				ExactlyOneOf: []string{"alias", "records"},
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				Set:          schema.HashString,
			},

			"allow_overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// Route 53 supports CREATE, DELETE, and UPSERT actions. We use UPSERT, and
	// AWS dynamically determines if a record should be created or updated.
	// Amazon Route 53 can update an existing resource record set only when all
	// of the following values match: Name, Type and SetIdentifier
	// See http://docs.aws.amazon.com/Route53/latest/APIReference/API_ChangeResourceRecordSets.html
	diags diag.Diagnostics

	if !d.HasChange("type") && !d.HasChange("set_identifier") {
		// If neither type nor set_identifier changed we use UPSERT,
		// for resource update here we simply fall through to
		// our resource create function.
		return append(diags, resourceRecordCreate(ctx, d, meta)...)
	}

	// Otherwise, we delete the existing record and create a new record within
	// a transactional change.
	conn := meta.(*conns.AWSClient).Route53Conn()
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
	// - failover_routing_policy
	// - geolocation_routing_policy
	// - latency_routing_policy
	// - multivalue_answer_routing_policy
	// - weighted_routing_policy

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
	rec := resourceRecordBuildSet(d, aws.StringValue(zoneRecord.HostedZone.Name))

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

	respRaw, err := ChangeRecordSet(ctx, conn, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route 53 Record (%s): updating record set: : %s", d.Id(), err)
	}

	changeInfo := respRaw.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo

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
		return sdkdiag.AppendErrorf(diags, "updating Route 53 Record (%s): updating record set: waiting for completion: %s", d.Id(), err)
	}

	if _, err := findRecord(ctx, d, meta); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route 53 Record (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn()
	zone := CleanZoneID(d.Get("zone_id").(string))

	zoneRecord, err := FindHostedZoneByID(ctx, conn, zone)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Record (%s): getting Hosted Zone (%s): %s", d.Id(), zone, err)
	}

	// Build the record
	rec := resourceRecordBuildSet(d, aws.StringValue(zoneRecord.HostedZone.Name))

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

	req := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(CleanZoneID(aws.StringValue(zoneRecord.HostedZone.Id))),
		ChangeBatch:  changeBatch,
	}

	log.Printf("[DEBUG] Creating resource records for zone: %s, name: %s\n\n%s",
		zone, aws.StringValue(rec.Name), req)

	respRaw, err := ChangeRecordSet(ctx, conn, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Record (%s): updating record set: %s", d.Id(), err)
	}

	changeInfo := respRaw.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo

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

	err = WaitForRecordSetToSync(ctx, conn, CleanChangeID(aws.StringValue(changeInfo.Id)))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Record (%s): updating record set: wating for completion: %s", d.Id(), err)
	}

	if _, err := findRecord(ctx, d, meta); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route 53 Record (%s): %s", d.Id(), err)
	}
	return diags
}

func ChangeRecordSet(ctx context.Context, conn *route53.Route53, input *route53.ChangeResourceRecordSetsInput) (interface{}, error) {
	var out *route53.ChangeResourceRecordSetsOutput
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.ChangeResourceRecordSetsWithContext(ctx, input)
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			log.Print("[DEBUG] Hosted Zone not found, retrying...")
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.ChangeResourceRecordSetsWithContext(ctx, input)
	}

	return out, err
}

func WaitForRecordSetToSync(ctx context.Context, conn *route53.Route53, requestId string) error {
	rand.Seed(time.Now().UTC().UnixNano())

	wait := resource.StateChangeConf{
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

func resourceRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var
	// If we don't have a zone ID, we're doing an import. Parse it from the ID.
	diags diag.Diagnostics

	if _, ok := d.GetOk("zone_id"); !ok {
		parts := ParseRecordID(d.Id())
		// We check that we have parsed the id into the correct number of segments.
		// We need at least 3 segments!
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return sdkdiag.AppendErrorf(diags, "importing aws_route_53 record. Please make sure the record ID is in the form ZONEID_RECORDNAME_TYPE_SET-IDENTIFIER (e.g. Z4KAPRWWNC7JR_dev.example.com_NS_dev), where SET-IDENTIFIER is optional")
		}

		d.Set("zone_id", parts[0])
		d.Set("name", parts[1])
		d.Set("type", parts[2])
		if parts[3] != "" {
			d.Set("set_identifier", parts[3])
		}
	}

	record, err := findRecord(ctx, d, meta)
	if err != nil {
		switch err {
		case errNoHostedZoneFound, errNoRecordsFound:
			log.Printf("[WARN] Route 53 Record (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		default:
			return sdkdiag.AppendErrorf(diags, "reading Route 53 Record (%s): %s", d.Id(), err)
		}
	}

	err = d.Set("records", FlattenResourceRecords(record.ResourceRecords, aws.StringValue(record.Type)))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting records: %s", err)
	}

	if alias := record.AliasTarget; alias != nil {
		name := NormalizeAliasName(aws.StringValue(alias.DNSName))
		d.Set("alias", []interface{}{
			map[string]interface{}{
				"zone_id":                aws.StringValue(alias.HostedZoneId),
				"name":                   name,
				"evaluate_target_health": aws.BoolValue(alias.EvaluateTargetHealth),
			},
		})
	}

	d.Set("ttl", record.TTL)

	if record.Failover != nil {
		v := []map[string]interface{}{{
			"type": aws.StringValue(record.Failover),
		}}
		if err := d.Set("failover_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting failover records for: %s, error: %s", d.Id(), err)
		}
	}

	if record.GeoLocation != nil {
		v := []map[string]interface{}{{
			"continent":   aws.StringValue(record.GeoLocation.ContinentCode),
			"country":     aws.StringValue(record.GeoLocation.CountryCode),
			"subdivision": aws.StringValue(record.GeoLocation.SubdivisionCode),
		}}
		if err := d.Set("geolocation_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting gelocation records for: %s, error: %s", d.Id(), err)
		}
	}

	if record.Region != nil {
		v := []map[string]interface{}{{
			"region": aws.StringValue(record.Region),
		}}
		if err := d.Set("latency_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting latency records for: %s, error: %s", d.Id(), err)
		}
	}

	if record.Weight != nil {
		v := []map[string]interface{}{{
			"weight": aws.Int64Value((record.Weight)),
		}}
		if err := d.Set("weighted_routing_policy", v); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting weighted records for: %s, error: %s", d.Id(), err)
		}
	}

	if err := d.Set("multivalue_answer_routing_policy", record.MultiValueAnswer); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting multivalue answer records for: %s, error: %s", d.Id(), err)
	}

	d.Set("set_identifier", record.SetIdentifier)
	d.Set("health_check_id", record.HealthCheckId)

	return diags
}

// findRecord takes a ResourceData struct for aws_resource_route53_record. It
// uses the referenced zone_id to query Route53 and find information on its
// records.
//
// If records are found, it returns the matching
// route53.ResourceRecordSet and nil for the error.
//
// If no hosted zone is found, it returns a nil recordset and errNoHostedZoneFound
// error.
//
// If no matching recordset is found, it returns nil and a errNoRecordsFound
// error.
//
// If there are other errors, it returns a nil recordset and passes on the
// error.
func findRecord(ctx context.Context, d *schema.ResourceData, meta interface{}) (*route53.ResourceRecordSet, error) {
	conn := meta.(*conns.AWSClient).Route53Conn()
	// Scan for a
	zone := CleanZoneID(d.Get("zone_id").(string))

	// get expanded name
	zoneRecord, err := conn.GetHostedZoneWithContext(ctx, &route53.GetHostedZoneInput{Id: aws.String(zone)})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			return nil, errNoHostedZoneFound
		}

		return nil, err
	}

	var name string
	// If we're dealing with a change of record name, but we're operating on the old, rather than
	// the new, resource, then we need to use the old name to find it (in order to delete it).
	if !d.IsNewResource() && d.HasChange("name") {
		oldName, _ := d.GetChange("name")
		name = oldName.(string)
	} else {
		name = d.Get("name").(string)
	}

	en := ExpandRecordName(name, aws.StringValue(zoneRecord.HostedZone.Name))
	log.Printf("[DEBUG] Expanded record name: %s", en)
	d.Set("fqdn", en)

	recordName := FQDN(strings.ToLower(en))
	recordType := d.Get("type").(string)
	recordSetIdentifier := d.Get("set_identifier").(string)

	// If this isn't a Weighted, Latency, Geo, or Failover resource with
	// a SetIdentifier we only need to look at the first record in the response since there can be
	// only one
	maxItems := "1"
	if recordSetIdentifier != "" {
		maxItems = "100"
	}

	lopts := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(CleanZoneID(zone)),
		StartRecordName: aws.String(recordName),
		StartRecordType: aws.String(recordType),
		MaxItems:        aws.String(maxItems),
	}

	log.Printf("[DEBUG] List resource records sets for zone: %s, opts: %s",
		zone, lopts)

	var record *route53.ResourceRecordSet

	// We need to loop over all records starting from the record we are looking for because
	// Weighted, Latency, Geo, and Failover resource record sets have a special option
	// called SetIdentifier which allows multiple entries with the same name and type but
	// a different SetIdentifier.
	// For all other records we are setting the maxItems to 1 so that we don't return extra
	// unneeded records.
	err = conn.ListResourceRecordSetsPagesWithContext(ctx, lopts, func(resp *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		for _, recordSet := range resp.ResourceRecordSets {
			responseName := strings.ToLower(CleanRecordName(*recordSet.Name))
			responseType := strings.ToUpper(aws.StringValue(recordSet.Type))

			if recordName != responseName {
				continue
			}
			if recordType != responseType {
				continue
			}
			if aws.StringValue(recordSet.SetIdentifier) != recordSetIdentifier {
				continue
			}

			record = recordSet
			return false
		}

		nextRecordName := strings.ToLower(CleanRecordName(aws.StringValue(resp.NextRecordName)))
		nextRecordType := strings.ToUpper(aws.StringValue(resp.NextRecordType))

		if nextRecordName != recordName {
			return false
		}

		if nextRecordType != recordType {
			return false
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errNoRecordsFound
	}
	return record, nil
}

func resourceRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn()
	// Get the records
	rec, err := findRecord(ctx, d, meta)
	if err != nil {
		switch err {
		case errNoHostedZoneFound, errNoRecordsFound:
			return diags
		default:
			return sdkdiag.AppendErrorf(diags, "deleting Route 53 Record (%s): %s", d.Id(), err)
		}
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

	zone := CleanZoneID(d.Get("zone_id").(string))

	req := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(CleanZoneID(zone)),
		ChangeBatch:  changeBatch,
	}

	respRaw, err := DeleteRecordSet(ctx, conn, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route 53 Record (%s): deleting record set: %s", d.Id(), err)
	}

	changeInfo := respRaw.(*route53.ChangeResourceRecordSetsOutput).ChangeInfo
	if changeInfo == nil {
		log.Printf("[INFO] No ChangeInfo Found. Waiting for Sync not required")
		return diags
	}

	if err := WaitForRecordSetToSync(ctx, conn, CleanChangeID(aws.StringValue(changeInfo.Id))); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route 53 Record (%s): deleting record set: waiting for completion: %s", d.Id(), err)
	}
	return diags
}

func DeleteRecordSet(ctx context.Context, conn *route53.Route53, input *route53.ChangeResourceRecordSetsInput) (interface{}, error) {
	out, err := conn.ChangeResourceRecordSetsWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, route53.ErrCodeInvalidChangeBatch) {
		return out, nil
	}

	return out, err
}

func resourceRecordBuildSet(d *schema.ResourceData, zoneName string) *route53.ResourceRecordSet {
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

	if v, ok := d.GetOk("failover_routing_policy"); ok {
		records := v.([]interface{})
		failover := records[0].(map[string]interface{})

		rec.Failover = aws.String(failover["type"].(string))
	}

	if v, ok := d.GetOk("health_check_id"); ok {
		rec.HealthCheckId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weighted_routing_policy"); ok {
		records := v.([]interface{})
		weight := records[0].(map[string]interface{})

		rec.Weight = aws.Int64(int64(weight["weight"].(int)))
	}

	if v, ok := d.GetOk("set_identifier"); ok {
		rec.SetIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("latency_routing_policy"); ok {
		records := v.([]interface{})
		latency := records[0].(map[string]interface{})

		rec.Region = aws.String(latency["region"].(string))
	}

	if v, ok := d.GetOk("geolocation_routing_policy"); ok {
		geolocations := v.([]interface{})
		geolocation := geolocations[0].(map[string]interface{})

		rec.GeoLocation = &route53.GeoLocation{
			ContinentCode:   nilString(geolocation["continent"].(string)),
			CountryCode:     nilString(geolocation["country"].(string)),
			SubdivisionCode: nilString(geolocation["subdivision"].(string)),
		}
		log.Printf("[DEBUG] Creating geolocation: %#v", geolocation)
	}

	if v, ok := d.GetOk("multivalue_answer_routing_policy"); ok {
		rec.MultiValueAnswer = aws.Bool(v.(bool))
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
		var recTypeIndex int = -1
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
