// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-swagger/go-swagger/generator"
	"github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
)

type shared struct {
	Spec          flags.Filename `long:"spec" short:"f" description:"the spec file to use" default:"./swagger.json"`
	APIPackage    string         `long:"api-package" short:"a" description:"the package to save the operations" default:"operations"`
	ModelPackage  string         `long:"model-package" short:"m" description:"the package to save the models" default:"models"`
	ServerPackage string         `long:"server-package" short:"s" description:"the package to save the server specific code" default:"restapi"`
	ClientPackage string         `long:"client-package" short:"c" description:"the package to save the client specific code" default:"client"`
	Target        flags.Filename `long:"target" short:"t" default:"./" description:"the base directory for generating the files"`
	TemplateDir   flags.Filename `long:"template-dir" short:"T" description:"alternative template override directory"`
	ConfigFile    flags.Filename `long:"config-file" short:"C" description:"configuration file to use for overriding template options"`
}

// Server the command to generate an entire server application
type Server struct {
	shared
	Name           string   `long:"name" short:"A" description:"the name of the application, defaults to a mangled value of info.title"`
	Operations     []string `long:"operation" short:"O" description:"specify an operation to include, repeat for multiple"`
	Tags           []string `long:"tags" description:"the tags to include, if not specified defaults to all"`
	Principal      string   `long:"principal" short:"P" description:"the model to use for the security principal"`
	DefaultScheme  string   `long:"default-scheme" description:"the default scheme for this API" default:"http"`
	Models         []string `long:"model" short:"M" description:"specify a model to include, repeat for multiple"`
	SkipModels     bool     `long:"skip-models" description:"no models will be generated when this flag is specified"`
	SkipOperations bool     `long:"skip-operations" description:"no operations will be generated when this flag is specified"`
	SkipSupport    bool     `long:"skip-support" description:"no supporting files will be generated when this flag is specified"`
	ExcludeMain    bool     `long:"exclude-main" description:"exclude main function, so just generate the library"`
	ExcludeSpec    bool     `long:"exclude-spec" description:"don't embed the swagger specification"`
	WithContext    bool     `long:"with-context" description:"handlers get a context as first arg"`
	DumpData       bool     `long:"dump-data" description:"when present dumps the json for the template generator instead of generating files"`
}

// Execute runs this command
func (s *Server) Execute(args []string) error {
	var cfg *viper.Viper
	if string(s.ConfigFile) != "" {
		apt, err := filepath.Abs(string(s.ConfigFile))
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("trying to read config from", apt)
		v, err := generator.ReadConfig(apt)
		if err != nil {
			return err
		}
		cfg = v
	}
	if os.Getenv("DEBUG") != "" || os.Getenv("SWAGGER_DEBUG") != "" {
		if cfg != nil {
			cfg.Debug()
		} else {
			log.Println("NO config read")
		}
	}

	opts := &generator.GenOpts{
		Spec:              string(s.Spec),
		Target:            string(s.Target),
		APIPackage:        s.APIPackage,
		ModelPackage:      s.ModelPackage,
		ServerPackage:     s.ServerPackage,
		ClientPackage:     s.ClientPackage,
		Principal:         s.Principal,
		DefaultScheme:     s.DefaultScheme,
		IncludeModel:      !s.SkipModels,
		IncludeValidator:  !s.SkipModels,
		IncludeHandler:    !s.SkipOperations,
		IncludeParameters: !s.SkipOperations,
		IncludeResponses:  !s.SkipOperations,
		IncludeMain:       !s.ExcludeMain,
		IncludeSupport:    !s.SkipSupport,
		ExcludeSpec:       s.ExcludeSpec,
		TemplateDir:       string(s.TemplateDir),
		WithContext:       s.WithContext,
		DumpData:          s.DumpData,
		Models:            s.Models,
		Operations:        s.Operations,
		Tags:              s.Tags,
		Name:              s.Name,
	}

	if err := opts.EnsureDefaults(false); err != nil {
		return err
	}

	if cfg != nil {
		var def generator.LanguageDefinition
		if err := cfg.Unmarshal(&def); err != nil {
			return err
		}
		def.ConfigureOpts(opts)
	}

	if err := generator.GenerateServer(s.Name, s.Models, s.Operations, opts); err != nil {
		return err
	}

	rp, err := filepath.Rel(".", opts.Target)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, `Generation completed!

For this generation to compile you need to have some packages in your GOPATH:

  * github.com/go-openapi/runtime
  * github.com/tylerb/graceful
  * github.com/jessevdk/go-flags
  * golang.org/x/net/context

You can get these now with: go get -u -f %s/...
`, rp)

	return nil
}
