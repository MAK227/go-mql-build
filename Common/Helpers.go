package Common

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	ini "gopkg.in/ini.v1"
)

type Info struct {
	ScriptName string
	Type       string
	Message    string
	FileName   string
	Line       int
	Char       int
	Code       int
}

type Diagnostic struct {
	info          []Info
	totalErrors   int
	totalWarnings int
	elapsedTime   string
}

var Spinners = []spinner.Type{
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

var Bold = lipgloss.
	NewStyle().
	Bold(true)

var FaintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(catppuccin.Mocha.Overlay0().Hex))

func Keyvals(m map[string]string) []interface{} {
	var keyvals []interface{}
	for k, v := range m {
		keyvals = append(keyvals, k, v)
	}
	return keyvals
}

var Logger *log.Logger = log.New(os.Stderr)

func InitLogger() {
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

	Logger.SetStyles(styles)
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

func ParseLogFile(outputStr string, status int, mode string) (diagnostics Diagnostic) {
	scanner := bufio.NewScanner(strings.NewReader(outputStr))

	for scanner.Scan() {
		line := scanner.Text()

		// if empty line, skip
		if strings.TrimSpace(line) == "" {
			continue
		}

		// : information: result 0 errors, 0 warnings, 18 msec elapsed

		if mode == "syntax" {
			if strings.HasPrefix(line, " : information: result") {
				result := strings.Split(line, ",")
				fmt.Sscanf(result[0], ": information: result %d errors", &diagnostics.totalErrors)
				fmt.Sscanf(result[1], "%d warnings", &diagnostics.totalWarnings)
				if len(result) > 2 {
					fmt.Sscanf(result[2], "%s elapsed", &diagnostics.elapsedTime)
				} else {
					diagnostics.elapsedTime = "0"
				}
				continue
			}
		} else if mode == "compile" {
			if strings.HasPrefix(line, "Result:") {
				result := strings.Split(line, ",")
				fmt.Sscanf(result[0], "Result: %d errors", &diagnostics.totalErrors)
				fmt.Sscanf(result[1], "%d warnings", &diagnostics.totalWarnings)
				if len(result) > 2 {
					fmt.Sscanf(result[2], "%s elapsed", &diagnostics.elapsedTime)
				} else {
					diagnostics.elapsedTime = "0"
				}
				continue
			}
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
		diagnostics.info = append(diagnostics.info, info)
	}

	succesMsg := "Compilation successful!"
	failMsg := "Failed to compile"

	if mode == "syntax" {
		succesMsg = "Syntax check successful!"
		failMsg = "Syntax check failed"
	}

	if status == 0 && diagnostics.totalErrors == 0 {
		log.Print(Bold.
			Foreground(
				lipgloss.Color(catppuccin.Mocha.Green().Hex),
			).
			Render(succesMsg),
		)
	} else {
		log.Print(Bold.
			Foreground(
				lipgloss.Color(catppuccin.Mocha.Red().Hex),
			).
			Render(failMsg),
		)
	}

	return diagnostics
}

func PrintDiagnostics(diagnostics Diagnostic, readFileCache map[string][]string) {
	fmt.Println()
	for _, info := range diagnostics.info {

		if info.Type == "" {
			continue
		}

		if info.Type == "information" {
			fileName := strings.ReplaceAll(info.FileName, "\\", "/")
			Logger.Info(cases.Title(language.English).String(strings.Split(info.Message, " ")[0]), "Script", fileName)
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

	if diagnostics.totalWarnings > 0 {
		Logger.Warn("Warnings", "Total", diagnostics.totalWarnings)
		fmt.Println()
	}

	if diagnostics.totalErrors > 0 {
		Logger.Error("Errors", "Total", diagnostics.totalErrors)
		fmt.Println()
	}
	Logger.Info("Elapsed Time", "ms", diagnostics.elapsedTime)
}

func BuildCompileTarget(target string) (compileTarget map[string]string, logfile string) {
	cfg, err := ini.Load("../config/terminal.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	targetPath := strings.Split(target, "/")
	logfile = strings.Split(targetPath[len(targetPath)-1], ".")[0] + ".log"

	lang := "MQL4"
	broker := cfg.Section("Settings").Key("LastScanServer").String()

	compileTarget = map[string]string{
		"target":   target,
		"Broker":   broker,
		"Language": lang,
	}

	return compileTarget, logfile
}
