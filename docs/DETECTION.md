# Image Detection API Documentation

Advanced AI-powered object detection for images using Google Gemini, AWS Rekognition, and OpenAI Vision APIs.

## Table of Contents

- [Overview](#overview)
- [Supported Providers](#supported-providers)
- [Setup & Authentication](#setup--authentication)
- [Quick Start](#quick-start)
- [Detection Features](#detection-features)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Pricing Considerations](#pricing-considerations)
- [Troubleshooting](#troubleshooting)

## Overview

The detection package provides a unified interface for multiple AI vision providers, allowing you to:
- Detect objects and labels in images
- Extract text (OCR)
- Detect faces and facial attributes
- Analyze image properties (colors, quality, sharpness)
- Check for inappropriate content
- Get natural language descriptions

All providers return results in a consistent format, making it easy to switch between providers or compare results.

## Supported Providers

| Provider | API Key Required | Features |
|----------|------------------|----------|
| **Google Gemini** | `GEMINI_API_KEY` | Labels, Text, Faces, Description, Web detection, Landmarks |
| **AWS Rekognition** | AWS credentials | Labels, Text, Faces, Image Properties, Moderation |
| **OpenAI Vision** | `OPENAI_API_KEY` | Labels, Description, Text, Faces (via GPT-4o) |

## Setup & Authentication

### Google Gemini

1. Get API key from [Google AI Studio](https://aistudio.google.com/)
2. Set environment variable:
```bash
export GEMINI_API_KEY="your-api-key"
```

### AWS Rekognition

AWS uses the standard credential chain:

**Option 1: Environment Variables**
```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"
```

**Option 2: AWS CLI Configuration**
```bash
aws configure
```

**Option 3: AWS Profile**
```bash
export AWS_PROFILE="myprofile"
```

**Option 4: IAM Roles** (automatic on EC2/ECS/Lambda)

### OpenAI Vision

1. Get API key from [OpenAI Platform](https://platform.openai.com/)
2. Set environment variable:
```bash
export OPENAI_API_KEY="sk-..."
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
	// Load an image
	img, err := imgx.Load("photo.jpg")
	if err != nil {
		log.Fatal(err)
	}

	// Detect objects using Gemini
	ctx := context.Background()
	result, err := img.Detect(ctx, "gemini")
	if err != nil {
		log.Fatal(err)
	}

	// Print detected labels
	fmt.Println("Detected objects:")
	for _, label := range result.Labels {
		fmt.Printf("- %s (%.1f%% confidence)\n",
			label.Name, label.Confidence*100)
	}
}
```

## Detection Features

### Available Features

```go
const (
	FeatureLabels      Feature = "labels"       // Object/label detection
	FeatureObjects     Feature = "objects"      // Alias for labels
	FeatureText        Feature = "text"         // OCR text extraction
	FeatureFaces       Feature = "faces"        // Face detection
	FeatureDescription Feature = "description"  // Natural language description
	FeatureWeb         Feature = "web"          // Web entities (Gemini only)
	FeatureLandmarks   Feature = "landmarks"    // Landmark detection (Gemini only)
	FeatureProperties  Feature = "properties"   // Image properties (AWS only)
	FeatureSafeSearch  Feature = "safesearch"   // Content moderation
)
```

### Feature Support Matrix

| Feature | Gemini | AWS | OpenAI |
|---------|--------|-----|--------|
| Labels | ✅ | ✅ | ✅ |
| Text (OCR) | ✅ | ✅ | ✅ |
| Faces | ✅ | ✅ | ✅ |
| Description | ✅ | ❌ | ✅ |
| Web Detection | ✅ | ❌ | ❌ |
| Landmarks | ✅ | ❌ | ❌ |
| Properties | ❌ | ✅ | ❌ |
| SafeSearch/Moderation | ✅ | ✅ | ✅ |

## API Reference

### Core Types

#### DetectionResult

```go
type DetectionResult struct {
	Provider      string                 // Provider used (gemini, aws, openai)
	Labels        []Label                // Detected objects/labels
	Description   string                 // Natural language description
	Text          []TextBlock            // Extracted text
	Faces         []Face                 // Detected faces
	Web           *WebDetection          // Web entities (Gemini)
	BoundingBoxes []BoundingBox          // Object locations
	Properties    map[string]string      // Image properties (AWS)
	Confidence    float32                // Overall confidence (0.0-1.0)
	ProcessedAt   time.Time              // Processing timestamp
	RawResponse   string                 // Raw API response (if requested)
}
```

#### Label

```go
type Label struct {
	Name       string   // Label name (e.g., "Dog", "Car")
	Confidence float32  // Confidence score (0.0-1.0)
	Categories []string // Parent categories
}
```

#### TextBlock

```go
type TextBlock struct {
	Text       string       // Extracted text
	Confidence float32      // Confidence score (0.0-1.0)
	BoundingBox *BoundingBox // Text location
	Language   string       // Detected language (if available)
}
```

#### Face

```go
type Face struct {
	Confidence  float32            // Confidence score (0.0-1.0)
	BoundingBox *BoundingBox       // Face location
	Landmarks   []FaceLandmark     // Facial landmarks (eyes, nose, etc.)
	Emotions    map[string]float32 // Emotion scores
	AgeRange    string             // Estimated age range
	Gender      string             // Detected gender
}
```

### Detection Options

```go
type DetectOptions struct {
	Features           []Feature // Features to detect
	MaxResults         int       // Maximum labels to return (default: 10)
	MinConfidence      float32   // Minimum confidence threshold (0.0-1.0, default: 0.5)
	CustomPrompt       string    // Custom prompt (Gemini/OpenAI)
	IncludeRawResponse bool      // Include raw API response
}

// Create default options
opts := detection.DefaultDetectOptions()

// Customize options
opts := &detection.DetectOptions{
	Features:      []detection.Feature{detection.FeatureLabels, detection.FeatureText},
	MaxResults:    20,
	MinConfidence: 0.7,
}
```

### Methods

#### Image.Detect()

```go
func (img *Image) Detect(ctx context.Context, provider string,
	opts ...*detection.DetectOptions) (*detection.DetectionResult, error)
```

High-level method on `*imgx.Image` instances.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `provider`: Provider name ("gemini", "google", "aws", "rekognition", "openai")
- `opts`: Optional detection options

**Returns:**
- `*DetectionResult`: Detection results
- `error`: Error if detection fails

#### Provider Interface

```go
type Provider interface {
	Detect(ctx context.Context, img *image.NRGBA,
		opts *DetectOptions) (*DetectionResult, error)
	Name() string
	IsConfigured() bool
}
```

Direct provider access for advanced use cases:

```go
provider, err := detection.GetProvider("gemini")
if err != nil {
	log.Fatal(err)
}

result, err := provider.Detect(ctx, img.ToNRGBA(), opts)
```

## Examples

### Basic Detection

```go
img, _ := imgx.Load("photo.jpg")
ctx := context.Background()

result, err := img.Detect(ctx, "gemini")
if err != nil {
	log.Fatal(err)
}

for _, label := range result.Labels {
	fmt.Printf("%s: %.1f%%\n", label.Name, label.Confidence*100)
}
```

### Multiple Features

```go
opts := &detection.DetectOptions{
	Features: []detection.Feature{
		detection.FeatureLabels,
		detection.FeatureText,
		detection.FeatureFaces,
	},
	MaxResults:    15,
	MinConfidence: 0.7,
}

result, err := img.Detect(ctx, "aws", opts)
if err != nil {
	log.Fatal(err)
}

// Labels
fmt.Println("Objects:", len(result.Labels))

// Text extraction
fmt.Println("Text found:")
for _, text := range result.Text {
	fmt.Printf("- %s\n", text.Text)
}

// Faces
fmt.Printf("Found %d faces\n", len(result.Faces))
```

### AWS Image Properties

```go
// Get image quality metrics and dominant colors
opts := &detection.DetectOptions{
	Features: []detection.Feature{detection.FeatureProperties},
}

result, err := img.Detect(ctx, "aws", opts)
if err != nil {
	log.Fatal(err)
}

// Access properties
fmt.Println("Brightness:", result.Properties["brightness"])
fmt.Println("Sharpness:", result.Properties["sharpness"])
fmt.Println("Contrast:", result.Properties["contrast"])
fmt.Println("Dominant colors:", result.Properties["dominant_colors"])
fmt.Println("Color 1 (hex):", result.Properties["color_1_hex"])
fmt.Println("Color 1 (rgb):", result.Properties["color_1_rgb"])
```

### Custom Prompt (Gemini/OpenAI)

```go
opts := &detection.DetectOptions{
	CustomPrompt: "Is there a dog in this image? What breed might it be?",
}

result, err := img.Detect(ctx, "gemini", opts)
if err != nil {
	log.Fatal(err)
}

fmt.Println("Description:", result.Description)
```

### Compare Multiple Providers

```go
providers := []string{"gemini", "aws", "openai"}

for _, provider := range providers {
	result, err := img.Detect(ctx, provider)
	if err != nil {
		fmt.Printf("%s error: %v\n", provider, err)
		continue
	}

	fmt.Printf("\n%s results:\n", strings.ToUpper(provider))
	for _, label := range result.Labels[:5] {
		fmt.Printf("  - %s (%.1f%%)\n", label.Name, label.Confidence*100)
	}
}
```

### Error Handling

```go
result, err := img.Detect(ctx, "aws")
if err != nil {
	// Check for specific error types
	if errors.Is(err, detection.ErrProviderNotConfigured) {
		log.Fatal("AWS credentials not configured. Run: aws configure")
	}

	// Check error message for details
	if strings.Contains(err.Error(), "invalid credentials") {
		log.Fatal("AWS credentials are invalid or expired")
	}

	log.Fatal(err)
}
```

### With Context Timeout

```go
// Set 10 second timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := img.Detect(ctx, "gemini")
if err != nil {
	if errors.Is(err, context.DeadlineExceeded) {
		log.Fatal("Detection timed out")
	}
	log.Fatal(err)
}
```

### Batch Processing

```go
images := []string{"photo1.jpg", "photo2.jpg", "photo3.jpg"}
ctx := context.Background()

for _, imagePath := range images {
	img, err := imgx.Load(imagePath)
	if err != nil {
		log.Printf("Failed to load %s: %v", imagePath, err)
		continue
	}

	result, err := img.Detect(ctx, "gemini")
	if err != nil {
		log.Printf("Failed to detect %s: %v", imagePath, err)
		continue
	}

	fmt.Printf("\n%s:\n", imagePath)
	for i, label := range result.Labels {
		if i >= 3 { break } // Top 3 labels
		fmt.Printf("  %s (%.1f%%)\n", label.Name, label.Confidence*100)
	}
}
```

## Best Practices

### 1. Choose the Right Provider

- **Gemini**: Best for general-purpose detection, custom prompts, and web entity detection
- **AWS Rekognition**: Best for production workloads, image properties, and when you need reliable face detection
- **OpenAI Vision**: Best for natural language descriptions and when you need GPT-4o's reasoning

### 2. Set Appropriate Confidence Thresholds

```go
// For critical applications, use higher confidence
opts := &detection.DetectOptions{
	MinConfidence: 0.8, // Only high-confidence results
}

// For exploratory analysis, use lower confidence
opts := &detection.DetectOptions{
	MinConfidence: 0.3, // Catch more possibilities
}
```

### 3. Handle Rate Limits

```go
import "time"

for _, img := range images {
	result, err := img.Detect(ctx, "gemini")
	if err != nil {
		if strings.Contains(err.Error(), "rate limit") {
			time.Sleep(2 * time.Second)
			result, err = img.Detect(ctx, "gemini") // Retry
		}
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
	}
	// Process result...
}
```

### 4. Use Context for Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Cancel on Ctrl+C
go func() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	cancel()
}()

result, err := img.Detect(ctx, "gemini")
```

### 5. Cache Results

```go
type DetectionCache struct {
	mu    sync.RWMutex
	cache map[string]*detection.DetectionResult
}

func (c *DetectionCache) Get(key string) (*detection.DetectionResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result, ok := c.cache[key]
	return result, ok
}

func (c *DetectionCache) Set(key string, result *detection.DetectionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = result
}
```

## Pricing Considerations

### AWS Rekognition

**Important**: The `properties` feature has separate pricing.

- **Labels only**: `--features labels` → Standard DetectLabels pricing
- **Properties only**: `--features properties` → Image Properties pricing only
- **Both**: `--features labels,properties` → Charged for BOTH APIs

Example costs (as of 2024):
- DetectLabels: ~$0.001 per image (first 1M images/month)
- Image Properties: Additional charge when combined with labels
- DetectText, DetectFaces: Separate pricing

**Recommendation**: Only request properties when needed to minimize costs.

### Gemini

- Free tier available with rate limits
- Pay-as-you-go pricing after free tier
- Custom prompts may use more tokens

### OpenAI Vision

- Charged per API call
- GPT-4o has different pricing than GPT-4
- Image size affects cost

## Troubleshooting

### Common Issues

#### 1. "Provider not configured"

**Gemini:**
```bash
# Check if key is set
echo $GEMINI_API_KEY

# Set the key
export GEMINI_API_KEY="your-key"
```

**AWS:**
```bash
# Test credentials
aws sts get-caller-identity

# If that fails, run
aws configure
```

**OpenAI:**
```bash
# Check if key is set
echo $OPENAI_API_KEY

# Set the key
export OPENAI_API_KEY="sk-..."
```

#### 2. "Invalid AWS credentials"

This error means your AWS credentials are incorrect or expired.

```bash
# Verify credentials are correct
aws sts get-caller-identity

# If using temporary credentials, they may have expired
# Get new credentials from your IAM administrator

# Check which credential source is being used
AWS_PROFILE=default aws sts get-caller-identity
```

#### 3. "Access denied" (AWS)

Your IAM user/role lacks necessary permissions.

Required IAM permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rekognition:DetectLabels",
        "rekognition:DetectText",
        "rekognition:DetectFaces",
        "rekognition:DetectModerationLabels"
      ],
      "Resource": "*"
    }
  ]
}
```

#### 4. Rate Limiting

If you hit rate limits, implement exponential backoff:

```go
func detectWithRetry(img *imgx.Image, ctx context.Context, provider string) (*detection.DetectionResult, error) {
	maxRetries := 3
	baseDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		result, err := img.Detect(ctx, provider)
		if err == nil {
			return result, nil
		}

		if !strings.Contains(err.Error(), "rate limit") {
			return nil, err
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i))
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("max retries exceeded")
}
```

#### 5. Large Images

If you get errors about image size:

```go
// Resize large images before detection
if img.Bounds().Dx() > 2048 || img.Bounds().Dy() > 2048 {
	img = img.Fit(2048, 2048, imgx.Lanczos)
}

result, err := img.Detect(ctx, "gemini")
```

## Additional Resources

- [Google Gemini API Documentation](https://ai.google.dev/docs)
- [AWS Rekognition Documentation](https://docs.aws.amazon.com/rekognition/)
- [OpenAI Vision API Documentation](https://platform.openai.com/docs/guides/vision)
- [imgx GitHub Repository](https://github.com/razzkumar/imgx)
- [Example Code](../examples/detection/main.go)

## See Also

- [CLI Documentation](./CLI.md) - Command-line usage
- [Main README](../README.md) - Library overview
- [Versioning](./VERSIONING.md) - Version management
