//go:build generate
// +build generate

package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/names"
)

const namesDataFile = "../../../names/names_data.csv"

// DocPrefix tests/column needs to be reworked for compatibility with tfproviderdocs
type DocPrefix struct {
	HumanFriendly  string
	DocPrefixRegex []string
	ResourceRegex  string
}

var allDocs int // currently skipping this test
var allChecks int

func main() {
	fmt.Printf("Checking %s\n", strings.TrimPrefix(namesDataFile, "../../../"))

	f, err := os.Open(namesDataFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	docPrefixes := []DocPrefix{} // test to be reworked

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColHumanFriendly] == "" {
			log.Fatal("in names_data.csv, HumanFriendly cannot be blank")
		}

		// TODO: Check for duplicates in HumanFriendly, ProviderPackageActual,
		// ProviderPackageCorrect, ProviderNameUpper, GoV1ClientName,
		// ResourcePrefixActual, ResourcePrefixCorrect, FilePrefix, DocPrefix

		if l[names.ColAWSCLIV2Command] != "" && strings.Replace(l[names.ColAWSCLIV2Command], "-", "", -1) != l[names.ColAWSCLIV2CommandNoDashes] {
			log.Fatalf("in names_data.csv, for service %s, AWSCLIV2CommandNoDashes must be the same as AWSCLIV2Command without dashes (%s)", l[names.ColHumanFriendly], strings.Replace(l[names.ColAWSCLIV2Command], "-", "", -1))
		}

		if l[names.ColProviderPackageCorrect] != "" && l[names.ColAWSCLIV2CommandNoDashes] != "" && l[names.ColGoV2Package] != "" {
			if len(l[names.ColAWSCLIV2CommandNoDashes]) < len(l[names.ColGoV2Package]) && l[names.ColProviderPackageCorrect] != l[names.ColAWSCLIV2CommandNoDashes] {
				log.Fatalf("in names_data.csv, for service %s, ProviderPackageCorrect must be shorter of AWSCLIV2CommandNoDashes (%s) and GoV2Package (%s)", l[names.ColHumanFriendly], l[names.ColAWSCLIV2CommandNoDashes], l[names.ColGoV2Package])
			}

			if len(l[names.ColAWSCLIV2CommandNoDashes]) > len(l[names.ColGoV2Package]) && l[names.ColProviderPackageCorrect] != l[names.ColGoV2Package] {
				log.Fatalf("in names_data.csv, for service %s, ProviderPackageCorrect must be shorter of AWSCLIV2CommandNoDashes (%s) and GoV2Package (%s)", l[names.ColHumanFriendly], l[names.ColAWSCLIV2CommandNoDashes], l[names.ColGoV2Package])
			}
		}

		if l[names.ColAWSCLIV2CommandNoDashes] == "" && l[names.ColGoV2Package] == "" && l[names.ColExclude] == "" {
			log.Fatalf("in names_data.csv, for service %s, if Exclude is blank, either AWSCLIV2CommandNoDashes or GoV2Package must have values", l[names.ColHumanFriendly])
		}

		if l[names.ColProviderPackageActual] != "" && l[names.ColProviderPackageCorrect] == "" {
			log.Fatalf("in names_data.csv, for service %s, ProviderPackageActual can't be non-blank if ProviderPackageCorrect is blank", l[names.ColHumanFriendly])
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" && l[names.ColExclude] == "" {
			log.Fatalf("in names_data.csv, for service %s, ProviderPackageActual and ProviderPackageCorrect cannot both be blank unless Exclude is non-blank", l[names.ColHumanFriendly])
		}

		if l[names.ColProviderPackageCorrect] != "" && l[names.ColProviderPackageActual] == l[names.ColProviderPackageCorrect] {
			log.Fatalf("in names_data.csv, for service %s, ProviderPackageActual should only be used if different from ProviderPackageCorrect", l[names.ColHumanFriendly])
		}

		packageToUse := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			packageToUse = l[names.ColProviderPackageActual]
		}

		if l[names.ColAliases] != "" && packageToUse != "" {
			p := strings.Split(l[names.ColAliases], ";")

			for _, v := range p {
				if v == packageToUse {
					log.Fatalf("in names_data.csv, for service %s, Aliases should not include ProviderPackageActual, if not blank, or ProviderPackageCorrect, if not blank and ProviderPackageActual is blank", l[names.ColHumanFriendly])
				}
			}
		}

		if l[names.ColSDKVersion] != "1" && l[names.ColSDKVersion] != "2" && l[names.ColExclude] == "" {
			log.Fatalf("in names_data.csv, for service %s, SDKVersion must have a value if Exclude is blank", l[names.ColHumanFriendly])
		}

		if l[names.ColSDKVersion] == "1" && (l[names.ColGoV1Package] == "" || l[names.ColGoV1ClientName] == "") {
			log.Fatalf("in names_data.csv, for service %s, SDKVersion is set to 1 so neither GoV1Package nor GoV1ClientName can be blank", l[names.ColHumanFriendly])
		}

		if l[names.ColSDKVersion] == "2" && l[names.ColGoV2Package] == "" {
			log.Fatalf("in names_data.csv, for service %s, SDKVersion is set to 2 so GoV2Package cannot be blank", l[names.ColHumanFriendly])
		}

		if l[names.ColResourcePrefixCorrect] != "" && l[names.ColResourcePrefixCorrect] != fmt.Sprintf("aws_%s_", l[names.ColProviderPackageCorrect]) {
			log.Fatalf("in names_data.csv, for service %s, ResourcePrefixCorrect should be aws_<package>_, where <package> is ProviderPackageCorrect", l[names.ColHumanFriendly])
		}

		if l[names.ColResourcePrefixCorrect] != "" && l[names.ColResourcePrefixActual] == l[names.ColResourcePrefixCorrect] {
			log.Fatalf("in names_data.csv, for service %s, ResourcePrefixActual should not be the same as ResourcePrefixCorrect, set ResourcePrefixActual to blank", l[names.ColHumanFriendly])
		}

		if l[names.ColSplitPackageRealPackage] != "" && (l[names.ColProviderPackageCorrect] == "" || l[names.ColFilePrefix] == "" || l[names.ColResourcePrefixActual] == "") {
			log.Fatalf("in names_data.csv, for service %s, if SplitPackageRealPackage has a value, ProviderPackageCorrect, ResourcePrefixActual and FilePrefix must have values", l[names.ColHumanFriendly])
		}

		if l[names.ColSplitPackageRealPackage] == "" && l[names.ColFilePrefix] != "" {
			log.Fatalf("in names_data.csv, for service %s, if SplitPackageRealPackge is blank, FilePrefix must also be blank", l[names.ColHumanFriendly])
		}

		if l[names.ColBrand] != "AWS" && l[names.ColBrand] != "Amazon" && l[names.ColBrand] != "" {
			log.Fatalf("in names_data.csv, for service %s, Brand must be AWS, Amazon, or blank; found %s", l[names.ColHumanFriendly], l[names.ColBrand])
		}

		if (l[names.ColExclude] == "" || (l[names.ColExclude] != "" && l[names.ColAllowedSubcategory] != "")) && l[names.ColDocPrefix] == "" {
			log.Fatalf("in names_data.csv, for service %s, DocPrefix cannot be blank unless Exclude is non-blank and AllowedSubcategory is blank", l[names.ColHumanFriendly])
		}

		checkAllLowercase(l[names.ColHumanFriendly], "AWSCLIV2Command", l[names.ColAWSCLIV2Command])
		checkAllLowercase(l[names.ColHumanFriendly], "AWSCLIV2CommandNoDashes", l[names.ColAWSCLIV2CommandNoDashes])
		checkAllLowercase(l[names.ColHumanFriendly], "GoV1Package", l[names.ColGoV1Package])
		checkAllLowercase(l[names.ColHumanFriendly], "GoV2Package", l[names.ColGoV2Package])
		checkAllLowercase(l[names.ColHumanFriendly], "ProviderPackageActual", l[names.ColProviderPackageActual])
		checkAllLowercase(l[names.ColHumanFriendly], "ProviderPackageCorrect", l[names.ColProviderPackageCorrect])
		checkAllLowercase(l[names.ColHumanFriendly], "SplitPackageRealPackage", l[names.ColSplitPackageRealPackage])
		checkAllLowercase(l[names.ColHumanFriendly], "Aliases", l[names.ColAliases])
		checkAllLowercase(l[names.ColHumanFriendly], "ResourcePrefixActual", l[names.ColResourcePrefixActual])
		checkAllLowercase(l[names.ColHumanFriendly], "ResourcePrefixCorrect", l[names.ColResourcePrefixCorrect])
		checkAllLowercase(l[names.ColHumanFriendly], "FilePrefix", l[names.ColFilePrefix])
		checkAllLowercase(l[names.ColHumanFriendly], "DocPrefix", l[names.ColDocPrefix])

		checkNotAllLowercase(l[names.ColHumanFriendly], "ProviderNameUpper", l[names.ColProviderNameUpper])
		checkNotAllLowercase(l[names.ColHumanFriendly], "GoV1ClientName", l[names.ColGoV1ClientName])
		checkNotAllLowercase(l[names.ColHumanFriendly], "HumanFriendly", l[names.ColHumanFriendly])

		if l[names.ColExclude] == "" && l[names.ColAllowedSubcategory] != "" {
			log.Fatalf("in names_data.csv, for service %s, AllowedSubcategory can only be non-blank if Exclude is non-blank", l[names.ColHumanFriendly])
		}

		if l[names.ColExclude] != "" && l[names.ColNote] == "" {
			log.Fatalf("in names_data.csv, for service %s, if Exclude is not blank, include a Note why", l[names.ColHumanFriendly])
		}

		if l[names.ColExclude] != "" && l[names.ColAllowedSubcategory] == "" {
			continue
		}

		rre := l[names.ColResourcePrefixActual]

		if rre == "" {
			rre = l[names.ColResourcePrefixCorrect]
		}

		docPrefixes = append(docPrefixes, DocPrefix{
			HumanFriendly:  l[names.ColHumanFriendly],
			DocPrefixRegex: strings.Split(l[names.ColDocPrefix], ";"),
			ResourceRegex:  rre,
		})

		allChecks++
	}
	fmt.Printf("  Performed %d checks on names_data.csv, 0 errors.\n", (allChecks * 36))

	// DocPrefix needs to be reworked for compatibility with tfproviderdocs, in the meantime skip
	err = checkDocDir("../../../website/docs/r/", docPrefixes)
	if err != nil {
		log.Fatalf("while checking resource doc dir: %s", err)
	}

	err = checkDocDir("../../../website/docs/d/", docPrefixes)
	if err != nil {
		log.Fatalf("while checking data source doc dir: %s", err)
	}

	fmt.Printf("  Checked %d documentation files to ensure filename prefix, resource name, label regex, and subcategory match, 0 errors.\n", allDocs)
}

