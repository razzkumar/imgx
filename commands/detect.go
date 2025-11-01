package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/razzkumar/imgx/detection"
	"github.com/urfave/cli/v3"
)

// DetectCommand creates the detect command
func DetectCommand() *cli.Command {
	return &cli.Command{
		Name:  "detect",
		Usage: "Detect objects in images using AI vision APIs",
		Description: `Perform object detection using Google Gemini, AWS Rekognition,
or OpenAI Vision APIs.

The detection results include labels, confidence scores, and can optionally include
text detection (OCR), face detection, and more depending on the provider.

Supported Providers:
  gemini          Google Gemini API (default, requires GEMINI_API_KEY)
  google          Alias for gemini
  aws             AWS Rekognition (uses AWS credential chain)
  openai          OpenAI Vision API (requires OPENAI_API_KEY)

Setup:
  Gemini:    Get API key from https://aistudio.google.com/
             export GEMINI_API_KEY="your-api-key"

  AWS:       Uses standard AWS credential chain. Configure with any of:
             - Environment variables:
               export AWS_ACCESS_KEY_ID="your-key"
               export AWS_SECRET_ACCESS_KEY="your-secret"
               export AWS_REGION="us-east-1"
             - AWS CLI configuration: aws configure
             - IAM roles (for EC2, ECS, Lambda)
             - Shared credentials file (~/.aws/credentials)

  OpenAI:    export OPENAI_API_KEY="sk-..."

Examples:
  # Detect objects using Gemini (simplest setup)
  imgx detect --provider gemini input.jpg

  # Using "google" alias (same as gemini)
  imgx detect --provider google input.jpg

  # Detect with specific features
  imgx detect --provider gemini --features labels,text input.jpg

  # Custom prompt (Gemini/OpenAI)
  imgx detect --provider gemini --prompt "Is there a dog in this image?" input.jpg

  # Output as JSON
  imgx detect --provider aws --json input.jpg

  # Higher confidence threshold
  imgx detect --provider gemini --confidence 0.8 input.jpg`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "provider",
				Aliases:  []string{"p"},
				Usage:    "Detection provider: gemini, google (alias), aws, openai",
				Value:    "gemini",
				Required: false,
			},
			&cli.StringFlag{
				Name:    "features",
				Aliases: []string{"f"},
				Usage:   "Features to detect: labels,text,faces,web,description (comma-separated)",
				Value:   "labels",
			},
			&cli.IntFlag{
				Name:    "max-results",
				Aliases: []string{"m"},
				Usage:   "Maximum number of labels to return",
				Value:   10,
			},
			&cli.Float64Flag{
				Name:    "confidence",
				Aliases: []string{"c"},
				Usage:   "Minimum confidence threshold (0.0-1.0)",
				Value:   0.5,
			},
			&cli.StringFlag{
				Name:  "prompt",
				Usage: "Custom prompt for Gemini/OpenAI (overrides --features)",
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output results as JSON",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "raw",
				Usage: "Include raw API response in output",
				Value: false,
			},
		},
		Action: detectAction,
	}
}

func detectAction(ctx context.Context, cmd *cli.Command) error {
	// Validate input
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	provider := cmd.String("provider")

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Prepare detection options
	opts := &detection.DetectOptions{
		Features:           detection.ParseFeatures(cmd.String("features")),
		MaxResults:         cmd.Int("max-results"),
		MinConfidence:      float32(cmd.Float64("confidence")),
		CustomPrompt:       cmd.String("prompt"),
		IncludeRawResponse: cmd.Bool("raw"),
	}

	// Perform detection
	result, err := img.Detect(ctx, provider, opts)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	// Output results
	if cmd.Bool("json") {
		return outputDetectionJSON(result)
	}

	return outputDetectionPretty(result, provider)
}

