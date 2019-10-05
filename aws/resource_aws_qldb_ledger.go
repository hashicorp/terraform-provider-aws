package aws

import (
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-sdk-go/service/qldb"
    "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsQLDBLedger() *schema.Resource {
    return &schema.Resource{
        Create: resourceAwsQLDBLedgerCreate,
        Read:   resourceAwsQLDBLedgerRead,
        Update: resourceAwsQLDBLedgerUpdate,
        Delete: resourceAwsQLDBLedgerDelete,
        Importer: &schema.ResourceImporter{
            State: resourceAwsQLDBLedgerImport,
        },

        SchemaVersion: 0,

        Schema: map[string]*schema.Schema{
            "arn": {
                Type:     schema.TypeString,
                Computed: true,
            },

            "name": {
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
                ForceNew: true,
            },

            "name_prefix": {
                Type:          schema.TypeString,
                Optional:      true,
                ForceNew:      true,
                ConflictsWith: []string{"name"},
            },

            "permissions_mode": {
                Type:         schema.TypeString,
                Optional:     true,
                ForceNew:     true,
                ValidateFunc: validation.StringInSlice([]string{qldb.PermissionsModeAllowAll}, false),
                Default:      qldb.PermissionsModeAllowAll, // Delete this line when AWS will support fetching permissions_mode
            },

            "deletion_protection": {
                Type:     schema.TypeBool,
                Optional: true,
                ForceNew: false,
                Default:  true,
            },

            "tags": tagsSchema(),
        },
    }
}

func resourceAwsQLDBLedgerCreate(d *schema.ResourceData, meta interface{}) error {
    conn := meta.(*AWSClient).qldbconn

    var name string
    if v, ok := d.GetOk("name"); ok {
        name = v.(string)
    } else if v, ok := d.GetOk("name_prefix"); ok {
        name = resource.PrefixedUniqueId(v.(string))
    } else {
        name = resource.UniqueId()
    }

    if len(name) > 31 {
        log.Printf("[DEBUG] QLDB Ledger name is bigger than 32 characters, getting the first 32 characters %v -> %v", name, name[:32])
        name = name[:32]
    }

    d.Set("name", name)

    // Create the QLDB Ledger
    createOpts := &qldb.CreateLedgerInput{
        Name:               aws.String(d.Get("name").(string)),
        PermissionsMode:    aws.String(d.Get("permissions_mode").(string)),
        DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
        Tags:               tagsFromMapQLDBCreate(d.Get("tags").(map[string]interface{})),
    }

    //createOpts.

    log.Printf("[DEBUG] QLDB Ledger create config: %#v", *createOpts)
    qldbResp, err := conn.CreateLedger(createOpts)
    if err != nil {
        return fmt.Errorf("Error creating QLDB Ledger: %s", err)
    }

    // Set QLDB ledger name
    d.SetId(*qldbResp.Name)

    log.Printf("[INFO] QLDB Ledger name: %s", d.Id())

    // Update our attributes and return
    return resourceAwsQLDBLedgerRead(d, meta)
}

func resourceAwsQLDBLedgerRead(d *schema.ResourceData, meta interface{}) error {
    conn := meta.(*AWSClient).qldbconn

    // Refresh the QLDB state
    qldbRaw, _, err := QLDBLedgerStateRefreshFunc(conn, d.Id())()
    if err != nil {
        return err
    }
    if qldbRaw == nil {
        d.SetId("")
        return nil
    }

    // QLDB stuff
    qldbLedger := qldbRaw.(*qldb.DescribeLedgerOutput)
    d.Set("name", qldbLedger.Name)
    d.Set("deletion_protection", aws.BoolValue(qldbLedger.DeletionProtection))

    // ARN
    d.Set("arn", qldbLedger.Arn)

    // Setting the permissions_mode manually because GO AWS SDK currently
    // does not return the set permissions_mode
    d.Set("permissions_mode", qldb.PermissionsModeAllowAll)

    return nil
}

func resourceAwsQLDBLedgerUpdate(d *schema.ResourceData, meta interface{}) error {
    conn := meta.(*AWSClient).qldbconn

    // Turn on partial mode
    d.Partial(true)

    if d.HasChange("deletion_protection") {
        val := d.Get("deletion_protection").(bool)
        modifyOpts := &qldb.UpdateLedgerInput{
            Name:               aws.String(d.Id()),
            DeletionProtection: aws.Bool(val),
        }
        log.Printf(
            "[INFO] Modifying deletion_protection QLDB attribute for %s: %#v",
            d.Id(), modifyOpts)
        if _, err := conn.UpdateLedger(modifyOpts); err != nil {

            return err
        }

        d.SetPartial("deletion_protection")
    }

    if err := setTagsQLDB(conn, d); err != nil {
        return err
    } else {
        d.SetPartial("tags")
    }

    d.Partial(false)
    return resourceAwsQLDBLedgerRead(d, meta)
}

func resourceAwsQLDBLedgerDelete(d *schema.ResourceData, meta interface{}) error {
    conn := meta.(*AWSClient).qldbconn
    deleteLedgerOpts := &qldb.DeleteLedgerInput{
        Name: aws.String(d.Id()),
    }
    log.Printf("[INFO] Deleting QLDB Ledger: %s", d.Id())

    err := resource.Retry(5*time.Minute, func() *resource.RetryError {
        _, err := conn.DeleteLedger(deleteLedgerOpts)
        if err == nil {
            return nil
        }

        qldberr, ok := err.(awserr.Error)

        if ok && strings.Contains(qldberr.Code(), ".NotFound") {
            return nil
        }
        if isAWSErr(err, "DependencyViolation", "") {
            return resource.RetryableError(err)
        }
        return resource.NonRetryableError(fmt.Errorf("Error deleting QLDB Ledger: %s", err))
    })
    if isResourceTimeoutError(err) {
        _, err = conn.DeleteLedger(deleteLedgerOpts)
        if err != nil {
            qldberr, ok := err.(awserr.Error)
            if !ok && !strings.Contains(qldberr.Code(), ".NotFound") {
                return err
            }
        }
        return nil
    }

    if err != nil {
        return fmt.Errorf("Error deleting QLDB Ledger: %s", err)
    }
    return nil
}

// QLDBLedgerStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// a QLDB Ledger.
func QLDBLedgerStateRefreshFunc(conn *qldb.QLDB, id string) resource.StateRefreshFunc {
    return func() (interface{}, string, error) {
        describeQLDBOpts := &qldb.DescribeLedgerInput{
            Name: aws.String(id),
        }
        resp, err := conn.DescribeLedger(describeQLDBOpts)
        if err != nil {
            if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "ResourceNotFoundException" {
                resp = nil
            } else {
                log.Printf("Error on QLDBLedgerStateRefresh: %s", err)
                return nil, "", err
            }
        }

        if resp == nil {
            // Sometimes AWS just has consistency issues and doesn't see
            // our instance yet. Return an empty state.
            return nil, "", nil
        }

        return resp, *resp.State, nil
    }
}

func resourceAwsQLDBLedgerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
    return []*schema.ResourceData{d}, nil
}
