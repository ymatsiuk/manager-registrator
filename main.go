package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	consulapi "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

var (
	e             error
	managerFilter = types.NodeListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "role", Value: "manager"}),
	}
	federationService = &consulapi.AgentService{}
	swarmService      = &consulapi.CatalogRegistration{
		Service: federationService,
	}
)

func main() {
	level := getEnvOrFallback("LOG_LEVEL", "info")
	setLogger(level)

	federationService.Service = getEnvOrFallback("CONSUL_SERVICE_NAME", "prometheus-federation")
	federationService.Port, e = strconv.Atoi(getEnvOrFallback("CONSUL_SERVICE_PORT", "9100"))
	if e != nil {
		log.Fatal("Failed to parse CONSUL_SERVICE_PORT. Should be integer.")
	}
	swarmService.Node = getEnvOrFallback("CONSUL_SERVICE_NODE", "swarm-manager")
	swarmService.Datacenter = getEnvOrFallback("CONSUL_DC", "dc1")

	log.Debug("Filter is set to: ", managerFilter)

	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	catalog := client.Catalog()

	for {
		serviceReg := createCatReg()
		_, e := catalog.Register(serviceReg, nil)
		if e == nil {
			log.Info("Service registered in Consul")
		} else {
			log.Error("Connection to consul failed: ", e)
		}
		log.Info("Go to sleep. Retry in 10 minutes")
		time.Sleep(time.Minute * 10)
	}
}

func createCatReg() *consulapi.CatalogRegistration {
	docker, _ := client.NewEnvClient()
	nodes, e := docker.NodeList(context.Background(), managerFilter)
	if e != nil {
		log.Error("Failed to fetch data from docker. ", e)
	}

	for _, v := range nodes {
		if v.ManagerStatus.Leader {
			federationService.Address = v.Status.Addr
			swarmService.Address = v.Description.Hostname
			log.Info("Current swarm leader: ", v.Description.Hostname)
		}
	}
	log.Debug(swarmService)
	return swarmService
}

func setLogger(level string) {
	customJSONFormatter := &log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg:  "message",
			log.FieldKeyTime: "@timestamp",
		},
	}
	log.SetFormatter(customJSONFormatter)
	logLevel, e := log.ParseLevel(level)
	if e != nil {
		log.Fatal("Failed to set log level: ", e)
	}
	log.SetLevel(logLevel)
}

func getEnvOrFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
