package auth

// Messenger sends messages to a Telegram user by their Telegram ID.
type Messenger interface {
	SendMessage(telegramID int64, text string) error
}
