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
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/masudur-rahman/khorcha-pati/api"
	"github.com/masudur-rahman/khorcha-pati/api/web"
	"github.com/masudur-rahman/khorcha-pati/configs"
	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/services/all"

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
		configs.SeedAdminUser()

		bot, err := api.TeleBotRoutes()
		if err != nil {
			log.Fatalln(err)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		cfg := configs.TrackerConfig.Server
		messenger := api.NewBotMessenger(bot)
		uow := configs.GetUnitOfWork()
		botUsername := cfg.BotUsername
		if botUsername == "" && bot.Me != nil {
			botUsername = bot.Me.Username
		}
		all.InitiateWebServices(messenger, cfg.JWTSecret, cfg.RefreshSecret, botUsername, cfg.BaseURL, uow, logr.DefaultLogger)

		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		router := web.NewRouter(cfg.JWTSecret, cfg.CORSOrigin)
		webSrv := &http.Server{Addr: addr, Handler: router, ReadHeaderTimeout: 10 * time.Second}

		go func() {
			log.Printf("Backend server started at %s", addr)
			if err := webSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Backend server error: %v", err)
			}
		}()

		go pingHealthzAPIPeriodically(addr)
		log.Println("Khorcha-Pati started")

		go bot.Start()

		<-ctx.Done()
		log.Println("Shutting down...")
		bot.Stop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := webSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Web server shutdown error: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func pingHealthzAPIPeriodically(addr string) {
	logger := logr.DefaultLogger
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Fallback to local address if BASE_URL is not set
		baseURL = "http://" + addr
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		logger.Errorw("failed to parse base url for health check", "error", err.Error())
		return
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
