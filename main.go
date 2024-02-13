package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
		os.Exit(1)
		return
	}

	// Create an S3 client
	client = s3.NewFromConfig(cfg)

	if os.Getenv("GPG_PATH") == "" {
		fmt.Println("GPG_PATH variable required")
		os.Exit(1)
	}
	privateKey, _ = os.ReadFile(os.Getenv("GPG_PATH") + "/private.gpg")
	publicKey, _ = os.ReadFile(os.Getenv("GPG_PATH") + "/public.gpg")
	privatePassword, _ = os.ReadFile(os.Getenv("GPG_PATH") + "/password")

	if os.Getenv("BUCKET_NAME") == "" {
		fmt.Println("BUCKET_NAME variable required")
		os.Exit(1)
	}

}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()                   // empty engine
	router.Use(DefaultStructuredLogger()) // adds our new middleware
	router.Use(gin.Recovery())            // adds the default recovery middleware
	router.GET("/.well-known/terraform.json", _well_known)
	router.GET("/v1/providers/:namespace/:name/versions", versions)
	router.GET("/v1/providers/:namespace/:name/:version/download/:os/:arch", _package)
	router.POST("/v1/providers/:namespace/:name/:version/upload/:os/:arch", save)
	log.Info().Msg("Serving HTTP on port 8080")
	router.Run("0.0.0.0:8080")
}
