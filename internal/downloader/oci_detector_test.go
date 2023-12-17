package downloader

import "testing"

func TestOCIDetector_Detect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"should detect azurecr",
			"user.azurecr.io/plugin:tag",
			"oci://user.azurecr.io/plugin:tag",
		},
		{
			"should detect gcr",
			"gcr.io/example/plugin:tag",
			"oci://gcr.io/example/plugin:tag",
		},
		{
			"should detect ghcr",
			"ghcr.io/example/plugin:tag",
			"oci://ghcr.io/example/plugin:tag",
		},
		{
			"should detect ecr",
			"123456789012.dkr.ecr.us-east-1.amazonaws.com/example/plugin:tag",
			"oci://123456789012.dkr.ecr.us-east-1.amazonaws.com/example/plugin:tag",
		},
		{
			"should detect gitlab",
			"registry.gitlab.com/example/plugin:tag",
			"oci://registry.gitlab.com/example/plugin:tag",
		},
		{
			"should add latest tag",
			"user.azurecr.io/plugin",
			"oci://user.azurecr.io/plugin:latest",
		},
		{
			"should detect 127.0.0.1:5000 as most likely being an OCI registry",
			"127.0.0.1:5000/plugin:tag",
			"oci://127.0.0.1:5000/plugin:tag",
		},
		{
			"should detect 127.0.0.1:5000 as most likely being an OCI registry and tag it properly if no tag is supplied",
			"127.0.0.1:5000/plugin",
			"oci://127.0.0.1:5000/plugin:latest",
		},
		{
			"should detect localhost:5000 as most likely being an OCI registry and tag it properly if no tag is supplied",
			"localhost:5000/plugin",
			"oci://localhost:5000/plugin:latest",
		},
		{
			"should detect Quay",
			"quay.io/example/plugin:tag",
			"oci://quay.io/example/plugin:tag",
		},
		{
			"should detect localhost:32123/plugin:tag as most likely being an OCI registry",
			"localhost:32123/plugin:tag",
			"oci://localhost:32123/plugin:tag",
		},
		{
			"should detect 127.0.0.1:32123/plugin:tag as most likely being an OCI registry",
			"127.0.0.1:32123/plugin:tag",
			"oci://127.0.0.1:32123/plugin:tag",
		},
		{
			"should detect ::1:32123/plugin:tag as most likely being an OCI registry",
			"::1:32123/plugin:tag",
			"oci://::1:32123/plugin:tag",
		},
	}
	pwd := "/pwd"
	d := &OCIDetector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok, err := d.Detect(tt.input, pwd)
			if err != nil {
				t.Fatalf("OCIDetector.Detect() error = %v", err)
			}
			if !ok {
				t.Fatal("OCIDetector.Detect() not ok, should have detected")
			}
			if out != tt.expected {
				t.Errorf("OCIDetector.Detect() output = %v, want %v", out, tt.expected)
			}
		})
	}
}
