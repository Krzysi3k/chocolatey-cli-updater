package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/theckman/yacspin"
)

func main() {
	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[11],
		Suffix:          " choco",
		SuffixAutoColon: true,
		Message:         "checking new updates...",
		StopCharacter:   "✓",
		StopColors:      []string{"fgGreen"},
	}

	ignoredFullPath := flag.String("d", "None", "directory path to ignored packages file")
	flag.Parse()

	spinner, err := yacspin.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if *ignoredFullPath == "None" {
		log.Fatal("please provide full path to ignored packages\n")
	}

	spinner.Start()
	apps := checkAppUpdates(ignoredFullPath)
	spinner.Stop()
	if apps == nil {
		color.HiGreen("there is nothing to upgrade.")
		return
	}
	updateApps(apps)
}

func checkAppUpdates(ignoredFullPath *string) []string {
	cmd := exec.Command("choco", "outdated")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	ignored := readIgnoredPackages(ignoredFullPath)
	var apps []string
	output := string(out)
	lines := strings.Split(output, "\n")
	for _, v := range lines {
		if strings.Contains(v, `|`) && !strings.Contains(v, "Output is package name") {
			if !isInSlice(v, ignored) {
				apps = append(apps, v)
			}
		}
	}
	return apps
}

func readIgnoredPackages(ignoredFullPath *string) []string {
	ignoredLines := make([]string, 0, 10)
	file, err := os.Open(*ignoredFullPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ignoredLines = append(ignoredLines, scanner.Text())
	}
	return ignoredLines
}

func isInSlice(line string, lines []string) bool {
	for _, l := range lines {
		if strings.Contains(line, l) {
			return true
		}
	}
	return false
}

func updateApps(apps []string) {
	apps = append(apps, "exit")
	responses := []string{}
	prompt := &survey.MultiSelect{
		Message:  "select app to run upgrade:",
		Options:  apps,
		PageSize: 20,
	}
	survey.AskOne(prompt, &responses)
	for _, i := range responses {
		if i == "exit" {
			fmt.Println("upgrade aborted")
			return
		}
	}
	color.HiMagenta("apps to be upgraded:")
	for _, i := range responses {
		color.HiGreen(i)
	}

	for _, i := range responses {
		name := strings.Split(i, `|`)[0]
		cmd := exec.Command("choco", "upgrade", "-y", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}
}
