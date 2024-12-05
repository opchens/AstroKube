/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"AstroKube/cmd/app"
	"AstroKube/cmd/app/options"
	astrov1 "AstroKube/pkg/apis/core/v1"
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	"syscall"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(astrov1.AddToScheme(scheme))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		cancel()
	}()

	var opts options.ServerRunOptions
	options.SetDefaultOpts(&opts)

	command := app.NewAstroLetCommand(ctx, opts)
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
