package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	modelsEndpoint = "https://api.siliconflow.com/v1/models"
	baseURL        = "https://api.siliconflow.com/v1"

	opencodeProviderKey = "siliconflow"
	opencodeType        = "openai"
	opencodeAPIKey      = "${SILICONFLOW_API_KEY}"

	crushSchemaURL   = "https://charm.land/crush.json"
	crushProviderKey = "siliconflow"
	crushType        = "openai"
	crushAPIKey      = "$SILICONFLOW_API_KEY"
)

var (
	inputModalities  = []string{"text", "image", "video", "audio"}
	outputModalities = []string{"text"}
)

type apiResponse struct {
	Data []apiModel `json:"data"`
}

type apiModel struct {
	ID string `json:"id"`
}

type openCodeConfig struct {
	SiliconFlow openCodeProvider `json:"siliconflow"`
}

type openCodeProvider struct {
	Type    string                   `json:"type"`
	BaseURL string                   `json:"baseURL"`
	APIKey  string                   `json:"apiKey"`
	Models  map[string]openCodeModel `json:"models"`
}

type openCodeModel struct {
	Name       string     `json:"name"`
	Modalities modalities `json:"modalities"`
}

type modalities struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

type crushConfig struct {
	Schema    string                   `json:"$schema"`
	Providers map[string]crushProvider `json:"providers"`
}

type crushProvider struct {
	Type    string       `json:"type"`
	BaseURL string       `json:"base_url"`
	APIKey  string       `json:"api_key"`
	Models  []crushModel `json:"models"`
}

type crushModel struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	CostPer1MIn         float64 `json:"cost_per_1m_in,omitempty"`
	CostPer1MOut        float64 `json:"cost_per_1m_out,omitempty"`
	CostPer1MInCached   float64 `json:"cost_per_1m_in_cached,omitempty"`
	CostPer1MOutCached  float64 `json:"cost_per_1m_out_cached,omitempty"`
	ContextWindow       int     `json:"context_window,omitempty"`
	DefaultMaxTokens    int     `json:"default_max_tokens,omitempty"`
	CanReason           bool    `json:"can_reason,omitempty"`
	SupportsAttachments bool    `json:"supports_attachments,omitempty"`
}

func main() {
	genOpenCode := flag.Bool("gen-opencode", false, "generate an OpenCode-compatible SiliconFlow provider config instead of printing the raw API response")
	genCrush := flag.Bool("gen-crush", false, "generate a Crush-compatible SiliconFlow provider config instead of printing the raw API response")
	flag.Parse()

	if *genOpenCode && *genCrush {
		fmt.Fprintln(os.Stderr, "ERROR: --gen-opencode and --gen-crush cannot be used together")
		os.Exit(1)
	}

	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ERROR: SILICONFLOW_API_KEY is not set")
		os.Exit(1)
	}

	body, err := fetchModels(apiKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *genOpenCode:
		config, err := generateOpenCodeConfig(body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, err := os.Stdout.Write(config); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: writing output: %v\n", err)
			os.Exit(1)
		}
	case *genCrush:
		config, err := generateCrushConfig(body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, err := os.Stdout.Write(config); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: writing output: %v\n", err)
			os.Exit(1)
		}
	default:
		printRawResponse(body)
	}
}

func fetchModels(apiKey string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, modelsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ERROR: API returned status %d\nResponse: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	return body, nil
}

func generateOpenCodeConfig(body []byte) ([]byte, error) {
	ids, err := parseModelIDs(body)
	if err != nil {
		return nil, err
	}

	models := make(map[string]openCodeModel, len(ids))
	for _, id := range ids {
		models[id] = openCodeModel{
			Name: id,
			Modalities: modalities{
				Input:  append([]string(nil), inputModalities...),
				Output: append([]string(nil), outputModalities...),
			},
		}
	}

	if len(models) == 0 {
		return nil, errors.New("ERROR: no models found in API response")
	}

	config := openCodeConfig{
		SiliconFlow: openCodeProvider{
			Type:    opencodeType,
			BaseURL: baseURL,
			APIKey:  opencodeAPIKey,
			Models:  models,
		},
	}

	// encoding/json sorts map keys, but sort the values here too so generated
	// modalities remain stable even if the source slices are changed later.
	for id, model := range config.SiliconFlow.Models {
		sort.Strings(model.Modalities.Input)
		sort.Strings(model.Modalities.Output)
		config.SiliconFlow.Models[id] = model
	}

	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ERROR: encoding OpenCode config: %w", err)
	}

	return append(encoded, '\n'), nil
}

func generateCrushConfig(body []byte) ([]byte, error) {
	ids, err := parseModelIDs(body)
	if err != nil {
		return nil, err
	}

	models := make([]crushModel, 0, len(ids))
	for _, id := range ids {
		models = append(models, crushModel{
			ID:   id,
			Name: id,
		})
	}

	if len(models) == 0 {
		return nil, errors.New("ERROR: no models found in API response")
	}

	config := crushConfig{
		Schema: crushSchemaURL,
		Providers: map[string]crushProvider{
			crushProviderKey: {
				Type:    crushType,
				BaseURL: baseURL,
				APIKey:  crushAPIKey,
				Models:  models,
			},
		},
	}

	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ERROR: encoding Crush config: %w", err)
	}

	return append(encoded, '\n'), nil
}

func parseModelIDs(body []byte) ([]string, error) {
	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ERROR: parsing API response: %w", err)
	}

	seen := make(map[string]struct{}, len(response.Data))
	for _, model := range response.Data {
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}

		seen[id] = struct{}{}
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	return ids, nil
}

func printRawResponse(body []byte) {
	if len(body) == 0 {
		return
	}

	os.Stdout.Write(body)
	if body[len(body)-1] != '\n' {
		os.Stdout.WriteString("\n")
	}
}
