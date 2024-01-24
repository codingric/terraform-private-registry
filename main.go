package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg.Region = "ap-southeast-2"
	if err != nil {
		fmt.Println("Error loading AWS configuration:", err)
		return
	}

	// Create an S3 client
	client = s3.NewFromConfig(cfg)
}

func main() {

	router := gin.Default()
	router.GET("/.well-known/terraform.json", _well_known)
	router.GET("/v1/providers/:namespace/:name/versions", versions)
	router.GET("/v1/providers/:namespace/:name/:version/download/:os/:arch", _package)

	router.Run("0.0.0.0:8080")
}
