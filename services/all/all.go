package all

import (
	"strings"
	"sync"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	authmod "github.com/masudur-rahman/khorcha-pati/modules/auth"
	authrepo "github.com/masudur-rahman/khorcha-pati/repos/auth"
	"github.com/masudur-rahman/khorcha-pati/repos/budgets"
	"github.com/masudur-rahman/khorcha-pati/repos/event"
	"github.com/masudur-rahman/khorcha-pati/repos/transaction"
	"github.com/masudur-rahman/khorcha-pati/repos/user"
	"github.com/masudur-rahman/khorcha-pati/repos/wallets"
	"github.com/masudur-rahman/khorcha-pati/services"
	authsvc "github.com/masudur-rahman/khorcha-pati/services/auth"
	budgetsvc "github.com/masudur-rahman/khorcha-pati/services/budgets"
	eventsvc "github.com/masudur-rahman/khorcha-pati/services/event"
	summarysvc "github.com/masudur-rahman/khorcha-pati/services/summary"
	txnsvc "github.com/masudur-rahman/khorcha-pati/services/transaction"
	usersvc "github.com/masudur-rahman/khorcha-pati/services/user"
	walletsvc "github.com/masudur-rahman/khorcha-pati/services/wallets"

	"github.com/masudur-rahman/styx"
)

type Services struct {
	User    services.ProfileService
	Wallet  services.WalletService
	Contact services.ContactService
	Txn     services.TransactionService
	Event   services.EventService
	Budget  services.BudgetService
	Summary services.SummaryService
	Auth    services.AuthService
}

var (
	svc   *Services
	svcMu sync.RWMutex
)

// webConfig stores web service init params for re-initialization on DB reconnect.
type webConfig struct {
	messenger     authmod.Messenger
	jwtSecret     string
	refreshSecret string
	botUsername   string
	baseURL       string
}

var webCfg *webConfig

func GetServices() *Services {
	svcMu.RLock()
	defer svcMu.RUnlock()
	return svc
}

func InitiateSQLServices(uow styx.UnitOfWork, logger logr.Logger) {
	userRepo := user.NewSQLUserRepository(uow.SQL, logger)
	walletRepo := wallets.NewSQLWalletRepository(uow.SQL, logger)
	contactRepo := user.NewSQLContactRepository(uow.SQL, logger)
	txnRepo := transaction.NewSQLTransactionRepository(uow.SQL, logger)
	eventRepo := event.NewSQLEventRepository(uow.SQL, logger)
	budgetRepo := budgets.NewSQLBudgetRepository(uow.SQL, logger)

	userSvc := usersvc.NewProfileService(userRepo)
	walletSvc := walletsvc.NewWalletService(walletRepo)
	contactSvc := usersvc.NewContactService(contactRepo)
	txnSvc := txnsvc.NewTxnService(uow, walletRepo, contactRepo, txnRepo, eventRepo)
	eventSvc := eventsvc.NewEventService(eventRepo)
	budgetSvc := budgetsvc.NewBudgetService(budgetRepo, txnRepo)
	summarySvc := summarysvc.NewSummaryService(txnRepo, walletRepo, budgetRepo)

	// Preserve Auth service across DB reconnects
	var existingAuth services.AuthService
	svcMu.RLock()
	if svc != nil {
		existingAuth = svc.Auth
	}
	svcMu.RUnlock()

	newSvc := &Services{
		User:    userSvc,
		Wallet:  walletSvc,
		Contact: contactSvc,
		Txn:     txnSvc,
		Event:   eventSvc,
		Budget:  budgetSvc,
		Summary: summarySvc,
		Auth:    existingAuth,
	}

	// Re-initialize auth with new repos if web services were configured
	if webCfg != nil {
		ar := authrepo.NewSQLAuthRepository(uow.SQL, logger)
		newSvc.Auth = authsvc.NewAuthService(
			userRepo, ar, webCfg.messenger,
			webCfg.jwtSecret, webCfg.refreshSecret, webCfg.botUsername, webCfg.baseURL,
			logger,
		)
	}

	svcMu.Lock()
	svc = newSvc
	svcMu.Unlock()
}

// InitiateWebServices wires the auth service when the web dashboard is enabled.
func InitiateWebServices(
	messenger authmod.Messenger,
	jwtSecret, refreshSecret, botUsername, baseURL string,
	uow styx.UnitOfWork,
	logger logr.Logger,
) {
	botUsername = strings.TrimSpace(strings.TrimPrefix(botUsername, "@"))
	webCfg = &webConfig{
		messenger:     messenger,
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		botUsername:   botUsername,
		baseURL:       baseURL,
	}

	userRepo := user.NewSQLUserRepository(uow.SQL, logger)
	ar := authrepo.NewSQLAuthRepository(uow.SQL, logger)

	svcMu.Lock()
	svc.Auth = authsvc.NewAuthService(
		userRepo, ar, messenger,
		jwtSecret, refreshSecret, botUsername, baseURL,
		logger,
	)
	svcMu.Unlock()
}
