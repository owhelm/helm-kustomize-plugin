package parser

import (
	"strings"
	"testing"
)

func TestParseManifests_KustomizePluginDataDetection(t *testing.T) {
	tests := []struct {
		name                    string
		input                   string
		wantKustomizePluginData bool
		wantOtherResourcesCount int
	}{
		{
			name: "valid KustomizePluginData resource",
			input: `---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
metadata:
  name: test
files:
  test.yaml: content
`,
			wantKustomizePluginData: true,
			wantOtherResourcesCount: 0,
		},
		{
			name: "wrong apiVersion",
			input: `---
apiVersion: v1
kind: KustomizePluginData
metadata:
  name: test
`,
			wantKustomizePluginData: false,
			wantOtherResourcesCount: 1,
		},
		{
			name: "wrong kind",
			input: `---
apiVersion: helm.plugin.kustomize/v1
kind: ConfigMap
metadata:
  name: test
`,
			wantKustomizePluginData: false,
			wantOtherResourcesCount: 1,
		},
		{
			name: "missing apiVersion",
			input: `---
kind: KustomizePluginData
metadata:
  name: test
`,
			wantKustomizePluginData: false,
			wantOtherResourcesCount: 1,
		},
		{
			name: "missing kind",
			input: `---
apiVersion: helm.plugin.kustomize/v1
metadata:
  name: test
`,
			wantKustomizePluginData: false,
			wantOtherResourcesCount: 1,
		},
		{
			name: "empty resource",
			input: `---
{}
`,
			wantKustomizePluginData: false,
			wantOtherResourcesCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseManifests([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseManifests() error = %v, want nil", err)
			}

			hasKustomizePluginData := result.KustomizePluginData != nil
			if hasKustomizePluginData != tt.wantKustomizePluginData {
				t.Errorf("KustomizePluginData presence = %v, want %v", hasKustomizePluginData, tt.wantKustomizePluginData)
			}

			if len(result.OtherResources) != tt.wantOtherResourcesCount {
				t.Errorf("OtherResources count = %d, want %d", len(result.OtherResources), tt.wantOtherResourcesCount)
			}
		})
	}
}

func TestParseManifests_EmptyInput(t *testing.T) {
	input := []byte("")
	result, err := ParseManifests(input)

	if err != nil {
		t.Fatalf("ParseManifests() error = %v, want nil", err)
	}

	if result.KustomizePluginData != nil {
		t.Errorf("Expected no KustomizePluginData, got %v", result.KustomizePluginData)
	}

	if len(result.OtherResources) != 0 {
		t.Errorf("Expected no OtherResources, got %d", len(result.OtherResources))
	}
}

func TestParseManifests_NoKustomizePluginData(t *testing.T) {
	input := []byte(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`)

	result, err := ParseManifests(input)
	if err != nil {
		t.Fatalf("ParseManifests() error = %v, want nil", err)
	}

	if result.KustomizePluginData != nil {
		t.Errorf("Expected no KustomizePluginData, got %v", result.KustomizePluginData)
	}

	if len(result.OtherResources) != 2 {
		t.Errorf("Expected 2 OtherResources, got %d", len(result.OtherResources))
	}
}

func TestParseManifests_WithKustomizePluginData(t *testing.T) {
	input := []byte(`---
apiVersion: v1
kind: Service
metadata:
  name: test-service
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
metadata:
  name: kustomize-files
files:
  kustomization.yaml: |
    resources:
    - all.yaml
  patch.yaml: |
    apiVersion: apps/v1
    kind: Deployment
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`)

	result, err := ParseManifests(input)
	if err != nil {
		t.Fatalf("ParseManifests() error = %v, want nil", err)
	}

	if result.KustomizePluginData == nil {
		t.Fatal("Expected KustomizePluginData, got nil")
	}

	if result.KustomizePluginData.APIVersion != APIVersion {
		t.Errorf("Expected apiVersion %s, got %s", APIVersion, result.KustomizePluginData.APIVersion)
	}

	if result.KustomizePluginData.Kind != Kind {
		t.Errorf("Expected kind %s, got %s", Kind, result.KustomizePluginData.Kind)
	}

	if len(result.KustomizePluginData.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.KustomizePluginData.Files))
	}

	if _, ok := result.KustomizePluginData.Files["kustomization.yaml"]; !ok {
		t.Error("Expected kustomization.yaml file")
	}

	if _, ok := result.KustomizePluginData.Files["patch.yaml"]; !ok {
		t.Error("Expected patch.yaml file")
	}

	// Should have 2 other resources (Service and Deployment)
	if len(result.OtherResources) != 2 {
		t.Errorf("Expected 2 OtherResources, got %d", len(result.OtherResources))
	}
}

func TestParseManifests_MultipleKustomizePluginData(t *testing.T) {
	input := []byte(`---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
metadata:
  name: first
files:
  file1.yaml: content1
---
apiVersion: helm.plugin.kustomize/v1
kind: KustomizePluginData
metadata:
  name: second
files:
  file2.yaml: content2
`)

	_, err := ParseManifests(input)
	if err == nil {
		t.Fatal("Expected error for multiple KustomizePluginData, got nil")
	}

	if !strings.Contains(err.Error(), "multiple KustomizePluginData") {
		t.Errorf("Expected error about multiple KustomizePluginData, got: %v", err)
	}
}

func TestParseManifests_InvalidYAML(t *testing.T) {
	input := []byte(`---
this is not: valid: yaml: structure
  bad indentation
`)

	_, err := ParseManifests(input)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestMarshalResources(t *testing.T) {
	resources := []map[string]any{
		{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]any{
				"name": "test-service",
			},
		},
		{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name": "test-deployment",
			},
		},
	}

	data, err := MarshalResources(resources)
	if err != nil {
		t.Fatalf("MarshalResources() error = %v, want nil", err)
	}

	expected := `apiVersion: v1
kind: Service
metadata:
    name: test-service
---
apiVersion: apps/v1
kind: Deployment
metadata:
    name: test-deployment
`

	got := string(data)
	if got != expected {
		t.Errorf("MarshalResources() output mismatch\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestMarshalResources_Empty(t *testing.T) {
	var resources []map[string]any

	data, err := MarshalResources(resources)
	if err != nil {
		t.Fatalf("MarshalResources() error = %v, want nil", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty output, got %d bytes", len(data))
	}
}
