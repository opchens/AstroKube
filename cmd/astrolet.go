/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"AstroKube/cmd/app"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
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
