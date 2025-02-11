package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/K4rian/kfdsl/cmd"
	"github.com/K4rian/kfdsl/internal/services/kfserver"
	"github.com/K4rian/kfdsl/internal/settings"
)

func main() {
	// Build the root command and execute it
	rootCmd := cmd.BuildRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	sett := settings.Get()

	// Print all settings
	sett.Print()

	// Start SteamCMD, if enabled
	if !sett.NoSteam.Value() {
		if err := startSteamCMD(sett, ctx); err != nil {
			fmt.Printf("ERROR: [SteamCMD]: %v\n", err)
			os.Exit(1)
		}
	}

	var server *kfserver.KFServer
	var err error

	defer func() {
		fmt.Println("\nShutting down the server...")
		cancel()

		if server != nil && server.IsRunning() {
			server.Wait()
		}

		fmt.Println("The server has been stopped.")
	}()

	// Start the Killing Floor Dedicated Server
	server, err = startGameServer(sett, ctx)
	if err != nil {
		fmt.Printf("ERROR: [KFServer]: %v\n", err)
		return
	}
	fmt.Printf("> Killing Floor Server started from '%s'\n", server.RootDirectory())

	<-signalChan
}
