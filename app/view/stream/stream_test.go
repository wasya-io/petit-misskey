package stream

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/infrastructure/websocket"
	"github.com/wasya-io/petit-misskey/model/misskey"
)

// TestStreamFunctionality はStreamモデルの機能をテストします
// テスト実行時に実際の挙動を確認するためのものです
func TestStreamFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("長時間実行テストをスキップします")
	}

	// テスト用インスタンス設定
	instance := &setting.Instance{
		BaseUrl:     "https://example.com",
		UserName:    "testuser",
		AccessToken: "test-token",
	}

	// モックWebSocketクライアント
	mockClient := &MockWebSocketClient{
		startCalled: false,
		stopCalled:  false,
	}

	// メッセージチャネル
	msgCh := make(chan tea.Msg)

	// Streamモデルの作成
	model := NewModel(instance, mockClient, msgCh)

	// 終了シグナルの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// メッセージ送信ゴルーチン
	go func() {
		defer wg.Done()
		sendTestMessages(ctx, msgCh)
	}()

	// メッセージ処理ゴルーチン
	go func() {
		defer wg.Done()
		processModelUpdates(ctx, model, msgCh)
	}()

	// Ctrl+Cで終了できるように
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("テスト実行中... Ctrl+Cで終了")
	select {
	case <-sigCh:
		fmt.Println("シグナルを受信しました。終了します...")
		cancel()
	case <-time.After(30 * time.Second):
		fmt.Println("30秒経過したため、テストを終了します")
		cancel()
	}

	wg.Wait()
	fmt.Println("テスト終了")
}

// MockWebSocketClient はテスト用のモックWebSocketクライアントです
type MockWebSocketClient struct {
	websocket.Client
	startCalled bool
	stopCalled  bool
}

func (m *MockWebSocketClient) Start() error {
	m.startCalled = true
	fmt.Println("MockWebSocketClient: Start() called")
	return nil
}

func (m *MockWebSocketClient) Stop() {
	m.stopCalled = true
	fmt.Println("MockWebSocketClient: Stop() called")
}

// sendTestMessages はテスト用のメッセージをチャネルに送信します
func sendTestMessages(ctx context.Context, msgCh chan tea.Msg) {
	// 最初に接続メッセージを送信
	msgCh <- websocket.WebSocketConnectedMsg{}
	fmt.Println("接続メッセージを送信しました")

	// テスト用のノートデータを作成
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			// 通常の投稿を作成
			note := createTestNote(i)
			msgCh <- websocket.NoteMessage{Note: note}
			fmt.Printf("ノートメッセージ %d を送信しました\n", i)
		}
	}

	// リノート付きの投稿を作成
	select {
	case <-ctx.Done():
		return
	case <-time.After(1 * time.Second):
		renote := createTestRenote()
		msgCh <- websocket.NoteMessage{Note: renote}
		fmt.Println("リノートメッセージを送信しました")
	}

	// エラーメッセージ
	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Second):
		msgCh <- websocket.WebSocketErrorMsg{Err: fmt.Errorf("テストエラー")}
		fmt.Println("エラーメッセージを送信しました")
	}

	// 切断メッセージ
	select {
	case <-ctx.Done():
		return
	case <-time.After(2 * time.Second):
		msgCh <- websocket.WebSocketDisconnectedMsg{Err: nil}
		fmt.Println("切断メッセージを送信しました")
	}
}

// createTestNote はテスト用のノートを作成します
func createTestNote(index int) *misskey.Note {
	now := time.Now()
	return &misskey.Note{
		Body: misskey.NoteContainer{
			ID:   fmt.Sprintf("note-id-%d", index),
			Type: "note",
			Body: misskey.NoteBody{
				ID:        fmt.Sprintf("note-id-%d", index),
				CreatedAt: now,
				User: misskey.NoteUser{
					ID:       fmt.Sprintf("user-id-%d", index),
					Name:     fmt.Sprintf("ユーザー%d", index),
					Username: fmt.Sprintf("user%d", index),
				},
				Text: fmt.Sprintf("これはテストノート%dです", index),
				// 他のフィールドは省略
			},
		},
	}
}

// createTestRenote はテスト用のリノートを作成します
func createTestRenote() *misskey.Note {
	now := time.Now()
	return &misskey.Note{
		Body: misskey.NoteContainer{
			ID:   "renote-id",
			Type: "note",
			Body: misskey.NoteBody{
				ID:        "renote-id",
				CreatedAt: now,
				User: misskey.NoteUser{
					ID:       "renote-user",
					Name:     "リノートユーザー",
					Username: "renote_user",
				},
				Text: "これはリノートです",
				Renote: misskey.RenoteContent{
					ID:        "original-note",
					CreatedAt: now,
					UserID:    "original-user",
					User: misskey.NoteUser{
						ID:       "original-user",
						Name:     "オリジナルユーザー",
						Username: "original_user",
					},
					Text: "これはオリジナルノートです",
					// 他のフィールドは省略
				},
				// 他のフィールドは省略
			},
		},
		// 他のフィールドは省略
	}
}

// processModelUpdates はモデルの更新を処理して結果を標準出力に表示します
func processModelUpdates(ctx context.Context, model *Model, msgCh chan tea.Msg) {
	// 初期化
	_ = model.Init()
	fmt.Println("モデルを初期化しました")
	fmt.Println("----------- 初期表示 -----------")
	fmt.Println(model.View())
	fmt.Println("---------------------------------")

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgCh:
			// モデルを更新
			updatedModel, _ := model.Update(msg)

			// 型アサーションで戻り値をStreamモデルに変換
			if streamModel, ok := updatedModel.(Model); ok {
				model = &streamModel

				// 更新後の表示を出力
				fmt.Println("\n----------- 表示更新 -----------")
				fmt.Println(model.View())
				fmt.Println("---------------------------------")
			}

			// エラーメッセージの処理
			switch m := msg.(type) {
			case websocket.WebSocketErrorMsg:
				fmt.Printf("エラーメッセージを受信しました: %v\n", m.Err)
			case websocket.WebSocketDisconnectedMsg:
				fmt.Println("切断メッセージを受信しました")
			}
		}
	}
}

// TestFormatNote はformatNote関数の単体テスト
func TestFormatNote(t *testing.T) {
	// 1. 通常の投稿のテスト
	normalNote := createTestNote(1)
	formatted := formatNote(normalNote)
	if formatted == "" {
		t.Error("フォーマットされた通常ノートが空です")
	}
	t.Logf("通常ノートのフォーマット結果:\n%s", formatted)

	// 2. リノートのテスト
	renoteNote := createTestRenote()
	formatted = formatNote(renoteNote)
	if formatted == "" {
		t.Error("フォーマットされたリノートが空です")
	}
	t.Logf("リノートのフォーマット結果:\n%s", formatted)
}