// outputDetectionJSON outputs detection results as JSON
func outputDetectionJSON(result *detection.DetectionResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// outputDetectionPretty outputs detection results in a human-readable format
func outputDetectionPretty(result *detection.DetectionResult, provider string) error {
	fmt.Printf("=== Object Detection Results (%s) ===\n\n", result.Provider)

	// Labels
	if len(result.Labels) > 0 {
		fmt.Println("Labels:")
		for i, label := range result.Labels {
			fmt.Printf("  %d. %s (%.1f%% confidence)\n", i+1, label.Name, label.Confidence*100)
		}
		fmt.Println()
	}

	// Description
	if result.Description != "" {
		fmt.Println("Description:")
		// Word wrap description at 80 characters
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
		fmt.Println()
	}

	// Text (OCR)
	if len(result.Text) > 0 {
		fmt.Println("Detected Text:")
		for i, text := range result.Text {
			if text.Confidence > 0 {
				fmt.Printf("  %d. \"%s\" (%.1f%% confidence)\n", i+1, text.Text, text.Confidence*100)
			} else {
				fmt.Printf("  %d. \"%s\"\n", i+1, text.Text)
			}
		}
		fmt.Println()
	}

	// Faces
	if len(result.Faces) > 0 {
		fmt.Printf("Faces Detected: %d\n", len(result.Faces))
		for i, face := range result.Faces {
			fmt.Printf("  Face %d:\n", i+1)
			if face.Confidence > 0 {
				fmt.Printf("    Confidence: %.1f%%\n", face.Confidence*100)
			}
			if face.JoyLikelihood != "" {
				fmt.Printf("    Joy: %s\n", face.JoyLikelihood)
			}
			if face.SorrowLikelihood != "" {
				fmt.Printf("    Sorrow: %s\n", face.SorrowLikelihood)
			}
			if face.AngerLikelihood != "" {
				fmt.Printf("    Anger: %s\n", face.AngerLikelihood)
			}
			if face.Gender != "" {
				fmt.Printf("    Gender: %s\n", face.Gender)
			}
			if face.AgeRange != "" {
				fmt.Printf("    Age Range: %s\n", face.AgeRange)
			}
		}
		fmt.Println()
	}

	// Web Detection
	if result.Web != nil {
		if len(result.Web.WebEntities) > 0 {
			fmt.Println("Web Entities:")
			for i, entity := range result.Web.WebEntities {
				if i >= 5 {
					break // Limit to top 5
				}
				fmt.Printf("  - %s (score: %.2f)\n", entity.Description, entity.Score)
			}
			fmt.Println()
		}

		if len(result.Web.BestGuessLabels) > 0 {
			fmt.Println("Best Guess Labels:")
			for _, label := range result.Web.BestGuessLabels {
				fmt.Printf("  - %s\n", label)
			}
			fmt.Println()
		}
	}

	// Properties
	if len(result.Properties) > 0 {
		fmt.Println("Properties:")
		for key, value := range result.Properties {
			fmt.Printf("  %s: %s\n", key, value)
		}
		fmt.Println()
	}

	// Bounding Boxes
	if len(result.BoundingBoxes) > 0 {
		fmt.Printf("Objects with Locations: %d\n", len(result.BoundingBoxes))
		for i, bbox := range result.BoundingBoxes {
			if i >= 10 {
				break // Limit to top 10
			}
			fmt.Printf("  %d. %s (%.1f%% confidence) at x=%.2f, y=%.2f, w=%.2f, h=%.2f\n",
				i+1, bbox.Label, bbox.Confidence*100,
				bbox.Box.X, bbox.Box.Y, bbox.Box.Width, bbox.Box.Height)
		}
		fmt.Println()
	}

	// Overall confidence
	if result.Confidence > 0 {
		fmt.Printf("Overall Confidence: %.1f%%\n", result.Confidence*100)
	}

	// Raw response (if requested)
	if result.RawResponse != "" {
		fmt.Println("\n=== Raw API Response ===")
		fmt.Println(result.RawResponse)
	}

	// Timestamp
	fmt.Printf("\nProcessed at: %s\n", result.ProcessedAt.Format("2006-01-02 15:04:05"))

	return nil
}
