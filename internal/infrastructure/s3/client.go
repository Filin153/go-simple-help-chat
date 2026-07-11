package s3

import (
	"bytes"
	"context"
	"errors"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	PathStyle bool
}

type Client struct {
	client *awss3.Client
	bucket string
}

func newS3Client(ctx context.Context, cfg S3Config) (*awss3.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	client := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}

		o.UsePathStyle = cfg.PathStyle
	})

	return client, nil
}

func (c *Client) Save(ctx context.Context, folder string, file []byte) (string, error) {
	if c.bucket == "" {
		return "", errors.New("s3 bucket is empty")
	}

	key := path.Join(folder, uuid.NewString())
	_, err := c.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(file),
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

func (c *Client) Delete(ctx context.Context, objectPath string) error {
	if c.bucket == "" {
		return errors.New("s3 bucket is empty")
	}

	_, err := c.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectPath),
	})
	return err
}
