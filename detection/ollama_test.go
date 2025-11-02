package detection

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"image/color"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestNewOllamaProviderDefaults ensures the provider uses default host/model
func TestNewOllamaProviderDefaults(t *testing.T) {
	t.Setenv("IMGX_OLLAMA_HOST", "")
	t.Setenv("OLLAMA_HOST", "")
	t.Setenv("IMGX_OLLAMA_MODEL", "")

	provider, err := NewOllamaProvider()
	if err != nil {
		t.Fatalf("NewOllamaProvider() error = %v", err)
	}

	if provider.endpoint != defaultOllamaHost {
		t.Errorf("endpoint = %q, want %q", provider.endpoint, defaultOllamaHost)
	}

	if provider.model != defaultOllamaModel {
		t.Errorf("model = %q, want %q", provider.model, defaultOllamaModel)
	}
}

// TestOllamaDetectSuccess verifies a successful detection flow
func TestOllamaDetectSuccess(t *testing.T) {
	var capturedRequest ollamaGenerateRequest

	t.Setenv("IMGX_OLLAMA_HOST", "http://mock.local")
	t.Setenv("OLLAMA_HOST", "")
	t.Setenv("IMGX_OLLAMA_MODEL", "custom-model")

	provider, err := NewOllamaProvider()
	if err != nil {
		t.Fatalf("NewOllamaProvider() error = %v", err)
	}

	provider.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/api/generate" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request: %v", err)
			}
			_ = r.Body.Close()

			if err := json.Unmarshal(body, &capturedRequest); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if capturedRequest.Model != "custom-model" {
				t.Errorf("request model = %q, want %q", capturedRequest.Model, "custom-model")
			}
			if capturedRequest.Stream {
				t.Error("expected stream=false")
			}
			if capturedRequest.Format != "json" {
				t.Errorf("expected format=json, got %q", capturedRequest.Format)
			}
			if len(capturedRequest.Images) != 1 {
				t.Fatalf("expected 1 image, got %d", len(capturedRequest.Images))
			}
			if _, err := base64.StdEncoding.DecodeString(capturedRequest.Images[0]); err != nil {
				t.Fatalf("image payload is not valid base64: %v", err)
			}

			responseBody := `{"model":"custom-model","response":"{\"labels\":[{\"name\":\"cat\",\"confidence\":0.92}],\"description\":\"A cat\"}","done":true}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(responseBody)),
			}, nil
		}),
	}

	img := CreateTestImage(8, 8, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	opts := &DetectOptions{
		Features:           []Feature{FeatureLabels},
		MaxResults:         5,
		MinConfidence:      0.5,
		IncludeRawResponse: true,
	}

	result, err := provider.Detect(context.Background(), img, opts)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	AssertDetectionResult(t, result)

	if len(result.Labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(result.Labels))
	}

	if result.Labels[0].Name != "cat" {
		t.Errorf("label name = %q, want %q", result.Labels[0].Name, "cat")
	}

	if result.Labels[0].Confidence != 0.92 {
		t.Errorf("label confidence = %f, want 0.92", result.Labels[0].Confidence)
	}

	if result.Description != "A cat" {
		t.Errorf("description = %q, want %q", result.Description, "A cat")
	}

	if result.Properties["model"] != "custom-model" {
		t.Errorf("properties model = %q, want %q", result.Properties["model"], "custom-model")
	}

	if result.RawResponse == "" {
		t.Error("expected raw response to be captured")
	}
}

// TestOllamaDetectHTTPError ensures HTTP errors are surfaced
func TestOllamaDetectHTTPError(t *testing.T) {
	t.Setenv("IMGX_OLLAMA_HOST", "http://mock.local")
	t.Setenv("OLLAMA_HOST", "")
	t.Setenv("IMGX_OLLAMA_MODEL", "custom-model")

	provider, err := NewOllamaProvider()
	if err != nil {
		t.Fatalf("NewOllamaProvider() error = %v", err)
	}

	provider.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("server error")),
			}, nil
		}),
	}

	img := CreateTestImage(8, 8, color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	_, err = provider.Detect(context.Background(), img, DefaultDetectOptions())
	if err == nil {
		t.Fatal("Detect() expected error, got nil")
	}
}

// TestOllamaDetectBadJSON ensures invalid JSON payloads return errors
func TestOllamaDetectBadJSON(t *testing.T) {
	t.Setenv("IMGX_OLLAMA_HOST", "http://mock.local")
	t.Setenv("OLLAMA_HOST", "")
	t.Setenv("IMGX_OLLAMA_MODEL", "custom-model")

	provider, err := NewOllamaProvider()
	if err != nil {
		t.Fatalf("NewOllamaProvider() error = %v", err)
	}

	provider.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			responseBody := `{"model":"custom-model","response":123}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(responseBody)),
			}, nil
		}),
	}

	img := CreateTestImage(8, 8, color.NRGBA{R: 0, G: 255, B: 0, A: 255})

	_, err = provider.Detect(context.Background(), img, DefaultDetectOptions())
	if err == nil {
		t.Fatal("Detect() expected error for invalid response, got nil")
	}
}
