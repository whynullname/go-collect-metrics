package encryptionmiddleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
)

func RSAMiddleware(cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return rsaDecryptionMiddleware(next, cfg.RSAKey)
	}
}

func rsaDecryptionMiddleware(next http.Handler, privateKey *rsa.PrivateKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if privateKey == nil {
			next.ServeHTTP(w, r)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		decryptedBody, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, bodyBytes)
		if err != nil {
			logger.Log.Errorf("Error while decrypt body %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
		r.ContentLength = int64(len(decryptedBody))
		next.ServeHTTP(w, r)
	})
}
