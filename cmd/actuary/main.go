package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/diogomonica/actuary/audit"
	"github.com/diogomonica/actuary/audit/container/images"
	"github.com/diogomonica/actuary/audit/container/runtime"
	"github.com/diogomonica/actuary/audit/dockerconf"
	"github.com/diogomonica/actuary/audit/dockerfiles"
	"github.com/diogomonica/actuary/audit/dockerhost"
	"github.com/diogomonica/actuary/audit/dockersecops"
	"github.com/diogomonica/actuary/oututils"
	"github.com/diogomonica/actuary/profileutils"
	"github.com/docker/engine-api/client"
)

var profile = flag.String("profile", "", "Actuary profile file path")
var output = flag.String("output", "", "output filename")
var outputType = flag.String("type", "json", "output type - XML or JSON")
var tomlProfile profileutils.Profile
var clientHeaders map[string]string
var results []audit.Result
var actions map[string]audit.Check

func init() {
	flag.StringVar(profile, "f", "", "Actuary profile file path")
	flag.StringVar(output, "o", "", "output filename")
	flag.StringVar(outputType, "", "json", "output type - XML or JSON")

	clientHeaders = make(map[string]string)
	clientHeaders["User-Agent"] = "engine-api-cli-1.0"
}

func main() {
	var cmdArgs []string
	var hash string
	var auditName string

	flag.Parse()
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.20", nil, clientHeaders)
	if err != nil {
		log.Fatalf("Unable to connect to Docker daemon: %s", err)
	}

	cmdArgs = flag.Args()
	if len(cmdArgs) == 1 {
		hash = cmdArgs[0]
		tomlProfile, err = profileutils.GetFromURL(hash)
		if err != nil {
			log.Fatalf("Unable to fetch profile. Exiting...")
		}
	} else if len(cmdArgs) == 0 {
		_, err := os.Stat(*profile)
		if os.IsNotExist(err) {
			log.Fatalf("Invalid profile path: %s", *profile)
		}
		tomlProfile = profileutils.GetFromFile(*profile)
	} else {
		log.Fatalf("Unsupported number of arguments. Use -h for help")
	}

	//loop through the audits
	for category := range tomlProfile.Audit {
		switch auditName = tomlProfile.Audit[category].Name; auditName {
		case "Host Configuration":
			actions = dockerhost.GetAuditDefinitions()
		case "Docker daemon configuration":
			actions = dockerconf.GetAuditDefinitions()
		case "Docker daemon configuration files":
			actions = dockerfiles.GetAuditDefinitions()
		case "Container Images and Build File":
			actions = images.GetAuditDefinitions()
		case "Container Runtime":
			actions = runtime.GetAuditDefinitions()
		case "Docker Security Operations":
			actions = dockersecops.GetAuditDefinitions()
		default:
			log.Panicf("No audit category named: %s", auditName)
			continue
		}
		log.Printf("Running Audit: %s", auditName)
		checks := tomlProfile.Audit[category].Checklist
		//cross-reference checks
		for _, check := range checks {
			if _, ok := actions[check]; ok {
				res := actions[check](cli)
				results = append(results, res)
				oututils.ConsolePrint(res)
			} else {
				log.Panicf("No check named %s", check)
			}
		}
	}

	if *output != "" {
		rep := oututils.CreateReport(*output)
		rep.Results = results
		switch strings.ToLower(*outputType) {
		case "json":
			rep.WriteJSON()
		case "xml":
			rep.WriteXML()
		}
	}
}