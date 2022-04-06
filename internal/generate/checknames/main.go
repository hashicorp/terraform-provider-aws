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
	"strings"
)

const (
	// column indices of CSV
	//awsCLIV2Command         = 0
	//awsCLIV2CommandNoDashes = 1
	//goV1Package             = 2
	//goV2Package             = 3
	//providerPackageActual   = 4
	//providerPackageCorrect  = 5
	//splitPackageRealPackage = 6
	//aliases                 = 7
	//providerNameUpper       = 8
	//goV1ClientName          = 9
	//skipClientGenerate      = 10
	//sdkVersion              = 11
	//resourcePrefixActual    = 12
	//resourcePrefixCorrect   = 13
	//filePrefix              = 14
	//docPrefix               = 15
	//humanFriendly           = 16
	//brand                   = 17
	//exclude                 = 18
	//allowedSubcategory      = 19
	//deprecatedEnvVar        = 20
	//envVar                  = 21
	//note                    = 22
	awsCLIV2Command         = 0
	awsCLIV2CommandNoDashes = 1
	goV1Package             = 2
	goV2Package             = 3
	providerPackageActual   = 4
	providerPackageCorrect  = 5
	splitPackageRealPackage = 6
	aliases                 = 7
	providerNameUpper       = 8
	goV1ClientName          = 9
	sdkVersion              = 11
	resourcePrefixActual    = 12
	resourcePrefixCorrect   = 13
	filePrefix              = 14
	docPrefix               = 15
	humanFriendly           = 16
	brand                   = 17
	exclude                 = 18
	allowedSubcategory      = 19
	note                    = 22
)

var allDocs int
var allChecks int

