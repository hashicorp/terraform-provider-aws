// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/names/data"
)

const (
	lineOffset = 2 // 1 for skipping header line + 1 to translate from 0-based to 1-based index
)

// DocPrefix tests/column needs to be reworked for compatibility with tfproviderdocs
type DocPrefix struct {
	HumanFriendly  string
	DocPrefixRegex []string
	ResourceRegex  string
}

var allDocs int // currently skipping this test
var allChecks int

func main() {
	fmt.Println("Checking service data")

	data, err := data.ReadAllServiceData()

	if err != nil {
		log.Fatalf("error reading service data: %s", err)
	}

	docPrefixes := []DocPrefix{} // test to be reworked

	for i, l := range data {
		if l.HumanFriendly() == "" {
			log.Fatalf("in service data, line %d, HumanFriendly cannot be blank", i+lineOffset)
		}

		// TODO: Check for duplicates in HumanFriendly, ProviderPackageActual,
		// ProviderPackageCorrect, ProviderNameUpper, GoV1ClientTypeName,
		// ResourcePrefixActual, ResourcePrefixCorrect, FilePrefix, DocPrefix

		if l.AWSCLIV2Command() != "" && strings.Replace(l.AWSCLIV2Command(), "-", "", -1) != l.AWSCLIV2CommandNoDashes() {
			log.Fatalf("in service data, line %d, for service %s, AWSCLIV2CommandNoDashes must be the same as AWSCLIV2Command without dashes (%s)", i+lineOffset, l.HumanFriendly(), strings.Replace(l.AWSCLIV2Command(), "-", "", -1))
		}

		if l.ProviderPackageCorrect() != "" && l.AWSCLIV2CommandNoDashes() != "" && l.GoV2Package() != "" {
			if len(l.AWSCLIV2CommandNoDashes()) < len(l.GoV2Package()) && l.ProviderPackageCorrect() != l.AWSCLIV2CommandNoDashes() {
				log.Fatalf("in service data, line %d, for service %s, ProviderPackageCorrect must be shorter of AWSCLIV2CommandNoDashes (%s) and GoV2Package (%s)", i+lineOffset, l.HumanFriendly(), l.AWSCLIV2CommandNoDashes(), l.GoV2Package())
			}

			if len(l.AWSCLIV2CommandNoDashes()) > len(l.GoV2Package()) && l.ProviderPackageCorrect() != l.GoV2Package() {
				log.Fatalf("in service data, line %d, for service %s, ProviderPackageCorrect must be shorter of AWSCLIV2CommandNoDashes (%s) and GoV2Package (%s)", i+lineOffset, l.HumanFriendly(), l.AWSCLIV2CommandNoDashes(), l.GoV2Package())
			}
		}

		if l.AWSCLIV2CommandNoDashes() == "" && l.GoV2Package() == "" && !l.Exclude() {
			log.Fatalf("in service data, line %d, for service %s, if Exclude is blank, either AWSCLIV2CommandNoDashes or GoV2Package must have values", i+lineOffset, l.HumanFriendly())
		}

		packageToUse := l.ProviderPackageCorrect()

		if l.ProviderPackageActual() != "" {
			packageToUse = l.ProviderPackageActual()
		}

		if l.ResourcePrefixCorrect() != "" && l.ResourcePrefixCorrect() != fmt.Sprintf("aws_%s_", l.ProviderPackageCorrect()) {
			log.Fatalf("in service data, line %d, for service %s, ResourcePrefixCorrect should be aws_<package>_, where <package> is ProviderPackageCorrect", i+lineOffset, l.HumanFriendly())
		}

		if p := l.Aliases(); len(p) > 0 && packageToUse != "" {
			for _, v := range p {
				if v == packageToUse {
					log.Fatalf("in service data, line %d, for service %s, Aliases should not include ProviderPackageActual, if not blank, or ProviderPackageCorrect, if not blank and ProviderPackageActual is blank", i+lineOffset, l.HumanFriendly())
				}
			}
		}

		if !l.ClientSDKV1() && !l.ClientSDKV2() && !l.Exclude() {
			log.Fatalf("in service data, line %d, for service %s, at least one of ClientSDKV1 or ClientSDKV2 must have a value if Exclude is blank", i+lineOffset, l.HumanFriendly())
		}

		if l.ClientSDKV1() && (l.GoV1Package() == "" || l.GoV1ClientTypeName() == "") {
			log.Fatalf("in service data, line %d, for service %s, SDKVersion is set to 1 so neither GoV1Package nor GoV1ClientTypeName can be blank", i+lineOffset, l.HumanFriendly())
		}

		if l.ClientSDKV2() && l.GoV2Package() == "" {
			log.Fatalf("in service data, line %d, for service %s, SDKVersion is set to 2 so GoV2Package cannot be blank", i+lineOffset, l.HumanFriendly())
		}

		if l.ResourcePrefixCorrect() == "" && !l.Exclude() {
			log.Fatalf("in service data, line %d, for service %s, ResourcePrefixCorrect must have a value if Exclude is blank", i+lineOffset, l.HumanFriendly())
		}

		if l.ResourcePrefixCorrect() != "" && l.ResourcePrefixActual() == l.ResourcePrefixCorrect() {
			log.Fatalf("in service data, line %d, for service %s, ResourcePrefixActual should not be the same as ResourcePrefixCorrect, set ResourcePrefixActual to blank", i+lineOffset, l.HumanFriendly())
		}

		if l.SplitPackageRealPackage() != "" && (l.ProviderPackageCorrect() == "" || l.FilePrefix() == "" || l.ResourcePrefixActual() == "") {
			log.Fatalf("in service data, line %d, for service %s, if SplitPackageRealPackage has a value, ProviderPackageCorrect, ResourcePrefixActual and FilePrefix must have values", i+lineOffset, l.HumanFriendly())
		}

		if l.SplitPackageRealPackage() == "" && l.FilePrefix() != "" {
			log.Fatalf("in service data, line %d, for service %s, if SplitPackageRealPackge is blank, FilePrefix must also be blank", i+lineOffset, l.HumanFriendly())
		}

		if l.Brand() != "AWS" && l.Brand() != "Amazon" && l.Brand() != "" {
			log.Fatalf("in service data, line %d, for service %s, Brand must be AWS, Amazon, or blank; found %s", l.HumanFriendly(), i, l.Brand())
		}

		if (!l.Exclude() || (l.Exclude() && l.AllowedSubcategory() != "")) && len(l.DocPrefix()) == 0 {
			log.Fatalf("in service data, line %d, for service %s, DocPrefix cannot be blank unless Exclude is non-blank and AllowedSubcategory is blank", i+lineOffset, l.HumanFriendly())
		}

		checkAllLowercase(i, l.HumanFriendly(), "AWSCLIV2Command", l.AWSCLIV2Command())
		checkAllLowercase(i, l.HumanFriendly(), "AWSCLIV2CommandNoDashes", l.AWSCLIV2CommandNoDashes())
		checkAllLowercase(i, l.HumanFriendly(), "GoV1Package", l.GoV1Package())
		checkAllLowercase(i, l.HumanFriendly(), "GoV2Package", l.GoV2Package())
		checkAllLowercase(i, l.HumanFriendly(), "ProviderPackageActual", l.ProviderPackageActual())
		checkAllLowercase(i, l.HumanFriendly(), "ProviderPackageCorrect", l.ProviderPackageCorrect())
		checkAllLowercase(i, l.HumanFriendly(), "SplitPackageRealPackage", l.SplitPackageRealPackage())
		checkAllLowercase(i, l.HumanFriendly(), "Aliases", l.Aliases()...)
		checkAllLowercase(i, l.HumanFriendly(), "ResourcePrefixActual", l.ResourcePrefixActual())
		checkAllLowercase(i, l.HumanFriendly(), "ResourcePrefixCorrect", l.ResourcePrefixCorrect())
		checkAllLowercase(i, l.HumanFriendly(), "FilePrefix", l.FilePrefix())
		checkAllLowercase(i, l.HumanFriendly(), "DocPrefix", l.DocPrefix()...)

		checkNotAllLowercase(i, l.HumanFriendly(), "ProviderNameUpper", l.ProviderNameUpper())
		checkNotAllLowercase(i, l.HumanFriendly(), "GoV1ClientTypeName", l.GoV1ClientTypeName())
		checkNotAllLowercase(i, l.HumanFriendly(), "HumanFriendly", l.HumanFriendly())

		if !l.Exclude() && l.AllowedSubcategory() != "" {
			log.Fatalf("in service data, line %d, for service %s, AllowedSubcategory can only be non-blank if Exclude is non-blank", i+lineOffset, l.HumanFriendly())
		}

		if l.Exclude() && l.Note() == "" {
			log.Fatalf("in service data, line %d, for service %s, if Exclude is not blank, include a Note why", i+lineOffset, l.HumanFriendly())
		}

		if l.Exclude() && l.AllowedSubcategory() == "" {
			continue
		}

		deprecatedEnvVar := l.DeprecatedEnvVar() != ""
		tfAwsEnvVar := l.TFAWSEnvVar() != ""
		if deprecatedEnvVar != tfAwsEnvVar {
			log.Fatalf("in service data, line %d, for service %s, either both DeprecatedEnvVar and TFAWSEnvVar must be specified or neither can be", i+lineOffset, l.HumanFriendly())
		}

		if l.SDKID() == "" && !l.Exclude() {
			log.Fatalf("in service data, line %d, for service %s, SDKID is required unless Exclude is set", i+lineOffset, l.HumanFriendly())
		}

		if l.EndpointAPICall() == "" && !l.NotImplemented() && !l.Exclude() {
			log.Fatalf("in service data, line %d, for service %s, EndpointAPICall is required for unless NotImplemented or Exclude is set", i+lineOffset, l.HumanFriendly())
		}

		rre := l.ResourcePrefixActual()

		if rre == "" {
			rre = l.ResourcePrefixCorrect()
		}

		docPrefixes = append(docPrefixes, DocPrefix{
			HumanFriendly:  l.HumanFriendly(),
			DocPrefixRegex: l.DocPrefix(),
			ResourceRegex:  rre,
		})

		allChecks++
	}
	fmt.Printf("  Performed %d checks on service data, 0 errors.\n", (allChecks * 40))

	var fileErrs bool

	// DocPrefix needs to be reworked for compatibility with tfproviderdocs, in the meantime skip
	err = checkDocDir("../../../website/docs/r/", docPrefixes)
	if err != nil {
		fileErrs = true
		log.Printf("while checking resource doc dir: %s", err)
	}

	err = checkDocDir("../../../website/docs/d/", docPrefixes)
	if err != nil {
		fileErrs = true
		log.Printf("while checking data source doc dir: %s", err)
	}

	if fileErrs {
		os.Exit(1)
	}

	fmt.Printf("  Checked %d documentation files to ensure filename prefix, resource name, label regex, and subcategory match, 0 errors.\n", allDocs)
}

