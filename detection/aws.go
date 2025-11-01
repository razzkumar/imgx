package detection

import (
	"context"
	"fmt"
	"image"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

// AWSProvider implements the Provider interface for AWS Rekognition
type AWSProvider struct {
	client     *rekognition.Client
	cfg        aws.Config
	credSource string // Source of credentials for debugging
}

// NewAWSProvider creates a new AWS Rekognition provider instance
// It uses the default AWS credential chain which checks in order:
// 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)
// 2. AWS credentials file (~/.aws/credentials)
// 3. AWS config file (~/.aws/config)
// 4. IAM roles for Amazon EC2, ECS, or Lambda
func NewAWSProvider() (*AWSProvider, error) {
	ctx := context.Background()

	// Load AWS configuration using default config loader
	// This automatically handles:
	// - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, etc.)
	// - Shared config/credentials files (~/.aws/config, ~/.aws/credentials)
	// - IAM roles (EC2, ECS, Lambda, etc.)
	// - SSO configurations
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to load AWS config. Ensure you have AWS credentials configured via environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION) or AWS CLI (aws configure): %v", ErrProviderNotConfigured, err)
	}

	// Verify credentials are available by retrieving them
	// This ensures we fail fast if credentials are not properly configured
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve AWS credentials. Ensure you have AWS credentials configured via environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY) or AWS CLI (aws configure): %v", ErrProviderNotConfigured, err)
	}

	// Check if credentials are empty
	if creds.AccessKeyID == "" {
		return nil, fmt.Errorf("%w: AWS credentials not found. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables or configure AWS CLI", ErrProviderNotConfigured)
	}

	// Validate region is set
	if cfg.Region == "" {
		return nil, fmt.Errorf("%w: AWS region not configured. Set AWS_REGION environment variable or configure it via AWS CLI (aws configure)", ErrProviderNotConfigured)
	}

	// Create Rekognition client
	client := rekognition.NewFromConfig(cfg)

	// Store credential source info for debugging
	credSource := creds.Source
	if credSource == "" {
		credSource = "unknown"
	}

	return &AWSProvider{
		client:     client,
		cfg:        cfg,
		credSource: credSource,
	}, nil
}

// Name returns the provider name
func (a *AWSProvider) Name() string {
	return "aws"
}

// IsConfigured checks if the provider is properly configured
func (a *AWSProvider) IsConfigured() bool {
	// If the provider was successfully initialized, it's configured
	return a.client != nil
}

// Detect performs object detection using AWS Rekognition
func (a *AWSProvider) Detect(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
	if opts == nil {
		opts = DefaultDetectOptions()
	}

	startTime := time.Now()

	// Convert image to JPEG bytes
	imgBytes, err := imageToJPEGBytes(img)
	if err != nil {
		return nil, NewDetectionError("aws", "failed to encode image", err)
	}

	result := &DetectionResult{
		Provider:    "aws",
		Labels:      []Label{},
		Text:        []TextBlock{},
		Faces:       []Face{},
		Properties:  make(map[string]string),
		ProcessedAt: startTime,
	}

	// Check if image properties are requested
	enableImageProperties := false
	hasLabelsFeature := false
	for _, feature := range opts.Features {
		if feature == FeatureProperties {
			enableImageProperties = true
		}
		if feature == FeatureLabels || feature == FeatureObjects {
			hasLabelsFeature = true
		}
	}

	// If only properties are requested without labels, we still need to call detectLabels
	// but with only IMAGE_PROPERTIES enabled (not GENERAL_LABELS)
	labelsProcessed := false

	// Perform detection based on requested features
	for _, feature := range opts.Features {
		switch feature {
		case FeatureLabels, FeatureObjects:
			if !labelsProcessed {
				if err := a.detectLabels(ctx, imgBytes, result, opts, enableImageProperties); err != nil {
					return nil, err
				}
				labelsProcessed = true
			}

		case FeatureProperties:
			// If properties requested without labels, still call detectLabels but with IMAGE_PROPERTIES only
			if !labelsProcessed && !hasLabelsFeature {
				if err := a.detectLabelsImagePropertiesOnly(ctx, imgBytes, result); err != nil {
					return nil, err
				}
				labelsProcessed = true
			}

		case FeatureText:
			if err := a.detectText(ctx, imgBytes, result); err != nil {
				return nil, err
			}

		case FeatureFaces:
			if err := a.detectFaces(ctx, imgBytes, result); err != nil {
				return nil, err
			}

		case FeatureSafeSearch:
			if err := a.detectModeration(ctx, imgBytes, result); err != nil {
				return nil, err
			}
		}
	}

	// Calculate overall confidence
	if len(result.Labels) > 0 {
		var totalConf float32
		for _, label := range result.Labels {
			totalConf += label.Confidence
		}
		result.Confidence = totalConf / float32(len(result.Labels))
	}

	return result, nil
}

