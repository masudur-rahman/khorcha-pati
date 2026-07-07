package web

import (
	"net/http"
	"time"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter creates a chi router with auth endpoints and CORS.
func NewRouter(jwtSecret, corsOrigin string) chi.Router {
	r := chi.NewRouter()
	r.Use(RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware(corsOrigin))

	r.Get("/healthz", HandleHealthz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/request-otp", HandleRequestOTP)
		r.Post("/auth/verify-otp", HandleVerifyOTP)
		r.Post("/auth/qr/init", HandleQRInit)
		r.Get("/auth/qr/status", HandleQRPoll)
		r.Get("/auth/qr/redirect", HandleQRRedirect)
		r.Post("/auth/magic-link", HandleVerifyMagicLink)
		r.Post("/auth/refresh", HandleRefresh)
		r.Post("/auth/logout", HandleLogout)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(JWTAuth(jwtSecret))

			r.Get("/transactions", HandleListTransactions)
			r.Post("/transactions", HandleCreateTransaction)
			r.Put("/transactions/{id}", HandleUpdateTransaction)
			r.Delete("/transactions/{id}", HandleDeleteTransaction)

			r.Get("/wallets", HandleListWallets)
			r.Post("/wallets", HandleCreateWallet)
			r.Get("/contacts", HandleListContacts)
			r.Post("/contacts", HandleCreateContact)

			r.Get("/budgets", HandleListBudgets)
			r.Post("/budgets", HandleSetBudget)
			r.Delete("/budgets/{categoryID}", HandleDeleteBudget)
			r.Get("/budgets/alerts", HandleBudgetAlerts)

			r.Get("/summary/charts", HandleChartData)
			r.Get("/summary/report", HandleGetReport)
			r.Get("/summary/report-data", HandleGetReportData)
			r.Get("/categories", HandleListCategories)
			r.Get("/subcategories", HandleListSubcategories)
			r.Get("/profile", HandleGetProfile)
			r.Put("/profile", HandleUpdateProfile)

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(AdminAuth)
				r.Get("/admin/stats", HandleAdminStats)
				r.Get("/admin/users", HandleAdminUsers)
				r.Get("/admin/users/{id}", HandleAdminUserDetail)
				r.Patch("/admin/users/{id}/activate", HandleAdminSetUserActive)

				r.Get("/admin/ai-cache", HandleAdminListAICache)
				r.Post("/admin/ai-cache", HandleAdminCreateAICache)
				r.Put("/admin/ai-cache/{id}", HandleAdminUpdateAICache)
				r.Delete("/admin/ai-cache/{id}", HandleAdminDeleteAICache)
			})
		})
	})

	return r
}

// RequestLogger is a middleware that logs the start and end of each request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		logr.DefaultLogger.Infow("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start),
			"remote", r.RemoteAddr,
		)
	})
}

func corsMiddleware(origin string) func(next http.Handler) http.Handler {
	origins := []string{"http://localhost:5173", "http://localhost:3000"}
	if origin != "" && origin != "*" {
		origins = append(origins, origin)
	}
	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
