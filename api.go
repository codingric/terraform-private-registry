package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
