package Common

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/rand"
)

func compileMetaEditor(target, logfile string) (outputStr string, status int) {
	// MT4 should be on portable mode

	// for linux and mac
	cmd := exec.Command("wine", "../metaeditor.exe", "/compile:"+target, "/log:"+logfile)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("../metaeditor.exe", "/compile:"+target, "/log:"+logfile)
	}

	// check the status of the command
	cmd.Run()

	// read the log file
	logFile, err := os.ReadFile(logfile)
	if err != nil {
		return "", 1
	}

	logFileUTF8, err := DecodeUTF16(logFile)
	if err != nil {
		return "", 1
	}

	return logFileUTF8, 0
}

func Compile(target string, logfile string, compileTarget map[string]string) (outputStr string, status int) {
	fmt.Println()
	Logger.Info("Compiling", Keyvals(compileTarget)...)
	fmt.Println()

	rand.Seed(uint64(time.Now().Nanosecond()))
	randomSpinner := Spinners[rand.Intn(len(Spinners))]

	runCompileCmd := func() {
		outputStr, status = compileMetaEditor(target, logfile)
	}
	err := spinner.New().
		Type(randomSpinner).
		Title(SpinnerStyle.
			Render(
				fmt.Sprintf(
					"Compiling %s",
					SpinnerTitleStyle.Render(target),
				),
			),
		).
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("#F780E2")).PaddingLeft(1)).
		Action(runCompileCmd).
		Run()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println()

	return outputStr, status
}
