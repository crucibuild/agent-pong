/* Copyright (C) 2016 Christophe Camel, Jonathan Pigrée
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/crucibuild/agent-pong/schema"
	"github.com/crucibuild/sdk-agent-go/agentiface"
	"github.com/crucibuild/sdk-agent-go/agentimpl"
)

var Resources http.FileSystem

type PongAgent struct {
	*agentimpl.Agent
}

func MustOpenResources(path string) []byte {
	file, err := Resources.Open(path)

	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadAll(file)

	if err != nil {
		panic(err)
	}

	return content
}

func NewPongAgent() (agentiface.Agent, error) {
	var agentSpec map[string]interface{}

	manifest := MustOpenResources("/resources/manifest.json")

	err := json.Unmarshal(manifest, &agentSpec)

	if err != nil {
		return nil, err
	}

	impl, err := agentimpl.NewAgent(agentimpl.NewManifest(agentSpec))

	if err != nil {
		return nil, err
	}

	agent := &PongAgent{
		impl,
	}

	if err := agent.init(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *PongAgent) register(rawAvroSchema string) error {
	s, err := agentimpl.LoadAvroSchema(rawAvroSchema, a)
	if err != nil {
		return err
	}

	a.SchemaRegister(s)

	return nil
}

func (a *PongAgent) init() error {
	// register schemas:
	var content []byte
	content = MustOpenResources("/schema/header.avro")
	if err := a.register(string(content[:])); err != nil {
		return err
	}

	content = MustOpenResources("/schema/test-command.avro")
	if err := a.register(string(content[:])); err != nil {
		return err
	}

	content = MustOpenResources("/schema/tested-event.avro")
	if err := a.register(string(content[:])); err != nil {
		return err
	}

	// register types
	// register types
	if _, err := a.TypeRegister(agentimpl.NewTypeFromType("crucibuild/agent-pong#tested-event", schema.TestedEventType)); err != nil {
		return err
	}
	if _, err := a.TypeRegister(agentimpl.NewTypeFromType("crucibuild/agent-pong#test-command", schema.TestCommandType)); err != nil {
		return err
	}

	// register state callback
	a.RegisterStateCallback(a.onStateChange)

	return nil
}

func (a *PongAgent) onStateChange(state agentiface.State) error {
	switch state {
	case agentiface.StateConnected:
		if _, err := a.RegisterCommandCallback("crucibuild/agent-pong#test-command", a.onTestCommand); err != nil {
			return err
		}
	}
	return nil

}

func (a *PongAgent) onTestCommand(ctx agentiface.CommandCtx) error {
	cmd := ctx.Message().(*schema.TestCommand)

	a.Info(fmt.Sprintf("Received test-command: '%s' '%s' '%d' ", cmd.Foo.Z, cmd.Value, cmd.X))

	// reply with a tested event

	return ctx.SendEvent(&schema.TestedEvent{Value: "pong"})
}
