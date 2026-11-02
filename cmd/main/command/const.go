package command

import (
	"github.com/ryo-kagawa/go-cwindows/cwindows"
	"golang.org/x/sys/windows"
)

const (
	sessionName = "ProcessTuner"
)

var (
	sessionNameLPWSTR cwindows.LPWSTR
	sessionNameWCHAR  []cwindows.WCHAR
)

func init() {
	sessionNameUTF16Ptr, _ := windows.UTF16PtrFromString(sessionName)
	sessionNameLPWSTR = cwindows.LPWSTR((*cwindows.WCHAR)(sessionNameUTF16Ptr))
	sessionNameUTF16, _ := windows.UTF16FromString(sessionName)
	sessionNameWCHAR = cwindows.UTF16ToWCHAR(sessionNameUTF16)
}
