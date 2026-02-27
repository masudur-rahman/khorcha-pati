package handlers

import (
	"gopkg.in/telebot.v3"
)

// HandleUndo reverses the most recent active transaction for the calling user.
// It soft-deletes the transaction and reverses any wallet / contact balance
// changes that were applied when it was originally created.
//
// TODO: Implement TransactionService.Undo(userID int64) (*models.Transaction, error)
// that soft-deletes the last active transaction and reverts wallet/contact balances.
// See refactor guide §4.1c for the recommended compensating-action pattern.
//
// Register in api/tele.go:
//
//	bot.Handle("/undo", HandleUndo)
func HandleUndo(c telebot.Context) error {
	// TODO: call all.GetServices().Transaction.Undo(int64(c.Sender().ID))
	// once the Undo method is implemented on TransactionService.
	return c.Send("⚠️ Undo is not yet implemented. Please implement TransactionService.Undo() first.")
}
