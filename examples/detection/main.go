package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/razzkumar/imgx"
	"github.com/razzkumar/imgx/detection"
)

func main() {
	// Check if an image path is provided
	if len(os.Args) < 2 {
		fmt.Println("Detection API Example")
		fmt.Println("=====================")
		fmt.Println()
		fmt.Println("This example demonstrates the imgx detection API with various providers and features.")
		fmt.Println()
		fmt.Println("Usage: go run main.go <image-path>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run main.go photo.jpg")
		fmt.Println("  go run main.go ../../testdata/branches.png")
		fmt.Println()
		fmt.Println("Setup:")
		fmt.Println("  Gemini:  export GEMINI_API_KEY=\"your-key\"")
		fmt.Println("  AWS:     aws configure  (or set AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)")
		fmt.Println("  OpenAI:  export OPENAI_API_KEY=\"sk-...\"")
		os.Exit(1)
	}

	imagePath := os.Args[1]

	// Load the image
	img, err := imgx.Load(imagePath)
	if err != nil {
		log.Fatalf("Failed to load image: %v", err)
	}

	fmt.Printf("Loaded: %s (%dx%d)\n", imagePath, img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	ctx := context.Background()

	// Example 1: Basic detection with Gemini (default)
	fmt.Println("Example 1: Basic Label Detection with Gemini")
	fmt.Println(strings.Repeat("-", 80))
	basicDetection(ctx, img)

	// Example 2: AWS with multiple features
	fmt.Println("\nExample 2: AWS with Multiple Features (labels, text, faces)")
	fmt.Println(strings.Repeat("-", 80))
	awsMultiFeature(ctx, img)

	// Example 3: AWS Image Properties
	fmt.Println("\nExample 3: AWS Image Properties (colors, quality)")
	fmt.Println(strings.Repeat("-", 80))
	awsImageProperties(ctx, img)

	// Example 4: Custom prompt with Gemini
	fmt.Println("\nExample 4: Custom Prompt with Gemini")
	fmt.Println(strings.Repeat("-", 80))
	customPrompt(ctx, img)

	// Example 5: Compare all providers
	fmt.Println("\nExample 5: Compare All Providers")
	fmt.Println(strings.Repeat("-", 80))
	compareProviders(ctx, img)

	// Example 6: JSON output
	fmt.Println("\nExample 6: JSON Output")
	fmt.Println(strings.Repeat("-", 80))
	jsonOutput(ctx, img)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("All examples completed!")
}

// Example 1: Basic label detection with default settings
func basicDetection(ctx context.Context, img *imgx.Image) {
	result, err := img.Detect(ctx, "gemini")
	if err != nil {
		handleError("Gemini", err)
		return
	}

	fmt.Println("✓ Detection successful")
	fmt.Printf("Provider: %s\n", result.Provider)
	fmt.Printf("Labels found: %d\n", len(result.Labels))

	if len(result.Labels) > 0 {
		fmt.Println("\nTop 5 labels:")
		count := len(result.Labels)
		if count > 5 {
			count = 5
		}
		for i := 0; i < count; i++ {
			label := result.Labels[i]
			fmt.Printf("  %d. %s (%.1f%% confidence)\n", i+1, label.Name, label.Confidence*100)
		}
	}
}

// Example 2: AWS with multiple features
func awsMultiFeature(ctx context.Context, img *imgx.Image) {
	opts := &detection.DetectOptions{
		Features: []detection.Feature{
			detection.FeatureLabels,
			detection.FeatureText,
			detection.FeatureFaces,
		},
		MaxResults:    10,
		MinConfidence: 0.7,
	}

	result, err := img.Detect(ctx, "aws", opts)
	if err != nil {
		handleError("AWS", err)
		return
	}

	fmt.Println("✓ Detection successful")
	fmt.Printf("Labels: %d\n", len(result.Labels))
	fmt.Printf("Text blocks: %d\n", len(result.Text))
	fmt.Printf("Faces: %d\n", len(result.Faces))

	if len(result.Labels) > 0 {
		fmt.Println("\nTop 3 labels:")
		count := len(result.Labels)
		if count > 3 {
			count = 3
		}
		for i := 0; i < count; i++ {
			label := result.Labels[i]
			fmt.Printf("  - %s (%.1f%%)\n", label.Name, label.Confidence*100)
		}
	}

	if len(result.Text) > 0 {
		fmt.Println("\nDetected text:")
		count := len(result.Text)
		if count > 3 {
			count = 3
		}
		for i := 0; i < count; i++ {
			text := result.Text[i]
			fmt.Printf("  - \"%s\" (%.1f%%)\n", text.Text, text.Confidence*100)
		}
	}

	if len(result.Faces) > 0 {
		fmt.Println("\nFace details:")
		for i, face := range result.Faces {
			details := []string{}
			if face.Confidence > 0 {
				details = append(details, fmt.Sprintf("confidence: %.1f%%", face.Confidence*100))
			}
			if face.AgeRange != "" {
				details = append(details, fmt.Sprintf("age: %s", face.AgeRange))
			}
			if face.Gender != "" {
				details = append(details, fmt.Sprintf("gender: %s", face.Gender))
			}
			fmt.Printf("  Face %d: %s\n", i+1, strings.Join(details, ", "))
		}
	}
}

// Example 3: AWS Image Properties
func awsImageProperties(ctx context.Context, img *imgx.Image) {
	opts := &detection.DetectOptions{
		Features: []detection.Feature{detection.FeatureProperties},
	}

	result, err := img.Detect(ctx, "aws", opts)
	if err != nil {
		handleError("AWS", err)
		return
	}

	fmt.Println("✓ Detection successful")

	if len(result.Properties) > 0 {
		fmt.Println("\nImage Properties:")

		// Quality metrics
		if val, ok := result.Properties["brightness"]; ok {
			fmt.Printf("  Brightness: %s\n", val)
		}
		if val, ok := result.Properties["sharpness"]; ok {
			fmt.Printf("  Sharpness: %s\n", val)
		}
		if val, ok := result.Properties["contrast"]; ok {
			fmt.Printf("  Contrast: %s\n", val)
		}

		// Dominant colors
		if val, ok := result.Properties["dominant_colors"]; ok {
			fmt.Printf("\n  Dominant Colors: %s\n", val)

			// Show hex and RGB values for first 3 colors
			for i := 1; i <= 3; i++ {
				hexKey := fmt.Sprintf("color_%d_hex", i)
				rgbKey := fmt.Sprintf("color_%d_rgb", i)
				if hex, hexOk := result.Properties[hexKey]; hexOk {
					if rgb, rgbOk := result.Properties[rgbKey]; rgbOk {
						fmt.Printf("    Color %d: %s %s\n", i, hex, rgb)
					}
				}
			}
		}
	}
}

// Example 4: Custom prompt with Gemini
func customPrompt(ctx context.Context, img *imgx.Image) {
	opts := &detection.DetectOptions{
		CustomPrompt: "Describe this image in detail, including objects, colors, and overall composition.",
	}

	result, err := img.Detect(ctx, "gemini", opts)
	if err != nil {
		handleError("Gemini", err)
		return
	}

	fmt.Println("✓ Detection successful")
	if result.Description != "" {
		fmt.Println("\nDescription:")
		// Word wrap at 80 characters
		words := strings.Fields(result.Description)
		line := "  "
		for _, word := range words {
			if len(line)+len(word)+1 > 80 {
				fmt.Println(line)
				line = "  " + word
			} else {
				if line != "  " {
					line += " "
				}
				line += word
			}
		}
		if line != "  " {
			fmt.Println(line)
		}
	}
}

// Example 5: Compare all providers
func compareProviders(ctx context.Context, img *imgx.Image) {
	providers := []string{"gemini", "aws", "openai"}
	results := make(map[string]*detection.DetectionResult)

	for _, provider := range providers {
		result, err := img.Detect(ctx, provider)
		if err != nil {
			if errors.Is(err, detection.ErrProviderNotConfigured) {
				fmt.Printf("%s: Not configured (skipped)\n", strings.ToUpper(provider))
				continue
			}
			fmt.Printf("%s: Error - %v\n", strings.ToUpper(provider), err)
			continue
		}
		results[provider] = result
		fmt.Printf("%s: ✓ Found %d labels\n", strings.ToUpper(provider), len(result.Labels))
	}

	if len(results) > 0 {
		fmt.Println("\nTop labels by provider:")
		for _, provider := range providers {
			result, ok := results[provider]
			if !ok {
				continue
			}
			fmt.Printf("\n  %s:\n", strings.ToUpper(provider))
			count := len(result.Labels)
			if count > 3 {
				count = 3
			}
			for i := 0; i < count; i++ {
				label := result.Labels[i]
				fmt.Printf("    - %s (%.1f%%)\n", label.Name, label.Confidence*100)
			}
		}
	}
}

// Example 6: JSON output
func jsonOutput(ctx context.Context, img *imgx.Image) {
	result, err := img.Detect(ctx, "gemini")
	if err != nil {
		handleError("Gemini", err)
		return
	}

	fmt.Println("✓ Detection successful")
	fmt.Println("\nJSON output (pretty-printed):")

	data, err := json.MarshalIndent(result, "  ", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	// Show first 500 characters to keep output manageable
	jsonStr := string(data)
	if len(jsonStr) > 500 {
		jsonStr = jsonStr[:500] + "..."
	}
	fmt.Println("  " + strings.ReplaceAll(jsonStr, "\n", "\n  "))
}

// handleError provides helpful error messages
func handleError(provider string, err error) {
	if errors.Is(err, detection.ErrProviderNotConfigured) {
		fmt.Printf("⚠ %s is not configured\n\n", provider)
		fmt.Println("Setup instructions:")
		switch strings.ToLower(provider) {
		case "gemini":
			fmt.Println("  export GEMINI_API_KEY=\"your-key\"")
			fmt.Println("  Get API key from: https://aistudio.google.com/")
		case "aws":
			fmt.Println("  Option 1: aws configure")
			fmt.Println("  Option 2: export AWS_ACCESS_KEY_ID=\"your-key\"")
			fmt.Println("            export AWS_SECRET_ACCESS_KEY=\"your-secret\"")
			fmt.Println("            export AWS_REGION=\"us-east-1\"")
		case "openai":
			fmt.Println("  export OPENAI_API_KEY=\"sk-...\"")
			fmt.Println("  Get API key from: https://platform.openai.com/")
		}
		return
	}

	fmt.Printf("✗ Detection failed: %v\n", err)
}
