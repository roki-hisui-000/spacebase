package config

import (
	"os"
	"testing"
)

// テスト用にcfgを一時的に設定するヘルパー関数
func setTestConfig(testCfg Config) {
	cfg = testCfg
}

// テスト用に現在のconfigを取得するヘルパー関数
func getTestConfig() Config {
	return cfg
}

func TestAdminUser(t *testing.T) {
	originalCfg := getTestConfig()   // 元のconfigを保存
	defer setTestConfig(originalCfg) // テスト後に元に戻す

	// 1. Config構造体の値が優先されること
	setTestConfig(Config{AdminUser: "test_user_from_config"})
	if AdminUser() != "test_user_from_config" {
		t.Errorf("Expected AdminUser to be 'test_user_from_config', got %s", AdminUser())
	}

	// 2. 環境変数の値が取得されること (Configが空の場合)
	setTestConfig(Config{}) // Configをリセット
	os.Setenv("ADMIN_USER", "env_user")
	if AdminUser() != "env_user" {
		t.Errorf("Expected AdminUser to be 'env_user', got %s", AdminUser())
	}
	os.Unsetenv("ADMIN_USER") // 環境変数をクリーンアップ

	// 3. 環境変数もConfigも空の場合、デフォルト値が返されること
	setTestConfig(Config{}) // Configをリセット
	if AdminUser() != "admin" {
		t.Errorf("Expected AdminUser to be 'admin', got %s", AdminUser())
	}
}

func TestAdminPass(t *testing.T) {
	originalCfg := getTestConfig()
	defer setTestConfig(originalCfg)

	// 1. Config構造体の値が優先されること
	setTestConfig(Config{AdminPass: "test_pass_from_config"})
	if AdminPass() != "test_pass_from_config" {
		t.Errorf("Expected AdminPass to be 'test_pass_from_config', got %s", AdminPass())
	}

	// 2. 環境変数の値が取得されること (Configが空の場合)
	setTestConfig(Config{})
	os.Setenv("ADMIN_PASS", "env_pass")
	if AdminPass() != "env_pass" {
		t.Errorf("Expected AdminPass to be 'env_pass', got %s", AdminPass())
	}
	os.Unsetenv("ADMIN_PASS")

	// 3. 環境変数もConfigも空の場合、デフォルト値が返されること
	setTestConfig(Config{})
	if AdminPass() != "admin_pass" {
		t.Errorf("Expected AdminPass to be 'admin_pass', got %s", AdminPass())
	}
}