func main() {
	f, err := os.Open("../../../names/names_data.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]string)

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[humanFriendly] == "" {
			log.Fatal("in names_data.csv, HumanFriendly cannot be blank")
		}

		if l[awsCLIV2Command] != "" && strings.Replace(l[awsCLIV2Command], "-", "", -1) != l[awsCLIV2CommandNoDashes] {
			log.Fatalf("in names_data.csv, for service %s, AWSCLIV2CommandNoDashes must be the same as AWSCLIV2Command without dashes (%s)", l[humanFriendly], strings.Replace(l[awsCLIV2Command], "-", "", -1))
		}

		if l[providerPackageCorrect] != "" && l[awsCLIV2CommandNoDashes] != "" && l[goV2Package] != "" {
			if len(l[awsCLIV2CommandNoDashes]) < len(l[goV2Package]) && l[providerPackageCorrect] != l[awsCLIV2CommandNoDashes] {
				log.Fatalf("in names_data.csv, for service %s, ProviderPackageCorrect must be shorter of AWSCLIV2CommandNoDashes (%s) and GoV2Package (%s)", l[humanFriendly], l[awsCLIV2CommandNoDashes], l[goV2Package])
			}

			if len(l[awsCLIV2CommandNoDashes]) > len(l[goV2Package]) && l[providerPackageCorrect] != l[goV2Package] {
				log.Fatalf("in names_data.csv, for service %s, ProviderPackageCorrect must be shorter of AWSCLIV2CommandNoDashes (%s) and GoV2Package (%s)", l[humanFriendly], l[awsCLIV2CommandNoDashes], l[goV2Package])
			}
		}

		if l[awsCLIV2CommandNoDashes] == "" && l[goV2Package] == "" && l[exclude] == "" {
			log.Fatalf("in names_data.csv, for service %s, if Exclude is blank, either AWSCLIV2CommandNoDashes or GoV2Package must have values", l[humanFriendly])
		}

		if l[providerPackageActual] != "" && l[providerPackageCorrect] == "" {
			log.Fatalf("in names_data.csv, for service %s, ProviderPackageActual can't be non-blank if ProviderPackageCorrect is blank", l[humanFriendly])
		}

		if l[providerPackageActual] == "" && l[providerPackageCorrect] == "" && l[exclude] == "" {
			log.Fatalf("in names_data.csv, for service %s, ProviderPackageActual and ProviderPackageCorrect cannot both be blank unless Exclude is non-blank", l[humanFriendly])
		}

		if l[providerPackageCorrect] != "" && l[providerPackageActual] == l[providerPackageCorrect] {
			log.Fatalf("in names_data.csv, for service %s, ProviderPackageActual should only be used if different from ProviderPackageCorrect", l[humanFriendly])
		}

		packageToUse := l[providerPackageCorrect]

		if l[providerPackageActual] != "" {
			packageToUse = l[providerPackageActual]
		}

		if l[aliases] != "" && packageToUse != "" {
			p := strings.Split(l[aliases], ";")

			for _, v := range p {
				if v == packageToUse {
					log.Fatalf("in names_data.csv, for service %s, Aliases should not include ProviderPackageActual, if not blank, or ProviderPackageCorrect, if not blank and ProviderPackageActual is blank", l[humanFriendly])
				}
			}
		}

		if l[sdkVersion] != "1" && l[sdkVersion] != "2" && l[exclude] == "" {
			log.Fatalf("in names_data.csv, for service %s, SDKVersion must have a value if Exclude is blank", l[humanFriendly])
		}

		if l[sdkVersion] == "1" && (l[goV1Package] == "" || l[goV1ClientName] == "") {
			log.Fatalf("in names_data.csv, for service %s, SDKVersion is set to 1 so neither GoV1Package nor GoV1ClientName can be blank", l[humanFriendly])
		}

		if l[sdkVersion] == "2" && l[goV2Package] == "" {
			log.Fatalf("in names_data.csv, for service %s, SDKVersion is set to 2 so GoV2Package cannot be blank", l[humanFriendly])
		}

		if l[resourcePrefixCorrect] != "" && l[resourcePrefixCorrect] != fmt.Sprintf("aws_%s_", l[providerPackageCorrect]) {
			log.Fatalf("in names_data.csv, for service %s, ResourcePrefixCorrect should be aws_<package>_, where <package> is ProviderPackageCorrect", l[humanFriendly])
		}

		if l[resourcePrefixCorrect] != "" && l[resourcePrefixActual] == l[resourcePrefixCorrect] {
			log.Fatalf("in names_data.csv, for service %s, ResourcePrefixActual should not be the same as ResourcePrefixCorrect, set ResourcePrefixActual to blank", l[humanFriendly])
		}

		if l[splitPackageRealPackage] != "" && (l[providerPackageCorrect] == "" || l[filePrefix] == "" || l[resourcePrefixActual] == "") {
			log.Fatalf("in names_data.csv, for service %s, if SplitPackageRealPackage has a value, ProviderPackageCorrect, ResourcePrefixActual and FilePrefix must have values", l[humanFriendly])
		}

		if l[splitPackageRealPackage] == "" && l[filePrefix] != "" {
			log.Fatalf("in names_data.csv, for service %s, if SplitPackageRealPackge is blank, FilePrefix must also be blank", l[humanFriendly])
		}

		if l[brand] != "AWS" && l[brand] != "Amazon" && l[brand] != "" {
			log.Fatalf("in names_data.csv, for service %s, Brand must be AWS, Amazon, or blank; found %s", l[humanFriendly], l[brand])
		}

		if (l[exclude] == "" || (l[exclude] != "" && l[allowedSubcategory] != "")) && l[docPrefix] == "" {
			log.Fatalf("in names_data.csv, for service %s, DocPrefix cannot be blank unless Exclude is non-blank and AllowedSubcategory is blank", l[humanFriendly])
		}

		checkAllLowercase(l[humanFriendly], "AWSCLIV2Command", l[awsCLIV2Command])
		checkAllLowercase(l[humanFriendly], "AWSCLIV2CommandNoDashes", l[awsCLIV2CommandNoDashes])
		checkAllLowercase(l[humanFriendly], "GoV1Package", l[goV1Package])
		checkAllLowercase(l[humanFriendly], "GoV2Package", l[goV2Package])
		checkAllLowercase(l[humanFriendly], "ProviderPackageActual", l[providerPackageActual])
		checkAllLowercase(l[humanFriendly], "ProviderPackageCorrect", l[providerPackageCorrect])
		checkAllLowercase(l[humanFriendly], "SplitPackageRealPackage", l[splitPackageRealPackage])
		checkAllLowercase(l[humanFriendly], "Aliases", l[aliases])
		checkAllLowercase(l[humanFriendly], "ResourcePrefixActual", l[resourcePrefixActual])
		checkAllLowercase(l[humanFriendly], "ResourcePrefixCorrect", l[resourcePrefixCorrect])
		checkAllLowercase(l[humanFriendly], "FilePrefix", l[filePrefix])
		checkAllLowercase(l[humanFriendly], "DocPrefix", l[docPrefix])

		checkNotAllLowercase(l[humanFriendly], "ProviderNameUpper", l[providerNameUpper])
		checkNotAllLowercase(l[humanFriendly], "GoV1ClientName", l[goV1ClientName])
		checkNotAllLowercase(l[humanFriendly], "HumanFriendly", l[humanFriendly])

		if l[exclude] == "" && l[allowedSubcategory] != "" {
			log.Fatalf("in names_data.csv, for service %s, AllowedSubcategory can only be non-blank if Exclude is non-blank", l[humanFriendly])
		}

		if l[exclude] != "" && l[note] == "" {
			log.Fatalf("in names_data.csv, for service %s, if Exclude is not blank, include a Note why", l[humanFriendly])
		}

		if l[exclude] != "" && l[allowedSubcategory] == "" {
			continue
		}

		m[l[docPrefix]] = l[humanFriendly]
		allChecks++
	}
	fmt.Printf("Performed %d checks on names_data.csv, 0 errors.\n", (allChecks * 36))

	err = checkDocDir("../../../website/docs/r/", m)
	if err != nil {
		log.Fatalf("while checking resource doc dir: %s", err)
	}

	err = checkDocDir("../../../website/docs/d/", m)
	if err != nil {
		log.Fatalf("while checking data source doc dir: %s", err)
	}

	fmt.Printf("Checked %d documentation files for matching prefix and subcategory, 0 errors.\n", allDocs)
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

func checkDocDir(dir string, data map[string]string) error {
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

		p := strings.Split(fh.Name(), "_")
		if len(p) < 2 {
			return fmt.Errorf("file (%s) has no prefix", fh.Name())
		}

		prefix := fmt.Sprintf("%s_", p[0])

		scanner := bufio.NewScanner(f)
		var line int
		for scanner.Scan() {
			switch line {
			case 0:
				if scanner.Text() != "---" {
					return fmt.Errorf("file (%s) doesn't start like doc file", fh.Name())
				}
			case 1:
				sc := scanner.Text()
				sc = strings.TrimSuffix(strings.TrimPrefix(sc, "subcategory: \""), "\"")
				if data[prefix] != sc {
					return fmt.Errorf("file (%s) subcategory (%s) doesn't match file name prefix (%s)", fh.Name(), sc, prefix)
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
