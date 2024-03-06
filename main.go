package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	file = flag.String("file", "", "Log file")

	// [1] - time
	// [2] - caller
	// [3] - line no
	// [4] - message
	linePat = regexp.MustCompile(` *(\w+ \d+, \d+ @ \d+:\d+:\d+\.\d+)\s+([\w\.]+)\s+(\d+)\s+:\s+(.*)`)
	timePat = `Jan 2, 2006 15:04:05.000`

	maxMsgLength = 500
	minTime      = 0
)

func main() {
	flag.Parse()
	if *file == "" {
		print("file is required")
		return
	}

	buf, err := os.ReadFile(*file)
	if err != nil {
		panic(fmt.Errorf("failed to read file %v, %v", *file, err))
	}

	ctn := string(buf)
	lines := strings.Split(ctn, "\n")
	logLines := make([]LogLine, 0, 100)
	for _, l := range lines {
		if l == "" {
			continue
		}
		ll, err := parseLogLine(l)
		if err != nil {
			continue
		}
		logLines = append(logLines, ll)
	}

	first := LogLine{}
	last := LogLine{}
	var max int64 = 0
	var prev LogLine
	for i := range logLines {
		l := logLines[i]
		if i < 1 {
			first = l
			prev = l
		} else {
			if i == len(logLines)-1 {
				last = l
			}
			gap := prev.Time.Sub(l.Time).Milliseconds()
			if gap > max {
				max = gap
			}
			prev = l
		}
	}
	total := first.Time.Sub(last.Time).Milliseconds()

	prev = LogLine{}
	for i := range logLines {
		l := logLines[i]
		if i < 1 {
			prev = l
		} else {
			gap := prev.Time.Sub(l.Time).Milliseconds()
			if gap < int64(minTime) {
				prev = l
				continue
			}

			calc := int((float64(gap) / float64(max)) * 30)
			pad := strings.Repeat("-", calc)
			fmt.Printf("%v|\n", pad)
			fmt.Printf("%v| > took: %vms (%.2f%%)\n", pad, gap, float64(gap)/float64(total)*100)
			fmt.Printf("%v|\n", pad)
			prev = l
		}
		fmt.Printf("%v, %v %v: %v\n", l.Time.String(), l.Caller, l.LineNo, l.Message)
	}
	fmt.Printf("\n\nTotal: %vms\n\n", total)
}

func parseLogLine(line string) (LogLine, error) {
	matches := linePat.FindStringSubmatch(line)
	if matches == nil {
		return LogLine{}, fmt.Errorf("doesn't match pattern")
	}

	matches[1] = strings.ReplaceAll(matches[1], " @ ", " ")
	time, ep := time.ParseInLocation(timePat, matches[1], time.Local)
	if ep != nil {
		return LogLine{}, fmt.Errorf("time format illegal, %v", ep)
	}

	msg := matches[4]
	msgRu := []rune(msg)
	if len(msgRu) > maxMsgLength {
		msg = string(msgRu[:maxMsgLength+1])
	}

	return LogLine{
		Time:    time,
		Caller:  matches[2],
		LineNo:  matches[3],
		Message: msg,
	}, nil
}

type LogLine struct {
	Time    time.Time
	Caller  string
	LineNo  string
	Message string
}