// detectLabels performs label detection and optionally image properties detection
// enableImageProperties: when true, also detects image properties (colors, quality, etc.)
func (a *AWSProvider) detectLabels(ctx context.Context, imgBytes []byte, result *DetectionResult, opts *DetectOptions, enableImageProperties bool) error {
	input := &rekognition.DetectLabelsInput{
		Image: &types.Image{
			Bytes: imgBytes,
		},
		MaxLabels:     aws.Int32(int32(opts.MaxResults)),
		MinConfidence: aws.Float32(opts.MinConfidence * 100), // AWS uses 0-100 scale
	}

	// Configure Features to enable GENERAL_LABELS and/or IMAGE_PROPERTIES
	// According to AWS docs, use Features field to enable image properties
	if enableImageProperties {
		// Enable both GENERAL_LABELS and IMAGE_PROPERTIES
		input.Features = []types.DetectLabelsFeatureName{
			types.DetectLabelsFeatureNameGeneralLabels,
			types.DetectLabelsFeatureNameImageProperties,
		}
		// Configure image properties settings
		input.Settings = &types.DetectLabelsSettings{
			ImageProperties: &types.DetectLabelsImagePropertiesSettings{
				MaxDominantColors: 10, // Get up to 10 dominant colors
			},
		}
	} else {
		// Only GENERAL_LABELS (default behavior when Features is not set)
		// We can omit Features field and it will default to GENERAL_LABELS
	}

	output, err := a.client.DetectLabels(ctx, input)
	if err != nil {
		return a.enhanceAWSError("label detection failed", err)
	}

	// Parse labels
	for _, label := range output.Labels {
		if label.Name != nil && label.Confidence != nil {
			l := Label{
				Name:       *label.Name,
				Confidence: *label.Confidence / 100.0, // Convert to 0-1 scale
			}

			// Add parent categories
			if len(label.Parents) > 0 {
				categories := make([]string, 0, len(label.Parents))
				for _, parent := range label.Parents {
					if parent.Name != nil {
						categories = append(categories, *parent.Name)
					}
				}
				l.Categories = categories
			}

			result.Labels = append(result.Labels, l)
		}
	}

	// Parse image properties if available
	if output.ImageProperties != nil {
		a.parseImageProperties(output.ImageProperties, result)
	}

	return nil
}

// detectText performs text detection (OCR)
func (a *AWSProvider) detectText(ctx context.Context, imgBytes []byte, result *DetectionResult) error {
	input := &rekognition.DetectTextInput{
		Image: &types.Image{
			Bytes: imgBytes,
		},
	}

	output, err := a.client.DetectText(ctx, input)
	if err != nil {
		return a.enhanceAWSError("text detection failed", err)
	}

	for _, detection := range output.TextDetections {
		if detection.DetectedText != nil && detection.Type == types.TextTypesLine {
			textBlock := TextBlock{
				Text: *detection.DetectedText,
				Type: string(detection.Type),
			}

			if detection.Confidence != nil {
				textBlock.Confidence = *detection.Confidence / 100.0
			}

			// Add bounding box if available
			if detection.Geometry != nil && detection.Geometry.BoundingBox != nil {
				bb := detection.Geometry.BoundingBox
				textBlock.BoundingBox = &Box{
					X:      aws.ToFloat32(bb.Left),
					Y:      aws.ToFloat32(bb.Top),
					Width:  aws.ToFloat32(bb.Width),
					Height: aws.ToFloat32(bb.Height),
				}
			}

			result.Text = append(result.Text, textBlock)
		}
	}

	return nil
}

