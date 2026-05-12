// Package main is the Galileo API gateway entry point.
//
// Stage 0 skeleton: boots, prints, exits. Real wiring (LiteLLM proxy,
// tenant resolver, budget cap, Opik logging, JWT verification) lands in
// Week 3 per docs/plans/STAGE_0_PLAN.md.
package main

import (
	"fmt"
	"log"
	"os"
)

const serviceName = "galileo-gateway"

func main() {
	log.SetPrefix(serviceName + " ")
	log.SetFlags(log.LstdFlags | log.LUTC)

	fmt.Println(serviceName + " stage0 skeleton")

	// Stage 0: no HTTP server yet. Week 3 wires the real gateway.
	if _, err := fmt.Fprintln(os.Stdout, "ready"); err != nil {
		log.Fatalf("write: %v", err)
	}
}
