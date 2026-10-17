// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Resource
// ↓
// ResourceLocation
// ↓
// Client(s)
// ↓
// SDK Package
// ↓
// AWS Service
// ↓
// Model Path

// Final Index should have:
// ResourceMap map: tfResourceName --> ResourceServiceInfo
//
//	ResourceServiceInfo {
//	   FileName string  // AWS service file name (ie models/<service>/<date>/<service-date>.json)
//		  Err string (any error that occurred (ie multiple clients, service file not found, etc))
//	}
package serviceindex

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	resourceAnnotationRE = regexp.MustCompile(`@(SDKResource|FrameworkResource)\((?:"([^"]+)"|([a-zA-Z0-9_]+))(?:,\s*name="([^"]+)")?`)
	clientRE             = regexp.MustCompile(`(?:Meta\(\)|AWSClient\))\.([A-Za-z0-9]+Client)\(`)
	awsClientMethodRE    = regexp.MustCompile(`func\s+\(c\s+\*AWSClient\)\s+([A-Za-z0-9]+Client)\([^)]*\)\s+\*([A-Za-z0-9_]+)\.Client`)
	githubTreeURL        = "https://api.github.com/repos/aws/api-models-aws/git/trees/main?recursive=1"
	httpTimeout          = 30 * time.Second
)

// --- Building Resource Index (connecting the TF resource name to the file service it's described in) ---

type ResourceLocation struct {
	Resource          string 
	Service           string
	File              string
	TFResourceSegment string
}

// Building the ResourceIndex and FileConnIndex
// ResourceIndex = map ResourceName --> ResourceLocation
// FileConnIndex = map FileName --> List of AWS Clients used inside file
type ServiceIndexes struct {
	ResourceIndex map[string][]ResourceLocation
	FileConnIndex map[string][]string
}

func BuildServiceIndexes(repoRoot string) (*ServiceIndexes, error) {
	serviceRoot := filepath.Join(repoRoot, "internal", "service")

	indexes := &ServiceIndexes{
		ResourceIndex: make(map[string][]ResourceLocation),
		FileConnIndex: make(map[string][]string),
	}

	err := filepath.WalkDir(serviceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Check if SDKResource or FrameworkResource
		matches := resourceAnnotationRE.FindAllStringSubmatch(string(content), -1)
		if len(matches) == 0 {
			return nil
		}

		relPath, err := filepath.Rel(serviceRoot, path)
		if err != nil {
			return err
		}
		service := filepath.Dir(relPath)
		if service == "." {
			service = ""
		}
		for _, match := range matches {
			resourceName := match[2]
			if resourceName == "" {
				resourceName = match[3]
			}

			TFResourceNameSegment := strings.ReplaceAll(
				match[4],
				" ",
				"",
)

			indexes.ResourceIndex[resourceName] = append(
				indexes.ResourceIndex[resourceName],
				ResourceLocation{
					Resource:          resourceName,
					Service:           service,
					File:              relPath,
					TFResourceSegment: TFResourceNameSegment,
				},
			)
		}

		// Build FileConnIndex
		clients := extractClients(content)
		if len(clients) > 0 {
			indexes.FileConnIndex[relPath] = clients
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return indexes, nil
}

func extractClients(content []byte) []string {
	matches := clientRE.FindAllStringSubmatch(string(content), -1)
	if len(matches) == 0 {
		return nil
	}

	clientSet := make(map[string]struct{})

	for _, match := range matches {
		clientSet[match[1]] = struct{}{}
	}

	clients := make([]string, 0, len(clientSet))

	for client := range clientSet {
		clients = append(clients, client)
	}

	sort.Strings(clients)

	return clients
}

// --- Building the client index (making the map between the AWS api client name to the AWS service name) ---
func BuildAWSClientIndex(repoRoot string) (map[string]string, error) {

	files := []string{
		filepath.Join(repoRoot, "internal", "conns", "awsclient_gen.go"),
		filepath.Join(repoRoot, "internal", "conns", "awsclient.go"),
	}

	index := make(map[string]string)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		matches := awsClientMethodRE.FindAllStringSubmatch(string(content), -1)

		for _, match := range matches {
			clientName := match[1]
			sdkPackage := match[2]

			if _, ok := index[clientName]; !ok {
				index[clientName] = sdkPackage
			}
		}
	}
	
	return index, nil
}

// --- Building the AWS service index ---
// Index holds discovered AWS service model metadata and Terraform-to-AWS
// service name mappings used during model resolution.
type AWSServiceIndex struct {
	models   map[string]string // AWS service name → latest date folder (e.g. "amp" → "2020-08-01")
	services map[string]string // TF service name → AWS directory name overrides (obtained from mapping file)
	aliases  map[string]string // AWS service names w/o hyphens → AWS service names w/ hyphens (e.g. "appautoscaling" → "app-auto-scaling")
}

// FIXME: maybe remove AWSServiceIndex.services

// githubTreeResponse is the response from the GitHub Trees API.
type githubTreeResponse struct {
	Tree []githubTreeEntry `json:"tree"`
}

// githubTreeEntry is one element of the GitHub Trees API response.
type githubTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

func BuildAWSServiceIndex() (*AWSServiceIndex, error) {
	models, err := fetchLatestDates()
	if err != nil {
		return nil, err
	}

	aliases := make(map[string]string)
	for service := range models {
		if strings.Contains(service, "-") {
			aliases[strings.ReplaceAll(service, "-", "")] = service
		}
	}

	return &AWSServiceIndex{models: models, aliases: aliases}, nil
}

func fetchLatestDates() (map[string]string, error) {
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest(http.MethodGet, githubTreeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", githubTreeURL, err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching service index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("service index URL not found: %s", githubTreeURL)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching service index: unexpected status %s", resp.Status)
	}

	var treeResp githubTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return nil, fmt.Errorf("decoding service index response: %w", err)
	}

	models := make(map[string]string)

	for _, entry := range treeResp.Tree {
		parts := strings.Split(entry.Path, "/")

		// Expected:
		// models/<service>/service/<date>/<file>.json
		if len(parts) != 5 {
			continue
		}
		if parts[0] != "models" {
			continue
		}
		if parts[2] != "service" {
			continue
		}
		if !strings.HasSuffix(parts[4], ".json") {
			continue
		}

		service := parts[1]
		date := parts[3]

		// ISO dates sort lexicographically.
		if current, ok := models[service]; !ok || date > current {
			models[service] = date
		}
	}

	return models, nil
}

func (idx *AWSServiceIndex) HasService(serviceName string) bool {
	_, ok := idx.models[serviceName]
	return ok
}

func (idx *AWSServiceIndex) ResolveModelPath(awsServiceName string) (modelPath, namespace string, err error) {
	date, ok := idx.models[awsServiceName]
	if !ok {
		return "", "", fmt.Errorf("service %q not found", awsServiceName)
	}

	modelPath = fmt.Sprintf("models/%s/service/%s/%s-%s.json", awsServiceName, date, awsServiceName, date)
	namespace = fmt.Sprintf("com.amazonaws.%s", strings.ReplaceAll(awsServiceName, "-", ""))
	return modelPath, namespace, nil
}
