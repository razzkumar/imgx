# Detection Test Data

This directory contains test fixtures and sample data used by the detection package tests.

## Structure

```
testdata/
├── README.md           # This file
└── responses/          # Sample API responses
    ├── gemini_labels.json   # Google Gemini response fixture
    ├── aws_labels.json      # AWS Rekognition response fixture
    └── openai_labels.json   # OpenAI Vision response fixture
```

## Response Fixtures

The `responses/` directory contains JSON fixtures representing typical API responses from each provider. These are used in unit tests to mock provider behavior without making actual API calls.

### gemini_labels.json
Sample Google Gemini API response with labels, categories, and descriptions.

### aws_labels.json
Sample AWS Rekognition response with labels and confidence scores (0-100 scale).

### openai_labels.json
Sample OpenAI Vision API response with labels and natural language descriptions.

## Usage in Tests

Load fixtures using the test helper:

```go
data := LoadFixtureResponse(t, "gemini_labels.json")
```

## Adding New Fixtures

When adding new test fixtures:

1. Place JSON files in the appropriate subdirectory
2. Use realistic data structures matching actual API responses
3. Include various edge cases (empty results, max results, etc.)
4. Document the fixture purpose in comments or this README
