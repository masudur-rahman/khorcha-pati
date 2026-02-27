// Package telegram provides shared Telegram bot utilities.
package telegram

import (
	"fmt"
	"strings"
)

const MaxMessageLen = 4000 // safe margin below Telegram's hard 4096-byte limit

// SplitMessage splits a long string into chunks that each fit within
// MaxMessageLen, breaking only on newline boundaries.
// Use this before any bot.Send() call to avoid silent message truncation.
//
// Usage in a handler:
//
//	for _, chunk := range telegram.SplitMessage(text) {
//	    if err := c.Send(chunk, telebot.ModeMarkdown); err != nil {
//	        return err
//	    }
//	}
func SplitMessage(text string) []string {
	if len(text) <= MaxMessageLen {
		return []string{text}
	}
	var chunks []string
	var buf strings.Builder
	for _, line := range strings.Split(text, "\n") {
		// +1 for the newline we are about to add
		if buf.Len()+len(line)+1 > MaxMessageLen {
			if buf.Len() > 0 {
				chunks = append(chunks, buf.String())
				buf.Reset()
			}
		}
		buf.WriteString(line + "\n")
	}
	if buf.Len() > 0 {
		chunks = append(chunks, buf.String())
	}
	return chunks
}

// FormatAmount formats a float64 as a currency string with 2 decimal places.
func FormatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}
