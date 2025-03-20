/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/view"
	"github.com/wasya-io/petit-misskey/view/meta"
)

// metaCmd represents the meta command
var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "インスタンスのメタデータを表示します",
	Long: `インスタンスのメタデータを取得して表示します。
インスタンスキーを指定してください。

使用例:
  petit-misskey meta --key="your-instance-key"`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("meta run")

		key, _ := cmd.Flags().GetString("key")
		if key == "" {
			fmt.Println("エラー: インスタンスキーが指定されていません。--keyフラグを使用してインスタンスキーを指定してください。")
			fmt.Println("使用例: petit-misskey meta --key=\"your-instance-key\"")
			return
		}

		setting := setting.NewUserSetting()       // ユーザ設定を呼び出す
		instance := setting.GetInstanceByKey(key) // ユーザ設定からインスタンスの接続情報を呼び出す

		model := meta.InitializeModel(instance) // initializerでmodelを作る

		view.Run(model) // modelをrunnerに渡す
	},
}

func init() {
	rootCmd.AddCommand(metaCmd)

	// Here you will define your flags and configuration settings.
	// フラグの定義
	metaCmd.Flags().StringP("key", "k", "", "インスタンスキー（必須）")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// metaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// metaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
