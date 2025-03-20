package setting

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/wasya-io/petit-misskey/model/misskey"
)

// TODO: tomlへのアクセスだけを行うようにする
// TODO: アカウント情報の返却、追加などのusecaseに近い処理はaccountsに移管する

type (
	UserSetting struct {
		value    *Value
		filepath string
	}
	Value struct {
		Instances map[string]Instance `toml:"instance"`
	}

	Instance struct {
		BaseUrl     string              `toml:"baseurl" validate:"required"`
		UserName    string              `toml:"username" validate:"required"`
		AccessToken misskey.AccessToken `toml:"token" validate:"required"`
	}
)

var ProviderSet = wire.NewSet(
	NewUserSetting,
)

func NewUserSetting() *UserSetting {
	// TODO: once.Doをかける
	settingDir, _ := os.UserConfigDir()
	settingPath := filepath.Join(settingDir, "petit-misskey.toml")

	value, _ := readValue(settingPath)

	return &UserSetting{
		value:    value,
		filepath: settingPath,
	}
}

func readValue(settingPath string) (*Value, error) {
	var data Value
	if _, err := os.Stat(settingPath); err != nil {
		return nil, nil
	}
	_, err := toml.DecodeFile(settingPath, &data)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &data, nil
}

// 設定ファイルの書き込み
func (s *UserSetting) WriteValue(instances map[string]Instance) error {
	var file *os.File
	file, err := os.Create(s.filepath)
	if err != nil {
		return errors.WithStack(err)
	}

	defer file.Close()

	v := &Value{
		Instances: instances,
	}
	err = toml.NewEncoder(file).Encode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// インスタンス情報の読み出し
func (s *UserSetting) GetInstances() map[string]Instance {
	return s.value.Instances
}

// 特定インスタンス情報の読み出し
func (s *UserSetting) GetInstanceByKey(key string) *Instance {
	var instance Instance
	var exists bool
	if instance, exists = s.value.Instances[key]; !exists {
		return nil
	}
	return &instance
}
