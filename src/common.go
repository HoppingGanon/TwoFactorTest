package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"crypto/rand"
	"encoding/base64"

	"github.com/labstack/echo"
)

// ======================================
// ユーザーから受け取るデータの構造体

// ユーザーから送付される認証情報
type AuthData struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

// ユーザーから送付される二段階認証情報
type TowFactorData struct {
	TempToken string `json:"tempToken"`
	OneTime   string `json:"oneTime"`
}

// テスト用のhtmlファイルに渡す環境変数の値
type ViewEnv struct {
	EnvAuthBaseUri   string
	EnvSmtpToAddress string
}

// ======================================

// ヘッダからBearerトークンを抜き出す関数
func getBearer(c echo.Context) (string, bool) {
	auth := c.Request().Header.Get("Authorization")
	typeStr := substring(auth, 0, 7)

	if typeStr != "Bearer " {
		return "", false
	}
	return substring(auth, 7, len(auth)-7), true
}

// SHA256のハッシュをバイナリで返す
func getSHA256Bytes(s string) []byte {
	r := sha256.Sum256([]byte(s))
	return r[:]
}

// SHA256の文字列(hex)をバイナリで返す
func getSHA256(s string) string {
	return hex.EncodeToString(getSHA256Bytes(s))
}

// 文字列を抜き出す
func substring(s string, start int, count int) string {
	if len(s) < start {
		return ""
	} else if len(s) < start+count {
		return s[start:]
	} else {
		return s[start : start+count]
	}
}

// ランダムなハッシュをBase64形式で生成する
func getRandomBase64() (string, error) {
	// 1億通りの乱数を生成
	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		return "", nil
	}
	// ハッシュをBase64にして返す
	hashbyte := getSHA256Bytes(fmt.Sprintf("%s", n))
	return base64.StdEncoding.EncodeToString(hashbyte), nil
}

// envLoad 環境変数のロード
func loadEnv() {
	// 開発環境のファイルを読み込む
	fmt.Println("\n 【環境変数の設定】")
	envAuthServerPort = getEnv("AUTH_SERVER_PORT", envAuthServerPort)
	envAuthBaseUri = getEnv("AUTH_BASE_URI", envAuthBaseUri)
	envSmtpHost = getEnv("SMTP_HOST", envSmtpHost)
	envSmtpPort = getEnv("SMTP_PORT", envSmtpPort)
	envSmtpUser = getEnv("SMTP_USER", envSmtpUser)
	envSmtpPassword = getEnv("SMTP_PASSWORD", envSmtpPassword)
	envSmtpFromAddress = getEnv("SMTP_FROM_ADDRESS", envSmtpFromAddress)
	envSmtpToAddress = getEnv("SMTP_TO_ADDRESS", envSmtpToAddress)
}

func getEnv(name string, defaultValue string) string {
	fmt.Printf("%-20s", name)
	if os.Getenv(name) == "" {
		fmt.Printf(" = %s\n", defaultValue)
		return defaultValue
	}
	fmt.Printf(" = %s\n", os.Getenv(name))
	return os.Getenv(name)
}
