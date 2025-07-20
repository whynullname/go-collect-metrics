package rsareader

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/whynullname/go-collect-metrics/internal/logger"
)

var ErrEmptyKeyPath = errors.New("RSA key path is empty")

func ReadPublicRSAKey(path string) (*rsa.PublicKey, error) {
	if path == "" {
		return nil, ErrEmptyKeyPath
	}

	data, err := readPEMFile(path)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PublicKey(data)
	if err != nil {
		logger.Log.Errorf("failed to parse RSA public key: %v", err)
		return nil, err
	}

	return key, nil
}

func ReadPrivateRSAKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, ErrEmptyKeyPath
	}

	data, err := readPEMFile(path)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		logger.Log.Errorf("failed to parse RSA private key: %v", err)
		return nil, err
	}

	return key, nil
}

func readPEMFile(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Errorf("failed to read file %s: %v", path, err)
		return nil, err
	}

	block, _ := pem.Decode(body)
	if block == nil {
		logger.Log.Errorf("no PEM block found in file: %s", path)
		return nil, err
	}

	return block.Bytes, nil
}
