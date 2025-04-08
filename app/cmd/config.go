/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/model/misskey"
	"github.com/wasya-io/petit-misskey/service/accounts"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Petit-Misskeyの設定を管理します",
	Long: `Petit-Misskeyの設定を対話的に管理します。
インスタンスの追加、一覧表示、削除などが可能です。

使用例:
  petit-misskey config`,
	Run: func(cmd *cobra.Command, args []string) {
		runConfigManager()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

// runConfigManager は設定管理の対話型インターフェースを実行します
func runConfigManager() {
	// 設定ファイルのパスを取得
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("設定ディレクトリの取得に失敗しました: %v\n", err)
		os.Exit(1)
	}
	configPath := filepath.Join(configDir, "petit-misskey.toml")

	// ユーザー設定と関連サービスの初期化
	userSetting := setting.NewUserSetting()
	accountService := accounts.NewService(userSetting)

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
			addInstance(scanner, userSetting, accountService)
		case "2":
			listInstances(userSetting)
		case "3":
			deleteInstance(scanner, userSetting)
		case "4":
			err := saveConfig(configPath, userSetting)
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

// saveConfig は設定をファイルに保存します
func saveConfig(path string, userSetting *setting.UserSetting) error {
	// 既存のインスタンス設定を取得
	instances := userSetting.GetInstances()

	// 設定ディレクトリが存在しない場合は作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("設定ディレクトリの作成に失敗: %w", err)
	}

	// インスタンスの変更をUserSettingに書き込む
	if err := userSetting.WriteValue(instances); err != nil {
		return fmt.Errorf("設定の保存に失敗: %w", err)
	}

	return nil
}

// addInstance はユーザーにインスタンス情報を入力してもらい、設定に追加します
func addInstance(scanner *bufio.Scanner, userSetting *setting.UserSetting, accountService *accounts.Service) {
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
		if userSetting.GetInstanceByKey(instanceKey) != nil {
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
	newInstance := setting.Instance{
		BaseUrl:     baseURL,
		UserName:    username,
		AccessToken: misskey.AccessToken(token),
	}

	// アカウントサービスを使って追加
	err := accountService.Add(instanceKey, newInstance)
	if err != nil {
		if err == accounts.ErrAccountAlreadyExists {
			fmt.Printf("エラー: インスタンス「%s」は既に登録されています。\n", instanceKey)
		} else {
			fmt.Printf("エラー: インスタンスの追加に失敗しました: %v\n", err)
		}
		return
	}

	fmt.Printf("インスタンス「%s」を追加しました。\n", instanceKey)
}

// listInstances は設定されているインスタンスの一覧を表示します
func listInstances(userSetting *setting.UserSetting) {
	fmt.Println("\n--- 登録済みインスタンス一覧 ---")

	instances := userSetting.GetInstances()
	if len(instances) == 0 {
		fmt.Println("登録されているインスタンスはありません。")
		return
	}

	for name, instance := range instances {
		fmt.Printf("- %s\n", name)
		fmt.Printf("  URL: %s\n", instance.BaseUrl)
		fmt.Printf("  ユーザー名: %s\n", instance.UserName)
		fmt.Printf("  トークン: %s\n", maskToken(string(instance.AccessToken)))
	}
}

// deleteInstance はインスタンスを削除します
func deleteInstance(scanner *bufio.Scanner, userSetting *setting.UserSetting) {
	fmt.Println("\n--- インスタンスの削除 ---")

	instances := userSetting.GetInstances()
	if len(instances) == 0 {
		fmt.Println("登録されているインスタンスはありません。")
		return
	}

	fmt.Println("削除可能なインスタンス:")
	for name := range instances {
		fmt.Printf("- %s\n", name)
	}

	fmt.Print("\n削除するインスタンス名を入力してください: ")
	scanner.Scan()
	instanceKey := scanner.Text()

	instance := userSetting.GetInstanceByKey(instanceKey)
	if instance == nil {
		fmt.Println("指定されたインスタンスは存在しません。")
		return
	}

	fmt.Printf("インスタンス「%s」を削除します。よろしいですか？ (y/N): ", instanceKey)
	scanner.Scan()
	confirm := scanner.Text()

	if strings.ToLower(confirm) == "y" {
		// インスタンスのマップを取得して更新
		instances := userSetting.GetInstances()
		delete(instances, instanceKey)

		// 更新したマップを書き込み
		if err := userSetting.WriteValue(instances); err != nil {
			fmt.Printf("エラー: インスタンスの削除に失敗しました: %v\n", err)
			return
		}

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
