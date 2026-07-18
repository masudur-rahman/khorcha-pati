package all

import (
	"strings"
	"sync"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	authmod "github.com/masudur-rahman/khorcha-pati/modules/auth"
	accessrepo "github.com/masudur-rahman/khorcha-pati/repos/access"
	authrepo "github.com/masudur-rahman/khorcha-pati/repos/auth"
	"github.com/masudur-rahman/khorcha-pati/repos/budgets"
	"github.com/masudur-rahman/khorcha-pati/repos/event"
	"github.com/masudur-rahman/khorcha-pati/repos/transaction"
	"github.com/masudur-rahman/khorcha-pati/repos/user"
	"github.com/masudur-rahman/khorcha-pati/repos/wallets"
	"github.com/masudur-rahman/khorcha-pati/services"
	accesssvc "github.com/masudur-rahman/khorcha-pati/services/access"
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
	Access  services.AccessService
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

// GetMessenger returns the auth Messenger instance.
func GetMessenger() authmod.Messenger {
	if webCfg == nil {
		return nil
	}
	return webCfg.messenger
}

// BotUsername returns the effective bot username (config override or live bot identity).
func BotUsername() string {
	if webCfg == nil {
		return ""
	}
	return webCfg.botUsername
}

// makeAccessCheck builds the web-login gate: restricted instances reject
// non-allowed users with the admin-configured redirect text.
func makeAccessCheck(acc services.AccessService) func(*models.Profile) error {
	return func(p *models.Profile) error {
		if !acc.IsRestricted() || acc.IsUserAllowed(p.Username, p.TelegramID) {
			return nil
		}
		return models.StatusError{Status: 403, Message: acc.RestrictedReplyText()}
	}
}

func InitiateSQLServices(uow styx.UnitOfWork, logger logr.Logger) {
	userRepo := user.NewSQLUserRepository(uow.SQL, logger)
	walletRepo := wallets.NewSQLWalletRepository(uow.SQL, logger)
	contactRepo := user.NewSQLContactRepository(uow.SQL, logger)
	txnRepo := transaction.NewSQLTransactionRepository(uow.SQL, logger)
	eventRepo := event.NewSQLEventRepository(uow.SQL, logger)
	budgetRepo := budgets.NewSQLBudgetRepository(uow.SQL, logger)

	userSvc := usersvc.NewProfileService(userRepo)
	walletSvc := walletsvc.NewWalletService(uow, walletRepo, txnRepo)
	contactSvc := usersvc.NewContactService(uow, contactRepo, txnRepo)
	txnSvc := txnsvc.NewTxnService(uow, walletRepo, contactRepo, txnRepo, eventRepo)
	eventSvc := eventsvc.NewEventService(eventRepo)
	budgetSvc := budgetsvc.NewBudgetService(budgetRepo, txnRepo)
	summarySvc := summarysvc.NewSummaryService(txnRepo, walletRepo, budgetRepo)
	accessSvc := accesssvc.NewAccessService(accessrepo.NewSQLAccessRepository(uow.SQL, logger), logger)

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
		Access:  accessSvc,
	}

	// Re-initialize auth with new repos if web services were configured
	if webCfg != nil {
		ar := authrepo.NewSQLAuthRepository(uow.SQL, logger)
		a := authsvc.NewAuthService(
			userRepo, ar, webCfg.messenger,
			webCfg.jwtSecret, webCfg.refreshSecret, webCfg.botUsername, webCfg.baseURL,
			logger,
		)
		a.SetAccessCheck(makeAccessCheck(accessSvc))
		newSvc.Auth = a
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
	a := authsvc.NewAuthService(
		userRepo, ar, messenger,
		jwtSecret, refreshSecret, botUsername, baseURL,
		logger,
	)
	if svc.Access != nil {
		a.SetAccessCheck(makeAccessCheck(svc.Access))
	}
	svc.Auth = a
	svcMu.Unlock()
}