// detectFaces performs face detection
func (a *AWSProvider) detectFaces(ctx context.Context, imgBytes []byte, result *DetectionResult) error {
	input := &rekognition.DetectFacesInput{
		Image: &types.Image{
			Bytes: imgBytes,
		},
		Attributes: []types.Attribute{types.AttributeAll},
	}

	output, err := a.client.DetectFaces(ctx, input)
	if err != nil {
		return a.enhanceAWSError("face detection failed", err)
	}

	for _, faceDetail := range output.FaceDetails {
		face := Face{}

		if faceDetail.Confidence != nil {
			face.Confidence = *faceDetail.Confidence / 100.0
		}

		// Emotions
		if len(faceDetail.Emotions) > 0 {
			// Get the emotion with highest confidence
			maxEmotion := faceDetail.Emotions[0]
			for _, emotion := range faceDetail.Emotions {
				if aws.ToFloat32(emotion.Confidence) > aws.ToFloat32(maxEmotion.Confidence) {
					maxEmotion = emotion
				}
			}

			emotionType := string(maxEmotion.Type)
			switch maxEmotion.Type {
			case types.EmotionNameHappy:
				face.JoyLikelihood = fmt.Sprintf("%.1f%%", aws.ToFloat32(maxEmotion.Confidence))
			case types.EmotionNameSad:
				face.SorrowLikelihood = fmt.Sprintf("%.1f%%", aws.ToFloat32(maxEmotion.Confidence))
			case types.EmotionNameAngry:
				face.AngerLikelihood = fmt.Sprintf("%.1f%%", aws.ToFloat32(maxEmotion.Confidence))
			case types.EmotionNameSurprised:
				face.SurpriseLikelihood = fmt.Sprintf("%.1f%%", aws.ToFloat32(maxEmotion.Confidence))
			}

			// Store primary emotion in properties
			result.Properties["primary_emotion"] = emotionType
		}

		// Gender
		if faceDetail.Gender != nil && faceDetail.Gender.Value != "" {
			face.Gender = string(faceDetail.Gender.Value)
		}

		// Age range
		if faceDetail.AgeRange != nil {
			if faceDetail.AgeRange.Low != nil && faceDetail.AgeRange.High != nil {
				face.AgeRange = fmt.Sprintf("%d-%d", *faceDetail.AgeRange.Low, *faceDetail.AgeRange.High)
			}
		}

		// Bounding box
		if faceDetail.BoundingBox != nil {
			bb := faceDetail.BoundingBox
			face.BoundingBox = &Box{
				X:      aws.ToFloat32(bb.Left),
				Y:      aws.ToFloat32(bb.Top),
				Width:  aws.ToFloat32(bb.Width),
				Height: aws.ToFloat32(bb.Height),
			}
		}

		// Landmarks
		if len(faceDetail.Landmarks) > 0 {
			landmarks := make([]Landmark, 0, len(faceDetail.Landmarks))
			for _, lm := range faceDetail.Landmarks {
				if lm.Type != "" {
					landmarks = append(landmarks, Landmark{
						Type: string(lm.Type),
						X:    aws.ToFloat32(lm.X),
						Y:    aws.ToFloat32(lm.Y),
					})
				}
			}
			face.Landmarks = landmarks
		}

		result.Faces = append(result.Faces, face)
	}

	return nil
}

// detectModeration performs content moderation (safe search)
func (a *AWSProvider) detectModeration(ctx context.Context, imgBytes []byte, result *DetectionResult) error {
	input := &rekognition.DetectModerationLabelsInput{
		Image: &types.Image{
			Bytes: imgBytes,
		},
		MinConfidence: aws.Float32(50.0),
	}

	output, err := a.client.DetectModerationLabels(ctx, input)
	if err != nil {
		return a.enhanceAWSError("moderation detection failed", err)
	}

	// Store moderation labels in properties
	for _, label := range output.ModerationLabels {
		if label.Name != nil && label.Confidence != nil {
			key := fmt.Sprintf("moderation_%s", *label.Name)
			value := fmt.Sprintf("%.1f%%", *label.Confidence)
			result.Properties[key] = value

			conf := aws.ToFloat32(label.Confidence)
			result.Moderation = append(result.Moderation, ModerationLabel{
				Name:       *label.Name,
				Parent:     aws.ToString(label.ParentName),
				Confidence: conf / 100.0,
				Severity:   fmt.Sprintf("%.1f%%", conf),
			})
		}
	}

	if len(result.Moderation) > 0 && result.SafeSearch == nil {
		result.SafeSearch = &SafeSearchSummary{
			Labels: result.Moderation,
		}
	}

	return nil
}

