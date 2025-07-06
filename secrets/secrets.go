package secrets

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"os"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength  = 16
	nonceLength = 12
)

// stores API secrets
type secrets struct {
	Username        string
	IntegrationCode string
	Secret          string
}

// api secrets, filepath, and mutex
type SecretsCollection struct {
	sync.RWMutex
	secrets  secrets
	FilePath string
}

// returns true if secrets have been loaded to memory
func (sc *SecretsCollection) SetSecrets(IntegrationCode, secret, username string) {
	sc.RLock()
	defer sc.RUnlock()
	sc.secrets.IntegrationCode = IntegrationCode
	sc.secrets.Secret = secret
	sc.secrets.Username = username
}

// returns true if filepath is present on disk
func (sc *SecretsCollection) GetSecrets() (IntegrationCode, secret, username string) {
	sc.RLock()
	defer sc.RUnlock()
	return sc.secrets.IntegrationCode, sc.secrets.Secret, sc.secrets.Username
}

func (sc *SecretsCollection) SecretsAreLoaded() bool {
	sc.RLock()
	defer sc.RUnlock()
	return sc.secrets.Secret != "" && sc.secrets.IntegrationCode != "" && sc.secrets.Username != ""
}

// returns true if filepath is present on disk
func (sc *SecretsCollection) EncFilePresent() bool {
	sc.RLock()
	defer sc.RUnlock()
	if _, err := os.Stat(sc.FilePath); err == nil {
		return true
	}
	return false
}

// uses password to decrypt sc.FilePath to sc.Secrets
func (sc *SecretsCollection) DecryptSecrets(password []byte, maxExpectedBytes int) error {
	sc.Lock()
	defer sc.Unlock()
	// Open the encrypted file
	encryptedData, err := os.ReadFile(sc.FilePath)
	if err != nil {
		return err
	}

	// Extract salt, nonce, ciphertext
	if len(encryptedData) < saltLength+nonceLength {
		return fmt.Errorf("file too short")
	}
	salt := encryptedData[:saltLength]
	nonce := encryptedData[saltLength : saltLength+nonceLength]
	ciphertext := encryptedData[saltLength+nonceLength:]

	// Derive key
	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	// Decode gob from plaintext into result
	decoder := gob.NewDecoder(bytes.NewReader(plaintext))
	err = decoder.Decode(&sc.secrets)
	if err != nil {
		return err
	}

	return nil
}

// encrypts and saves sc.Secrets to sc.FilePath
func (sc *SecretsCollection) EncryptToDisk(password []byte) error {
	sc.RLock()
	defer sc.RUnlock()
	// First, gob-encode the data into a memory buffer
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(sc.secrets); err != nil {
		return err
	}
	plaintext := buffer.Bytes()

	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	key := deriveKey(password, salt)

	// Create AES-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Random nonce
	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	// Encrypt the plaintext
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Create the final output: salt || nonce || ciphertext
	output := append(salt, nonce...)
	output = append(output, ciphertext...)

	// Write to file
	return os.WriteFile(sc.FilePath, output, 0600)
}

// deriveKey derives a key from the password and salt using Argon2id.
func deriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32) // 32 bytes = 256 bits for AES-256
}
