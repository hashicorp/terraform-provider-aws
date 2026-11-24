// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package permissions

import "strings"

// boto3ClientAliases maps AWS SDK for Go v2 package names, as used in sdkmap's
// sdk_calls, to the Boto3 client names used in service reference data. Most
// packages need no entry: they match their Boto3 client name once dashes are
// removed, e.g. "ssoadmin" and "sso-admin". The rest are listed here.
var boto3ClientAliases = map[string]string{
	"cloudwatchlogs":                  "logs",
	"cognitoidentityprovider":         "cognito-idp",
	"configservice":                   "config",
	"costandusagereportservice":       "cur",
	"costexplorer":                    "ce",
	"databasemigrationservice":        "dms",
	"directoryservice":                "ds",
	"directoryservicedata":            "ds-data",
	"elasticloadbalancing":            "elb",
	"elasticloadbalancingv2":          "elbv2",
	"elasticsearchservice":            "es",
	"eventbridge":                     "events",
	"lexmodelbuildingservice":         "lex-models",
	"lexmodelsv2":                     "lexv2-models",
	"serverlessapplicationrepository": "serverlessrepo",
	"sfn":                             "stepfunctions",
}

// boto3Client returns the normalized Boto3 client name for an AWS SDK for Go v2
// package name.
func boto3Client(goSDKPackage string) string {
	if alias, ok := boto3ClientAliases[goSDKPackage]; ok {
		return normalizeClient(alias)
	}
	return normalizeClient(goSDKPackage)
}

// normalizeClient removes the dashes from a Boto3 client name, e.g.
// "sso-admin" becomes "ssoadmin", so it compares equal to the Go package name.
func normalizeClient(name string) string {
	return strings.ReplaceAll(name, "-", "")
}
