package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

func NewS3Client(ctx context.Context, cfg S3Config) (*s3.Client, error) {
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

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}

		o.UsePathStyle = cfg.PathStyle
	})

	return client, nil
}

func main() {
	cfg := S3Config{
		Endpoint:  "http://localhost:9000",
		Region:    "localhost",
		Bucket:    "shc-img",
		AccessKey: "admin",
		SecretKey: "adminadmin",
		PathStyle: true,
	}

	sss, err := NewS3Client(context.TODO(), cfg)

	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("README.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = sss.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.Bucket),
		Key:    aws.String("README.md"),
		Body:   f,
	})

	if err != nil {
		log.Fatal(err)
	}
}
