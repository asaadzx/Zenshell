package prompt

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func HexToANSI(hex string) string {
	h := strings.TrimPrefix(hex, "#")
	if len(h) != 6 {
		return "\033[0m"
	}

	var r, g, b int
	if _, err := fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b); err != nil {
		return "\033[0m"
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func Format(format, user, host, cwd, colorHex string, lastExit int) string {
	if format == "" {
		format = "[%u@%h %d]$ "
	}

	now := time.Now()

	prompt := format
	prompt = strings.ReplaceAll(prompt, "%u", user)
	prompt = strings.ReplaceAll(prompt, "%h", host)
	prompt = strings.ReplaceAll(prompt, "%d", cwd)
	prompt = strings.ReplaceAll(prompt, "%t", now.Format("15:04"))
	prompt = strings.ReplaceAll(prompt, "%T", now.Format("15:04:05"))

	if lastExit == 0 {
		prompt = strings.ReplaceAll(prompt, "%?", "0")
	} else {
		prompt = strings.ReplaceAll(prompt, "%?", fmt.Sprintf("%d", lastExit))
	}

	root := "$"
	if os.Geteuid() == 0 {
		root = "#"
	}
	prompt = strings.ReplaceAll(prompt, "%$", root)

	color := HexToANSI(colorHex)
	return color + prompt + "\033[0m"
}
