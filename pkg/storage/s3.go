package storage
import (
"context"
"io"
"os"
"time"

"github.com/aws/aws-sdk-go-v2/aws"
"github.com/aws/aws-sdk-go-v2/config"
"github.com/aws/aws-sdk-go-v2/credentials"
"github.com/aws/aws-sdk-go-v2/service/s3"
)
type S3Client struct { client *s3.Client; presignClient *s3.PresignClient; bucket string }
func NewS3Client(ctx context.Context, region, endpoint, accessKey, secretKey, bucket string) (*S3Client, error) {
cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
if err != nil { return nil, err }
client := s3.NewFromConfig(cfg, func(o *s3.Options) { if endpoint != "" { o.BaseEndpoint = aws.String(endpoint); o.UsePathStyle = true } })
return &S3Client{ client: client, presignClient: s3.NewPresignClient(client), bucket: bucket }, nil
}
func (s *S3Client) UploadFile(ctx context.Context, key, filePath, contentType string) error {
file, err := os.Open(filePath)
if err != nil { return err }
defer file.Close()
_, err = s.client.PutObject(ctx, &s3.PutObjectInput{ Bucket: aws.String(s.bucket), Key: aws.String(key), Body: file, ContentType: aws.String(contentType) })
return err
}
func (s *S3Client) DownloadFile(ctx context.Context, key, destPath string) error {
out, err := s.client.GetObject(ctx, &s3.GetObjectInput{ Bucket: aws.String(s.bucket), Key: aws.String(key) })
if err != nil { return err }
defer out.Body.Close()
file, err := os.Create(destPath)
if err != nil { return err }
defer file.Close()
_, err = io.Copy(file, out.Body)
return err
}
func (s *S3Client) GeneratePresignedURL(ctx context.Context, key string, lifetime time.Duration) (string, error) {
req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{ Bucket: aws.String(s.bucket), Key: aws.String(key) }, s3.WithPresignExpires(lifetime))
if err != nil { return "", err }
return req.URL, nil
}
// GeneratePresignedPutURL presigns an upload (PUT) so admins can push a raw IPA straight to S3.
func (s *S3Client) GeneratePresignedPutURL(ctx context.Context, key string, lifetime time.Duration) (string, error) {
req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{ Bucket: aws.String(s.bucket), Key: aws.String(key) }, s3.WithPresignExpires(lifetime))
if err != nil { return "", err }
return req.URL, nil
}
// CopyObject server-side copies one key to another within the bucket (used by the signing worker
// to promote a raw IPA into the signed slot).
func (s *S3Client) CopyObject(ctx context.Context, srcKey, dstKey string) error {
_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{ Bucket: aws.String(s.bucket), CopySource: aws.String(s.bucket + "/" + srcKey), Key: aws.String(dstKey) })
return err
}
// ObjectExists reports whether a key is present in the bucket.
func (s *S3Client) ObjectExists(ctx context.Context, key string) bool {
_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{ Bucket: aws.String(s.bucket), Key: aws.String(key) })
return err == nil
}
