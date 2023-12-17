// Adapted from https://github.com/open-policy-agent/conftest/tree/v0.47.0/downloader
package downloader

import (
	"context"
	"fmt"
	"strings"

	getter "github.com/hashicorp/go-getter"
)

var detectors = []getter.Detector{
	new(OCIDetector),
	new(getter.GitHubDetector),
	new(getter.GitDetector),
	new(getter.BitBucketDetector),
	new(getter.S3Detector),
	new(getter.GCSDetector),
	new(getter.FileDetector),
}

var getters = map[string]getter.Getter{
	"file":  new(getter.FileGetter),
	"git":   new(getter.GitGetter),
	"gcs":   new(getter.GCSGetter),
	"hg":    new(getter.HgGetter),
	"s3":    new(getter.S3Getter),
	"oci":   new(OCIGetter),
	"http":  new(getter.HttpGetter),
	"https": new(getter.HttpGetter),
}

// Download downloads the given url into the given destination directory.
func Download(ctx context.Context, dst string, url string) error {
	opts := []getter.ClientOption{}

	detectedURL, err := Detect(url, dst)
	if err != nil {
		return fmt.Errorf("detecting url: %w", err)
	}

	client := &getter.Client{
		Ctx:       ctx,
		Src:       detectedURL,
		Dst:       dst,
		Pwd:       dst,
		Mode:      getter.ClientModeAny,
		Detectors: detectors,
		Getters:   getters,
		Options:   opts,
	}

	if err := client.Get(); err != nil {
		return fmt.Errorf("client get: %w", err)
	}

	return nil
}

// Detect determines whether a url is a known source url from which we can download files.
// If a known source is found, the url is formatted, otherwise an error is returned.
func Detect(url string, dst string) (string, error) {
	// localhost is not considered a valid scheme for the detector which
	// causes pull commands that reference localhost to error.
	//
	// To allow for localhost to be used, replace the localhost reference
	// with the IP address.
	if strings.Contains(url, "localhost") {
		url = strings.ReplaceAll(url, "localhost", "127.0.0.1")
	}

	result, err := getter.Detect(url, dst, detectors)
	if err != nil {
		return "", fmt.Errorf("detect: %w", err)
	}

	return result, nil
}
