package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db            *sql.DB
	encryptionKey []byte
)

type subscription struct {
	ID              int `json:"id"`
	SubscriptionInfo string `json:"subscription_info"`
	// Add other fields as needed
}

func main() {
	// Generate a secure, random key for AES encryption
	key, err := generateRandomKey(32)
	if err != nil {
		fmt.Println("Error generating random key:", err)
		return
	}
	encryptionKey = key

	port := getPort()
	router := gin.Default()
	router.GET("/subscription", getSubscription)
	router.POST("/subscription", postSubscription)

	err = initDB()
	
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	defer db.Close()

	router.Run("localhost:" + port)
}

func generateRandomKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func getPort() string {
	port := os.Getenv("APPLICATION_PORT")
	if port == "" {
		port = "3535" // Default port if not set
	}
	return port
}

func initDB() error {
	// Set your MySQL database connection details here
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, databaseName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	fmt.Println("Connected to MySQL database")
	return nil
}

func encrypt(data string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(data))

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decrypt(encrypted string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func getSubscription(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM subscription_table")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while querying the database"})
		return
	}
	defer rows.Close()

	var subscriptions []subscription
	for rows.Next() {
		var s subscription
		err := rows.Scan(&s.ID, &s.SubscriptionInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while scanning rows"})
			return
		}

		decryptedInfo, err := decrypt(s.SubscriptionInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decrypting data"})
			return
		}

		s.SubscriptionInfo = decryptedInfo
		subscriptions = append(subscriptions, s)
	}

	c.JSON(http.StatusOK, subscriptions)
}

func postSubscription(c *gin.Context) {
	var newSubscription subscription

	// Parse JSON request body into newSubscription
	if err := c.BindJSON(&newSubscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Encrypt subscription_info only if it is not an empty string
	if newSubscription.SubscriptionInfo != "" {
		encryptedInfo, err := encrypt(newSubscription.SubscriptionInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error encrypting data"})
			return
		}
		newSubscription.SubscriptionInfo = encryptedInfo
	}

	// Check if there are existing records in the subscription_table
	var rowCount int
	err := db.QueryRow("SELECT COUNT(*) FROM subscription_table").Scan(&rowCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking existing records"})
		return
	}

	// Delete records only if there are existing records
	if rowCount > 0 {
		_, err := db.Exec("DELETE FROM subscription_table")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting existing records"})
			return
		}
	} else {
		fmt.Println("No existing records to delete")
	}

	// Insert new subscription data into the database
	_, err = db.Exec("INSERT INTO subscription_table (id, subscription_info) VALUES (?, ?)", newSubscription.ID, newSubscription.SubscriptionInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting data into the database"})
		return
	}

	c.JSON(http.StatusCreated, newSubscription)
}
