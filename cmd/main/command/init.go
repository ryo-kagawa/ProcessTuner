package command

import "github.com/ryo-kagawa/go-cwindows/cwindows"

func init() {
	cwindows.InitKernel32CloseHandle()
	cwindows.InitKernel32CreateToolhelp32Snapshot()
	cwindows.InitKernel32OpenProcess()
	cwindows.InitKernel32Process32First()
	cwindows.InitKernel32Process32Next()
	cwindows.InitKernel32SetPriorityClass()
	cwindows.InitKernel32SetProcessAffinityMask()
	cwindows.InitSechostControlTrace()
	cwindows.InitSechostEnableTraceEx2()
	cwindows.InitSechostOpenTraceFromRealTimeLogger()
	cwindows.InitSechostProcessTrace()
	cwindows.InitSechostStartTrace()
}
