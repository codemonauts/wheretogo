package main

import (
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type RegexData struct {
	Match  string `json:"match"`
	Target string `json:"target"`
}
type WebhookData struct {
	Endpoint string `json:"endpoint"`
}

type Entry struct {
	Name   string            `json:"name"`
	Domain string            `json:"domain"`
	Type   string            `json:"type"`
	Path   string            `json:"path"`
	Data   map[string]string `json:"data"`
}

type Config struct {
	Entrys []Entry `json:"entrys"`
}

func loadConfig(bucketName string, bucketRegion string) *Config {
	tmpFile, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal("Couldn't create temp file: ", err)
	}

	log.Debug("Created tempfile %q for config file", tmpFile.Name())

	defer os.Remove(tmpFile.Name())

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(bucketRegion)},
	)

	downloader := s3manager.NewDownloader(sess)

	log.Info("Loading configfile from s3://%s/config.yml in %s", bucketName, bucketRegion)

	_, err = downloader.Download(tmpFile,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("config.yml"),
		})

	if err != nil {
		log.Fatalf("Unable to download configuration from S3: %s", err)
	}

	var configData []byte
	configData, err = ioutil.ReadAll(tmpFile)
	if err != nil {
		log.Fatalf("Couldn't read downloaded config file: %s", err)
	}
	c := &Config{}
	err = yaml.Unmarshal(configData, &c)
	check(err)

	return c
}
