# misskey クライアントを Go で作ってみるテスト

## Introduction

[misskey](https://join.misskey.page/ja-JP/)のエンドポイントを Go で実行してみるテストです。

## Features

### できたこと

現在は CUI 含めてユーザ入力を受け付けるインターフェースがないため、すべてテストコードから実行する必要があります。  
Dev Container を採用しているため、実行に際して Go のインストールは必須ではありません。(VSCode, Docker は必要です)

- `/meta` エンドポイントの実行
- `/notes/create` エンドポイントの実行とノートの作成
- API キー、ベース URL は設定ファイル(yaml)に退避

### 実行手順

WIP

## TODO

### やること

- アカウントの CRUD(toml へ書き込み)
- 通知(notifications)の取得
- ノート一覧の見た目改善
- home と local の指定

- カスタム絵文字はいったん諦めましょう

## NOTE

- dev container をリビルドした場合に cobra-cli の再インストールが必要かも
  - 毎回実行する or dev container の機能でどうにかする
- go install github.com/google/wire/cmd/wire@latest もどうにかする

## cobra + bubbletea + layered architecture

```
.
├── cmd ... cobra が実行するコマンドのエントリポイント  
│ ├── root.go ... サブコマンドなしで実行されるルートコマンド  
│ └── {command-name}.go ... サブコマンド  
├── config ... ユーザ操作で変更できないアプリケーションの設定  
├── domain ... 実装に要求する usecase に応じた interface 定義を配置  
├── infrastructure ... infrastructure 層. 通信やファイルなどの実体を操作する機能を配置  
│ ├── misskey ... Misskey の API 操作(stream 除く)  
│ ├── setting ... ユーザ操作で変更可能な設定ファイル処理  
│ └── websocket ... Misskey の stream API 操作  
│ 　 └── template ... text/template で用いるテンプレートファイルを配置  
├── model ... データ構造を配置  
│ └── misskey ... Misskey API で受け渡しするデータ構造  
├── service ... usecase 層  
├── test ... テストコードで用いるユーティリティやヘルパークラス、データを配置  
├── util ... 汎用的に使われるごく短い処理  
└── view ... presentation 層. 処理に必要な interface に依存し、tea.Model を実装する構造体を配置  
　 ├── {view model name} ... コマンドに対応して動作する tea.Model 層  
　 └── runner.go ... tea の実行処理(汎用)  
```

### 依存の方向

cmd(User Interface) -> view(usecase + presentation) -> service(usecase) -> domain <- infrastructure

## 参考

https://misskey.io/cli

https://github.com/mikuta0407/misskey-cli/tree/main
