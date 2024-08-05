package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"golang.org/x/exp/rand"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var spinners = []spinner.Type{
	spinner.Line,
	spinner.Dots,
	spinner.MiniDot,
	spinner.Jump,
	spinner.Points,
	spinner.Pulse,
	spinner.Globe,
	spinner.Moon,
	spinner.Monkey,
	spinner.Meter,
	spinner.Hamburger,
	spinner.Ellipsis,
}

var SpinnerStyle = lipgloss.
	NewStyle().
	Padding(0, 1).
	Foreground(lipgloss.Color(lipgloss.Color(catppuccin.Mocha.Green().Hex)))

var SpinnerTitleStyle = lipgloss.
	NewStyle().
	Padding(0, 1).
	Margin(0, 1).
	Foreground(lipgloss.Color(catppuccin.Latte.Base().Hex)).
	Background(lipgloss.Color(catppuccin.Latte.Green().Hex))

var bold = lipgloss.
	NewStyle().
	Bold(true)

var FaintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(catppuccin.Mocha.Overlay0().Hex))

func keyvals(m map[string]string) []interface{} {
	var keyvals []interface{}
	for k, v := range m {
		keyvals = append(keyvals, k, v)
	}
	return keyvals
}

func DecodeUTF16(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", fmt.Errorf("must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}

func compile(target, logfile string) (outputStr string, status int) {
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

type Info struct {
	ScriptName    string
	Type          string
	Message       string
	FileName      string
	ElapsedTime   string
	Line          int
	Char          int
	Code          int
	TotalErrors   int
	TotalWarnings int
}

var readFileCache map[string][]string

func CenterString(str string, width int, color string) string {
	spaces := int(float64(width-len(str)) / 2)
	fg := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	return fg.
		Render("╭"+strings.Repeat("─", spaces)) +
		lipgloss.
			NewStyle().
			Bold(true).
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("#000000")).
			Render(str) +
		fg.
			Render(strings.Repeat("─", width-(spaces+len(str)))+"╮")
}

