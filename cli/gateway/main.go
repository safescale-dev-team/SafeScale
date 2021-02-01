/*
 * Copyright 2018-2020, CS Systemes d'Information, http://csgroup.eu
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"context"
	"net/http"
	"syscall"


	"github.com/dlespiau/covertool/pkg/exit"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"

	"github.com/CS-SI/SafeScale/lib/protocol"
	_ "github.com/CS-SI/SafeScale/lib/server"
	"github.com/CS-SI/SafeScale/lib/server/iaas"
	app2 "github.com/CS-SI/SafeScale/lib/utils/app"
	"github.com/CS-SI/SafeScale/lib/utils/debug"
	"github.com/CS-SI/SafeScale/lib/utils/debug/tracing"
)

var profileCloseFunc = func() {}

const (
	defaultDaemonHost string = "localhost" // By default, safescaled only listen on localhost
	defaultDaemonPort string = "8080"
	defaultGrpcPort string = "50051"
)


func cleanup(onAbort bool) {
	if onAbort {
		fmt.Println("Cleaning up...")
	}
	profileCloseFunc()
	exit.Exit(1)
}

// newGateway returns a new gateway server which translates HTTP into gRPC.
func newGateway(ctx context.Context, conn *grpc.ClientConn, opts []gwruntime.ServeMuxOption) (http.Handler, error) {

	mux := gwruntime.NewServeMux(opts...)

	for _, f := range []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
		protocol.RegisterBucketServiceHandler,
		protocol.RegisterClusterServiceHandler,
		protocol.RegisterHostServiceHandler,
		protocol.RegisterImageServiceHandler,
//		protocol.RegisterJobServiceHandler,
		protocol.RegisterNetworkServiceHandler,
		protocol.RegisterSubnetServiceHandler,
//		protocol.RegisterSecurityGroupServiceHandler,
//		protocol.RegisterShareServiceHandler,
//		protocol.RegisterSshServiceHandler,
//		protocol.RegisterTemplateServiceHandler,
		protocol.RegisterTenantServiceHandler,
		protocol.RegisterVolumeServiceHandler,
	} {
		if err := f(ctx, mux, conn); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

func work(c *cli.Context) {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		cleanup(true)
	}()

	// NOTE: is it the good behavior ? Shouldn't we fail ?
	// If trace settings cannot be registered, report it but do not fail
	err := tracing.RegisterTraceSettings(appTrace)
	if err != nil {
		logrus.Errorf(err.Error())
	}

	logrus.Infoln("Checking configuration")
	_, err = iaas.GetTenantNames()
	if err != nil {
		logrus.Fatalf(err.Error())
	}

	listen := assembleListenString(c)
	endpoint := assembleEndpointString(c)
	
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	logrus.Infof("Connecting to grpc backend on '%s'", endpoint)
	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("Failed to connect to grpc: %v", err)
		return
	}
	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			logrus.Errorf("Failed to close a client connection to the gRPC server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/openapiv2/", openAPIServer("lib/protocol"))
	

	var opts []gwruntime.ServeMuxOption
	gw, err := newGateway(ctx, conn, opts)
	if err != nil {
		logrus.Errorf("Failed to initialize gateway, %v", err)
		return
	}
	mux.Handle("/", gw)

	s := &http.Server{
		Addr:    "localhost:8080",
		Handler: allowCORS(mux),
	}
	go func() {
		<-ctx.Done()
		logrus.Infof("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			logrus.Errorf("Failed to shutdown http server: %v", err)
		}
	}()
	
	logrus.Infof("Listening http on '%s'", listen)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Errorf("Failed to listen and serve: %v", err)
		return
	}
}

// assembleListenString constructs the listen string we will use in net.Listen()
func assembleListenString(c *cli.Context) string {
	// Get listen from parameters
	listen := c.String("listen")
	if listen != "" {
		// Validate port part of the content of listen...
		parts := strings.Split(listen, ":")
		switch len(parts) {
		case 1:
			listen = parts[0] + ":" + defaultDaemonPort
		case 2:
			num, err := strconv.Atoi(parts[1])
			if err != nil || num <= 0 {
				logrus.Warningf("Parameter 'listen' content is invalid (port cannot be '%s'): ignored.", parts[1])
			}
		default:
			logrus.Warningf("Parameter 'listen' content is invalid, ignored.")
		}
	}
	// At last, if listen is empty, build it from defaults
	if listen == "" {
		listen = defaultDaemonHost + ":" + defaultDaemonPort
	}
	return listen
}

// assembleEndpointString constructs the endpoint string we will use to send grpc requests
func assembleEndpointString(c *cli.Context) string {
	// Get listen from parameters
	listen := c.String("endpoint")
	if listen == "" {
		listen = os.Getenv("SAFESCALED_LISTEN")
	}
	if listen != "" {
		// Validate port part of the content of listen...
		parts := strings.Split(listen, ":")
		switch len(parts) {
		case 1:
			listen = parts[0] + ":" + defaultGrpcPort
		case 2:
			num, err := strconv.Atoi(parts[1])
			if err != nil || num <= 0 {
				logrus.Warningf("Parameter 'listen' content is invalid (port cannot be '%s'): ignored.", parts[1])
			}
		default:
			logrus.Warningf("Parameter 'listen' content is invalid, ignored.")
		}
	}
	// if listen is empty, get the port from env
	if listen == "" {
		if port := os.Getenv("SAFESCALED_PORT"); port != "" {
			num, err := strconv.Atoi(port)
			if err != nil || num <= 0 {
				logrus.Warningf("Environment variable 'SAFESCALED_PORT' contains invalid content ('%s'): ignored.", port)
			} else {
				listen = defaultDaemonHost + ":" + port
			}
		}
	}
	// At last, if listen is empty, build it from defaults
	if listen == "" {
		listen = defaultDaemonHost + ":" + defaultGrpcPort
	}
	return listen
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := cli.NewApp()
	app.Name = "safescale-gatewayd"
	app.Usage = "safescale-gatewayd [OPTIONS]"
	app.Version = Version + ", build " + Revision + " compiled with " + runtime.Version() + " (" + BuildDate + ")"

	app.Authors = []*cli.Author{
		{
			Name:  "CS-SI",
			Email: "safescale@csgroup.eu",
		},
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "Print program version",
	}

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Increase verbosity",
		},
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Usage:   "Show debug information",
		},
		&cli.StringFlag{
			Name:    "listen",
			Aliases: []string{"l"},
			Usage:   "Listen on specified port `IP:PORT` (default: localhost:8080)",
		},
		&cli.StringFlag{
			Name:    "endpoint",
			Aliases: []string{"e"},
			Usage:   "Safescale grpc daemon `IP:PORT` (default: localhost:50051)",
		},
	}

	app.Before = func(c *cli.Context) error {
		// Sets profiling
		if c.IsSet("profile") {
			what := c.String("profile")
			profileCloseFunc = debug.Profile(what)
		}

		if strings.Contains(path.Base(os.Args[0]), "-cover") {
			logrus.SetLevel(logrus.TraceLevel)
			app2.Verbose = true
		} else {
			logrus.SetLevel(logrus.WarnLevel)
		}

		if c.Bool("verbose") {
			logrus.SetLevel(logrus.InfoLevel)
			app2.Verbose = true
		}
		if c.Bool("debug") {
			if c.Bool("verbose") {
				logrus.SetLevel(logrus.TraceLevel)
			} else {
				logrus.SetLevel(logrus.DebugLevel)
			}
			app2.Debug = true
		}
		return nil
	}

	app.Action = func(c *cli.Context) error {
		work(c)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}

	cleanup(false)
}

