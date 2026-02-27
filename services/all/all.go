package all

import (
	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/repos/wallets"
	"github.com/masudur-rahman/expense-tracker-bot/repos/event"
	"github.com/masudur-rahman/expense-tracker-bot/repos/transaction"
	"github.com/masudur-rahman/expense-tracker-bot/repos/user"
	"github.com/masudur-rahman/expense-tracker-bot/services"
	walletsvc "github.com/masudur-rahman/expense-tracker-bot/services/wallets"
	eventsvc "github.com/masudur-rahman/expense-tracker-bot/services/event"
	txnsvc "github.com/masudur-rahman/expense-tracker-bot/services/transaction"
	usersvc "github.com/masudur-rahman/expense-tracker-bot/services/user"

	"github.com/masudur-rahman/styx"
)

type Services struct {
	User    services.ProfileService
	Wallet  services.WalletService
	Contact services.ContactService
	Txn     services.TransactionService
	Event   services.EventService
}

var svc *Services

func GetServices() *Services {
	return svc
}

func InitiateSQLServices(uow styx.UnitOfWork, logger logr.Logger) {
	userRepo := user.NewSQLUserRepository(uow.SQL, logger)
	walletRepo := wallets.NewSQLWalletRepository(uow.SQL, logger)
	contactRepo := user.NewSQLContactRepository(uow.SQL, logger)
	txnRepo := transaction.NewSQLTransactionRepository(uow.SQL, logger)
	eventRepo := event.NewSQLEventRepository(uow.SQL, logger)

	userSvc := usersvc.NewProfileService(userRepo)
	walletSvc := walletsvc.NewWalletService(walletRepo)
	contactSvc := usersvc.NewContactService(contactRepo)
	txnSvc := txnsvc.NewTxnService(uow, walletRepo, contactRepo, txnRepo, eventRepo)
	eventSvc := eventsvc.NewEventService(eventRepo)

	svc = &Services{
		User:    userSvc,
		Wallet:  walletSvc,
		Contact: contactSvc,
		Txn:     txnSvc,
		Event:   eventSvc,
	}
}
