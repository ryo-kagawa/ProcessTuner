package command

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/ryo-kagawa/ProcessTuner/settings"
	"github.com/ryo-kagawa/go-cwindows/cwindows"
	"github.com/ryo-kagawa/go-utils/commandline"
)

type EVENT_TRACE_PROPERTIES_V2Data struct {
	Props      cwindows.EVENT_TRACE_PROPERTIES_V2
	LoggerName [128]cwindows.WCHAR
}

type Command struct{}

var _ = (commandline.RootCommand)(Command{})

func (Command) Execute([]string) (string, error) {
	settings, err := settings.LoadSettings()
	if err != nil {
		return "", err
	}

	eventTracePropertiesData := EVENT_TRACE_PROPERTIES_V2Data{
		Props: cwindows.EVENT_TRACE_PROPERTIES_V2{
			Wnode: cwindows.WNODE_HEADER{
				BufferSize:    cwindows.ULONG(unsafe.Sizeof(EVENT_TRACE_PROPERTIES_V2Data{})),
				ClientContext: cwindows.QueryPerformanceCounter,
				Flags:         cwindows.WNODE_FLAG_TRACED_GUID | cwindows.WNODE_FLAG_VERSIONED_PROPERTIES,
			},
			LogFileMode:      cwindows.EVENT_TRACE_REAL_TIME_MODE,
			LoggerNameOffset: cwindows.ULONG(unsafe.Sizeof(cwindows.EVENT_TRACE_PROPERTIES_V2{})),
		},
	}
	copy(eventTracePropertiesData.LoggerName[:], sessionNameWCHAR)
	cwindows.SechostControlTrace[cwindows.PEVENT_TRACE_PROPERTIES_V2](
		0,
		cwindows.LPCWSTR(sessionNameLPWSTR),
		&eventTracePropertiesData.Props,
		cwindows.EVENT_TRACE_CONTROL_STOP,
	)
	traceId := cwindows.CONTROLTRACE_ID(0)
	traceResult := cwindows.SechostStartTrace[cwindows.PEVENT_TRACE_PROPERTIES_V2](
		&traceId,
		cwindows.LPCWSTR(sessionNameLPWSTR),
		&eventTracePropertiesData.Props,
	)
	if traceResult != cwindows.ULONG(cwindows.ERROR_SUCCESS) {
		return "", fmt.Errorf("error StartTrace: %d", traceResult)
	}
	defer cwindows.SechostControlTrace[cwindows.PEVENT_TRACE_PROPERTIES_V2](
		traceId,
		cwindows.LPCWSTR(sessionNameLPWSTR),
		&eventTracePropertiesData.Props,
		cwindows.EVENT_TRACE_CONTROL_STOP,
	)

	traceResult = cwindows.SechostEnableTraceEx2(
		traceId,
		&cwindows.MicrosoftWindowsKernelProcess,
		cwindows.EVENT_CONTROL_CODE_ENABLE_PROVIDER,
		cwindows.TRACE_LEVEL_INFORMATION,
		cwindows.MicrosoftWindowsKernelProcessWINEVENT_KEYWORD_PROCESS,
		0,
		0,
		&cwindows.ENABLE_TRACE_PARAMETERS{
			Version: cwindows.ENABLE_TRACE_PARAMETERS_VERSION_2,
		},
	)
	if traceResult != cwindows.ULONG(cwindows.ERROR_SUCCESS) {
		return "", fmt.Errorf("error EnableTraceEx2: %d", traceResult)
	}
	options := cwindows.ETW_OPEN_TRACE_OPTIONS{}
	options.SetEventCallback(
		func(EventRecord cwindows.PEVENT_RECORD) {
			if EventRecord.EventHeader.EventDescriptor.Id == cwindows.MicrosoftWindowsKernelProcessEventIdProcessStart {
				userData := cwindows.NewMicrosoftWindowsKernelProcessProcessStartEvent(EventRecord.UserData, EventRecord.UserDataLength)
				processName := filepath.Base(cwindows.WCHARToString(userData.ImageName))
				process, found := settings.GetTargetProcess(processName)
				if found {
					processHandle, err := cwindows.Kernel32OpenProcess(
						cwindows.PROCESS_QUERY_LIMITED_INFORMATION|cwindows.PROCESS_SET_INFORMATION,
						cwindows.FALSE,
						cwindows.DWORD(userData.ProcessID),
					)
					if processHandle == cwindows.HANDLE(0) {
						fmt.Printf("error OpenProcess: %v\n", err)
						return
					}
					defer cwindows.Kernel32CloseHandle(processHandle)
					if process.Priority != 0 {
						cwindows.Kernel32SetPriorityClass(processHandle, process.Priority)
					}
					if process.AffinityMask != 0 {
						cwindows.Kernel32SetProcessAffinityMask(processHandle, process.AffinityMask)
					}
				}
			}
		},
	)
	traceHandle, err := cwindows.SechostOpenTraceFromRealTimeLogger(
		cwindows.PCWSTR(sessionNameLPWSTR),
		&options,
		nil,
	)
	if traceHandle == cwindows.INVALID_PROCESSTRACE_HANDLE {
		return "", err
	}
	defer cwindows.Kernel32CloseHandle(cwindows.HANDLE(traceHandle))

	go func() {
		cwindows.SechostProcessTrace(
			(*cwindows.PROCESSTRACE_HANDLE)(new(cwindows.ULONG64(traceHandle))),
			1,
			&cwindows.FILETIME{},
			&cwindows.FILETIME{},
		)
	}()

	// シグナルハンドラの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Toolhelp APIを使用して既存プロセス一覧を取得し、対象プロセスの優先度とアフィニティマスクを変更
	err = modifyExistingProcesses(settings)
	if err != nil {
		return "", fmt.Errorf("Failed to modify existing processes: %v", err)
	}

	// シグナルを受け取るまで待機
	_ = <-sigChan
	return "finish", nil
}

