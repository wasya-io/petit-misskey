package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Instance はMisskeyインスタンスの設定情報を表します
type Instance struct {
	BaseUrl     string `toml:"baseurl"`
	UserName    string `toml:"username"`
	AccessToken string `toml:"token"`
}

// Config はTOML設定ファイルの構造を表します
type Config struct {
	Instances map[string]Instance `toml:"instance"`
}

func main() {
	// 設定ファイルのパスを取得
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("設定ディレクトリの取得に失敗しました: %v\n", err)
		os.Exit(1)
	}
	configPath := filepath.Join(configDir, "petit-misskey.toml")

	// 既存の設定を読み込み
	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Printf("警告: 既存の設定ファイルの読み込みに失敗しました: %v\n", err)
		// 新しい設定を初期化
		config = &Config{
			Instances: make(map[string]Instance),
		}
	}

	// 既存の設定がない場合は初期化
	if config == nil {
		config = &Config{
			Instances: make(map[string]Instance),
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n==== Petit-Misskey設定ツール ====")
		fmt.Println("1. インスタンスを追加")
		fmt.Println("2. インスタンス一覧を表示")
		fmt.Println("3. インスタンスを削除")
		fmt.Println("4. 設定を保存して終了")
		fmt.Println("5. 変更を破棄して終了")
		fmt.Print("操作を選択してください (1-5): ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			addInstance(scanner, config)
		case "2":
			listInstances(config)
		case "3":
			deleteInstance(scanner, config)
		case "4":
			err := saveConfig(configPath, config)
			if err != nil {
				fmt.Printf("設定の保存に失敗しました: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("設定を保存しました: %s\n", configPath)
			os.Exit(0)
		case "5":
			fmt.Println("変更を破棄して終了します。")
			os.Exit(0)
		default:
			fmt.Println("無効な選択です。1から5の数字を入力してください。")
		}
	}
}

// loadConfig は設定ファイルを読み込みます
func loadConfig(path string) (*Config, error) {
	// ファイルが存在しない場合は新規作成と判断
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, errors.Wrap(err, "TOML設定ファイルのデコードに失敗")
	}

	// インスタンスマップの初期化
	if config.Instances == nil {
		config.Instances = make(map[string]Instance)
	}

	return &config, nil
}

// saveConfig は設定をファイルに保存します
func saveConfig(path string, config *Config) error {
	// 設定ディレクトリが存在しない場合は作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "設定ディレクトリの作成に失敗")
	}

	// ファイルを作成またはトランケート
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "設定ファイルの作成に失敗")
	}
	defer file.Close()

	// TOMLエンコーダを使用して設定を書き込み
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return errors.Wrap(err, "設定のエンコードに失敗")
	}

	return nil
}

// addInstance はユーザーにインスタンス情報を入力してもらい、設定に追加します
func addInstance(scanner *bufio.Scanner, config *Config) {
	fmt.Println("\n--- インスタンスの追加 ---")

	// インスタンス名の入力
	var instanceKey string
	for {
		fmt.Print("インスタンス名 (例: misskey.io): ")
		scanner.Scan()
		instanceKey = scanner.Text()
		if instanceKey == "" {
			fmt.Println("インスタンス名は必須です。")
			continue
		}
		if _, exists := config.Instances[instanceKey]; exists {
			fmt.Println("そのインスタンス名は既に登録されています。別の名前を入力してください。")
			continue
		}
		break
	}

	// BaseURLの入力
	var baseURL string
	for {
		fmt.Print("ベースURL (例: https://misskey.io): ")
		scanner.Scan()
		baseURL = scanner.Text()
		if baseURL == "" {
			fmt.Println("ベースURLは必須です。")
			continue
		}
		// プロトコルが含まれていない場合はhttpsを追加
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "https://" + baseURL
		}
		break
	}

	// ユーザー名の入力
	var username string
	for {
		fmt.Print("ユーザー名: ")
		scanner.Scan()
		username = scanner.Text()
		if username == "" {
			fmt.Println("ユーザー名は必須です。")
			continue
		}
		break
	}

	// アクセストークンの入力
	var token string
	for {
		fmt.Print("アクセストークン: ")
		scanner.Scan()
		token = scanner.Text()
		if token == "" {
			fmt.Println("アクセストークンは必須です。")
			continue
		}
		break
	}

	// インスタンス情報を設定に追加
	config.Instances[instanceKey] = Instance{
		BaseUrl:     baseURL,
		UserName:    username,
		AccessToken: token,
	}

	fmt.Printf("インスタンス「%s」を追加しました。\n", instanceKey)
}

// listInstances は設定されているインスタンスの一覧を表示します
func listInstances(config *Config) {
	fmt.Println("\n--- 登録済みインスタンス一覧 ---")
	if len(config.Instances) == 0 {
		fmt.Println("登録されているインスタンスはありません。")
		return
	}

	for name, instance := range config.Instances {
		fmt.Printf("- %s\n", name)
		fmt.Printf("  URL: %s\n", instance.BaseUrl)
		fmt.Printf("  ユーザー名: %s\n", instance.UserName)
		fmt.Printf("  トークン: %s\n", maskToken(instance.AccessToken))
	}
}

// deleteInstance はインスタンスを削除します
func deleteInstance(scanner *bufio.Scanner, config *Config) {
	fmt.Println("\n--- インスタンスの削除 ---")
	if len(config.Instances) == 0 {
		fmt.Println("登録されているインスタンスはありません。")
		return
	}

	fmt.Println("削除可能なインスタンス:")
	for name := range config.Instances {
		fmt.Printf("- %s\n", name)
	}

	fmt.Print("\n削除するインスタンス名を入力してください: ")
	scanner.Scan()
	instanceKey := scanner.Text()

	if _, exists := config.Instances[instanceKey]; !exists {
		fmt.Println("指定されたインスタンスは存在しません。")
		return
	}

	fmt.Printf("インスタンス「%s」を削除します。よろしいですか？ (y/N): ", instanceKey)
	scanner.Scan()
	confirm := scanner.Text()

	if strings.ToLower(confirm) == "y" {
		delete(config.Instances, instanceKey)
		fmt.Printf("インスタンス「%s」を削除しました。\n", instanceKey)
	} else {
		fmt.Println("削除をキャンセルしました。")
	}
}

// maskToken はトークンを表示用にマスクします
func maskToken(token string) string {
	if len(token) <= 8 {
		return "********"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
