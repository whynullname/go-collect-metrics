package shamiddleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
)

const headerKey = "HashSHA256"

func HashSHA256(cfg *config.ServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return hashSHA256Middleware(next, cfg)
	}
}

func hashSHA256Middleware(next http.Handler, cfg *config.ServerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerHash := r.Header.Get(headerKey)
		if headerHash == "" {
			logger.Log.Infof("Header hash empty")
			next.ServeHTTP(w, r)
			return
		}
		decodedHash, err := hex.DecodeString(headerHash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		encodedBody := hmac.New(sha256.New, []byte(cfg.HashKey))
		encodedBody.Write(bodyBytes)
		next.ServeHTTP(w, r)
		if !hmac.Equal(decodedHash, encodedBody.Sum(nil)) {
			logger.Log.Infof("Bad header hash.\n")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Infof("New best request with sha hash!\n")
		w.Header().Set(headerKey, headerHash)
		w.WriteHeader(http.StatusOK)
	})
}
