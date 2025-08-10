package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config アプリケーションの設定構造体
type Config struct {
	// Firebase関連
	FirebaseCredentials string
	GCloudProject      string
	
	// GitHub関連
	GitHubToken    string
	GitHubUserName string
	
	// エミュレータ関連
	FirestoreEmulatorHost string
	FirebaseAuthEmulatorHost string
	
	// サーバー関連
	Port string
}

// LoadConfig 環境変数から設定を読み込みます
func LoadConfig() *Config {
	// .envファイルを読み込み（エラーは無視）
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env ファイルが見つかりません（本番環境では正常です）")
	}

	config := &Config{
		FirebaseCredentials:      os.Getenv("CREDENTIALS"),
		GCloudProject:           os.Getenv("GCLOUD_PROJECT"),
		GitHubToken:             os.Getenv("GITHUB_TOKEN"),
		GitHubUserName:          os.Getenv("GITHUB_USER_NAME"),
		FirestoreEmulatorHost:   os.Getenv("FIRESTORE_EMULATOR_HOST"),
		FirebaseAuthEmulatorHost: os.Getenv("FIREBASE_AUTH_EMULATOR_HOST"),
		Port:                    getEnvWithDefault("PORT", "8081"),
	}

	return config
}

// getEnvWithDefault 環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsEmulatorMode エミュレータモードかどうかを判定します
func (c *Config) IsEmulatorMode() bool {
	return c.FirestoreEmulatorHost != "" || c.FirebaseAuthEmulatorHost != ""
}
