package main

import (
	"encoding/json"
	"log"
	"os"
)

// A gateway API request that uses JSON content type constructed using "method request passthrough" template.
type GWInput struct {
	Body     json.RawMessage   `json:"body-json"`
	Param    GWParams          `json:"params"`
	StageVar map[string]string `json:"stage-variables"`
	Context  GWRequest         `json:"context"`
}

func GWInputFromJSON(js []byte) (in GWInput) {
	if err := json.Unmarshal(js, &in); err != nil {
		log.Panicf("Failed to deserialise GWInputFromJSON - %v", err)
	}
	return
}

// API request parameters.
type GWParams struct {
	Path   map[string]string `json:"path"`
	Query  map[string]string `json:"querystring"`
	Header map[string]string `json:"header"`
}

// API request endpoint and other context information.
type GWRequest struct {
	Method string `json:"http-method"`
	Path   string `json:"resource-path"`
	Stage  string `json:"stage"`
	IP     string `json:"source-ip"`
}

// All gateway API responses conform to this template.
type GWOutput struct {
	Status   int               `json:"status"`
	Header   map[string]string `json:"header"`
	BodyJSON interface{}       `json:"body-json"`
}

func (out GWOutput) ToJSON() []byte {
	// Header shall never be null, but body JSON can.
	if out.Header == nil {
		out.Header = map[string]string{}
	}
	js, err := json.Marshal(out)
	if err != nil {
		log.Panicf("Failed to serialise GWOutput '%+v': %+v", out, err)
	}
	return js
}

// Write an API response to stdout and exit with status 0. The return value is purely cosmetic for coding style.
func LambdaOutput(gwOutput GWOutput) int {
	os.Stdout.Write(gwOutput.ToJSON())
	os.Exit(0)
	return 0
}
