package misskey

import "time"

type (
	AccessToken string

	// api/meta
	Meta struct {
		AccessToken AccessToken `json:"i"`
		Detail      bool        `json:"detail"`
	}

	MetaResponse struct {
		MaintainerName string `json:"maintainerName"`
		Version        string `json:"version"`
		Name           string `json:"name"`
		BannerUrl      string `json:"bannerUrl"`
		IconUrl        string `json:"iconUrl"`
	}

	CreateNote struct {
		AccessToken AccessToken `json:"i"`
		Visibility  Visibility  `json:"visibility"`
		Text        string      `json:"text"`
	}

	CreateNoteResponse struct {
		CreatedNote Note `json:"createdNote"`
	}

	// misskey defined types
	CreatedNote struct {
		Id         string     `json:"id"`
		CreatedAt  string     `json:"createdAt"`
		UserId     string     `json:"userId"`
		Cw         *string    `json:"cw,omitempty"`
		User       User       `json:"user"`
		Visibility Visibility `json:"visibility"`
	}

	User struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		UserName  string `json:"username"`
		AvatarUrl string `json:"avatarUrl"`
		IsBot     bool   `json:"isBot"`
		IsCat     bool   `json:"isCat"`
	}

	// 新しい構造体定義
	// AvatarDecoration はアバターの装飾情報を表します
	AvatarDecoration struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	// BadgeRole はバッジロールの情報を表します
	BadgeRole struct {
		Name         string `json:"name"`
		IconURL      string `json:"iconUrl"`
		DisplayOrder int    `json:"displayOrder"`
	}

	// NoteUser はノート投稿者の情報を表します
	NoteUser struct {
		ID                string             `json:"id"`
		Name              string             `json:"name"`
		Username          string             `json:"username"`
		Host              any                `json:"host"`
		AvatarURL         string             `json:"avatarUrl"`
		AvatarBlurhash    string             `json:"avatarBlurhash"`
		AvatarDecorations []AvatarDecoration `json:"avatarDecorations"`
		IsBot             bool               `json:"isBot"`
		IsCat             bool               `json:"isCat"`
		Emojis            map[string]string  `json:"emojis"`
		OnlineStatus      string             `json:"onlineStatus"`
		BadgeRoles        []BadgeRole        `json:"badgeRoles"`
	}

	// FileProperties はファイルのプロパティ情報を表します
	FileProperties struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	// NoteFile はノートに添付されたファイル情報を表します
	NoteFile struct {
		ID           string         `json:"id"`
		CreatedAt    time.Time      `json:"createdAt"`
		Name         string         `json:"name"`
		Type         string         `json:"type"`
		Md5          string         `json:"md5"`
		Size         int            `json:"size"`
		IsSensitive  bool           `json:"isSensitive"`
		Blurhash     string         `json:"blurhash"`
		Properties   FileProperties `json:"properties"`
		URL          string         `json:"url"`
		ThumbnailURL string         `json:"thumbnailUrl"`
		Comment      any            `json:"comment"`
		FolderID     any            `json:"folderId"`
		Folder       any            `json:"folder"`
		UserID       any            `json:"userId"`
		User         any            `json:"user"`
	}

	// RenoteContent はリノートされた内容を表します
	RenoteContent struct {
		ID                       string            `json:"id"`
		CreatedAt                time.Time         `json:"createdAt"`
		UserID                   string            `json:"userId"`
		User                     NoteUser          `json:"user"`
		Text                     string            `json:"text"`
		Cw                       any               `json:"cw"`
		Visibility               string            `json:"visibility"`
		LocalOnly                bool              `json:"localOnly"`
		ReactionAcceptance       any               `json:"reactionAcceptance"`
		RenoteCount              int               `json:"renoteCount"`
		RepliesCount             int               `json:"repliesCount"`
		Reactions                map[string]int    `json:"reactions"`
		ReactionEmojis           map[string]string `json:"reactionEmojis"`
		ReactionAndUserPairCache []string          `json:"reactionAndUserPairCache"`
		Tags                     []string          `json:"tags"`
		FileIds                  []string          `json:"fileIds"`
		Files                    []NoteFile        `json:"files"`
		ReplyID                  any               `json:"replyId"`
		RenoteID                 any               `json:"renoteId"`
		ClippedCount             int               `json:"clippedCount"`
	}

	// NoteBody はノートの本文部分を表します
	NoteBody struct {
		ID                       string            `json:"id"`
		CreatedAt                time.Time         `json:"createdAt"`
		UserID                   string            `json:"userId"`
		User                     NoteUser          `json:"user"`
		Text                     string            `json:"text"`
		Cw                       any               `json:"cw"`
		Visibility               string            `json:"visibility"`
		LocalOnly                bool              `json:"localOnly"`
		ReactionAcceptance       any               `json:"reactionAcceptance"`
		RenoteCount              int               `json:"renoteCount"`
		RepliesCount             int               `json:"repliesCount"`
		Reactions                map[string]int    `json:"reactions"`
		ReactionEmojis           map[string]string `json:"reactionEmojis"`
		ReactionAndUserPairCache []any             `json:"reactionAndUserPairCache"`
		FileIds                  []any             `json:"fileIds"`
		Files                    []any             `json:"files"`
		ReplyID                  any               `json:"replyId"`
		RenoteID                 string            `json:"renoteId"`
		ClippedCount             int               `json:"clippedCount"`
		Renote                   RenoteContent     `json:"renote"`
	}

	// NoteContainer はノートのコンテナ部分を表します
	NoteContainer struct {
		ID   string   `json:"id"`
		Type string   `json:"type"`
		Body NoteBody `json:"body"`
	}

	// Note はMisskeyのノートを表します
	Note struct {
		Type string        `json:"type"`
		Body NoteContainer `json:"body"`
	}

	Visibility string
)
