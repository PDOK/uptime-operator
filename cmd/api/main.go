package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/PDOK/uptime-operator/internal/api"
	"github.com/PDOK/uptime-operator/internal/service"
	"github.com/PDOK/uptime-operator/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/peterbourgon/ff"
	"github.com/urfave/cli/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	var hostAddress string
	flag.StringVar(&hostAddress, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	providerFlags := util.GetProviderFlags()

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	if err := ff.Parse(flag.CommandLine, os.Args[1:], ff.WithEnvVarNoPrefix()); err != nil {
		os.Exit(1)
	}

	uptimeProviderID, uptimeProviderSettings := service.GetUptimeProviderSettings(*providerFlags)
	provider := service.GetProviderFromIDAndSettings(uptimeProviderID, uptimeProviderSettings)

	action := func(_ context.Context, _ *cli.Command) error {
		endpoints := api.Endpoints{
			UptimeProvider: provider,
		}

		router := gin.New()
		api.AddRoutes(router, endpoints)
		addr := hostAddress
		server := http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: 0,
		}
		err := server.ListenAndServe()
		return err
	}

	command := cli.Command{
		Name:      "uptime-operator-api",
		Usage:     "A REST API for the uptime-operator for queries and manual operations",
		UsageText: "A REST API for the uptime-operator for queries and manual operations",
		Flags:     nil,
		Action:    action,
	}
	err := command.Run(context.Background(), os.Args)

	if err != nil {
		fmt.Printf("Error occurred during execution: %s", err)
		os.Exit(1)
	}
}
