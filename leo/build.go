/*
 * Copyright 2018-2020 the original author or authors.
 *
 * COPY FROM paketo-community/rustup.git
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package leo

import (
	"fmt"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, nil)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	// create a second time so the configuration is only printed after we know the buildpack should run
	cr, err = libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	// install leo
	v, _ := cr.Resolve("BP_LEO_VERSION")
	libc, _ := cr.Resolve("BP_LEO_LIBC")

	leoDependency, err := dr.Resolve(fmt.Sprintf("leo-%s", libc), v)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
	}

	leoLayer := NewLeo(leoDependency, dc)
	leoLayer.Logger = b.Logger

	enableProcess, _ := cr.Resolve("BP_ENABLE_LEO_PROCESS")
	result.Processes, err = leoLayer.BuildProcessTypes(enableProcess)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to build list of process types\n%w", err)
	}
	result.Layers = append(result.Layers, leoLayer)

	return result, nil
}
