// Package main is the Galileo agent runner entry point.
//
// Stage 0 skeleton: boots, prints, exits. Week 3 wires Temporal client +
// HelloAgentWorkflow + CallLLMActivity per docs/plans/STAGE_0_PLAN.md.
package main

import (
	"fmt"
	"log"
)

const serviceName = "galileo-agent-runner"

func main() {
	log.SetPrefix(serviceName + " ")
	log.SetFlags(log.LstdFlags | log.LUTC)

	fmt.Println(serviceName + " stage0 skeleton")
}
