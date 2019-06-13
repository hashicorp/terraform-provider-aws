// This is a Fugue-specific read-only resource type that just grabs all
// information from the AWS IAM Credential Report.

package aws

import (
	"bytes"
	"log"
	"time"

	"encoding/csv"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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
						"access_key_1_active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"access_key_1_last_used_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_key_1_last_rotated": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_key_2_active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"access_key_2_last_used_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"access_key_2_last_rotated": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsIamCredentialReportUpdate(d *schema.ResourceData, meta interface{}) error {
	d.SetId("iam-credential-report")
	return resourceAwsIamCredentialReportRead(d, meta)
}

func resourceAwsIamCredentialReportRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	// Send a request to generate a credential report.
	generateReportInput := &iam.GenerateCredentialReportInput{}
	if _, err := iamconn.GenerateCredentialReport(generateReportInput); err != nil {
		return err
	}

	return resource.Retry(time.Duration(1)*time.Minute, func() *resource.RetryError {
		// Prepare a request to actually get the credential report.
		getReportInput := &iam.GetCredentialReportInput{}
		getReportOutput, err := iamconn.GetCredentialReport(getReportInput)
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

		// Store report in the resource state.
		d.Set("report", flattenCredentialReport(report))

		return nil
	})
}

func resourceAwsIamCredentialReportDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

type CredentialReport = []ReportRow

type ReportRow struct {
	User                   string
	PasswordEnabled        bool
	PasswordLastUsed       string
	PasswordLastChanged    string
	MfaActive              bool
	AccessKey1Active       bool
	AccessKey1LastUsedDate string
	AccessKey1LastRotated  string
	AccessKey2Active       bool
	AccessKey2LastUsedDate string
	AccessKey2LastRotated  string
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
	rows := make([]ReportRow, len(lines))
	for i, line := range lines {
		rows[i] = ReportRow{
			User:                   line[header["user"]],
			PasswordEnabled:        parseCsvBool(line[header["password_enabled"]]),
			PasswordLastUsed:       line[header["password_last_used"]],
			PasswordLastChanged:    line[header["password_last_changed"]],
			MfaActive:              parseCsvBool(line[header["mfa_active"]]),
			AccessKey1Active:       parseCsvBool(line[header["access_key_1_active"]]),
			AccessKey1LastUsedDate: line[header["access_key_1_last_used_date"]],
			AccessKey1LastRotated:  line[header["access_key_1_last_rotated"]],
			AccessKey2Active:       parseCsvBool(line[header["access_key_2_active"]]),
			AccessKey2LastUsedDate: line[header["access_key_2_last_used_date"]],
			AccessKey2LastRotated:  line[header["access_key_2_last_rotated"]],
		}
	}

	return rows, nil
}

func parseCsvBool(csv string) bool {
	return csv == "true"
}

func flattenCredentialReport(report CredentialReport) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(report))
	for _, row := range report {
		r := map[string]interface{}{
			"user":                        row.User,
			"password_enabled":            row.PasswordEnabled,
			"password_last_used":          row.PasswordLastUsed,
			"password_last_changed":       row.PasswordLastChanged,
			"mfa_active":                  row.MfaActive,
			"access_key_1_active":         row.AccessKey1Active,
			"access_key_1_last_used_date": row.AccessKey1LastUsedDate,
			"access_key_1_last_rotated":   row.AccessKey1LastRotated,
			"access_key_2_active":         row.AccessKey2Active,
			"access_key_2_last_used_date": row.AccessKey2LastUsedDate,
			"access_key_2_last_rotated":   row.AccessKey2LastRotated,
		}
		result = append(result, r)
	}
	return result
}
