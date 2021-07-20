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

type DomainConfig struct {
	Name       string `yaml:"name"`
	PathPrefix string `yaml:"pathPrefix"`
	Match      string `yaml:"match"`
	Target     string `yaml:"target"`
}

type Config map[string][]DomainConfig

func loadLocalConfig(filename string) *Config {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Could open local config file: ", err)
	}

	c := &Config{}
	err = yaml.Unmarshal(file, &c)
	check(err)
	return c
}

func loadConfigFromS3(bucketName string, bucketRegion string) *Config {
	tmpFile, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal("Couldn't create temp file: ", err)
	}

	defer os.Remove(tmpFile.Name())

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(bucketRegion)},
	)

	downloader := s3manager.NewDownloader(sess)

	log.Infof("Loading configfile from s3://%s/config.yml in %s", bucketName, bucketRegion)

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
