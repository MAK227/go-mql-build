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

func syntaxMetaEditor(target, logfile string, cfg *MQLConfig) (outputStr string, status int) {
	// MT4 should be on portable mode

	// for linux and mac

	cmd := exec.Command("wine", cfg.MetaEditorPath, "/compile:"+target, "/log:"+logfile, "/s")

	if runtime.GOOS == "windows" {
		cmd = exec.Command(cfg.MetaEditorPath, "/compile:"+target, "/log:"+logfile, "/s")
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

func SyntaxCheck(target string, logfile string, compileTarget map[string]string, cfg *MQLConfig) (outputStr string, status int) {
	fmt.Println()
	Logger.Info("Checking syntax", Keyvals(compileTarget)...)
	fmt.Println()

	rand.Seed(uint64(time.Now().Nanosecond()))
	randomSpinner := Spinners[rand.Intn(len(Spinners))]

	// var outputStr string
	// var status int

	runCompileCmd := func() {
		outputStr, status = syntaxMetaEditor(target, logfile, cfg)
	}
	err := spinner.New().
		Type(randomSpinner).
		Title(SpinnerStyle.
			Render(
				fmt.Sprintf(
					"Checking syntax %s",
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
