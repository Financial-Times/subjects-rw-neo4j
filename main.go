package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/Financial-Times/base-ft-rw-app-go"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/neo-utils-go"
	"github.com/Financial-Times/subjects-rw-neo4j/subjects"
	log "github.com/Sirupsen/logrus"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"strconv"
)

func init() {
	log.SetFormatter(new(log.JSONFormatter))
}

func main() {
	app := cli.App("subjetcs-rw-neo4j", "A RESTful API for managing Subjects in neo4j")
	neoURL := stringEnv("NEO_URL", "http://localhost:7474/db/data") //"neo4j endpoint URL"
	port := intEnv("PORT", 8080)                                    //"Port to listen on"
	batchSize := intEnv("BATCH_SIZE", 1024)                         //"Maximum number of statements to execute per batch"
	graphiteTCPAddress := stringEnv("GRAPHITE_TCP_ADDRESS", "")     //"Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)"
	graphitePrefix := stringEnv("GRAPHITE_PREFIX", "")              //"Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.subjects-rw-neo4j.1"
	logMetrics := boolEnv("LOG_METRICS", false)                     //"Whether to log metrics. Set to true if running locally and you want metrics output"

	app.Action = func() {
		db, err := neoism.Connect(neoURL)
		if err != nil {
			log.Errorf("Could not connect to neo4j, error=[%s]\n", err)
		}

		batchRunner := neoutils.NewBatchCypherRunner(neoutils.StringerDb{db}, batchSize)
		subjectsDriver := subjects.NewCypherSubjectsService(batchRunner, db)

		baseftrwapp.OutputMetricsIfRequired(graphiteTCPAddress, graphitePrefix, logMetrics)

		engs := map[string]baseftrwapp.Service{
			"subjects": subjectsDriver,
		}

		var checks []v1a.Check
		for _, e := range engs {
			checks = append(checks, makeCheck(e, batchRunner))
		}

		baseftrwapp.RunServer(engs,
			v1a.Handler("ft-subjects_rw_neo4j ServiceModule", "Writes 'subjects' to Neo4j, usually as part of a bulk upload done on a schedule", checks...),
			port)
	}

	app.Run(os.Args)
}

func makeCheck(service baseftrwapp.Service, cr neoutils.CypherRunner) v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Cannot read/write subjects via this writer",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "TODO - write panic guide",
		Severity:         1,
		TechnicalSummary: fmt.Sprintf("Cannot connect to Neo4j instance %s with at least one subject loaded in it", cr),
		Checker:          func() (string, error) { return "", service.Check() },
	}
}

func stringEnv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func intEnv(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Errorf("Could not convert %s:\"%s\" to int. Using default value: %d", key, v, def)
	}
	return i
}

func boolEnv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Errorf("Could not convert %s:\"%s\" to int. Using default value: %d", key, v, def)
	}
	return b
}
