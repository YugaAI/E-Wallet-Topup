package main

import (
	"ewallet-topup/cmd"
	"ewallet-topup/helpers"

	"go.temporal.io/sdk/client"
)

func main() {

	// load config
	helpers.SetupConfig()

	// load log
	helpers.SetupLogger()
	// load db
	helpers.SetupMySQL()

	temporalClient, err := client.Dial(client.Options{
		HostPort: helpers.GetEnv("TEMPORAL_HOST", "localhost:7233"),
	})
	if err != nil {
		panic(err)
	}
	defer temporalClient.Close()
	// run grpc
	//go cmd.ServeGRPC()

	// run http
	cmd.ServeHTTP(temporalClient)
}
