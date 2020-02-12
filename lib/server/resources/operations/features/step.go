/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
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

package features

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/CS-SI/SafeScale/lib/client"
	"github.com/CS-SI/SafeScale/lib/server/resources"
	"github.com/CS-SI/SafeScale/lib/server/resources/enums/installaction"
	"github.com/CS-SI/SafeScale/lib/system"
	"github.com/CS-SI/SafeScale/lib/utils"
	"github.com/CS-SI/SafeScale/lib/utils/cli/enums/outputs"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/data"
	"github.com/CS-SI/SafeScale/lib/utils/retry"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
	"github.com/CS-SI/SafeScale/lib/utils/temporal"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
)

const (
	targetHosts    = "hosts"
	targetMasters  = "masters"
	targetNodes    = "nodes"
	targetGateways = "gateways"
)

type stepResult struct {
	completed bool // if true, the script has been run to completion
	output    string
	success   bool  // if true, the script has been run successfully and the result is a success
	err       error // if an error occurred, contains the err
}

func (sr stepResult) Successful() bool {
	return sr.success
}

func (sr stepResult) Completed() bool {
	return sr.completed
}

func (sr stepResult) Error() error {
	return sr.err
}

func (sr stepResult) ErrorMessage() string {
	if sr.err != nil {
		return sr.err.Error()
	}
	return ""
}

// stepResults contains the errors of the step for each host target
type stepResults map[string]stepResult

// ErrorMessages returns a string containing all the errors registered
func (s stepResults) ErrorMessages() string {
	output := ""
	for h, k := range s {
		val := k.ErrorMessage()
		if val != "" {
			output += h + ": " + val + "\n"
		}
	}
	return output
}

// UncompletedEntries returns an array of string of all keys where the script
// to run action wasn't completed
func (s stepResults) UncompletedEntries() []string {
	var output []string
	for k, v := range s {
		if !v.Completed() {
			output = append(output, k)
		}
	}
	return output
}

// Successful tells if all the steps have been successful
func (s stepResults) Successful() bool {
	if len(s) == 0 {
		return false
	}
	for _, k := range s {
		if !k.Successful() {
			return false
		}
	}
	return true
}

// Completed tells if all the scripts corresponding to action have been completed.
func (s stepResults) Completed() bool {
	if len(s) == 0 {
		return false
	}
	for _, k := range s {
		if !k.Completed() {
			return false
		}
	}
	return true
}

type stepTargets map[string]string

// parse converts the content of specification file loaded inside struct to
// standardized values (0, 1 or *)
func (st stepTargets) parse() (string, string, string, string, error) {
	var (
		hostT, masterT, nodeT, gwT string
		ok                         bool
	)

	if hostT, ok = st[targetHosts]; ok {
		switch strings.ToLower(hostT) {
		case "":
			fallthrough
		case "false":
			fallthrough
		case "no":
			fallthrough
		case "none":
			fallthrough
		case "0":
			hostT = "0"
		case "yes":
			fallthrough
		case "true":
			fallthrough
		case "1":
			hostT = "1"
		default:
			return "", "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", hostT, targetHosts)
		}
	}

	if masterT, ok = st[targetMasters]; ok {
		switch strings.ToLower(masterT) {
		case "":
			fallthrough
		case "false":
			fallthrough
		case "no":
			fallthrough
		case "none":
			fallthrough
		case "0":
			masterT = "0"
		case "any":
			fallthrough
		case "one":
			fallthrough
		case "1":
			masterT = "1"
		case "all":
			fallthrough
		case "*":
			masterT = "*"
		default:
			return "", "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", masterT, targetMasters)
		}
	}

	if nodeT, ok = st[targetNodes]; ok {
		switch strings.ToLower(nodeT) {
		case "":
			fallthrough
		case "false":
			fallthrough
		case "no":
			fallthrough
		case "none":
			nodeT = "0"
		case "any":
			fallthrough
		case "one":
			fallthrough
		case "1":
			nodeT = "1"
		case "all":
			fallthrough
		case "*":
			nodeT = "*"
		default:
			return "", "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", nodeT, targetNodes)
		}
	}

	if gwT, ok = st[targetGateways]; ok {
		switch strings.ToLower(gwT) {
		case "":
			fallthrough
		case "false":
			fallthrough
		case "no":
			fallthrough
		case "none":
			fallthrough
		case "0":
			gwT = "0"
		case "any":
			fallthrough
		case "one":
			fallthrough
		case "1":
			gwT = "1"
		case "all":
			fallthrough
		case "*":
			gwT = "*"
		default:
			return "", "", "", "", fmt.Errorf("invalid value '%s' for target '%s'", gwT, targetGateways)
		}
	}

	if hostT == "0" && masterT == "0" && nodeT == "0" && gwT == "0" {
		return "", "", "", "", fmt.Errorf("no targets identified")
	}
	return hostT, masterT, nodeT, gwT, nil
}

