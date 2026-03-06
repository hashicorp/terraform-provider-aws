package awsops

import (
	"go/ast"
	"strings"
)

// discoverResources scans package comments for @SDKResource and @FrameworkResource annotations.
func discoverResources(files map[string]*ast.File, dir string) []resourceInfo {
	var resources []resourceInfo

	for filename, file := range files {
		for _, cg := range file.Comments {
			text := cg.Text()
			for _, line := range strings.Split(text, "\n") {
				line = strings.TrimSpace(line)
				if m := sdkResourceRe.FindStringSubmatch(line); m != nil {
					resources = append(resources, resourceInfo{
						Name:      m[1],
						Type:      "sdk",
						File:      filename,
						Package:   file.Name.Name,
						Directory: dir,
					})
				}
				if m := fwResourceRe.FindStringSubmatch(line); m != nil {
					resources = append(resources, resourceInfo{
						Name:      m[1],
						Type:      "framework",
						File:      filename,
						Package:   file.Name.Name,
						Directory: dir,
					})
				}
			}
		}
	}

	return resources
}
