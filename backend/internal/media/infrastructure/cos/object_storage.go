package cosstorage

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	cos "github.com/tencentyun/cos-go-sdk-v5"
)

type ObjectStorage struct {
	client              *cos.Client
	secretID, secretKey string
}

func NewObjectStorage(bucketURL, secretID, secretKey string) (ObjectStorage, error) {
	parsed, err := url.Parse(bucketURL)
	if err != nil {
		return ObjectStorage{}, err
	}
	base := &cos.BaseURL{BucketURL: parsed}
	client := cos.NewClient(base, &http.Client{Transport: &cos.AuthorizationTransport{SecretID: secretID, SecretKey: secretKey}})
	return ObjectStorage{client: client, secretID: secretID, secretKey: secretKey}, nil
}
func (s ObjectStorage) SignUpload(ctx context.Context, key, contentType string, expiresAt time.Time) (string, error) {
	headers := http.Header{"Content-Type": []string{contentType}}
	options := &cos.PresignedURLOptions{Query: &url.Values{}, Header: &headers}
	signed, err := s.client.Object.GetPresignedURL(ctx, http.MethodPut, key, s.secretID, s.secretKey, time.Until(expiresAt), options)
	if err != nil {
		return "", err
	}
	return signed.String(), nil
}
func (s ObjectStorage) VerifyUploaded(ctx context.Context, key, contentType string, maxSize int64) error {
	response, err := s.client.Object.Head(ctx, key, nil)
	if err != nil {
		if cos.IsNotFoundError(err) {
			return domain.ErrInvalidUpload
		}
		return fmt.Errorf("verify COS object: %w", err)
	}
	defer response.Body.Close()
	size, err := strconv.ParseInt(response.Header.Get("Content-Length"), 10, 64)
	actualType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if err != nil || size < 1 || size > maxSize || actualType != contentType {
		return domain.ErrInvalidUpload
	}
	return nil
}
func (s ObjectStorage) SignRead(ctx context.Context, key string, expiresAt time.Time) (string, error) {
	signed, err := s.client.Object.GetPresignedURL(ctx, http.MethodGet, key, s.secretID, s.secretKey, time.Until(expiresAt), nil)
	if err != nil {
		return "", err
	}
	return signed.String(), nil
}