// step is a struct containing the needed information to apply the installation
// step on all selected host targets
type step struct {
	// Worker is a back pointer to the caller
	Worker *worker
	// Name is the name of the step
	Name string
	// Action is the action of the step (check, add, remove)
	Action installaction.Enum
	// Targets contains the host targets to select
	Targets stepTargets
	// Script contains the script to execute
	Script string
	// WallTime contains the maximum time the step must run
	WallTime time.Duration
	// YamlKey contains the root yaml key on the specification file
	YamlKey string
	// OptionsFileContent contains the "options file" if it exists (for DCOS cluster for now)
	OptionsFileContent string
	// Serial tells if step can be performed in parallel on selected host or not
	Serial bool
}

// Run executes the step on all the concerned hosts
func (is *step) Run(hosts []resources.Host, v data.Map, s resources.FeatureSettings) (outcomes resources.UnitResults, err error) {
	outcomes = unitResults{}

	tracer := concurrency.NewTracer(is.Worker.feature.task, "", true).GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()
	nHosts := uint(len(hosts))
	defer temporal.NewStopwatch().OnExitLogWithLevel(
		fmt.Sprintf("Starting step '%s' on %d host%s...", is.Name, nHosts, utils.Plural(nHosts)),
		fmt.Sprintf("Ending step '%s' on %d host%s", is.Name, len(hosts), utils.Plural(nHosts)),
		logrus.DebugLevel,
	)()

	if is.Serial || s.Serialize {

		for _, h := range hosts {
			tracer.Trace("%s(%s):step(%s)@%s: starting", is.Worker.action.String(), is.Worker.feature.Name(), is.Name, h.Name)
			is.Worker.startTime = time.Now()

			cloneV := v.Clone()
			cloneV["HostIP"], err = h.PrivateIP(is.Worker.feature.task)
			if err != nil {
				return nil, err
			}
			cloneV["Hostname"] = h.Name
			cloneV, err = realizeVariables(cloneV)
			if err != nil {
				return nil, err
			}
			subtask, err := concurrency.NewTaskWithParent(is.Worker.feature.task)
			if err != nil {
				return nil, err
			}
			outcome, err := subtask.Run(is.taskRunOnHost, data.Map{"host": h, "variables": cloneV})
			if err != nil {
				return nil, err
			}
			outcomes.AddSingle(h.Name(), outcome.(resources.UnitResult))
			subtask.Close()
			// err = subtask.Reset()
			// if err != nil {
			// 	return nil, err
			// }

			if !outcomes.Successful() {
				if is.Worker.action == installaction.Check { // Checks can fail and it's ok
					tracer.Trace("%s(%s):step(%s)@%s finished in %s: not present: %s",
						is.Worker.action.String(), is.Worker.feature.Name(), is.Name, h.Name,
						temporal.FormatDuration(time.Since(is.Worker.startTime)), outcomes.ErrorMessages())
				} else { // other steps are expected to succeed
					tracer.Trace("%s(%s):step(%s)@%s failed in %s: %s",
						is.Worker.action.String(), is.Worker.feature.Name(), is.Name, h.Name,
						temporal.FormatDuration(time.Since(is.Worker.startTime)), outcomes.ErrorMessages())
				}
			} else {
				tracer.Trace("%s(%s):step(%s)@%s succeeded in %s.",
					is.Worker.action.String(), is.Worker.feature.Name(), is.Name, h.Name,
					temporal.FormatDuration(time.Since(is.Worker.startTime)))
			}
		}
	} else {
		subtasks := map[string]concurrency.Task{}
		for _, h := range hosts {
			tracer.Trace("%s(%s):step(%s)@%s: starting", is.Worker.action.String(), is.Worker.feature.Name(), is.Name, h.Name)
			is.Worker.startTime = time.Now()

			cloneV := v.Clone()
			cloneV["HostIP"], err = h.PrivateIP(is.Worker.feature.task)
			if err != nil {
				return nil, err
			}
			cloneV["Hostname"] = h.Name
			cloneV, err = realizeVariables(cloneV)
			if err != nil {
				return nil, err
			}
			subtask, err := concurrency.NewTaskWithParent(is.Worker.feature.task)
			if err != nil {
				return nil, err
			}

			subtask, err = subtask.Start(is.taskRunOnHost, data.Map{
				"host":      h,
				"variables": cloneV,
			})
			if err != nil {
				return nil, err
			}

			subtasks[h.Name()] = subtask
		}
		for k, s := range subtasks {
			outcome, err := s.Wait()
			if err != nil {
				log.Warn(tracer.TraceMessage(": %s(%s):step(%s)@%s finished after %s, but failed to recover result",
					is.Worker.action.String(), is.Worker.feature.Name(), is.Name, k, temporal.FormatDuration(time.Since(is.Worker.startTime))))
				continue
			}
			outcomes.AddSingle(k, outcome.(resources.UnitResult))

			if !outcomes.Successful() {
				if is.Worker.action == installaction.Check { // Checks can fail and it's ok
					tracer.Trace(": %s(%s):step(%s)@%s finished in %s: not present: %s",
						is.Worker.action.String(), is.Worker.feature.Name(), is.Name, k,
						temporal.FormatDuration(time.Since(is.Worker.startTime)), outcomes.ErrorMessages())
				} else { // other steps are expected to succeed
					tracer.Trace(": %s(%s):step(%s)@%s failed in %s: %s",
						is.Worker.action.String(), is.Worker.feature.Name(), is.Name, k,
						temporal.FormatDuration(time.Since(is.Worker.startTime)), outcomes.ErrorMessages())
				}
			} else {
				tracer.Trace("%s(%s):step(%s)@%s succeeded in %s.",
					is.Worker.action.String(), is.Worker.feature.Name(), is.Name, k,
					temporal.FormatDuration(time.Since(is.Worker.startTime)))
			}
		}
	}
	return outcomes, nil
}

