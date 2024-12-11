package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
	"github.com/joho/godotenv" // 用于加载 .env 文件
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// UploadAliyunOss 将文件上传到阿里云 OSS，并返回文件路径
func UploadAliyunOss(file *multipart.FileHeader) (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	bucketName := os.Getenv("OSS_BUCKET_NAME")
	endpoint := os.Getenv("OSS_ENDPOINT")
	accessKeyID := os.Getenv("OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")

	if bucketName == "" || endpoint == "" || accessKeyID == "" || accessKeySecret == "" {
		log.Fatal("Please ensure OSS_BUCKET_NAME, OSS_ENDPOINT, OSS_ACCESS_KEY_ID, and OSS_ACCESS_KEY_SECRET are set in the .env file.")
	}

	// 创建 OSS 客户端
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return "", fmt.Errorf("创建 OSS 客户端失败: %v", err)
	}

	// 获取 bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("获取 OSS Bucket 失败: %v", err)
	}

	// 打开文件流
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	// 获取文件扩展名并生成唯一文件名
	ext := strings.ToLower(filepath.Ext(file.Filename))             // 提取文件扩展名（如 .jpg）
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext) // 生成唯一文件名

	// 文件上传路径
	path := "avatar/" + uniqueFilename

	// 上传文件
	err = bucket.PutObject(path, src)
	if err != nil {
		return "", fmt.Errorf("文件上传到 OSS 失败: %v", err)
	}

	// 返回文件路径
	fileURL := fmt.Sprintf("https://%s.%s/%s", bucketName, endpoint, path)
	return fileURL, nil
}

// UploadFileToOSS 将文件上传到阿里云 OSS，并返回文件路径
func UploadFileToOSS(file *multipart.FileHeader) (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	bucketName := os.Getenv("OSS_BUCKET_NAME")
	endpoint := os.Getenv("OSS_ENDPOINT")
	accessKeyID := os.Getenv("OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")

	if bucketName == "" || endpoint == "" || accessKeyID == "" || accessKeySecret == "" {
		log.Fatal("Please ensure OSS_BUCKET_NAME, OSS_ENDPOINT, OSS_ACCESS_KEY_ID, and OSS_ACCESS_KEY_SECRET are set in the .env file.")
	}

	// 创建 OSS 客户端
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return "", fmt.Errorf("创建 OSS 客户端失败: %v", err)
	}

	// 获取 bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("获取 OSS Bucket 失败: %v", err)
	}

	// 打开文件流
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	// 获取文件扩展名并生成唯一文件名
	ext := strings.ToLower(filepath.Ext(file.Filename))             // 提取文件扩展名（如 .jpg）
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext) // 生成唯一文件名

	// 文件上传路径
	path := "note/" + uniqueFilename

	// 上传文件
	err = bucket.PutObject(path, src)
	if err != nil {
		return "", fmt.Errorf("文件上传到 OSS 失败: %v", err)
	}

	// 返回文件路径
	fileURL := fmt.Sprintf("https://%s.%s/%s", bucketName, endpoint, path)
	return fileURL, nil
}
