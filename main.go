package main

import (
	"asynchronous-encryption-go/encryption"
	"crypto/rsa"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PurchaseRequest struct {
	Amount string `json:"amount"`
}

type PurchaseResponse struct {
	DecryptedAmount string `json:"decrypted_amount"`
}

var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err error
	privateKey, publicKey, err = encryption.GenerateRSAKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate RSA key pair: %v", err)
	}
}

func main() {
	r := gin.Default()

	r.GET("/public-key", func(c *gin.Context) {
		pubPEM, err := encryption.ExportPublicKeyToPEM(publicKey)
		if err != nil {
			log.Printf("Failed to export public key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export public key"})
			return
		}
		c.String(http.StatusOK, string(pubPEM))
	})

	r.POST("/purchase", func(c *gin.Context) {
		var req PurchaseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Invalid request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		encryptedAmount, err := base64.StdEncoding.DecodeString(req.Amount)
		if err != nil {
			log.Printf("Invalid base64 encoding: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base64 encoding"})
			return
		}

		decryptedAmount, err := encryption.DecryptAmount(encryptedAmount, privateKey)
		if err != nil {
			log.Printf("Failed to decrypt amount: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt amount"})
			return
		}

		c.JSON(http.StatusOK, PurchaseResponse{
			DecryptedAmount: string(decryptedAmount),
		})
	})

	r.Run(":8080")
}