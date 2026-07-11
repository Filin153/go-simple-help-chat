package s3

import (
	"context"
)

func NewMiniO(ctx context.Context, url, login, password, bucket string) (*Client, error) {
	client, err := newS3Client(ctx, S3Config{
		Endpoint:  url,
		Region:    "local",
		Bucket:    bucket,
		AccessKey: login,
		SecretKey: password,
		PathStyle: true,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		bucket: bucket,
	}, nil
}
