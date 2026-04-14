/*
Copyright © 2023 Masudur Rahman <masudjuly02@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/api"
	"github.com/masudur-rahman/expense-tracker-bot/api/web"
	"github.com/masudur-rahman/expense-tracker-bot/configs"
	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := configs.Validate(); err != nil {
			log.Fatalln(err)
		}
		if err := configs.InitiateDatabaseConnection(cmd.Context()); err != nil {
			log.Fatalln(err)
		}
		configs.InitiateCache()
		configs.LoadAICacheIntoMemory()

		bot, err := api.TeleBotRoutes()
		if err != nil {
			log.Fatalln(err)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		var webSrv *http.Server
		if cfg := configs.TrackerConfig.WebDashboard; cfg.Enabled {
			messenger := api.NewBotMessenger(bot)
			uow := configs.GetUnitOfWork()
			all.InitiateWebServices(messenger, cfg.JWTSecret, cfg.RefreshSecret, cfg.BotUsername, uow, logr.DefaultLogger)

			port := cfg.Port
			if port == "" {
				port = ":8081"
			}
			router := web.NewRouter(cfg.JWTSecret, cfg.CORSOrigin)
			webSrv = &http.Server{Addr: port, Handler: router, ReadHeaderTimeout: 10 * time.Second}
			go func() {
				log.Printf("Web dashboard started at %s", port)
				if err := webSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("Web server error: %v", err)
				}
			}()
		}

		healthSrv := startHealthz()
		go pingHealthzAPIPeriodically()
		log.Println("Expense Tracker Bot started")

		go bot.Start()

		<-ctx.Done()
		log.Println("Shutting down...")
		bot.Stop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if webSrv != nil {
			if err := webSrv.Shutdown(shutdownCtx); err != nil {
				log.Printf("Web server shutdown error: %v", err)
			}
		}
		if err := healthSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Health server shutdown error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func pingHealthzAPIPeriodically() {
	logger := logr.DefaultLogger
	baseURL, ok := os.LookupEnv("BASE_URL")
	if !ok {
		return
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalln(err)
	}
	u.Path = path.Join(u.Path, "healthz")
	healthPath := u.String()
	logger.Infow("Health url provided", "url", healthPath)

	t20 := time.NewTicker(20 * time.Minute)
	for range t20.C {
		resp, err := http.Get(healthPath) //nolint:gosec // health check URL from config
		if err != nil {
			logger.Errorw("healthz api failed", "error", err.Error())
		} else {
			data, err := io.ReadAll(resp.Body)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			logger.Infow("healthz api", "status", resp.StatusCode, "msg", string(data), "error", errMsg)
		}
	}
}

func startHealthz() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)

	srv := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	logr.DefaultLogger.Infow("Health checker started at :8080/healthz")
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}
	}()
	return srv
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]string{"status": "ok", "db": "ok"}

	if err := configs.PingDatabase(); err != nil {
		resp["status"] = "degraded"
		resp["db"] = err.Error()
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp) //nolint:errchkjson
}
