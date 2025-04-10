#!/bin/bash

# Petit-Misskey ログファイル削除スクリプト
# このスクリプトはプロジェクト内のlog-*.jsonファイルを一括で削除します

echo "Petit-Misskey ログファイル削除ユーティリティ"
echo "-------------------------------------------"

# appディレクトリのパスを設定
APP_DIR="$(dirname "$0")/app"

# log-*.jsonファイルのパターンを設定
LOG_PATTERN="log-*.json"

# ログファイルの数を数える
LOG_COUNT=$(find "$APP_DIR" -maxdepth 1 -name "$LOG_PATTERN" | wc -l)

if [ "$LOG_COUNT" -eq 0 ]; then
    echo "削除対象のログファイルが見つかりませんでした。"
    exit 0
fi

echo "削除対象のログファイル数: $LOG_COUNT"
echo "削除対象："
find "$APP_DIR" -maxdepth 1 -name "$LOG_PATTERN" -exec basename {} \; | sort

# 確認を求める
read -p "これらのログファイルを削除してもよろしいですか？ (y/N): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "操作をキャンセルしました。"
    exit 0
fi

# ログファイルを削除
find "$APP_DIR" -maxdepth 1 -name "$LOG_PATTERN" -delete

echo "ログファイルの削除が完了しました。"
