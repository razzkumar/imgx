# Detection API Example

This example demonstrates the imgx detection API with various AI vision providers.

## Features Demonstrated

1. **Basic Label Detection** - Simple object detection with default settings
2. **AWS Multi-Feature** - Labels, text (OCR), and face detection together
3. **AWS Image Properties** - Image quality metrics and dominant colors
4. **Custom Prompts** - Natural language queries with Gemini/OpenAI
5. **Provider Comparison** - Test all providers and compare results
6. **JSON Output** - Structured data output format

## Usage

```bash
# Run the example with an image
go run main.go <image-path>

# Examples
go run main.go ../../testdata/flower.jpg
go run main.go ~/Pictures/photo.jpg
```

## Setup

The example works with three AI vision providers. Configure at least one to see results:

### Google Gemini (Recommended for quick testing)
```bash
export GEMINI_API_KEY="your-api-key"
```
Get your API key from: https://aistudio.google.com/

### AWS Rekognition
```bash
# Option 1: AWS CLI configuration
aws configure

# Option 2: Environment variables
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"
```

### OpenAI Vision
```bash
export OPENAI_API_KEY="sk-..."
```
Get your API key from: https://platform.openai.com/

## What It Does

The example loads the specified image and runs 6 different detection scenarios:

1. Detects objects/labels using Gemini with default settings
2. Uses AWS to detect labels, text, and faces with higher confidence threshold
3. Analyzes image quality (brightness, sharpness, contrast) and colors using AWS
4. Sends a custom prompt to Gemini for detailed description
5. Compares all three providers side-by-side
6. Demonstrates JSON output format

Each example gracefully handles missing credentials and provides setup instructions.

## Expected Output

```
Loaded: testdata/flower.jpg (350x525)
================================================================================

Example 1: Basic Label Detection with Gemini
--------------------------------------------------------------------------------
✓ Detection successful
Provider: gemini
Labels found: 10

Top 5 labels:
  1. Flower (95.2% confidence)
  2. Plant (92.8% confidence)
  3. Petal (89.3% confidence)
  ...
```

## Documentation

For complete API documentation, see:
- [Detection API Documentation](../../docs/DETECTION.md)
- [CLI Documentation](../../docs/CLI.md)
