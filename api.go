package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func _well_known(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"providers.v1": "/v1/providers/"})
}

func versions(c *gin.Context) {

	provider, err := loadProviderJson(c.Param("namespace"), c.Param("name"))
	if err != nil {
		if err.Error() == "provider not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	}

	resp := []gin.H{}
	for _, v := range provider.Version {
		plats := []gin.H{}
		for _, p := range v.Platforms {
			plats = append(plats, gin.H{"os": p.Os, "arch": p.Arch})
		}

		resp = append(resp, gin.H{
			"version":   v.Version,
			"protocols": v.Protocols,
			"platforms": plats,
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"versions": resp})
}

func save(c *gin.Context) {
	form, _ := c.MultipartForm()
	file_index := ""

	for i := range form.File {
		file_index = i
		break
	}

	file, header, _ := c.Request.FormFile(file_index)
	defer file.Close()
	uploaded := filepath.Base(header.Filename)

	content, _ := io.ReadAll(file)

	ns, name, version, _os, arch := c.Param("namespace"), c.Param("name"), c.Param("version"), c.Param("os"), c.Param("arch")
	provider, err := loadProviderJson(ns, name)
	if err != nil {
		provider = &VersionJson{}
	}

	binary_name := fmt.Sprintf("%s_%s", uploaded, version)
	zip_name := fmt.Sprintf("%s_%s_%s_%s.zip", uploaded, version, _os, arch)
	sum_name := fmt.Sprintf("%s_%s_SHA256SUMS", uploaded, version)
	sig_name := fmt.Sprintf("%s_%s_SHA256SUMS.sig", uploaded, version)
	zip_content, _ := compressFile(content, binary_name)

	saveToS3(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", ns, name, zip_name), zip_content)

	sum := generateShaSum(zip_content)
	sum_content := getFromS3(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", ns, name, sum_name))
	sum_content = []byte(fmt.Sprintf("%s%x  %s\n", sum_content, sum, zip_name))

	saveToS3(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", ns, name, sum_name), sum_content)

	sign_content, _ := signGpg(sum_content)
	saveToS3(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", ns, name, sig_name), sign_content)

	platform := Platform{
		Os:       _os,
		Arch:     arch,
		Filename: zip_name,
		Shasum:   fmt.Sprintf("%x", sum),
	}

	platform.SigningKey.GpgPublicKey = append(platform.SigningKey.GpgPublicKey, GpgPublicKey{
		KeyId:      publicKeyId(),
		AsciiArmor: string(publicKey),
		SourceUrl:  "https://registry.svc.dev.aims.altavec.com/",
	})

	_version := Version{
		Version:   version,
		Protocols: []string{"4.0", "5.1"},
	}

	if len(provider.Version) == 0 {
		_version.Platforms = append(_version.Platforms, platform)
		provider.Version = append(provider.Version, _version)
	} else {

		for vi, v := range provider.Version {
			if v.Version == version {
				for pi, p := range v.Platforms {
					if p.Os == c.Param("os") && p.Arch == c.Param("arch") {
						break
					} else {
						if pi == (len(v.Platforms) - 1) {
							v.Platforms = append(v.Platforms, platform)
							break
						}
					}
				}
			} else {
				if vi == (len(provider.Version) - 1) {
					_version.Platforms = append(_version.Platforms, platform)
					provider.Version = append(provider.Version, _version)
					break
				}
			}
		}
	}
	data, _ := json.Marshal(*provider)
	saveToS3(fmt.Sprintf("terraform-registry/providers/%s/%s.json", ns, name), data)

	c.IndentedJSON(http.StatusOK, provider)
}

func _package(c *gin.Context) {
	provider, err := loadProviderJson(c.Param("namespace"), c.Param("name"))
	if err != nil {
		if err.Error() == "provider not found" {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	}

	for _, v := range provider.Version {
		if v.Version == c.Param("version") {
			for _, p := range v.Platforms {
				if p.Os == c.Param("os") && p.Arch == c.Param("arch") {
					ext := fmt.Sprintf("%s_%s.zip", c.Param("os"), c.Param("arch"))
					p.DownloadUrl = generatePresignedURL(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", c.Param("namespace"), c.Param("name"), p.Filename))
					p.ShasumsSignatureUrl = generatePresignedURL(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", c.Param("namespace"), c.Param("name"), strings.Replace(p.Filename, ext, "SHA256SUMS.sig", -1)))
					p.ShasumsUrl = generatePresignedURL(fmt.Sprintf("terraform-registry/binaries/%s/%s/%s", c.Param("namespace"), c.Param("name"), strings.Replace(p.Filename, ext, "SHA256SUMS", -1)))
					c.IndentedJSON(http.StatusOK, p)
					return
				}
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "provider not found"})
}

func loadProviderJson(namespace string, name string) (*VersionJson, error) {
	filepath := fmt.Sprintf("terraform-registry/providers/%s/%s.json", namespace, name)
	req := &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(filepath),
	}

	resp, err := client.GetObject(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r VersionJson
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func saveToS3(path string, content []byte) {
	req := &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(path),
		Body:   bytes.NewReader(content),
	}
	result, err := client.PutObject(context.TODO(), req)
	if err != nil {
		log.Fatal("saveToS3", err, result)
	}
}

func getFromS3(path string) (content []byte) {
	req := &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(path),
	}
	resp, err := client.GetObject(context.TODO(), req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	content, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func generatePresignedURL(key string) string {
	presignClient := s3.NewPresignClient(client)
	presignedUrl, _ := presignClient.PresignGetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(os.Getenv("BUCKET_NAME")),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(time.Minute*15))
	return presignedUrl.URL
}