func modifyProcess(processId cwindows.DWORD, process settings.Process) {
	processHandle, err := cwindows.Kernel32OpenProcess(cwindows.PROCESS_SET_INFORMATION, cwindows.FALSE, processId)
	if err != nil {
		return
	}
	defer cwindows.Kernel32CloseHandle(processHandle)

	if process.Priority != 0 {
		cwindows.Kernel32SetPriorityClass(processHandle, process.Priority)
	}
	if process.AffinityMask != 0 {
		cwindows.Kernel32SetProcessAffinityMask(processHandle, process.AffinityMask)
	}
}

func modifyExistingProcesses(settings settings.Settings) error {
	handle, err := cwindows.Kernel32CreateToolhelp32Snapshot(cwindows.TH32CS_SNAPPROCESS, 0)
	if handle == cwindows.INVALID_HANDLE_VALUE {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}
	defer cwindows.Kernel32CloseHandle(handle)

	processEntry := cwindows.PROCESSENTRY32{
		Size: cwindows.DWORD(unsafe.Sizeof(cwindows.PROCESSENTRY32{})),
	}

	modifyCurrentProcessEntry := func(processEntry *cwindows.PROCESSENTRY32) {
		processName := cwindows.WCHARToString(processEntry.ExeFile[:])
		process, found := settings.GetTargetProcess(processName)
		if found {
			modifyProcess(processEntry.ProcessID, process)
		}
	}

	result, err := cwindows.Kernel32Process32First(handle, &processEntry)
	if result == cwindows.FALSE {
		return fmt.Errorf("failed to get first process: %v", err)
	}
	modifyCurrentProcessEntry(&processEntry)

	for {
		processEntry.Size = cwindows.DWORD(unsafe.Sizeof(cwindows.PROCESSENTRY32{}))
		nextResult, err := cwindows.Kernel32Process32Next(handle, &processEntry)
		if nextResult == cwindows.FALSE {
			if err == syscall.Errno(cwindows.ERROR_NO_MORE_FILES) {
				break
			}
			return fmt.Errorf("failed to get next process: %v", err)
		}
		modifyCurrentProcessEntry(&processEntry)
	}

	return nil
}
