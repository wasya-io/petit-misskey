package core

type Logger interface {
	Log(messageType string, message string)
	ReadyWithType(messageType string) LogEntry
	Flush()
	Close() // 追加: ロガーを正しく終了し、残りのログをフラッシュする
}

type LogEntry interface {
	WithType() LogEntry
	WithString() LogEntry
	WithInt() LogEntry
	Do(values ...interface{})
}
