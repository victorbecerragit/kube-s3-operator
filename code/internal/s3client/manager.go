package s3client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	s3v1alpha1 "github.com/victorbecerragit/kube-s3-operator/code/api/v1alpha1"
)

// BucketManager abstracts AWS S3 operations required by the controller.
type BucketManager interface {
	CreateBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (string, error)
	DeleteBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error
	IsBucketReady(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (bool, error)
	IsBucketDeleted(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (bool, error)
	ApplyLifecycleConfiguration(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error
}

type defaultManager struct {
	clients map[string]*s3.Client
	mu      sync.RWMutex
}

// NewManager returns a new BucketManager instance.
func NewManager() BucketManager {
	return &defaultManager{
		clients: make(map[string]*s3.Client),
	}
}

// getClient returns a cached S3 client for the given region, or creates a new one.
func (m *defaultManager) getClient(ctx context.Context, region string) (*s3.Client, error) {
	if region == "" {
		region = "us-west-2"
	}

	m.mu.RLock()
	client, exists := m.clients[region]
	m.mu.RUnlock()

	if exists {
		return client, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := m.clients[region]; exists {
		return client, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
	}

	client = s3.NewFromConfig(cfg)
	m.clients[region] = client
	return client, nil
}

func (m *defaultManager) CreateBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (string, error) {
	client, err := m.getClient(ctx, s3bkt.Spec.Region)
	if err != nil {
		return "", err
	}

	input := &s3.CreateBucketInput{
		Bucket:                     aws.String(s3bkt.Spec.Name),
		ObjectLockEnabledForBucket: aws.Bool(s3bkt.Spec.Locked),
	}

	region := s3bkt.Spec.Region
	if region == "" {
		region = "us-west-2"
	}
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	var output *s3.CreateBucketOutput
	var createErr error
	maxAttempts := 5
	for attempt := 0; attempt < maxAttempts; attempt++ {
		output, createErr = client.CreateBucket(ctx, input)
		if createErr == nil {
			break
		}

		errStr := createErr.Error()
		if strings.Contains(errStr, "OperationAborted") || strings.Contains(errStr, "BucketAlreadyExists") {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		return "", fmt.Errorf("failed to create S3 bucket: %w", createErr)
	}

	if createErr != nil {
		return "", fmt.Errorf("S3 CreateBucket API call failed after retries: %w", createErr)
	}

	location := ""
	if output != nil && output.Location != nil {
		location = *output.Location
	}
	return location, nil
}

func (m *defaultManager) DeleteBucket(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	client, err := m.getClient(ctx, s3bkt.Spec.Region)
	if err != nil {
		return err
	}

	_, err = client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchBucket" || apiErr.ErrorCode() == "NotFound" {
				return nil
			}
			if apiErr.ErrorCode() == "BucketNotEmpty" {
				return fmt.Errorf("bucket is not empty, cannot delete: %w", err)
			}
		}
		return fmt.Errorf("S3 DeleteBucket API call failed: %w", err)
	}

	return nil
}

func (m *defaultManager) IsBucketReady(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (bool, error) {
	client, err := m.getClient(ctx, s3bkt.Spec.Region)
	if err != nil {
		return false, err
	}

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchBucket") {
			return false, nil // Not ready yet
		}
		// Some other error
		return false, err
	}

	return true, nil
}

func (m *defaultManager) IsBucketDeleted(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) (bool, error) {
	client, err := m.getClient(ctx, s3bkt.Spec.Region)
	if err != nil {
		return false, err
	}

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s3bkt.Spec.Name),
	})
	if err != nil {
		var apiErr smithy.APIError
		// If it's a 404/NotFound, then it implies it has been deleted
		if errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchBucket") {
			return true, nil
		}
		// If we get an error that is not "NotFound", we can't be sure it's deleted. But typically HeadBucket returns 404 when deleted.
		// Wait, a smithy API error might not be the exact type. AWS SDK v2 usually uses specific types for not found.
		// For simplicity we check strings or just return true if it's an error and contains 404 or NotFound.
		errStr := err.Error()
		if strings.Contains(errStr, "NotFound") || strings.Contains(errStr, "NoSuchBucket") || strings.Contains(errStr, "404") {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (m *defaultManager) ApplyLifecycleConfiguration(ctx context.Context, s3bkt *s3v1alpha1.S3Bucket) error {
	if s3bkt.Spec.Lifecycle == nil || len(s3bkt.Spec.Lifecycle.Rules) == 0 {
		return nil
	}

	client, err := m.getClient(ctx, s3bkt.Spec.Region)
	if err != nil {
		return err
	}

	// Remove the region auto-detection logic to keep it simple and performant.
	// If the user deployed it in region A but specifies region B in spec, the controller should follow the spec or log error.

	rules := make([]s3types.LifecycleRule, 0, len(s3bkt.Spec.Lifecycle.Rules))
	for i, rule := range s3bkt.Spec.Lifecycle.Rules {
		status := rule.Status
		if status == "" {
			status = "Enabled"
		}

		awsStatus := s3types.ExpirationStatusEnabled
		if status == "Disabled" {
			awsStatus = s3types.ExpirationStatusDisabled
		}

		awsRule := s3types.LifecycleRule{
			Status: awsStatus,
		}

		if rule.ID != "" {
			awsRule.ID = aws.String(rule.ID)
		} else {
			awsRule.ID = aws.String(fmt.Sprintf("%s-rule-%d", s3bkt.Name, i))
		}

		awsRule.Filter = &s3types.LifecycleRuleFilterMemberPrefix{
			Value: rule.Prefix,
		}

		if rule.Expiration != nil && rule.Expiration.Days > 0 {
			awsRule.Expiration = &s3types.LifecycleExpiration{
				Days: aws.Int32(rule.Expiration.Days),
			}
		}

		if len(rule.Transitions) > 0 {
			awsRule.Transitions = make([]s3types.Transition, 0, len(rule.Transitions))
			for _, t := range rule.Transitions {
				if t.Days <= 0 || t.StorageClass == "" {
					continue
				}
				awsRule.Transitions = append(awsRule.Transitions, s3types.Transition{
					Days:         aws.Int32(t.Days),
					StorageClass: s3types.TransitionStorageClass(t.StorageClass),
				})
			}
		}

		if rule.NoncurrentVersionExpiration != nil && rule.NoncurrentVersionExpiration.Days > 0 {
			awsRule.NoncurrentVersionExpiration = &s3types.NoncurrentVersionExpiration{
				NoncurrentDays: aws.Int32(rule.NoncurrentVersionExpiration.Days),
			}
		}

		if len(rule.NoncurrentVersionTransitions) > 0 {
			awsRule.NoncurrentVersionTransitions = make([]s3types.NoncurrentVersionTransition, 0, len(rule.NoncurrentVersionTransitions))
			for _, t := range rule.NoncurrentVersionTransitions {
				if t.Days <= 0 || t.StorageClass == "" {
					continue
				}
				awsRule.NoncurrentVersionTransitions = append(awsRule.NoncurrentVersionTransitions, s3types.NoncurrentVersionTransition{
					NoncurrentDays: aws.Int32(t.Days),
					StorageClass:   s3types.TransitionStorageClass(t.StorageClass),
				})
			}
		}

		rules = append(rules, awsRule)
	}

	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(s3bkt.Spec.Name),
		LifecycleConfiguration: &s3types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	_, err = client.PutBucketLifecycleConfiguration(ctx, input)
	if err != nil {
		return fmt.Errorf("PutBucketLifecycleConfiguration failed: %w", err)
	}
	return nil
}
