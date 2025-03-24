/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wasya-io/petit-misskey/infrastructure/resolver"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/infrastructure/websocket"
	"github.com/wasya-io/petit-misskey/logger"
	"github.com/wasya-io/petit-misskey/view"
	"github.com/wasya-io/petit-misskey/view/stream"
)

// streamCmd represents the stream command
var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Misskey のストリーミングAPIを使ってタイムラインを表示します",
	Long: `Misskey のストリーミングAPIを使用してリアルタイムでタイムラインを
表示します。ホームタイムラインとローカルタイムラインを切り替えることができます。

使用例:
  petit-misskey stream --key="misskey.io"`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		if key == "" {
			fmt.Println("エラー: インスタンスキーが指定されていません。--keyフラグを使用してインスタンスキーを指定してください。")
			fmt.Println("使用例: petit-misskey stream --key=\"your-instance-key\"")
			return
		}

		setting := setting.NewUserSetting()       // ユーザ設定を呼び出す
		instance := setting.GetInstanceByKey(key) // ユーザ設定からインスタンスの接続情報を呼び出す
		if instance == nil {
			fmt.Printf("エラー: インスタンスキー '%s' が見つかりません。\n", key)
			return
		}

		resolver := resolver.NewMisskeyStreamUrlResolver()
		l := logger.New(true)                                                                          // ロガーを作成
		client, msgCh := websocket.NewClient(instance.BaseUrl, instance.AccessToken, resolver, nil, l) // websocketクライアントを作成

		model := stream.NewModel(instance, client, l, msgCh) // initializerでmodelを作る

		view.Run(model, l) // modelをrunnerに渡す
	},
}

func init() {
	rootCmd.AddCommand(streamCmd)
}
