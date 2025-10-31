package domain

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	pb "grpc-sample-minimal/proto"
)

var (
	s3BucketName = os.Getenv("S3_BUCKET_NAME")
	awsRegion = os.Getenv("AWS_REGION")
	localstackEndpoint = os.Getenv("LOCALSTACK_ENDPOINT")
)

type s3StorageService struct {
	s3Client *s3.Client
}

func NewS3StorageService() (StorageService, error) {
	// Load AWS config, prioritizing environment variables for credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
				config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: localstackEndpoint, SigningRegion: awsRegion},
						nil
				})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Ensure the S3 bucket exists
	_, err = s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(s3BucketName),
	})
	if err != nil {
		// Check if the error is that the bucket already exists
		if !isBucketAlreadyOwnedByYouError(err) {
			log.Printf("Warning: Failed to create S3 bucket %s: %v", s3BucketName, err)
		}
	}

	return &s3StorageService{s3Client: s3Client}, nil
}

func isBucketAlreadyOwnedByYouError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*types.BucketAlreadyOwnedByYou);
	ok {
		return true
	}
	return false
}

func (s *s3StorageService) UploadFile(ctx context.Context, filename string, content io.Reader) (*pb.FileUploadStatus, error) {
	// Convert io.Reader to bytes.Reader to make it seekable for S3 PutObject
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read content for S3 upload: %w", err)
	}
	reader := bytes.NewReader(contentBytes)

	// Build storage path with namespace prefix (documents/, media/, or others/)
	storagePath := BuildStoragePath(filename)

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(storagePath),
		Body:   reader,
		ContentLength: aws.Int64(int64(len(contentBytes))),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return &pb.FileUploadStatus{
		Filename: filename,
		Success:  true,
		Message:  fmt.Sprintf("File %s uploaded to S3 at %s", filename, storagePath),
	}, nil
}

func (s *s3StorageService) DownloadFile(ctx context.Context, filename string) (io.Reader, error) {
	// Build storage path with namespace prefix
	storagePath := BuildStoragePath(filename)

	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	return buf, nil
}

func (s *s3StorageService) DeleteFile(ctx context.Context, filename string) error {
	// Build storage path with namespace prefix
	storagePath := BuildStoragePath(filename)

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

func (s *s3StorageService) ListFiles(ctx context.Context) ([]*pb.FileInfo, error) {
	var files []*pb.FileInfo
	
	paginator := s3.NewListObjectsV2Paginator(s.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3BucketName),
	})
	
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}
		
		for _, obj := range page.Contents {
			// Skip directory prefixes (keys ending with "/")
			key := aws.ToString(obj.Key)
			if strings.HasSuffix(key, "/") {
				continue
			}
			
			// Extract namespace and filename from key (e.g., "documents/file.pdf" -> namespace: "documents/", filename: "file.pdf")
			var namespace, filename string
			if strings.HasPrefix(key, "documents/") {
				namespace = "documents"
				filename = strings.TrimPrefix(key, "documents/")
			} else if strings.HasPrefix(key, "media/") {
				namespace = "media"
				filename = strings.TrimPrefix(key, "media/")
			} else if strings.HasPrefix(key, "others/") {
				namespace = "others"
				filename = strings.TrimPrefix(key, "others/")
			} else {
				// Legacy files without namespace
				namespace = "others"
				filename = key
			}
			
			files = append(files, &pb.FileInfo{
				Filename:  filename,
				Namespace: namespace,
				Size:      aws.ToInt64(obj.Size),
			})
		}
	}
	
	return files, nil
}
