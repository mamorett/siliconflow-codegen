package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestGenerateQwencodeConfig(t *testing.T) {
	body := []byte(`{"data":[{"id":"b-model"},{"id":"a-model"},{"id":"  "}]}`)

	encoded, err := generateQwencodeConfig(body)
	if err != nil {
		t.Fatalf("generateQwencodeConfig returned error: %v", err)
	}

	var config qwencodeConfig
	if err := json.Unmarshal(encoded, &config); err != nil {
		t.Fatalf("generated config is not valid JSON: %v", err)
	}

	if len(config.OpenAI) != 2 {
		t.Fatalf("model count = %d, want 2", len(config.OpenAI))
	}

	wantIDs := []string{"a-model", "b-model"}
	for i, id := range wantIDs {
		model := config.OpenAI[i]
		if model.ID != id {
			t.Fatalf("model %d id = %q, want %q", i, model.ID, id)
		}
		if model.Name != id {
			t.Fatalf("model %d name = %q, want %q", i, model.Name, id)
		}
		if model.EnvKey != qwencodeAPIKey {
			t.Fatalf("model %d envKey = %q, want %q", i, model.EnvKey, qwencodeAPIKey)
		}
		if model.BaseURL != baseURL {
			t.Fatalf("model %d baseUrl = %q, want %q", i, model.BaseURL, baseURL)
		}
		if !model.GenerationConfig.Modalities["image"] {
			t.Fatalf("model %d image modality = false, want true", i)
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

func TestGenerateQwencodeConfigRejectsEmptyModelList(t *testing.T) {
	if _, err := generateQwencodeConfig([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("generateQwencodeConfig returned nil error for empty model list")
	}
}

func TestGenerateModelListJSON(t *testing.T) {
	body := []byte(`{"data":[{"id":"b-model"},{"id":"a-model"},{"id":"  "}]}`)

	encoded, err := generateModelListJSON(body)
	if err != nil {
		t.Fatalf("generateModelListJSON returned error: %v", err)
	}

	var ids []string
	if err := json.Unmarshal(encoded, &ids); err != nil {
		t.Fatalf("output is not a valid JSON array of strings: %v\nbody: %s", err, string(encoded))
	}

	want := []string{"a-model", "b-model"}
	if len(ids) != len(want) {
		t.Fatalf("ids length = %d, want %d", len(ids), len(want))
	}
	for i, id := range want {
		if ids[i] != id {
			t.Fatalf("ids[%d] = %q, want %q", i, ids[i], id)
		}
	}

	if len(encoded) == 0 || encoded[len(encoded)-1] != '\n' {
		t.Fatalf("output should end with a trailing newline, got %q", string(encoded))
	}
}

func TestGenerateModelListJSONRejectsEmptyModelList(t *testing.T) {
	if _, err := generateModelListJSON([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("generateModelListJSON returned nil error for empty model list")
	}
}

func TestRestartClaudeCodeRouterReturnsErrorWhenCommandMissing(t *testing.T) {
	// Isolate PATH to an empty directory so `ccr` is guaranteed not to be
	// found, regardless of what is installed on the test host.
	oldPath, hadPath := os.LookupEnv("PATH")
	defer func() {
		if hadPath {
			os.Setenv("PATH", oldPath)
		} else {
			os.Unsetenv("PATH")
		}
	}()

	emptyDir := t.TempDir()
	if err := os.Setenv("PATH", emptyDir); err != nil {
		t.Fatalf("failed to set PATH: %v", err)
	}

	if err := restartClaudeCodeRouter(); err == nil {
		t.Fatal("restartClaudeCodeRouter returned nil error when ccr is not in PATH")
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
func TestPromptForModelReturnsSelectedID(t *testing.T) {
	models := []string{"a-model", "b-model", "c-model"}
	input := bytes.NewBufferString("2\n")
	var output bytes.Buffer

	selected, err := promptForModel(models, input, &output)
	if err != nil {
		t.Fatalf("promptForModel returned error: %v", err)
	}

	if selected != "b-model" {
		t.Fatalf("selected = %q, want %q", selected, "b-model")
	}
	if !strings.Contains(output.String(), "1) a-model") {
		t.Fatalf("output does not list models:\n%s", output.String())
	}
}

func TestPromptForModelRejectsInvalidChoice(t *testing.T) {
	models := []string{"a-model", "b-model"}
	input := bytes.NewBufferString("bogus\n0\n5\n2\n")
	var output bytes.Buffer

	selected, err := promptForModel(models, input, &output)
	if err != nil {
		t.Fatalf("promptForModel returned error: %v", err)
	}

	if selected != "b-model" {
		t.Fatalf("selected = %q, want %q", selected, "b-model")
	}
}

func TestPromptForModelBlankCancels(t *testing.T) {
	models := []string{"a-model", "b-model"}
	input := bytes.NewBufferString("\n")
	var output bytes.Buffer

	if _, err := promptForModel(models, input, &output); err == nil {
		t.Fatal("promptForModel returned nil error for blank input")
	}
}

func TestWriteModelGridPutsIndicesInOrder(t *testing.T) {
	models := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot"}
	var output bytes.Buffer
	if err := writeModelGrid(&output, models, 3); err != nil {
		t.Fatalf("writeModelGrid returned error: %v", err)
	}

	for _, want := range []string{"1) alpha", "2) bravo", "3) charlie", "4) delta", "5) echo", "6) foxtrot"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("grid output missing %q:\n%s", want, output.String())
		}
	}
}

func TestChooseColumnCountPrefersSquareLayout(t *testing.T) {
	cases := []struct {
		name   string
		count  int
		idLen  int
		width  int
		expect int
	}{
		{name: "six models in 80 cols", count: 6, idLen: 20, width: 80, expect: 2},
		{name: "twelve models in 80 cols", count: 12, idLen: 12, width: 80, expect: 3},
		{name: "narrow terminal forces single column", count: 5, idLen: 60, width: 40, expect: 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := chooseColumnCount(c.count, c.idLen, c.width)
			if got != c.expect {
				t.Fatalf("chooseColumnCount(%d, %d, %d) = %d, want %d", c.count, c.idLen, c.width, got, c.expect)
			}
		})
	}
}

func TestPromptForModelRejectsEmptyList(t *testing.T) {
	if _, err := promptForModel(nil, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("promptForModel returned nil error for empty model list")
	}
}

func TestUpdateClaudeSettingsFile(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Test 1: File does not exist yet.
	err := updateClaudeSettingsFile(settingsPath, "deepseek-ai/DeepSeek-V3", "test-api-key")
	if err != nil {
		t.Fatalf("unexpected error updating non-existent settings file: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read created settings file: %v", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse created settings file JSON: %v", err)
	}

	envVal, ok := settings["env"]
	if !ok {
		t.Fatal("missing 'env' block in created settings")
	}
	envMap, ok := envVal.(map[string]interface{})
	if !ok {
		t.Fatal("'env' is not a JSON object")
	}

	if envMap["ANTHROPIC_BASE_URL"] != "http://localhost:3456" {
		t.Errorf("expected ANTHROPIC_BASE_URL to be 'http://localhost:3456', got %v", envMap["ANTHROPIC_BASE_URL"])
	}
	if envMap["ANTHROPIC_MODEL"] != "deepseek-ai/DeepSeek-V3" {
		t.Errorf("expected ANTHROPIC_MODEL to be 'deepseek-ai/DeepSeek-V3', got %v", envMap["ANTHROPIC_MODEL"])
	}
	if envMap["ANTHROPIC_API_KEY"] != "test-api-key" {
		t.Errorf("expected ANTHROPIC_API_KEY to be 'test-api-key', got %v", envMap["ANTHROPIC_API_KEY"])
	}
	if envMap["DISABLE_NON_ESSENTIAL_MODEL_CALLS"] != "1" {
		t.Errorf("expected DISABLE_NON_ESSENTIAL_MODEL_CALLS to be '1', got %v", envMap["DISABLE_NON_ESSENTIAL_MODEL_CALLS"])
	}
	if envMap["DISABLE_TELEMETRY"] != "1" {
		t.Errorf("expected DISABLE_TELEMETRY to be '1', got %v", envMap["DISABLE_TELEMETRY"])
	}
	if envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] != "1" {
		t.Errorf("expected CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS to be '1', got %v", envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"])
	}

	// Test 2: File exists and has unrelated settings that must be preserved.
	existingData := `{
  "customInstructions": "be concise",
  "env": {
    "OTHER_VAR": "keep-me",
    "ANTHROPIC_MODEL": "old-model"
  }
}`
	if err := os.WriteFile(settingsPath, []byte(existingData), 0644); err != nil {
		t.Fatalf("failed to write existing data: %v", err)
	}

	err = updateClaudeSettingsFile(settingsPath, "deepseek-ai/DeepSeek-Coder-V2-Instruct", "new-api-key")
	if err != nil {
		t.Fatalf("unexpected error updating existing settings file: %v", err)
	}

	data, err = os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read updated settings file: %v", err)
	}

	settings = nil
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse updated settings file JSON: %v", err)
	}

	if settings["customInstructions"] != "be concise" {
		t.Errorf("expected customInstructions to be preserved, got %v", settings["customInstructions"])
	}

	envVal, ok = settings["env"]
	if !ok {
		t.Fatal("missing 'env' block in updated settings")
	}
	envMap, ok = envVal.(map[string]interface{})
	if !ok {
		t.Fatal("'env' is not a JSON object")
	}

	if envMap["OTHER_VAR"] != "keep-me" {
		t.Errorf("expected OTHER_VAR to be preserved, got %v", envMap["OTHER_VAR"])
	}
	if envMap["ANTHROPIC_BASE_URL"] != "http://localhost:3456" {
		t.Errorf("expected ANTHROPIC_BASE_URL to be updated, got %v", envMap["ANTHROPIC_BASE_URL"])
	}
	if envMap["ANTHROPIC_MODEL"] != "deepseek-ai/DeepSeek-Coder-V2-Instruct" {
		t.Errorf("expected ANTHROPIC_MODEL to be updated, got %v", envMap["ANTHROPIC_MODEL"])
	}
	if envMap["ANTHROPIC_API_KEY"] != "new-api-key" {
		t.Errorf("expected ANTHROPIC_API_KEY to be updated, got %v", envMap["ANTHROPIC_API_KEY"])
	}
	if envMap["DISABLE_NON_ESSENTIAL_MODEL_CALLS"] != "1" {
		t.Errorf("expected DISABLE_NON_ESSENTIAL_MODEL_CALLS to be updated, got %v", envMap["DISABLE_NON_ESSENTIAL_MODEL_CALLS"])
	}
	if envMap["DISABLE_TELEMETRY"] != "1" {
		t.Errorf("expected DISABLE_TELEMETRY to be updated, got %v", envMap["DISABLE_TELEMETRY"])
	}
	if envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"] != "1" {
		t.Errorf("expected CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS to be updated, got %v", envMap["CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"])
	}

	// Test 3: Existing settings file has invalid JSON (must abort with error)
	invalidData := `{"env": {`
	if err := os.WriteFile(settingsPath, []byte(invalidData), 0644); err != nil {
		t.Fatalf("failed to write invalid data: %v", err)
	}

	err = updateClaudeSettingsFile(settingsPath, "some-model", "some-key")
	if err == nil {
		t.Fatal("expected error when updating settings file with invalid JSON, got nil")
	}

	// Test 4: Existing settings file has 'env' key that is not an object
	invalidEnvData := `{"env": "not-an-object"}`
	if err := os.WriteFile(settingsPath, []byte(invalidEnvData), 0644); err != nil {
		t.Fatalf("failed to write invalid env data: %v", err)
	}

	err = updateClaudeSettingsFile(settingsPath, "some-model", "some-key")
	if err == nil {
		t.Fatal("expected error when 'env' key is not an object, got nil")
	}
}

func TestUpdateClaudeCodeRouterFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Test 1: Config does not exist yet.
	err := updateClaudeCodeRouterFile(configPath, "deepseek-ai/DeepSeek-V3", "test-api-key")
	if err != nil {
		t.Fatalf("unexpected error updating non-existent router config file: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read created router config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse created router config file JSON: %v", err)
	}

	providersVal, ok := config["Providers"]
	if !ok {
		t.Fatal("missing 'Providers' block in created router config")
	}
	providersList, ok := providersVal.([]interface{})
	if !ok {
		t.Fatal("'Providers' is not a list")
	}
	if len(providersList) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providersList))
	}

	siliconflowProvider, ok := providersList[0].(map[string]interface{})
	if !ok {
		t.Fatal("provider is not a JSON object")
	}
	if siliconflowProvider["name"] != "siliconflow" {
		t.Errorf("expected provider name to be 'siliconflow', got %v", siliconflowProvider["name"])
	}
	if siliconflowProvider["api_base_url"] != "https://api.siliconflow.com/v1/chat/completions" {
		t.Errorf("expected api_base_url to be 'https://api.siliconflow.com/v1/chat/completions', got %v", siliconflowProvider["api_base_url"])
	}
	if siliconflowProvider["api_key"] != "test-api-key" {
		t.Errorf("expected api_key to be 'test-api-key', got %v", siliconflowProvider["api_key"])
	}

	transformerVal, ok := siliconflowProvider["transformer"]
	if !ok {
		t.Fatal("missing 'transformer' in provider")
	}
	transformerMap, ok := transformerVal.(map[string]interface{})
	if !ok {
		t.Fatal("'transformer' is not a JSON object")
	}
	useVal, ok := transformerMap["use"]
	if !ok {
		t.Fatal("missing 'use' in transformer")
	}
	useList, ok := useVal.([]interface{})
	if !ok || len(useList) != 1 || useList[0] != "OpenAI" {
		t.Errorf("expected transformer.use to contain 'OpenAI', got %v", useVal)
	}

	modelsVal, ok := siliconflowProvider["models"]
	if !ok {
		t.Fatal("missing 'models' in provider")
	}
	modelsList, ok := modelsVal.([]interface{})
	if !ok {
		t.Fatal("'models' is not a list")
	}
	if len(modelsList) != 1 || modelsList[0] != "deepseek-ai/DeepSeek-V3" {
		t.Errorf("expected models to contain 'deepseek-ai/DeepSeek-V3', got %v", modelsList)
	}

	routerVal, ok := config["Router"]
	if !ok {
		t.Fatal("missing 'Router' block")
	}
	routerMap, ok := routerVal.(map[string]interface{})
	if !ok {
		t.Fatal("'Router' is not a JSON object")
	}
	if routerMap["default"] != "siliconflow,deepseek-ai/DeepSeek-V3" {
		t.Errorf("expected Router.default to be 'siliconflow,deepseek-ai/DeepSeek-V3', got %v", routerMap["default"])
	}
	if routerMap["think"] != nil {
		t.Errorf("expected Router.think to be unset, got %v", routerMap["think"])
	}

	// Test 2: Reasoning model (containing "r1") updates "think" field
	err = updateClaudeCodeRouterFile(configPath, "deepseek-ai/DeepSeek-R1", "new-api-key")
	if err != nil {
		t.Fatalf("unexpected error updating existing router config file: %v", err)
	}

	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read updated router config: %v", err)
	}

	config = nil
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse updated config JSON: %v", err)
	}

	providersVal = config["Providers"]
	providersList = providersVal.([]interface{})
	siliconflowProvider = providersList[0].(map[string]interface{})
	modelsVal = siliconflowProvider["models"]
	modelsList = modelsVal.([]interface{})

	// Should contain both models now
	if len(modelsList) != 2 {
		t.Fatalf("expected 2 models, got %d (%v)", len(modelsList), modelsList)
	}
	if siliconflowProvider["api_key"] != "new-api-key" {
		t.Errorf("expected api_key to be updated to 'new-api-key', got %v", siliconflowProvider["api_key"])
	}

	transformerVal, ok = siliconflowProvider["transformer"]
	if !ok {
		t.Fatal("missing 'transformer' in provider after reasoning update")
	}
	transformerMap, ok = transformerVal.(map[string]interface{})
	if !ok {
		t.Fatal("'transformer' is not a JSON object")
	}
	useVal, ok = transformerMap["use"]
	if !ok {
		t.Fatal("missing 'use' in transformer")
	}
	useList, ok = useVal.([]interface{})
	if !ok || len(useList) != 2 || useList[0] != "OpenAI" || useList[1] != "reasoning" {
		t.Errorf("expected transformer.use to contain ['OpenAI', 'reasoning'] for reasoning model, got %v", useVal)
	}

	routerVal = config["Router"]
	routerMap = routerVal.(map[string]interface{})
	if routerMap["default"] != "siliconflow,deepseek-ai/DeepSeek-R1" {
		t.Errorf("expected Router.default to be 'siliconflow,deepseek-ai/DeepSeek-R1', got %v", routerMap["default"])
	}
	if routerMap["think"] != "siliconflow,deepseek-ai/DeepSeek-R1" {
		t.Errorf("expected Router.think to be 'siliconflow,deepseek-ai/DeepSeek-R1', got %v", routerMap["think"])
	}
}
