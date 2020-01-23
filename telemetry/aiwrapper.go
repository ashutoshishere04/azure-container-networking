// Copyright Microsoft. All rights reserved.
package telemetry

import (
	"fmt"
	"runtime"

	"github.com/Azure/azure-container-networking/aitelemetry"
	"github.com/Azure/azure-container-networking/log"
)

var (
	aiMetadata     string
	th             aitelemetry.TelemetryHandle
	gDisableTrace  bool
	gDisableMetric bool
)

const (
	// Wait time for AI to gracefully close AI telemetry session
	waitTimeInSecs = 10
)

func CreateAITelemetryHandle(aiConfig aitelemetry.AIConfig, disableAll, disableMetric, disableTrace bool) error {
	var err error

	if disableAll {
		log.Printf("Telemetry is disabled")
		return fmt.Errorf("Telmetry disabled")
	}

	th, err = aitelemetry.NewAITelemetry("", aiMetadata, aiConfig)
	if err != nil {
		return err
	}

	gDisableMetric = disableMetric
	gDisableTrace = disableTrace
	return nil
}

func SendAITelemetry(cnireport CNIReport) {
	if th == nil || gDisableTrace {
		return
	}

	var msg string
	if cnireport.ErrorMessage != "" {
		msg = cnireport.ErrorMessage
	} else {
		msg = cnireport.EventMessage
	}

	report := aitelemetry.Report{
		Message:          "CNI:" + msg,
		Context:          cnireport.ContainerName,
		CustomDimensions: make(map[string]string),
	}

	report.CustomDimensions[ContextStr] = cnireport.Context
	report.CustomDimensions[SubContextStr] = cnireport.SubContext
	report.CustomDimensions[VMUptimeStr] = cnireport.VMUptime
	report.CustomDimensions[OperationTypeStr] = cnireport.OperationType
	report.CustomDimensions[VersionStr] = cnireport.Version

	th.TrackLog(report)
}

func SendAIMetric(aiMetric AIMetric) {
	if th == nil || gDisableMetric {
		return
	}

	aiMetric.Metric.CustomDimensions[OSTypeStr] = runtime.GOOS
	th.TrackMetric(aiMetric.Metric)
}

func CloseAITelemetryHandle() {
	if th != nil {
		th.Close(waitTimeInSecs)
	}
}
