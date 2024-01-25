package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var (
	client          *s3.Client
	privateKey      []byte
	publicKey       []byte
	privatePassword []byte
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	cfg.Region = "ap-southeast-2"
	if err != nil {
		fmt.Println("Error loading AWS configuration:", err)
		return
	}

	// Create an S3 client
	client = s3.NewFromConfig(cfg)

	privateKey, _ = os.ReadFile(os.Getenv("GPG_PATH") + "/private.gpg")
	publicKey, _ = os.ReadFile(os.Getenv("GPG_PATH") + "/public.gpg")
	privatePassword, _ = os.ReadFile(os.Getenv("GPG_PATH") + "/password")

}

func main() {

	router := gin.Default()
	router.GET("/.well-known/terraform.json", _well_known)
	router.GET("/v1/providers/:namespace/:name/versions", versions)
	router.GET("/v1/providers/:namespace/:name/:version/download/:os/:arch", _package)
	router.POST("/v1/providers/:namespace/:name/:version/upload/:os/:arch", save)

	router.Run("0.0.0.0:8080")
}