func main() {
	target := os.Args[1]

	readFileCache = make(map[string][]string)

	// pwd, err := os.Getwd()
	// if err != nil {
	// 	panic(err)
	// }

	targetPath := strings.Split(target, "/")
	logfile := strings.Split(targetPath[len(targetPath)-1], ".")[0] + ".log"

	// path := strings.Split(pwd, "/")
	lang := "MQL4"
	broker := "MetaQuotes"
	platform := "MetaTrader"
	// broker := strings.Join(strings.Split(path[len(path)-2], " ")[0:2], " ")
	// platform := strings.Join(strings.Split(path[len(path)-2], " ")[1:], " ")

	styles := log.DefaultStyles()

	styles.Prefix = FaintStyle.Bold(true)
	styles.Caller = FaintStyle
	styles.Key = FaintStyle
	styles.Separator = FaintStyle
	styles.Value = lipgloss.NewStyle().Foreground(lipgloss.Color(catppuccin.Mocha.Sapphire().Hex))

	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(catppuccin.Latte.Teal().Hex)).
		Foreground(lipgloss.Color(catppuccin.Latte.Base().Hex))

	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(catppuccin.Latte.Yellow().Hex)).
		Foreground(lipgloss.Color(catppuccin.Latte.Base().Hex))

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(catppuccin.Latte.Red().Hex)).
		Foreground(lipgloss.Color(catppuccin.Latte.Base().Hex))

	compileTarget := map[string]string{
		"target":   target,
		"Broker":   broker,
		"Platform": platform,
		"Language": lang,
	}

	logger := log.New(os.Stderr)
	logger.SetStyles(styles)

	fmt.Println()
	logger.Info("Compiling", keyvals(compileTarget)...)
	fmt.Println()

	rand.Seed(uint64(time.Now().Nanosecond()))
	randomSpinner := spinners[rand.Intn(len(spinners))]

	var outputStr string
	var status int

	runCompileCmd := func() {
		outputStr, status = compile(target, logfile)
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

	if status == 0 {
		log.Print(bold.
			Foreground(
				lipgloss.Color(catppuccin.Mocha.Green().Hex),
			).
			Render("Compilation successful!"),
		)
	} else {
		log.Print(bold.
			Foreground(
				lipgloss.Color(catppuccin.Mocha.Red().Hex),
			).
			Render("Failed to compile"),
		)
	}

	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	var infos []Info
	var totalErrors, totalWarnings int
	var elapsedTime string

	for scanner.Scan() {
		line := scanner.Text()

		// if empty line, skip
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "Result:") {
			result := strings.Split(line, ",")
			fmt.Sscanf(result[0], "Result: %d errors", &totalErrors)
			fmt.Sscanf(result[1], "%d warnings", &totalWarnings)
			if len(result) > 2 {
				fmt.Sscanf(result[2], "%s elapsed", &elapsedTime)
			} else {
				elapsedTime = "0"
			}
			continue
		}

		info := Info{}
		if strings.Contains(line, "information:") {
			parts := strings.Split(line, ": information: ")
			info.FileName = strings.TrimSpace(parts[0])
			info.Type = "information"
			info.Message = strings.TrimSpace(parts[1])
		} else {
			re := regexp.MustCompile(`^(.*)\((\d+),(\d+)\) : (\w+) (\d+): (.*)$`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 7 {
				info.ScriptName = matches[1]
				fmt.Sscanf(matches[2], "%d", &info.Line)
				fmt.Sscanf(matches[3], "%d", &info.Char)
				info.Type = matches[4]
				fmt.Sscanf(matches[5], "%d", &info.Code)
				info.Message = matches[6]
			}
		}
		infos = append(infos, info)
	}

	/*
	   ╭──── Script: Scripts/pocketbase.mq4, Line: 191, Char: 20, Type: warning, Code: 43 ───╮
	   │                                                                                     │
	   │                                                                                     │
	   │  191 |     JWT_EXPIRY = json["exp"].ToInt();                                        │
	   │                       │                                                             │
	   │                       ╰─➤ possible loss of data due to type conversion              │
	   ╰─────────────────────────────────────────────────────────────────────────────────────╯
	*/

	fmt.Println()
	for _, info := range infos {

		if info.Type == "" {
			continue
		}

		if info.Type == "information" {
			fileName := strings.ReplaceAll(info.FileName, "\\", "/")
			logger.Info(cases.Title(language.English).String(strings.Split(info.Message, " ")[0]), "Script", fileName)
			fmt.Println()
		} else {

			fileName := strings.ReplaceAll(info.ScriptName, "\\", "/")
			header := fmt.Sprintf(
				" Script: %s | Char: %d | Type: %s | Code: %d ",
				info.ScriptName,
				info.Char,
				info.Type,
				info.Code,
			)

			// check if the file is already in the cache
			if readFileCache[fileName] == nil {
				// if not, read the file and save it into the cache
				fileContents, err := os.ReadFile(fileName)
				if err != nil {
					panic(err)
				}

				// add new entry to map
				readFileCache[fileName] = strings.Split(string(fileContents), "\n")
			}

			diagnosticLine := readFileCache[fileName][info.Line-1]

			// count number of spaces at the beginning of the line
			var leadingSpaces int
			for i := 0; i < len(diagnosticLine); i++ {
				if diagnosticLine[i] == ' ' {
					leadingSpaces++
				} else {
					break
				}
			}

			leadingSpaces = leadingSpaces - 1

			if leadingSpaces < 0 {
				leadingSpaces = 0
			}

			trimmedDiagnosticLine := diagnosticLine[leadingSpaces:]

			var chunkStart int

			if len(trimmedDiagnosticLine) > 69 {
				// where does the char fall

				chunkIndex := info.Char / 69
				chunkStart = chunkIndex * 69
				chunkEnd := chunkStart + 69
				chunkEnd = min(chunkEnd, len(trimmedDiagnosticLine))

				if (chunkEnd - chunkStart) < 69 {
					chunkStart = len(trimmedDiagnosticLine) - 69
					chunkStart = max(chunkStart, 0)
				}

				trimmedDiagnosticLine = trimmedDiagnosticLine[chunkStart:chunkEnd]

			}

			color := catppuccin.Mocha.Yellow().Hex
			if info.Type == "error" {
				color = catppuccin.Mocha.Red().Hex
			}

			repeatCount := info.Char - leadingSpaces - chunkStart + 1

			repeatCount = max(repeatCount, 1)

			codeBlock := "```cpp\n" +
				trimmedDiagnosticLine +
				"\n```\n" +
				strings.Repeat(" ", repeatCount) + "  │" +
				"\n" + strings.
				Repeat(" ", repeatCount) +
				"  ╰─➤ " +
				lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(info.Message)

			out, _ := glamour.Render(codeBlock, "dracula")

			outSplit := strings.Split(out, "\n")

			outSplit = append(outSplit[1:3], outSplit[4:]...)
			out = strings.Join(outSplit, "\n")
			out = lipgloss.
				NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(color)).
				Render(
					FaintStyle.Render(
						lipgloss.JoinHorizontal(
							lipgloss.Top,
							FaintStyle.Render(fmt.Sprintf("\n%5d", info.Line))+"\n",
							FaintStyle.Render("\n |\n"),
							strings.TrimRight(out, "\n"),
						),
					),
				)

			outSplit = strings.Split(out, "\n")

			out = strings.Join(outSplit[1:], "\n")

			outWidth := lipgloss.Width(out)

			centeredHeader := CenterString(header, outWidth-2, color)

			out = centeredHeader + "\n" + out

			fmt.Println(out)
		}
	}

	if totalWarnings > 0 {
		logger.Warn("Warnings", "Total", totalWarnings)
		fmt.Println()
	}

	if totalErrors > 0 {
		logger.Error("Errors", "Total", totalErrors)
		fmt.Println()
	}
	logger.Info("Elapsed Time", "ms", elapsedTime)
}
