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
	docPrefix          = 15
	humanFriendly      = 16
	exclude            = 18
	allowedSubcategory = 19
)

var allDocs int

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

		if l[exclude] != "" && l[allowedSubcategory] == "" {
			continue
		}

		m[l[docPrefix]] = l[humanFriendly]
	}

	err = checkDocDir("../../../website/docs/r/", m)
	if err != nil {
		log.Fatalf("while checking resource doc dir: %s", err)
	}

	err = checkDocDir("../../../website/docs/d/", m)
	if err != nil {
		log.Fatalf("while checking data source doc dir: %s", err)
	}

	fmt.Printf("Checked %d documentation files. 0 errors.\n", allDocs)
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
