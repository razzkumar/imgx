package detection

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

// AWSProvider implements the Provider interface for AWS Rekognition
type AWSProvider struct {
	client *rekognition.Client
	cfg    aws.Config
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
		return nil, fmt.Errorf("%w: failed to load AWS config. Ensure you have AWS credentials configured via environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION) or AWS CLI (aws configure)", ErrProviderNotConfigured)
	}

	// Verify credentials are available by retrieving them
	// This ensures we fail fast if credentials are not properly configured
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to retrieve AWS credentials. Ensure you have AWS credentials configured via environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY) or AWS CLI (aws configure)", ErrProviderNotConfigured)
	}

	// Check if credentials are empty
	if creds.AccessKeyID == "" {
		return nil, fmt.Errorf("%w: AWS credentials not found. Set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables or configure AWS CLI", ErrProviderNotConfigured)
	}

	// Create Rekognition client
	client := rekognition.NewFromConfig(cfg)

	return &AWSProvider{
		client: client,
		cfg:    cfg,
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

	// Perform detection based on requested features
	for _, feature := range opts.Features {
		switch feature {
		case FeatureLabels, FeatureObjects:
			if err := a.detectLabels(ctx, imgBytes, result, opts); err != nil {
				return nil, err
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

// detectLabels performs label detection
func (a *AWSProvider) detectLabels(ctx context.Context, imgBytes []byte, result *DetectionResult, opts *DetectOptions) error {
	input := &rekognition.DetectLabelsInput{
		Image: &types.Image{
			Bytes: imgBytes,
		},
		MaxLabels:     aws.Int32(int32(opts.MaxResults)),
		MinConfidence: aws.Float32(opts.MinConfidence * 100), // AWS uses 0-100 scale
	}

	output, err := a.client.DetectLabels(ctx, input)
	if err != nil {
		return NewDetectionError("aws", "label detection failed", err)
	}

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
		return NewDetectionError("aws", "text detection failed", err)
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
		return NewDetectionError("aws", "face detection failed", err)
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
		return NewDetectionError("aws", "moderation detection failed", err)
	}

	// Store moderation labels in properties
	for _, label := range output.ModerationLabels {
		if label.Name != nil && label.Confidence != nil {
			key := fmt.Sprintf("moderation_%s", *label.Name)
			value := fmt.Sprintf("%.1f%%", *label.Confidence)
			result.Properties[key] = value
		}
	}

	return nil
}

// Close closes the AWS client (no-op for AWS SDK v2)
func (a *AWSProvider) Close() error {
	// AWS SDK v2 doesn't require explicit cleanup
	return nil
}
