package main

import (
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func tesseract(filename string) (string, error) {
	cmd := exec.Command("tesseract", filename, "stdout", "quiet")
	retvBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	retv := string(retvBytes)
	multiNewlinesRegex := regexp.MustCompile("[\r\n]+")
	multiSpacesRegex := regexp.MustCompile(" +")
	retv = multiNewlinesRegex.ReplaceAllString(retv, "\n")
	retv = multiSpacesRegex.ReplaceAllString(retv, " ")
	retv = strings.TrimSpace(retv)
	return retv, nil
}

// spellcheckLine returns the proportion (0.0 - 1.0) of words in the line that are
// either a number or a word local ispell considers correctly spelled, and also
// are >2 chars long.
func spellcheckLine(line string) (float64, error) {
	correctWords := 0
	words := strings.Fields(line)
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(runCtx, "ispell")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return 0, err
	}
	for _, word := range words {
		if _, err := strconv.Atoi(word); err == nil {
			correctWords += 1
			continue
		}
		if len(word) <= 2 {
			continue
		}
		stdin.Write([]byte(word + "\n"))
	}
	stdin.Close()
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	outLines := strings.Split(string(out), "\n")
	for i, line := range outLines {
		if i == 0 {
			continue
		}
		if line == "word: ok" {
			correctWords += 1
		}
	}
	return float64(correctWords) / float64(len(words)), nil
}
