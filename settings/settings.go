package settings

import (
	"encoding/json/v2"
	"fmt"
	"os"

	"github.com/ryo-kagawa/go-cwindows/cwindows"
	"github.com/ryo-kagawa/go-utils/arrays"
)

const filename = "settings.json"

type Process struct {
	Name         string                 `json:"name"`
	AffinityMask cwindows.DWORD_PTR     `json:"affinityMask,omitzero"`
	Priority     cwindows.PriorityClass `json:"priority,omitzero"`
}

type Settings struct {
	Processes []Process `json:"processes"`
}

// settings.json から設定をロードします
func LoadSettings() (Settings, error) {
	settings := Settings{}
	file, err := os.Open(filename)
	if err != nil {
		return settings, fmt.Errorf("%sが見つかりません\n", filename)
	}
	defer file.Close()

	if err := json.UnmarshalRead(file, &settings); err != nil {
		return settings, fmt.Errorf("%sの内容が不正です\n", filename)
	}
	return settings, nil
}

func (s Settings) GetTargetProcess(processName string) (Process, bool) {
	return arrays.Find(
		s.Processes,
		func(process Process) bool {
			return process.Name == processName
		},
	)
}

func (s Settings) SaveSettings() error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ファイル作成に失敗: %v", err)
	}
	defer file.Close()

	if err := json.MarshalWrite(file, s); err != nil {
		return fmt.Errorf("JSONエンコードに失敗: %v", err)
	}
	if err != nil {
		return fmt.Errorf("JSONエンコードに失敗: %v", err)
	}

	return nil
}
