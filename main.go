package main

import (
	"encoding/json"
	"fmt"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Process a request sent by AWS API gateway.
func ProcessGatewayRequest() int {
	inputJS, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return LambdaOutput(GWOutput{
			Status:   http.StatusInternalServerError,
			Header:   nil,
			BodyJSON: ErrJSON{"err": fmt.Sprintf("Failed to read from stdin - %v", err)}})
	}
	gwInput := GWInputFromJSON(inputJS)
	fun, found := NewAPIImpl(nil).Routes[gwInput.Context.Method+gwInput.Context.Path]
	if !found {
		return LambdaOutput(GWOutput{
			Status:   http.StatusNotFound,
			Header:   nil,
			BodyJSON: ErrJSON{"err": fmt.Sprintf("No function found for '%s' '%s'", gwInput.Context.Method, gwInput.Context.Path)}})
	}
	return LambdaOutput(fun(gwInput))
}

func main() {
	if len(os.Args) == 1 {
		// Invoked without additional parameters by AWS API gateway
		ProcessGatewayRequest()
	} else {
		// Start an HTTP server and serve the APIs
		jwtPub, err := ioutil.ReadFile("jwt.pub")
		if err != nil {
			panic(err)
		}
		jwtKey, err := ioutil.ReadFile("jwt.key")
		if err != nil {
			panic(err)
		}
		// Read configuration parameters from file (first command line parameter)
		config, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		var stageVars map[string]string
		if err := json.Unmarshal(config, &stageVars); err != nil {
			panic(err)
		}
		stageVars["JWTPublicKey"] = string(jwtPub)
		stageVars["JWTPrivateKey"] = string(jwtKey)
		// Configure HTTP server and start
		apiimpl := NewAPIImpl(stageVars)
		svcMux := web.New()
		svcMux.Use(middleware.RequestID)
		svcMux.Use(middleware.Logger)
		svcMux.Use(middleware.Recoverer)
		svcMux.Use(AutomaticCORSOptions) // enable CORS for Swagger UI
		apiimpl.InstallMuxRoutes(svcMux)
		svcMux.Compile()
		log.Print("Will now start HTTP server")
		if err := graceful.ListenAndServe("0.0.0.0:12345", svcMux); err != nil {
			panic(err)
		}
	}
}