// taskRunOnHost ...
// Respects interface concurrency.TaskFunc
// func (is *step) runOnHost(host *protocol.Host, v Variables) Resources.UnitResult {
func (is *step) taskRunOnHost(task concurrency.Task, params concurrency.TaskParameters) (result concurrency.TaskResult, err error) {
	var (
		p  = data.Map{}
		ok bool
	)
	if params != nil {
		if p, ok = params.(data.Map); !ok {
			return nil, scerr.InvalidParameterError("params", "must be a 'data.Map'")
		}
	}

	// Get parameters
	host, ok := p["host"].(resources.Host)
	if !ok {
		return nil, scerr.InvalidParameterError("params['host']", "must be a 'resources.Host'")
	}
	variables, ok := p["variables"].(data.Map)
	if !ok {
		return nil, scerr.InvalidParameterError("params['variables'", "must be a 'data.Map'")
	}

	// Updates variables in step script
	command, err := replaceVariablesInString(is.Script, variables)
	if err != nil {
		return stepResult{err: fmt.Errorf("failed to finalize installer script for step '%s': %s", is.Name, err.Error())}, nil
	}

	// If options file is defined, upload it to the remote host
	if is.OptionsFileContent != "" {
		err := UploadStringToRemoteFile(is.OptionsFileContent, host, utils.TempFolder+"/options.json", "cladm:safescale", "ug+rw-x,o-rwx")
		if err != nil {
			return stepResult{err: err}, nil
		}
	}

	hidesOutput := strings.Contains(command, "set +x\n")
	if hidesOutput {
		command = strings.Replace(command, "set +x\n", "\n", 1)
		if strings.Contains(command, "exec 2>&1\n") {
			command = strings.Replace(command, "exec 2>&1\n", "exec 2>&7\n", 1)
		}
	}

	// Uploads then executes command
	filename := fmt.Sprintf("%s/feature.%s.%s_%s.sh", utils.TempFolder, is.Worker.feature.Name(), strings.ToLower(is.Action.String()), is.Name)
	err = UploadStringToRemoteFile(command, host, filename, "", "")
	if err != nil {
		return stepResult{err: err}, nil
	}

	if !hidesOutput {
		command = fmt.Sprintf("sudo chmod u+rx %s;sudo bash %s;exit ${PIPESTATUS}", filename, filename)
	} else {
		command = fmt.Sprintf("sudo chmod u+rx %s;sudo bash -c \"BASH_XTRACEFD=7 %s 7> /tmp/captured 2>&7\";echo ${PIPESTATUS} > /tmp/errc;cat /tmp/captured; sudo rm /tmp/captured;exit `cat /tmp/errc`", filename, filename)
	}

	// Executes the script on the remote host
	retcode, outrun, _, err := client.New().SSH.Run(task, host.Name(), command, outputs.COLLECT, temporal.GetConnectionTimeout(), is.WallTime)
	if err != nil {
		return stepResult{err: err, output: outrun}, nil
	}
	err = nil
	ok = retcode == 0
	if !ok {
		err = handleExecuteScriptReturn(retcode, outrun, "", err, "failure")
	}
	return stepResult{success: ok, completed: true, err: err, output: outrun}, nil
}

