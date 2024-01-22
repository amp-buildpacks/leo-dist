// Copyright (c) The Amphitheatre Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package leo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

// Leo will handle installing the `leo` tool & adding it to the PATH
type Leo struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
	Executor         effect.Executor
}

func NewLeo(dependency libpak.BuildpackDependency, cache libpak.DependencyCache) Leo {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Build:  true,
		Cache:  true,
		Launch: true,
	})
	return Leo{
		LayerContributor: contributor,
		Executor:         effect.NewExecutor(),
	}
}

func (r Leo) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	r.LayerContributor.Logger = r.Logger
	return r.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		bin := filepath.Join(layer.Path, "bin")

		r.Logger.Bodyf("Expanding %s to %s", artifact.Name(), bin)
		if err := crush.Extract(artifact, bin, 0); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand %s\n%w", artifact.Name(), err)
		}

		file := filepath.Join(bin, PlanEntryLeo)
		r.Logger.Bodyf("Setting %s as executable", file)
		if err := os.Chmod(file, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to chmod %s\n%w", file, err)
		}

		r.Logger.Bodyf("Setting %s in PATH", bin)
		if err := os.Setenv("PATH", sherpa.AppendToEnvVar("PATH", ":", bin)); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to set $PATH\n%w", err)
		}

		// compile contract
		r.Logger.Bodyf("Compiling contracts by '%s build'", PlanEntryLeo)
		if _, err := r.Execute(PlanEntryLeo, []string{"build"}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to compile contract\n%w", err)
		}

		// get version
		buf, err := r.Execute(PlanEntryLeo, []string{"-V"})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to get %s version: %s\n%w", PlanEntryLeo, buf.String(), err)
		}
		ver := strings.Split(strings.TrimSpace(buf.String()), " ")
		r.Logger.Bodyf("Checking %s version: %s", PlanEntryLeo, ver[1])

		return layer, nil
	})
}

func (r Leo) Execute(command string, args []string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if err := r.Executor.Execute(effect.Execution{
		Command: command,
		Args:    args,
		Stdout:  buf,
		Stderr:  buf,
	}); err != nil {
		return buf, fmt.Errorf("%s: %w", buf.String(), err)
	}
	return buf, nil
}

func (r Leo) Name() string {
	return r.LayerContributor.LayerName()
}
