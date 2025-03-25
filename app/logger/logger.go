package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/wasya-io/petit-misskey/config"
	"github.com/wasya-io/petit-misskey/domain/core"
)

// LogEntry はログのエントリを表す構造体
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	processor *Logger
}

// Logger はロギング機能を提供する構造体
type Logger struct {
	debugMode      bool
	entries        []LogEntry
	filePath       string
	maxBuffer      int
	maxEntries     int    // ファイルの最大エントリ数
	maxRotateFiles int    // 保持する最大ローテーションファイル数
	baseFilePath   string // ファイル名のベース部分（日時を含む）
	rotateCount    int    // 現在のローテーション回数
	entriesCount   int    // 現在のファイルのエントリ総数
	startTime      time.Time
	config         *config.Config
	mu             sync.Mutex // 複数のgoroutineからの同時アクセスを保護するmutex
}

// New は新しいLoggerインスタンスを作成する
func New(debugMode bool) *Logger {
	startTime := time.Now()
	cfg := config.NewConfig()

	// 日時を含むベースファイル名
	baseFilePath := fmt.Sprintf("log-%s", startTime.Format("20060102-150405"))
	filePath := baseFilePath + ".json"

	return &Logger{
		debugMode:      debugMode,
		entries:        make([]LogEntry, 0),
		filePath:       filePath,
		baseFilePath:   baseFilePath,
		maxBuffer:      5,
		maxEntries:     cfg.Log.MaxEntries,
		maxRotateFiles: cfg.Log.MaxRotationFiles,
		startTime:      startTime,
		config:         cfg,
	}
}

func (l *Logger) ReadyWithType(messageType string) core.LogEntry {
	return &LogEntry{
		Message:   "",
		Type:      messageType,
		processor: l,
	}
}

// Log はメッセージをログに記録する
func (l *Logger) Log(messageType string, message string) {
	if !l.debugMode {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   message,
		Type:      messageType,
	}
	l.entries = append(l.entries, entry)

	// バッファが一定量に達したらフラッシュ
	if len(l.entries) >= l.maxBuffer {
		l.flushUnsafe()
	}
}

// Flush は現在のログエントリをファイルに追記する（外部から呼び出し可能な安全版）
func (l *Logger) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.flushUnsafe()
}

// Close はロガーを終了し、残りのログをフラッシュする
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 終了ログを追加
	closeEntry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   "ロガーを終了します",
		Type:      "system",
	}
	l.entries = append(l.entries, closeEntry)

	// 残りのログをフラッシュ
	l.flushUnsafe()
}

// flushUnsafe は内部用のフラッシュ処理（ミューテックス取得済みの状態で呼び出すこと）
func (l *Logger) flushUnsafe() {
	if len(l.entries) == 0 {
		return
	}

	// ファイルが存在するか確認
	fileExists := false
	if _, err := os.Stat(l.filePath); err == nil {
		fileExists = true
	}

	var existingEntries []LogEntry

	if fileExists {
		// 既存のJSONファイルを読み込む
		existingData, err := os.ReadFile(l.filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ログファイルの読み込みに失敗しました: %v\n", err)
			return
		}

		// 有効なJSONかチェック
		if len(existingData) > 0 {
			if err := json.Unmarshal(existingData, &existingEntries); err != nil {
				// 不正なJSONの場合は新しく作り直す
				fmt.Fprintf(os.Stderr, "既存のログファイルが不正です。新しく作成します: %v\n", err)
				fileExists = false
				existingEntries = []LogEntry{}
			}
		}
	}

	// 新しいエントリと既存のエントリを結合
	allEntries := append(existingEntries, l.entries...)
	l.entriesCount = len(allEntries)

	// ローテーションが必要かどうかチェック
	if l.entriesCount > l.maxEntries {
		if err := l.rotateLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "ログローテーションに失敗しました: %v\n", err)
		} else {
			// ログローテーション成功後は現在のエントリのみ書き込む
			allEntries = l.entries
			l.entriesCount = len(allEntries)
		}
	}

	// ログをJSONとして処理
	data, err := json.MarshalIndent(allEntries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ログのシリアライズに失敗しました: %v\n", err)
		return
	}

	// ファイルを書き込みモードで開く（既存ファイルは上書き）
	file, err := os.Create(l.filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ログファイルのオープンに失敗しました: %v\n", err)
		return
	}
	defer file.Close()

	// JSONデータを書き込む
	if _, err := file.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "ログの書き込みに失敗しました: %v\n", err)
		return
	}

	// ログをクリア
	l.entries = make([]LogEntry, 0)
}

// rotateLogFile は現在のログファイルをローテーションする
func (l *Logger) rotateLogFile() error {
	// ローテーションファイルの一覧を取得
	filePattern := fmt.Sprintf("%s-*.json", l.baseFilePath)
	matches, err := filepath.Glob(filePattern)
	if err != nil {
		return fmt.Errorf("ローテーションファイルの検索に失敗しました: %w", err)
	}

	// 番号でソート
	rotationFiles := make(map[int]string)
	re := regexp.MustCompile(`-(\d{4})\.json$`)

	for _, match := range matches {
		submatch := re.FindStringSubmatch(match)
		if len(submatch) > 1 {
			if num, err := strconv.Atoi(submatch[1]); err == nil {
				rotationFiles[num] = match
			}
		}
	}

	// キーを昇順に並べる
	var keys []int
	for k := range rotationFiles {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// 最大数を超えた古いファイルを削除
	if len(keys) >= l.maxRotateFiles {
		for i := 0; i < len(keys)-(l.maxRotateFiles-1); i++ {
			oldestFile := rotationFiles[keys[i]]
			if err := os.Remove(oldestFile); err != nil {
				fmt.Fprintf(os.Stderr, "古いログファイルの削除に失敗しました: %v\n", err)
			}
		}
	}

	// 新しいローテーションファイル名を決定
	nextRotateNum := 0
	if len(keys) > 0 {
		nextRotateNum = keys[len(keys)-1] + 1
	}

	rotatedFilePath := fmt.Sprintf("%s-%04d.json", l.baseFilePath, nextRotateNum)

	// 現在のファイルをリネーム
	if _, err := os.Stat(l.filePath); err == nil {
		if err := os.Rename(l.filePath, rotatedFilePath); err != nil {
			return fmt.Errorf("ログファイルのリネームに失敗しました: %w", err)
		}

		// ローテーション情報を記録
		rotationEntry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Message:   fmt.Sprintf("ログファイルをローテーションしました: %s -> %s", l.filePath, rotatedFilePath),
			Type:      "system",
		}
		l.entries = append(l.entries, rotationEntry)
	}

	return nil
}

// SetDebugMode はデバッグモードの状態を設定する
func (l *Logger) SetDebugMode(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugMode = enabled
}

func (e *LogEntry) WithType() core.LogEntry {
	e.Message = e.Message + " %T"
	return e
}

func (e *LogEntry) WithString() core.LogEntry {
	e.Message = e.Message + " %s"
	return e
}

func (e *LogEntry) WithInt() core.LogEntry {
	e.Message = e.Message + " %d"
	return e
}

func (e *LogEntry) Do(values ...interface{}) {
	e.processor.Log(e.Type, fmt.Sprintf(e.Message, values...))
}
