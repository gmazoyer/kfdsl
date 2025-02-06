package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/K4rian/kfdsl/cmd"
	"github.com/K4rian/kfdsl/internal/services/kfserver"
	"github.com/K4rian/kfdsl/internal/services/redirectserver"
	"github.com/K4rian/kfdsl/internal/settings"
)

func shutdown(cancel context.CancelFunc, redirectServer *redirectserver.HTTPRedirectServer, gameServer *kfserver.KFServer) {
	fmt.Println("\nShutting down all services...")
	cancel()

	var wg sync.WaitGroup

	if redirectServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			redirectServer.Stop()
		}()
	}

	if gameServer != nil && gameServer.IsRunning() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gameServer.Wait()
		}()
	}

	wg.Wait()
	fmt.Println("All services have been stopped.")
}

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

	var (
		gameServer     *kfserver.KFServer
		redirectServer *redirectserver.HTTPRedirectServer
		err            error
	)

	defer shutdown(cancel, redirectServer, gameServer)

	// Start the Killing Floor Dedicated Server
	gameServer, err = startGameServer(sett, ctx)
	if err != nil {
		fmt.Printf("ERROR: [KFServer]: %v\n", err)
		return
	}
	fmt.Printf("> Killing Floor Server started from '%s'\n", gameServer.RootDirectory())

	// Start the HTTP Redirect Server, if enabled
	if sett.EnableRedirectServer.Value() {
		redirectServer, err = startRedirectServer(sett, ctx)
		if err != nil {
			fmt.Printf("ERROR: [RedirectServer]: %v\n", err)
			return
		}
		fmt.Printf("> HTTP Redirect Server serving '%s' on %s\n", redirectServer.RootDirectory(), redirectServer.Host())
	}

	<-signalChan
}