func checkAllLowercase(service, name, value string) {
	if value != "" && strings.ToLower(value) != value {
		log.Fatalf("in names_data.csv, for service %s, %s should not include uppercase letters (%s)", service, name, value)
	}
}

func checkNotAllLowercase(service, name, value string) {
	if value != "" && strings.ToLower(value) == value {
		log.Fatalf("in names_data.csv, for service %s, %s should be properly capitalized; it does not include uppercase letters (%s)", service, name, value)
	}
}

func checkDocDir(dir string, prefixes []DocPrefix) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("reading directory (%s): %s", dir, err)
	}

	for _, fh := range fs {
		if fh.IsDir() {
			continue
		}

		if !strings.HasSuffix(fh.Name(), ".markdown") {
			continue
		}

		allDocs++

		f, err := os.Open(filepath.Join(dir, fh.Name()))
		if err != nil {
			return fmt.Errorf("opening file (%s): %s", fh.Name(), err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var line int
		var rregex string
		for scanner.Scan() {
			switch line {
			case 0:
				if scanner.Text() != "---" {
					return fmt.Errorf("file (%s) doesn't start like doc file", fh.Name())
				}
			case 1:
				hf, rr, err := findHumanFriendly(fh.Name(), prefixes)
				if err != nil {
					return fmt.Errorf("checking file (%s): %w", fh.Name(), err)
				}

				rregex = rr

				sc := scanner.Text()
				sc = strings.TrimSuffix(strings.TrimPrefix(sc, "subcategory: \""), "\"")
				if hf != sc {
					return fmt.Errorf("file (%s) subcategory (%s) doesn't match file name prefix, expecting %s", fh.Name(), sc, hf)
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
			return fmt.Errorf("reading file (%s): %s", fh.Name(), err)
		}
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
