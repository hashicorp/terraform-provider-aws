// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releasesjson

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/hc-install/internal/httpclient"
)

type Downloader struct {
	Logger           *log.Logger
	VerifyChecksum   bool
	ArmoredPublicKey string
	BaseURL          string
}

func (d *Downloader) DownloadAndUnpack(ctx context.Context, pv *ProductVersion, binDir string, licenseDir string) (zipFilePath string, err error) {
	if len(pv.Builds) == 0 {
		return "", fmt.Errorf("no builds found for %s %s", pv.Name, pv.Version)
	}

	pb, ok := pv.Builds.FilterBuild(runtime.GOOS, runtime.GOARCH, "zip")
	if !ok {
		return "", fmt.Errorf("no ZIP archive found for %s %s %s/%s",
			pv.Name, pv.Version, runtime.GOOS, runtime.GOARCH)
	}

	var verifiedChecksum HashSum
	if d.VerifyChecksum {
		v := &ChecksumDownloader{
			BaseURL:          d.BaseURL,
			ProductVersion:   pv,
			Logger:           d.Logger,
			ArmoredPublicKey: d.ArmoredPublicKey,
		}
		verifiedChecksums, err := v.DownloadAndVerifyChecksums(ctx)
		if err != nil {
			return "", err
		}
		var ok bool
		verifiedChecksum, ok = verifiedChecksums[pb.Filename]
		if !ok {
			return "", fmt.Errorf("no checksum found for %q", pb.Filename)
		}
	}

	client := httpclient.NewHTTPClient()

	archiveURL := pb.URL
	if d.BaseURL != "" {
		// ensure that absolute download links from mocked responses
		// are still pointing to the mock server if one is set
		baseURL, err := url.Parse(d.BaseURL)
		if err != nil {
			return "", err
		}

		u, err := url.Parse(archiveURL)
		if err != nil {
			return "", err
		}
		u.Scheme = baseURL.Scheme
		u.Host = baseURL.Host
		archiveURL = u.String()
	}

	d.Logger.Printf("downloading archive from %s", archiveURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for %q: %w", archiveURL, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to download ZIP archive from %q: %s", archiveURL, resp.Status)
	}

	defer resp.Body.Close()

	pkgReader := resp.Body

	contentType := resp.Header.Get("content-type")
	if !contentTypeIsZip(contentType) {
		return "", fmt.Errorf("unexpected content-type: %s (expected any of %q)",
			contentType, zipMimeTypes)
	}

	expectedSize := resp.ContentLength

	pkgFile, err := ioutil.TempFile("", pb.Filename)
	if err != nil {
		return "", err
	}
	defer pkgFile.Close()
	pkgFilePath, err := filepath.Abs(pkgFile.Name())

	d.Logger.Printf("copying %q (%d bytes) to %s", pb.Filename, expectedSize, pkgFile.Name())

	var bytesCopied int64
	if d.VerifyChecksum {
		d.Logger.Printf("verifying checksum of %q", pb.Filename)
		h := sha256.New()
		r := io.TeeReader(resp.Body, pkgFile)

		bytesCopied, err = io.Copy(h, r)
		if err != nil {
			return "", err
		}

		calculatedSum := h.Sum(nil)
		if !bytes.Equal(calculatedSum, verifiedChecksum) {
			return pkgFilePath, fmt.Errorf(
				"checksum mismatch (expected: %x, got: %x)",
				verifiedChecksum, calculatedSum,
			)
		}
	} else {
		bytesCopied, err = io.Copy(pkgFile, pkgReader)
		if err != nil {
			return pkgFilePath, err
		}
	}

	d.Logger.Printf("copied %d bytes to %s", bytesCopied, pkgFile.Name())

	if expectedSize != 0 && bytesCopied != int64(expectedSize) {
		return pkgFilePath, fmt.Errorf(
			"unexpected size (downloaded: %d, expected: %d)",
			bytesCopied, expectedSize,
		)
	}

	r, err := zip.OpenReader(pkgFile.Name())
	if err != nil {
		return pkgFilePath, err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "..") {
			// While we generally trust the source ZIP file
			// we still reject path traversal attempts as a precaution.
			continue
		}
		srcFile, err := f.Open()
		if err != nil {
			return pkgFilePath, err
		}

		// Determine the appropriate destination file path
		dstDir := binDir
		if isLicenseFile(f.Name) && licenseDir != "" {
			dstDir = licenseDir
		}

		d.Logger.Printf("unpacking %s to %s", f.Name, dstDir)
		dstPath := filepath.Join(dstDir, f.Name)
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return pkgFilePath, err
		}

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return pkgFilePath, err
		}
		srcFile.Close()
		dstFile.Close()
	}

	return pkgFilePath, nil
}

// The production release site uses consistent single mime type
// but mime types are platform-dependent
// and we may use different OS under test
var zipMimeTypes = []string{
	"application/x-zip-compressed", // Windows
	"application/zip",              // Unix
}

func contentTypeIsZip(contentType string) bool {
	for _, mt := range zipMimeTypes {
		if mt == contentType {
			return true
		}
	}
	return false
}

// Enterprise products have a few additional license files
// that need to be extracted to a separate directory
var licenseFiles = []string{
	"EULA.txt",
	"TermsOfEvaluation.txt",
}

func isLicenseFile(filename string) bool {
	for _, lf := range licenseFiles {
		if lf == filename {
			return true
		}
	}
	return false
}
