/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"AstroKube/cmd/app"
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		cancel()
	}()

	command := app.NewAstroLetCommand(ctx)
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
