package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"kfdsl/cmd"
	"kfdsl/internal/services/kfserver"
	"kfdsl/internal/services/redirectserver"
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

	settings := cmd.GetSettings()

	// Print settings
	settings.Print()

	// Start SteamCMD, if enabled
	if !settings.NoSteam.Value() {
		if err := startSteamCMD(ctx); err != nil {
			fmt.Printf("ERROR: [SteamCMD]: %v\n", err)
			os.Exit(1)
		}
	}

	var gameServer *kfserver.KFServer
	var redirectServer *redirectserver.HTTPRedirectServer
	var err error

	defer shutdown(cancel, redirectServer, gameServer)

	// Start the Killing Floor Dedicated Server
	gameServer, err = startGameServer(ctx)
	if err != nil {
		fmt.Printf("ERROR: [KFServer]: %v\n", err)
		return
	}
	fmt.Printf("> Killing Floor Server started from '%s'\n", gameServer.RootDirectory())

	// Start the HTTP Redirect Server, if enabled
	if settings.EnableRedirectServer.Value() {
		redirectServer, err = startRedirectServer(ctx)
		if err != nil {
			fmt.Printf("ERROR: [RedirectServer]: %v\n", err)
			return
		}
		fmt.Printf("> HTTP Redirect Server serving directory '%s' on %s\n", redirectServer.RootDirectory(), redirectServer.Host())
	}

	/*
		// Test
		go func() {
			fmt.Printf("\n------ Let's wait 20 seconds then restart the server... ------\n\n")
			time.Sleep(20 * time.Second)
			fmt.Printf("\n------ Restarting... ------\n\n")
			if err := gameServer.Restart(); err != nil {
				fmt.Printf("KFServer restart error: %v\n", err)
			}
		}()
	*/
	<-signalChan
}
