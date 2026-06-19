package main

import (
	"encoding/json"
	"testing"
)

func TestGenerateOpenCodeConfig(t *testing.T) {
	body := []byte(`{"data":[{"id":"b-model"},{"id":"a-model"},{"id":"  "}]}`)

	encoded, err := generateOpenCodeConfig(body)
	if err != nil {
		t.Fatalf("generateOpenCodeConfig returned error: %v", err)
	}

	var config openCodeConfig
	if err := json.Unmarshal(encoded, &config); err != nil {
		t.Fatalf("generated config is not valid JSON: %v", err)
	}

	if config.SiliconFlow.Type != opencodeType {
		t.Fatalf("type = %q, want %q", config.SiliconFlow.Type, opencodeType)
	}
	if config.SiliconFlow.BaseURL != baseURL {
		t.Fatalf("baseURL = %q, want %q", config.SiliconFlow.BaseURL, baseURL)
	}
	if config.SiliconFlow.APIKey != opencodeAPIKey {
		t.Fatalf("apiKey = %q, want %q", config.SiliconFlow.APIKey, opencodeAPIKey)
	}
	if len(config.SiliconFlow.Models) != 2 {
		t.Fatalf("model count = %d, want 2", len(config.SiliconFlow.Models))
	}

	wantIDs := []string{"a-model", "b-model"}
	for _, id := range wantIDs {
		model, ok := config.SiliconFlow.Models[id]
		if !ok {
			t.Fatalf("missing model %q", id)
		}
		if model.Name != id {
			t.Fatalf("model %q name = %q, want %q", id, model.Name, id)
		}
	}
}

func TestGenerateCrushConfig(t *testing.T) {
	body := []byte(`{"data":[{"id":"b-model"},{"id":"a-model"},{"id":"  "}]}`)

	encoded, err := generateCrushConfig(body)
	if err != nil {
		t.Fatalf("generateCrushConfig returned error: %v", err)
	}

	var config crushConfig
	if err := json.Unmarshal(encoded, &config); err != nil {
		t.Fatalf("generated config is not valid JSON: %v", err)
	}

	if config.Schema != crushSchemaURL {
		t.Fatalf("schema = %q, want %q", config.Schema, crushSchemaURL)
	}

	provider, ok := config.Providers[crushProviderKey]
	if !ok {
		t.Fatalf("missing provider %q", crushProviderKey)
	}
	if provider.Type != crushType {
		t.Fatalf("type = %q, want %q", provider.Type, crushType)
	}
	if provider.BaseURL != baseURL {
		t.Fatalf("base_url = %q, want %q", provider.BaseURL, baseURL)
	}
	if provider.APIKey != crushAPIKey {
		t.Fatalf("api_key = %q, want %q", provider.APIKey, crushAPIKey)
	}
	if len(provider.Models) != 2 {
		t.Fatalf("model count = %d, want 2", len(provider.Models))
	}

	wantIDs := []string{"a-model", "b-model"}
	for i, id := range wantIDs {
		if provider.Models[i].ID != id {
			t.Fatalf("model %d id = %q, want %q", i, provider.Models[i].ID, id)
		}
		if provider.Models[i].Name != id {
			t.Fatalf("model %d name = %q, want %q", i, provider.Models[i].Name, id)
		}
	}
}

func TestGenerateOpenCodeConfigRejectsEmptyModelList(t *testing.T) {
	if _, err := generateOpenCodeConfig([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("generateOpenCodeConfig returned nil error for empty model list")
	}
}

func TestGenerateCrushConfigRejectsEmptyModelList(t *testing.T) {
	if _, err := generateCrushConfig([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("generateCrushConfig returned nil error for empty model list")
	}
}

func TestParseModelIDsDeduplicatesAndSorts(t *testing.T) {
	ids, err := parseModelIDs([]byte(`{"data":[{"id":"b"},{"id":"a"},{"id":"b"},{"id":"  "}]}`))
	if err != nil {
		t.Fatalf("parseModelIDs returned error: %v", err)
	}

	want := []string{"a", "b"}
	if len(ids) != len(want) {
		t.Fatalf("ids length = %d, want %d", len(ids), len(want))
	}
	for i, id := range want {
		if ids[i] != id {
			t.Fatalf("ids[%d] = %q, want %q", i, ids[i], id)
		}
	}
}