func handleExecuteScriptReturn(retcode int, stdout string, stderr string, err error, msg string) error {
	richErrc := fmt.Sprintf("%d", retcode)

	var collected []string
	if stdout != "" {
		errLines := strings.Split(stdout, "\n")
		for _, errline := range errLines {
			if strings.Contains(errline, "An error occurred") {
				collected = append(collected, errline)
			}
		}
	}
	if stderr != "" {
		errLines := strings.Split(stderr, "\n")
		for _, errline := range errLines {
			if strings.Contains(errline, "An error occurred") {
				collected = append(collected, errline)
			}
		}
	}

	if len(collected) > 0 {
		if err != nil {
			return scerr.Wrap(err, fmt.Sprintf("%s: failed with error code %s, std errors [%s]", msg, richErrc, strings.Join(collected, ";")))
		}
		return fmt.Errorf("%s: failed with error code %s, std errors [%s]", msg, richErrc, strings.Join(collected, ";"))
	}

	if err != nil {
		return scerr.Wrap(err, fmt.Sprintf("%s: failed with error code %s", msg, richErrc))
	}
	if retcode != 0 {
		return fmt.Errorf("%s: failed with error code %s", msg, richErrc)
	}

	return nil
}

// UploadFile uploads a file to remote host
func UploadFile(localpath string, host resources.Host, remotepath, owner, mode string) (err error) {
	if localpath == "" {
		return scerr.InvalidParameterError("localpath", "cannot be empty string")
	}
	if host == nil {
		return scerr.InvalidParameterError("host", "cannot be nil")
	}
	if remotepath == "" {
		return scerr.InvalidParameterError("remotepath", "cannot be empty string")
	}

	voidtask, err := concurrency.NewTask()
	if err != nil {
		return err
	}
	tracer := concurrency.NewTracer(voidtask, "", true).WithStopwatch().GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	retryErr := retry.WhileUnsuccessful(
		func() error {
			retcode, _, _, err := host.Push(voidtask, localpath, remotepath, "", "", temporal.GetExecutionTimeout())
			if err != nil {
				return err
			}
			if retcode != 0 {
				// If retcode == 1 (general copy error), retry. It may be a temporary network incident
				if retcode == 1 {
					// File may exist on target, try to remote it
					_, _, _, err = host.Run(voidtask, fmt.Sprintf("sudo rm -f %s", remotepath), temporal.GetBigDelay(), temporal.GetExecutionTimeout())
					if err == nil {
						return fmt.Errorf("file may exist on remote with inappropriate access rights, deleted it and retrying")
					}
					// If submission of removal of remote file fails, stop the retry and consider this as an unrecoverable network error
					return retry.StopRetryError("an unrecoverable network error has occurred", err)
				}
				if system.IsSCPRetryable(retcode) {
					err = fmt.Errorf("failed to copy file '%s' to '%s:%s' (retcode: %d=%s)", localpath, host.Name(), remotepath, retcode, system.SCPErrorString(retcode))
					return err
				}
				return nil
			}
			return nil
		},
		temporal.GetDefaultDelay(),
		temporal.GetLongOperationTimeout(),
	)
	if retryErr != nil {
		switch realErr := retryErr.(type) { // nolint
		case *retry.ErrStopRetry:
			return scerr.Wrap(realErr.Cause(), "failed to copy file to remote host '%s'", host.Name())
		case *retry.ErrTimeout:
			return scerr.Wrap(realErr, "timeout trying to copy temporary file to '%s:%s'", host.Name(), remotepath)
		}
		return retryErr
	}

	cmd := ""
	if owner != "" {
		cmd += `sudo chown ` + owner + ` "` + remotepath + `";`
	}
	if mode != "" {
		cmd += `sudo chmod ` + mode + ` "` + remotepath + `"`
	}

	var innerErr error
	retryErr = retry.WhileUnsuccessful(
		func() error {
			var retcode int
			retcode, _, _, innerErr = host.Run(voidtask, cmd, temporal.GetDefaultDelay(), temporal.GetExecutionTimeout())
			if innerErr != nil {
				return innerErr
			}
			if retcode != 0 {
				innerErr = scerr.NewError(fmt.Sprintf("failed to change rights of file '%s:%s' (retcode=%d)", host.Name(), remotepath, retcode), nil, nil)
				return nil
			}
			return nil
		},
		temporal.GetMinDelay(),
		temporal.GetContextTimeout(),
	)
	if retryErr != nil {
		switch retryErr.(type) {
		case retry.ErrTimeout:
			return scerr.Wrap(innerErr, "timeout trying to change rights of file '%s' on host '%s'", remotepath, host.Name())
		default:
			return scerr.Wrap(retryErr, "failed to change rights of file '%s' on host '%s'", remotepath, host.Name())
		}
	}
	return nil
}

