package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/cdn"
	"github.com/qiniu/go-sdk/v7/storage"
	"os"
	"sync"
	"time"
)

type MyPutRet struct {
	Key    string
	Hash   string
	Fsize  int
	Bucket string
	Name   string
}

func UploadFileToS3(localFiles []string) {
	awsConfig := GetAWSConfig()
	if awsConfig != nil {
		var wg sync.WaitGroup
		for _, awsInfo := range awsConfig {
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(awsInfo.Region), //桶所在的区域
				Credentials: credentials.NewStaticCredentials(
					awsInfo.AccessKey, // accessKey
					awsInfo.SecretKey, // secretKey
					""),               //sts的临时凭证
			})
			if err != nil {
				fmt.Errorf("new session error:%v", err)
				return
			}
			//upload file
			for _, l := range localFiles {
				wg.Add(1)
				go func(localFile string) {
					defer wg.Done()
					exist, _ := PathExists(localFile)
					if !exist {
						return
					}

					buffer, _ := os.ReadFile(localFile)
					key := awsInfo.KeyPrefix + localFile
					//fmt.Println("tokenList==", string(buffer))
					_, err := s3.New(sess).PutObject(&s3.PutObjectInput{
						Bucket: aws.String(awsInfo.Bucket), //桶名
						Key:    aws.String(key),
						Body:   bytes.NewReader(buffer),
					})
					if err != nil {
						fmt.Errorf("put file to s3 error:%v", err)
					}
					//c.log.Info("upload s3 info:", ret)
				}(l)
			}
			wg.Wait()
			//refresh dir
			callerReference := time.Now().String()
			svc := cloudfront.New(sess)
			paths := &cloudfront.Paths{
				Items:    []*string{aws.String(awsInfo.KeyPrefix + "*")},
				Quantity: aws.Int64(1),
			}

			input := &cloudfront.CreateInvalidationInput{DistributionId: aws.String(awsInfo.DistributionId),
				InvalidationBatch: &cloudfront.InvalidationBatch{
					CallerReference: aws.String(callerReference),
					Paths:           paths,
				}}
			ret, err := svc.CreateInvalidation(input)
			if err != nil {
				fmt.Errorf("create invalidation error:%v", err)
			}
			fmt.Errorf("s3 refresh dir:%v", ret)

		}
	}
}

func UpLoadFile2QiNiu(paths []string) {
	qiNiuConfig := GetQiNiuConfig()
	mac := qbox.NewMac(qiNiuConfig.AccessKey, qiNiuConfig.SecretKey)
	cdnManager := cdn.NewCdnManager(mac)
	for _, bucket := range qiNiuConfig.Bucket {
		//upToken := putPolicy.UploadToken(mac)
		cfg := storage.Config{
			UseHTTPS: true,
		}
		//bucketManager := storage.NewBucketManager(mac, &cfg)
		formUploader := storage.NewFormUploader(&cfg)
		ret := MyPutRet{}
		putExtra := storage.PutExtra{
			Params: map[string]string{
				"x:name": "github logo",
			},
		}
		for _, path := range paths {
			key := qiNiuConfig.KeyPrefix + path
			putPolicy := storage.PutPolicy{
				Scope: fmt.Sprintf("%s:%s", bucket, key),
			}
			upToken := putPolicy.UploadToken(mac)
			err := formUploader.PutFile(context.Background(), &ret, upToken, key, path, &putExtra)
			if err != nil {
				fmt.Errorf("PutFile Error:%v", err)
			}
			fmt.Println("upload info:", ret.Bucket, ret.Key, ret.Fsize, ret.Hash, ret.Name)
		}
	}

	_, err := cdnManager.RefreshDirs([]string{GetTokenListConfig().LogoPrefix, GetTokenListConfig().AwsLogoPrefix})
	if err != nil {
		fmt.Errorf("fetch dirs error:%v", err)
	}

}

func WriteInfoToFile(fileName string, fileInfo interface{}) error {
	listFile, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	defer listFile.Close()
	encoder := json.NewEncoder(listFile)
	err := encoder.Encode(fileInfo)
	return err
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}