// detectLabelsImagePropertiesOnly calls DetectLabels with only IMAGE_PROPERTIES enabled
// Used when user requests only properties without general labels
// This charges only for Image Properties API, not for general label detection
func (a *AWSProvider) detectLabelsImagePropertiesOnly(ctx context.Context, imgBytes []byte, result *DetectionResult) error {
	input := &rekognition.DetectLabelsInput{
		Image: &types.Image{
			Bytes: imgBytes,
		},
		// Only enable IMAGE_PROPERTIES, not GENERAL_LABELS
		// This way we're only charged for Image Properties API
		Features: []types.DetectLabelsFeatureName{
			types.DetectLabelsFeatureNameImageProperties,
		},
		Settings: &types.DetectLabelsSettings{
			ImageProperties: &types.DetectLabelsImagePropertiesSettings{
				MaxDominantColors: 10, // Get up to 10 dominant colors
			},
		},
	}

	output, err := a.client.DetectLabels(ctx, input)
	if err != nil {
		return a.enhanceAWSError("image properties detection failed", err)
	}

	// Parse image properties if available
	if output.ImageProperties != nil {
		a.parseImageProperties(output.ImageProperties, result)
	}

	return nil
}

// parseImageProperties parses AWS Rekognition image properties into our DetectionResult
func (a *AWSProvider) parseImageProperties(props *types.DetectLabelsImageProperties, result *DetectionResult) {
	ensureQuality := func() *ImageQuality {
		if result.ImageQuality == nil {
			result.ImageQuality = &ImageQuality{}
		}
		return result.ImageQuality
	}

	// Quality information
	if props.Quality != nil {
		quality := ensureQuality()
		if props.Quality.Brightness != nil {
			value := float32(*props.Quality.Brightness)
			result.Properties["brightness"] = fmt.Sprintf("%.2f", value)
			quality.Brightness = value
		}
		if props.Quality.Sharpness != nil {
			value := float32(*props.Quality.Sharpness)
			result.Properties["sharpness"] = fmt.Sprintf("%.2f", value)
			quality.Sharpness = value
		}
		if props.Quality.Contrast != nil {
			value := float32(*props.Quality.Contrast)
			result.Properties["contrast"] = fmt.Sprintf("%.2f", value)
			quality.Contrast = value
		}
	}

	// Dominant colors
	if len(props.DominantColors) > 0 {
		colors := make([]string, 0, len(props.DominantColors))
		for i, color := range props.DominantColors {
			if i >= 10 { // Limit to top 10 colors
				break
			}

			var (
				name       string
				percentage float32
				rgbValue   string
				hexValue   string
			)

			if color.SimplifiedColor != nil {
				name = *color.SimplifiedColor
			}
			if color.PixelPercent != nil {
				percentage = float32(*color.PixelPercent)
			}
			if color.Red != nil && color.Green != nil && color.Blue != nil {
				rgbValue = fmt.Sprintf("rgb(%d,%d,%d)", *color.Red, *color.Green, *color.Blue)
			}
			if color.CSSColor != nil {
				hexValue = *color.CSSColor
			}

			if name != "" {
				colorInfo := fmt.Sprintf("%s(%.1f%%)", name, percentage)
				colors = append(colors, colorInfo)
			}

			// Populate structured colors slice
			if name != "" || hexValue != "" || rgbValue != "" {
				result.Colors = append(result.Colors, ColorInfo{
					Name:       name,
					Hex:        hexValue,
					RGB:        rgbValue,
					Percentage: percentage,
				})
			}

			// Maintain legacy property keys
			if rgbValue != "" {
				rgbKey := fmt.Sprintf("color_%d_rgb", i+1)
				result.Properties[rgbKey] = rgbValue
			}
			if hexValue != "" {
				hexKey := fmt.Sprintf("color_%d_hex", i+1)
				result.Properties[hexKey] = hexValue
			}
		}
		if len(colors) > 0 {
			result.Properties["dominant_colors"] = strings.Join(colors, ", ")
			result.Properties["dominant_colors_count"] = fmt.Sprintf("%d", len(colors))
		}
	}

	// Foreground color block
	if props.Foreground != nil {
		quality := ensureQuality()
		if props.Foreground.Quality != nil {
			if props.Foreground.Quality.Brightness != nil {
				value := float32(*props.Foreground.Quality.Brightness)
				result.Properties["foreground_brightness"] = fmt.Sprintf("%.2f", value)
				quality.ForegroundBrightness = value
			}
			if props.Foreground.Quality.Sharpness != nil {
				value := float32(*props.Foreground.Quality.Sharpness)
				result.Properties["foreground_sharpness"] = fmt.Sprintf("%.2f", value)
				quality.ForegroundSharpness = value
			}
		}
		if len(props.Foreground.DominantColors) > 0 && props.Foreground.DominantColors[0].SimplifiedColor != nil {
			colorName := *props.Foreground.DominantColors[0].SimplifiedColor
			result.Properties["foreground_color"] = colorName
			quality.ForegroundColor = colorName
			result.Colors = append(result.Colors, ColorInfo{Name: colorName})
		}
	}

	// Background color block
	if props.Background != nil {
		quality := ensureQuality()
		if props.Background.Quality != nil {
			if props.Background.Quality.Brightness != nil {
				value := float32(*props.Background.Quality.Brightness)
				result.Properties["background_brightness"] = fmt.Sprintf("%.2f", value)
				quality.BackgroundBrightness = value
			}
			if props.Background.Quality.Sharpness != nil {
				value := float32(*props.Background.Quality.Sharpness)
				result.Properties["background_sharpness"] = fmt.Sprintf("%.2f", value)
				quality.BackgroundSharpness = value
			}
		}
		if len(props.Background.DominantColors) > 0 && props.Background.DominantColors[0].SimplifiedColor != nil {
			colorName := *props.Background.DominantColors[0].SimplifiedColor
			result.Properties["background_color"] = colorName
			quality.BackgroundColor = colorName
			result.Colors = append(result.Colors, ColorInfo{Name: colorName})
		}
	}

	// Ensure colors slice is deduplicated for readability
	if len(result.Colors) > 1 {
		unique := make([]ColorInfo, 0, len(result.Colors))
		seen := make(map[string]bool)
		for _, color := range result.Colors {
			key := fmt.Sprintf("%s|%s|%s|%0.2f", color.Name, color.Hex, color.RGB, color.Percentage)
			if !seen[key] {
				seen[key] = true
				unique = append(unique, color)
			}
		}
		result.Colors = unique
	}
}