// UploadStringToRemoteFile creates a file 'filename' on remote 'host' with the content 'content'
func UploadStringToRemoteFile(content string, host resources.Host, filename string, owner, mode string) error {
	if content == "" {
		return scerr.InvalidParameterError("content", "cannot be empty string")
	}
	if host == nil {
		return scerr.InvalidParameterError("host", "cannot be nil")
	}
	if filename == "" {
		return scerr.InvalidParameterError("filename", "cannot be empty string")
	}

	if forensics := os.Getenv("SAFESCALE_FORENSICS"); forensics != "" {
		_ = os.MkdirAll(utils.AbsPathify(fmt.Sprintf("$HOME/.safescale/forensics/%s", host.Name)), 0777)
		partials := strings.Split(filename, "/")
		dumpName := utils.AbsPathify(fmt.Sprintf("$HOME/.safescale/forensics/%s/%s", host.Name, partials[len(partials)-1]))

		err := ioutil.WriteFile(dumpName, []byte(content), 0644)
		if err != nil {
			logrus.Warnf("[TRACE] Forensics error creating %s", dumpName)
		}
	}

	f, err := system.CreateTempFileFromString(content, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %s", err.Error())
	}

	err = UploadFile(f.Name(), host, filename, owner, mode)
	_ = os.Remove(f.Name())
	return err
}

// realizeVariables replaces in every variable any template
func realizeVariables(variables data.Map) (data.Map, error) {
	cloneV := variables.Clone()

	for k, v := range cloneV {
		if variable, ok := v.(string); ok {
			varTemplate, err := template.New("realize_var").Parse(variable)
			if err != nil {
				return nil, fmt.Errorf("error parsing variable '%s': %s", k, err.Error())
			}
			buffer := bytes.NewBufferString("")
			err = varTemplate.Execute(buffer, variables)
			if err != nil {
				return nil, err
			}
			cloneV[k] = buffer.String()
		}
	}

	return cloneV, nil
}

func replaceVariablesInString(text string, v data.Map) (string, error) {
	tmpl, err := template.New("text").Parse(text)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %s", err.Error())
	}
	dataBuffer := bytes.NewBufferString("")
	err = tmpl.Execute(dataBuffer, v)
	if err != nil {
		return "", fmt.Errorf("failed to replace variables: %s", err.Error())
	}
	return dataBuffer.String(), nil
}
