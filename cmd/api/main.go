package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/PDOK/uptime-operator/internal/api"
	"github.com/PDOK/uptime-operator/internal/service"
	"github.com/PDOK/uptime-operator/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v3"
)

func main() {
	providerFlags := util.GetProviderFlags()
	uptimeProviderID, uptimeProviderSettings := service.GetUptimeProviderSettings(providerFlags)
	provider := service.GetProviderFromIdAndSettings(uptimeProviderID, uptimeProviderSettings)

	action := func(context context.Context, command *cli.Command) error {
		endpoints := api.Endpoints{
			UptimeProvider: provider,
		}

		router := gin.New()
		api.AddRoutes(router, endpoints)
		addr := ":" + strconv.Itoa(8080)
		server := http.Server{
			Addr:    addr,
			Handler: router,
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
