package infrastructure

import (
	"context"
	"net/http"
	"net/url"
	"time"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

type COSStorage struct {
	client              *cos.Client
	secretID, secretKey string
}

func NewCOSStorage(bucketURL, secretID, secretKey string) (COSStorage, error) {
	parsed, err := url.Parse(bucketURL)
	if err != nil {
		return COSStorage{}, err
	}
	base := &cos.BaseURL{BucketURL: parsed}
	client := cos.NewClient(base, &http.Client{Transport: &cos.AuthorizationTransport{SecretID: secretID, SecretKey: secretKey}})
	return COSStorage{client: client, secretID: secretID, secretKey: secretKey}, nil
}
func (s COSStorage) SignUpload(ctx context.Context, key, contentType string, expiresAt time.Time) (string, error) {
	options := &cos.PresignedURLOptions{Query: &url.Values{}}
	signed, err := s.client.Object.GetPresignedURL(ctx, http.MethodPut, key, s.secretID, s.secretKey, time.Until(expiresAt), options)
	if err != nil {
		return "", err
	}
	return signed.String(), nil
}
func (s COSStorage) SignRead(ctx context.Context, key string, expiresAt time.Time) (string, error) {
	signed, err := s.client.Object.GetPresignedURL(ctx, http.MethodGet, key, s.secretID, s.secretKey, time.Until(expiresAt), nil)
	if err != nil {
		return "", err
	}
	return signed.String(), nil
}