func checkAllLowercase(i int, service, name string, values ...string) {
	for _, value := range values {
		if value != "" && strings.ToLower(value) != value {
			log.Fatalf("in service data, line %d, for service %s, %s should not include uppercase letters (%s)", i+lineOffset, service, name, value)
		}
	}
}

func checkNotAllLowercase(i int, service, name, value string) {
	if value != "" && strings.ToLower(value) == value {
		log.Fatalf("in service data, line %d, for service %s, %s should be properly capitalized; it does not include uppercase letters (%s)", i+lineOffset, service, name, value)
	}
}

func checkDocDir(dir string, prefixes []DocPrefix) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("reading directory (%s): %s", dir, err)
	}

	var errs []error
	for _, fh := range fs {
		if fh.IsDir() {
			continue
		}

		if !strings.HasSuffix(fh.Name(), ".markdown") {
			continue
		}

		allDocs++

		if err := checkDocFile(dir, fh.Name(), prefixes); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func checkDocFile(dir, name string, prefixes []DocPrefix) error {
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("opening file (%s): %s", name, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var line int
	var rregex string
	for scanner.Scan() {
		switch line {
		case 0:
			if scanner.Text() != "---" {
				return fmt.Errorf("file (%s) doesn't start like doc file", name)
			}
		case 1:
			hf, rr, err := findHumanFriendly(name, prefixes)
			if err != nil {
				return fmt.Errorf("checking file (%s): %w", name, err)
			}

			rregex = rr

			sc := scanner.Text()
			sc = strings.TrimSuffix(strings.TrimPrefix(sc, "subcategory: \""), "\"")
			if hf != sc {
				return fmt.Errorf("file (%s) subcategory (%s) doesn't match file name prefix, expecting %s", name, sc, hf)
			}
		case 2:
			continue
		case 3:
			rn := scanner.Text()
			rn = strings.TrimSuffix(strings.TrimPrefix(rn, `page_title: "AWS: `), `"`)

			re, err := regexp.Compile(fmt.Sprintf(`^%s`, rregex))
			if err != nil {
				return fmt.Errorf("unable to compile resource regular expression pattern (%s): %s", rregex, err)
			}

			if !re.MatchString(rn) {
				return fmt.Errorf("resource regular expression (%s) does not match resource name (%s)", rregex, rn)
			}
		default:
			break
		}
		line++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file (%s): %s", name, err)
	}
	return nil
}

func findHumanFriendly(filename string, prefixes []DocPrefix) (string, string, error) {
	for _, v := range prefixes {
		for _, pf := range v.DocPrefixRegex {
			re, err := regexp.Compile(fmt.Sprintf(`^%s`, pf))
			if err != nil {
				return "", "", fmt.Errorf("unable to compile regular expression pattern %s: %s", pf, err)
			}

			if re.MatchString(filename) {
				return v.HumanFriendly, v.ResourceRegex, nil
			}
		}
	}

	return "", "", fmt.Errorf("could not find prefix in %v for file (%s)", prefixes, filename)
}
