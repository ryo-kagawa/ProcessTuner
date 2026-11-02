package command

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ryo-kagawa/ProcessTuner/settings"
	"github.com/ryo-kagawa/go-cwindows/cwindows"
	"github.com/ryo-kagawa/go-utils/commandline"
)

type Command struct{}

var _ = (commandline.RootCommand)(Command{})

var reader = bufio.NewReader(os.Stdin)

func (Command) Execute([]string) (string, error) {
	fmt.Println("========================================")
	fmt.Println(" Processor Affinity Config Generator")
	fmt.Println("========================================")

	settingsValue := settings.Settings{}
	for {
		fmt.Println("※終了する場合は、プロセス名に何も入力せずにEnterを押してください。")
		fmt.Print("プロセス名: ")
		processName, err := readInput()
		if err != nil {
			return "", err
		}
		if processName == "" {
			break
		}
		process := settings.Process{}
		process.Name = processName

		fmt.Println("プライオリティーを設定します")
		fmt.Println("未入力: 未設定")
		fmt.Println("1: 高")
		fmt.Println("2: 通常以上")
		fmt.Println("3: 通常")
		fmt.Println("4: 通常以下")
		fmt.Println("5: 低")
		fmt.Print("プライオリティー: ")
		priorityString, err := readInput()
		if err != nil {
			return "", err
		}
		if priorityString != "" {
			switch priorityString {
			case "1":
				process.Priority = cwindows.HIGH_PRIORITY_CLASS
			case "2":
				process.Priority = cwindows.ABOVE_NORMAL_PRIORITY_CLASS
			case "3":
				process.Priority = cwindows.NORMAL_PRIORITY_CLASS
			case "4":
				process.Priority = cwindows.BELOW_NORMAL_PRIORITY_CLASS
			case "5":
				process.Priority = cwindows.IDLE_PRIORITY_CLASS
			}
		}

		fmt.Print("アフィニティマスク（2進数）: ")
		affinityMaskString, err := readInput()
		if err != nil {
			return "", err
		}
		if affinityMaskString != "" {
			value, err := strconv.ParseUint(affinityMaskString, 2, 64)
			if err != nil {
				return "", err
			}
			process.AffinityMask = cwindows.DWORD_PTR(value)
		}

		settingsValue.Processes = append(settingsValue.Processes, process)
	}

	// 設定をファイルに保存
	if err := settingsValue.SaveSettings(); err != nil {
		return "", fmt.Errorf("[エラー] 設定の保存に失敗: %v\n", err)
	}

	return "設定ツールを終了します。", nil
}

func readInput() (string, error) {
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(value, "\r\n"), nil
}
