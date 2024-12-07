package article

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

// 这是 aws-sdk-go-v2,属实没搞明白
func TestS3(t *testing.T) {
	// 腾讯云对标 S3 的产品 COS
	cosId, ok := os.LookupEnv("COS_APP_ID")
	if !ok {
		t.Fatal("没有找到环境变量 COS_APP_ID")
	}
	cosKey, ok := os.LookupEnv("COS_APP_SECRET")
	if !ok {
		t.Fatal("没有找到环境变量 COS_APP_SECRET")
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 加载 AWS 配置，连接到 COS
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("ap-nanjing"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cosId, cosKey, "")),
		config.WithBaseEndpoint("https://cos.ap-nanjing.myqcloud.com"))
	if err != nil {
		t.Fatalf("无法加载 AWS 配置: %v", err)
	}

	// 创建 S3 客户端
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// 使用 BaseEndpoint 配置 COS 的 Endpoint
		o.BaseEndpoint = aws.String("https://webook-1312712974.cos.ap-nanjing.myqcloud.com")
		// 强制使用 /bucket/key 的形态
		o.UsePathStyle = true
	})

	// 上传文件到 COS
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("webook-1312712974"),
		Key:         aws.String("126"),
		Body:        strings.NewReader("测试内容"),
		ContentType: aws.String("text/plain;charset=utf-8"),
	})
	assert.NoError(t, err)

	// 下载文件
	res, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String("webook-1312712974"),
		Key:    aws.String("126"),
	})
	assert.NoError(t, err)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Log(err)
		}
	}(res.Body)

	// 读取文件内容
	data, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	// 打印文件内容
	t.Log(string(data))
}