// enhanceAWSError provides better error messages for common AWS errors
func (a *AWSProvider) enhanceAWSError(operation string, err error) error {
	errMsg := err.Error()

	// Check for common AWS authentication errors
	if strings.Contains(errMsg, "UnrecognizedClientException") ||
		strings.Contains(errMsg, "security token") ||
		strings.Contains(errMsg, "invalid") {
		return NewDetectionError("aws", fmt.Sprintf("%s: Invalid AWS credentials. "+
			"The credentials (from %s) are invalid, expired, or malformed. "+
			"Please verify your AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are correct. "+
			"If using temporary credentials, they may have expired. "+
			"Region: %s", operation, a.credSource, a.cfg.Region), err)
	}

	if strings.Contains(errMsg, "InvalidSignatureException") {
		return NewDetectionError("aws", fmt.Sprintf("%s: Invalid AWS signature. "+
			"Your AWS_SECRET_ACCESS_KEY may be incorrect. "+
			"Credential source: %s, Region: %s", operation, a.credSource, a.cfg.Region), err)
	}

	if strings.Contains(errMsg, "AccessDeniedException") ||
		strings.Contains(errMsg, "not authorized") {
		return NewDetectionError("aws", fmt.Sprintf("%s: Access denied. "+
			"Your AWS credentials don't have permission to use AWS Rekognition. "+
			"Ensure your IAM user/role has 'rekognition:DetectLabels', 'rekognition:DetectText', "+
			"and 'rekognition:DetectFaces' permissions. "+
			"Credential source: %s, Region: %s", operation, a.credSource, a.cfg.Region), err)
	}

	if strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "NoSuchBucket") {
		return NewDetectionError("aws", fmt.Sprintf("%s: Network or endpoint error. "+
			"Check your AWS_REGION (%s) is correct and supports Rekognition. "+
			"Available regions: us-east-1, us-west-2, eu-west-1, ap-southeast-1, etc.",
			operation, a.cfg.Region), err)
	}

	// Default error with region and credential info
	return NewDetectionError("aws", fmt.Sprintf("%s (Region: %s, Credentials from: %s)",
		operation, a.cfg.Region, a.credSource), err)
}

// Close closes the AWS client (no-op for AWS SDK v2)
func (a *AWSProvider) Close() error {
	// AWS SDK v2 doesn't require explicit cleanup
	return nil
}
