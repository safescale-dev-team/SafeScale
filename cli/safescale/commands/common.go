/*
 * Copyright 2018-2021, CS Systemes d'Information, http://csgroup.eu
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

package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/CS-SI/SafeScale/lib/client"
	"github.com/CS-SI/SafeScale/lib/protocol"
	clitools "github.com/CS-SI/SafeScale/lib/utils/cli"
	"github.com/CS-SI/SafeScale/lib/utils/cli/enums/exitcode"
	"github.com/CS-SI/SafeScale/lib/utils/temporal"
)

var (
	// Verbose tells if user asks more verbosity
	Verbose bool
	// Debug tells if user asks debug information
	Debug bool

	hostName     string
	hostInstance *protocol.Host
	featureName  string
)

func extractFeatureArgument(c *cli.Context) error {
	if c.NArg() < 2 {
		_ = cli.ShowSubcommandHelp(c)
		return clitools.ExitOnInvalidArgument("Missing mandatory argument FEATURENAME")
	}
	featureName = c.Args().Get(1)
	if featureName == "" {
		return clitools.ExitOnInvalidArgument("Invalid argument FEATURENAME")
	}
	return nil
}

// Use the hostnamePos argument of the command as a hostName and use it to get the host instance
func extractHostArgument(c *cli.Context, hostnamePos int) error {
	hostName = c.Args().Get(hostnamePos)
	if hostName == "" {
		return clitools.ExitOnInvalidArgument("argument HOSTNAME invalid")
	}

	clientSession, xerr := client.New(c.String("server"))
	if xerr != nil {
		return clitools.FailureResponse(clitools.ExitOnErrorWithMessage(exitcode.Run, xerr.Error()))
	}

	var err error
	hostInstance, err = clientSession.Host.Inspect(hostName, temporal.GetExecutionTimeout())
	if err != nil {
		// fmt.Printf("%s\n", err.Error()
		return clitools.ExitOnRPC(err.Error())
	}

	if hostInstance == nil {
		return clitools.ExitOnErrorWithMessage(exitcode.NotFound, fmt.Sprintf("Host '%s' not found", hostName))
	}

	return nil
}

// Use the 'nodePos'th argument of the command as a node reference and init hostName with it
func extractNodeArgument(c *cli.Context, nodePos int) error {
	hostName = c.Args().Get(nodePos)
	if hostName == "" {
		return clitools.ExitOnInvalidArgument("argument HOSTNAME invalid")
	}

	return nil
}
