package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/K4rian/kfdsl/cmd"
	"github.com/K4rian/kfdsl/internal/log"
	"github.com/K4rian/kfdsl/internal/services/kfserver"
	"github.com/K4rian/kfdsl/internal/settings"
)

func main() {
	// Build the root command and execute it
	rootCmd := cmd.BuildRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	// Get the settings
	sett := settings.Get()

	// Init the logger
	log.Init(
		sett.LogLevel.Value(),
		sett.LogFile.Value(),
		sett.LogFileFormat.Value(),
		sett.LogMaxSize.Value(),
		sett.LogMaxBackups.Value(),
		sett.LogMaxAge.Value(),
		sett.LogToFile.Value(),
	)

	// Create a cancel context
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Print all settings
	sett.Print()

	// Start SteamCMD, if enabled
	if !sett.NoSteam.Value() {
		if err := startSteamCMD(sett, ctx); err != nil {
			log.Logger.Error("SteamCMD raised an error", "err", err)
			os.Exit(1)
		}
	}

	var server *kfserver.KFServer

	defer func() {
		log.Logger.Info("Shutting down the KF Dedicated Server...")
		cancel()

		if server != nil && server.IsRunning() {
			if err := server.Wait(); err != nil {
				log.Logger.Error("KF Dedicated Server raised an error during shutdown", "err", err)
			}
		}
		log.Logger.Info("KF Dedicated Server has been stopped.")
	}()

	// Start the Killing Floor Dedicated Server
	server, err := startGameServer(sett, ctx)
	if err != nil {
		log.Logger.Error("KF Dedicated Server raised an error", "err", err)
		return
	}

	<-signalChan
	signal.Stop(signalChan)
}
