package main

import (
	"asynchronous-encryption-go/encryption"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	AMOUNT_MINIMUM = 4000
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

	cert, err := tls.LoadX509KeyPair("experimental/server.crt", "experimental/server.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}

	// Load CA cert to verify client certificates
	caCert, err := ioutil.ReadFile("experimental/ca.crt")
	if err != nil {
		log.Fatalf("server: read ca cert: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	
	r := gin.Default()

	r.GET("/public-key", func(c *gin.Context) {

		if c.Request.TLS != nil {
			log.Println("TLS connection")
			clientCert := c.Request.TLS.PeerCertificates
		

			if len(clientCert) <= 0 {
				log.Println("No client certificate")
				return
			}

			if len(clientCert)> 0 {

				for i, cert := range clientCert {
					log.Printf("Certificate #%d: Subject=%s, Issuer=%s\n", i+1, cert.Subject, cert.Issuer)
				}
				
				clientCertificate := clientCert[0]
			if clientCertificate.Issuer.String() != "CN=MyCA" { // Replace "CN=MyCA" with your CA's issuer details
				log.Println("Untrusted certificate issuer:", clientCertificate.Issuer.String())
			} else {
				log.Println("Certificate issuer is trusted:", clientCertificate.Issuer.String())
			}
			
		}
		}

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

		amountStr := string(decryptedAmount)

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			fmt.Println("Error parsing float:", err)
			return
		}

		if amount < AMOUNT_MINIMUM {
			log.Println("amount too low ", amount)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Amount too low"})
			return
		}

		
		c.JSON(http.StatusOK, PurchaseResponse{
			DecryptedAmount: amountStr,
		})
	})


	server := &http.Server{
		Addr:    ":8080", 
		Handler: r,
		TLSConfig: &tls.Config{
			ClientCAs:  caCertPool,   
			ClientAuth: tls.RequireAndVerifyClientCert, 
			Certificates: []tls.Certificate{cert}, 
		},
	}

	fmt.Println("Starting server on local dev")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("server: %s", err)
	}
}