package accounts

import (
	"errors"

	"github.com/wasya-io/petit-misskey/infrastructure/setting"
)

type (
	Service struct {
		setting *setting.UserSetting
	}
)

var ErrAccountAlreadyExists = errors.New("account already exists")

func NewService(setting *setting.UserSetting) *Service {
	return &Service{
		setting: setting,
	}
}

// アカウント情報の読み出し
func (s *Service) Get(key string) *setting.Instance {
	return s.setting.GetInstanceByKey(key)
}

// アカウント情報の追加
func (s *Service) Add(key string, account setting.Instance) error {
	instances := s.setting.GetInstances()

	if _, exists := instances[key]; exists {
		// 同じkeyが存在していたらエラーを返す
		return ErrAccountAlreadyExists
	}

	instances[key] = account

	s.setting.WriteValue(instances)

	return nil
}
