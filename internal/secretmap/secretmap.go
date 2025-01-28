package secretmap

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"sync"
)

// cannot be changed to integer non divisible by 8 atm
const secretLength int = 1032
const archWordSize int = 64
const maxArchUint int64 = (1 << (archWordSize - 1)) - 1

type secretMapHolder struct {
	mu        *sync.RWMutex
	secretMap map[string][]byte
}

type SecretMapHolder interface {
	Get(key string) ([]byte, error)
	Set(value []byte) (string, error)
}

func NewSecretMapHolder() SecretMapHolder {
	return &secretMapHolder{mu: &sync.RWMutex{}, secretMap: make(map[string][]byte)}
}

func generatePad() ([]byte, error) {
	b := make([]byte, secretLength)
	rand.Read(b)
	return b, nil
}

func getMapKeyFromPad(pad string) string {

	bytes := sha512.Sum512([]byte(pad))

	return hex.EncodeToString(bytes[:])
}

func (s *secretMapHolder) Get(base64MapKey string) ([]byte, error) {
	mapKey := getMapKeyFromPad(base64MapKey)
	decryptionKey := make([]byte, secretLength)

	_, err := base64.RawURLEncoding.Strict().Decode(decryptionKey, []byte(base64MapKey))
	if err != nil {
		slog.Error("Error decoding base64 key", "error", err)
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.secretMap[mapKey]
	if !ok {
		return nil, errors.New("no secret found")
	}
	decryptedSecret, err2 := encrypt(val, decryptionKey)
	if err2 != nil {
		slog.Error("Error decrypting secret", "error", err2)
		return nil, err2
	}
	delete(s.secretMap, mapKey)
	decryptedSecret = bytes.TrimRight(decryptedSecret, "\x00")

	return decryptedSecret, nil
}

func encrypt(value []byte, pad []byte) ([]byte, error) {
	if len(value) != len(pad) || len(value) != secretLength {
		return nil, errors.New("invalid pad or secret")
	}
	encrypted := make([]byte, len(pad))
	for i := 0; i < len(value); i += 1 {
		encrypted[i] = value[i] ^ pad[i]
	}
	return encrypted, nil
}

func (s *secretMapHolder) prepareValue(value []byte) ([]byte, error) {
	if len(value) < secretLength {
		value = append(value, make([]byte, secretLength-len(value))...)
		for i := len(value); i < secretLength; i++ {
			value[i] = 0
		}
	}
	if len(value) > secretLength {
		return nil, errors.New("secret value too long")
	}
	return value, nil
}

func (s *secretMapHolder) Set(value []byte) (string, error) {
	value, err := s.prepareValue(value)
	if err != nil {
		return "", err
	}

	pad, err := generatePad()
	if err != nil {
		slog.Error("Error generating secret pad", "error", err)
		return "", err
	}
	encrypted, err := encrypt(value, pad)
	if err != nil {
		slog.Error("Error while encrypting", "error", err)
		return "", err
	}
	base64EncryptionKey := base64.RawURLEncoding.Strict().EncodeToString(pad)
	mapKey := getMapKeyFromPad(base64EncryptionKey)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.secretMap[mapKey] = encrypted

	return base64EncryptionKey, nil
}
