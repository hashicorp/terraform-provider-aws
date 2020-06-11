// This is a Fugue-specific read-only resource type that just grabs all
// information from the AWS IAM Credential Report.

package aws

import (
	"bytes"
	"log"
	"regexp"
	"time"

	"encoding/csv"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsIamCredentialReport() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamCredentialReportUpdate,
		Read:   resourceAwsIamCredentialReportRead,
		Update: resourceAwsIamCredentialReportUpdate,
		Delete: resourceAwsIamCredentialReportDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"report": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"password_last_used": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_last_changed": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mfa_active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"mfa_virtual": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"access_keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"active": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"last_used_date": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"last_rotated": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsIamCredentialReportUpdate(d *schema.ResourceData, meta interface{}) error {
	// RM-2550: (eric-luminal) Fugue sets the ID of this resource as something
	// more human readable
	d.SetId("CredentialReport")
	return resourceAwsIamCredentialReportRead(d, meta)
}

func resourceAwsIamCredentialReportRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	// Send a request to generate a credential report.
	if _, err := iamconn.GenerateCredentialReport(nil); err != nil {
		return err
	}

	return resource.Retry(time.Duration(1)*time.Minute, func() *resource.RetryError {
		// Prepare a request to actually get the credential report.
		getReportOutput, err := iamconn.GetCredentialReport(nil)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				switch awserr.Code() {
				// Retry if it is still being generated.
				case "ReportInProgress":
					return resource.RetryableError(awserr)
				}
			}
			return resource.NonRetryableError(err)
		}

		// Parse report.
		log.Printf("[INFO]: Credential Report Content: %s", string(getReportOutput.Content))
		report, err := parseCsvCredentialReport(getReportOutput.Content)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		// Retrieve info about virtual MFA devices.
		listMfaOutput, err := iamconn.ListVirtualMFADevices(nil)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		// Run through the virtual MFA devices to create a set of users that
		// have them enabled.  The user names are constructed to match those in
		// the credential report.
		accountsWithVirtualMfa := map[string]bool{}
		serial, _ := regexp.Compile("^arn:aws:iam::[0-9]+:mfa/(.*)$")
		for _, virtualMfa := range listMfaOutput.VirtualMFADevices {
			match := serial.FindStringSubmatch(*virtualMfa.SerialNumber)
			if match != nil && len(match) > 1 {
				accountName := match[1]
				if accountName == "root-account-mfa-device" {
					accountName = "<root_account>"
				}

				accountsWithVirtualMfa[accountName] = true
			}
		}

		// Extend the report with the virtual MFA info.
		for _, row := range report {
			if _, ok := accountsWithVirtualMfa[row.User]; ok {
				row.MfaVirtual = true
			}
		}

		// Store report in the resource state.
		d.Set("report", flattenCredentialReport(report))

		return nil
	})
}

func resourceAwsIamCredentialReportDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

type CredentialReport = []*ReportRow

type ReportRow struct {
	User                string
	PasswordEnabled     bool
	PasswordLastUsed    string
	PasswordLastChanged string
	MfaActive           bool
	MfaVirtual          bool
	AccessKeys          []AccessKey
}

type AccessKey struct {
	Active       bool
	LastUsedDate string
	LastRotated  string
}

func parseCsvCredentialReport(content []byte) (CredentialReport, error) {
	reader := csv.NewReader(bytes.NewReader(content))

	// Parse header.
	header := map[string]int{}
	headerLine, err := reader.Read()
	if err != nil {
		return nil, err
	}
	for i, k := range headerLine {
		header[k] = i
	}

	// Parse rows into CSV.
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Copy rows into the datatype.
	rows := make([]*ReportRow, len(lines))
	for i, line := range lines {
		rows[i] = &ReportRow{
			User:                line[header["user"]],
			PasswordEnabled:     parseCsvBool(line[header["password_enabled"]]),
			PasswordLastUsed:    line[header["password_last_used"]],
			PasswordLastChanged: line[header["password_last_changed"]],
			MfaActive:           parseCsvBool(line[header["mfa_active"]]),
			AccessKeys: []AccessKey{
				{
					Active:       parseCsvBool(line[header["access_key_1_active"]]),
					LastUsedDate: line[header["access_key_1_last_used_date"]],
					LastRotated:  line[header["access_key_1_last_rotated"]],
				},
				{
					Active:       parseCsvBool(line[header["access_key_2_active"]]),
					LastUsedDate: line[header["access_key_2_last_used_date"]],
					LastRotated:  line[header["access_key_2_last_rotated"]],
				},
			},
		}
	}

	return rows, nil
}

func parseCsvBool(csv string) bool {
	return csv == "true"
}

func flattenCredentialReport(report CredentialReport) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, row := range report {
		m := map[string]interface{}{
			"user":                  row.User,
			"password_enabled":      row.PasswordEnabled,
			"password_last_used":    row.PasswordLastUsed,
			"password_last_changed": row.PasswordLastChanged,
			"mfa_active":            row.MfaActive,
			"mfa_virtual":           row.MfaVirtual,
			"access_keys":           flattenAccessKeys(row.AccessKeys),
		}
		out = append(out, m)
	}
	return out
}

func flattenAccessKeys(accessKeys []AccessKey) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	for _, accessKey := range accessKeys {
		m := map[string]interface{}{
			"active":         accessKey.Active,
			"last_used_date": accessKey.LastUsedDate,
			"last_rotated":   accessKey.LastRotated,
		}
		out = append(out, m)
	}
	return out
}
