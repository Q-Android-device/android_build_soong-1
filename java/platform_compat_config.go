// Copyright 2019 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package java

import (
	"android/soong/android"
)

func init() {
	android.RegisterModuleType("platform_compat_config", platformCompatConfigFactory)
}

type platformCompatConfigProperties struct {
	Src *string `android:"path"`
}

type platformCompatConfig struct {
	android.ModuleBase

	properties     platformCompatConfigProperties
	installDirPath android.InstallPath
	configFile     android.OutputPath
}

func (p *platformCompatConfig) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	rule := android.NewRuleBuilder()

	configFileName := p.Name() + ".xml"
	p.configFile = android.PathForModuleOut(ctx, configFileName).OutputPath
	path := android.PathForModuleSrc(ctx, String(p.properties.Src))

	// Use the empty config if the compat config file idoesn't exist (can happen if @ChangeId
	// annotation is not used).
	emptyConfig := `'<?xml version="1.0" encoding="UTF-8" standalone="no"?><config/>'`
	configPath := `compat/compat_config.xml`

	rule.Command().
		Text(`unzip`).
		Flag(`-l`).
		Input(path).
		Text(`| grep`).
		Flag(`-q`).
		Text(configPath).
		Text(`; if [ "$?" = "0" ] ; then`).
		Text(`unzip`).
		Flag(`-qp`).
		Input(path).
		Text(configPath).
		Text(`>`).
		Output(p.configFile).
		Text(`; else echo `).
		Text(emptyConfig).
		Text(`>`).
		Output(p.configFile).
		Text(`; fi`)

	p.installDirPath = android.PathForModuleInstall(ctx, "etc", "compatconfig")
	rule.Build(pctx, ctx, configFileName, "Extract compat/compat_config.xml and install it")

}

func (p *platformCompatConfig) AndroidMkEntries() android.AndroidMkEntries {
	return android.AndroidMkEntries{
		Class:      "ETC",
		OutputFile: android.OptionalPathForPath(p.configFile),
		Include:    "$(BUILD_PREBUILT)",
		ExtraEntries: []android.AndroidMkExtraEntriesFunc{
			func(entries *android.AndroidMkEntries) {
				entries.SetString("LOCAL_MODULE_PATH", "$(OUT_DIR)/"+p.installDirPath.RelPathString())
				entries.SetString("LOCAL_INSTALLED_MODULE_STEM", p.configFile.Base())
			},
		},
	}
}

func platformCompatConfigFactory() android.Module {
	module := &platformCompatConfig{}
	module.AddProperties(&module.properties)
	android.InitAndroidArchModule(module, android.DeviceSupported, android.MultilibFirst)
	return module
}
