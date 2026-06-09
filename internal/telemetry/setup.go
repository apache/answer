/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package telemetry

import (
	"context"
	"os"
	"time"

	otelconf "go.opentelemetry.io/contrib/otelconf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"github.com/segmentfault/pacman/log"
)

const defaultConfigPath = "configs/otel.yaml"

func SetupTelemetry(serviceVersion string) *Providers {
	configPath := os.Getenv("OTEL_CONFIG_FILE")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	b, err := os.ReadFile(configPath)
	if err != nil {
		log.Warnf("telemetry: config file %q not found, using no-op providers: %v", configPath, err)
		return noopProviders()
	}

	cfg, err := otelconf.ParseYAML(b)
	if err != nil {
		log.Warnf("telemetry: failed to parse otel config, using no-op providers: %v", err)
		return noopProviders()
	}

	injectRuntimeAttributes(cfg, serviceVersion)

	sdk, err := otelconf.NewSDK(otelconf.WithOpenTelemetryConfiguration(*cfg))
	if err != nil {
		log.Warnf("telemetry: failed to create SDK, using no-op providers: %v", err)
		return noopProviders()
	}

	otel.SetTracerProvider(sdk.TracerProvider())
	otel.SetMeterProvider(sdk.MeterProvider())
	otel.SetTextMapPropagator(sdk.Propagator())

	closer := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := sdk.Shutdown(ctx); err != nil {
			log.Errorf("telemetry: shutdown error: %v", err)
		}
	}

	return &Providers{
		TracerProvider: sdk.TracerProvider(),
		MeterProvider:  sdk.MeterProvider(),
		Closer:         closer,
	}
}

func injectRuntimeAttributes(cfg *otelconf.OpenTelemetryConfiguration, serviceVersion string) {
	if cfg.Resource == nil {
		cfg.Resource = &otelconf.Resource{}
	}

	instanceID := os.Getenv("HOSTNAME")
	if instanceID == "" {
		instanceID = "unknown"
	}

	cfg.Resource.Attributes = append(cfg.Resource.Attributes,
		otelconf.AttributeNameValue{Name: "service.version", Value: serviceVersion},
		otelconf.AttributeNameValue{Name: "service.instance.id", Value: instanceID},
	)
}

func noopProviders() *Providers {
	return &Providers{
		TracerProvider: tracenoop.NewTracerProvider(),
		MeterProvider:  noop.NewMeterProvider(),
		Closer:         func() {},
	}
}
