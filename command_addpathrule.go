package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/marcacohen/gcslock"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

type (
	loadbalancerURLMap struct {
		DefaultService string `yaml:"defaultService"`
		Description    string `yaml:"description"`
		HostRules      []struct {
			Hosts       []string `yaml:"hosts"`
			PathMatcher string   `yaml:"pathMatcher"`
		} `yaml:"hostRules"`
		Kind         string        `yaml:"kind"`
		Name         string        `yaml:"name"`
		PathMatchers []pathMatcher `yaml:"pathMatchers"`
		Region       string        `yaml:"region"`
		SelfLink     string        `yaml:"selfLink"`
	}
	pathMatcher struct {
		DefaultService string            `yaml:"defaultService"`
		Description    string            `yaml:"description"`
		Name           string            `yaml:"name"`
		PathRules      []pathsAndService `yaml:"pathRules"`
	}
	pathsAndService struct {
		Paths   []string `yaml:"paths"`
		Service string   `yaml:"service"`
	}
)

// findIndices returns (matcherIndex, ruleIndex )
func (m *loadbalancerURLMap) findIndices(pathMatcherName, fqnService string, verbose bool) (int, int) {
	ruleIndex := -1
	matcherIndex := -1
	for m, each := range m.PathMatchers {
		if each.Name == pathMatcherName {
			if verbose {
				log.Println("found existing path matcher:", pathMatcherName)
			}
			matcherIndex = m
		}
		for i, other := range each.PathRules {
			if other.Service == fqnService {
				if verbose {
					log.Println("found existing path rule:", fqnService)
				}
				ruleIndex = i
			}
		}
	}
	return matcherIndex, ruleIndex
}

func (m *loadbalancerURLMap) patchPathsAndService(isRemove bool, pathMatcherName, fqnService string, paths []string, verbose bool) error {
	// check if exists based on service
	matcherIndex, ruleIndex := m.findIndices(pathMatcherName, fqnService, verbose)
	if matcherIndex == -1 {
		err := fmt.Errorf("no path-matcher found with name [%s]", pathMatcherName)
		printError(err.Error())
		return errAbort
	}
	if isRemove {
		// Delete
		rules := m.PathMatchers[matcherIndex].PathRules
		copy(rules[ruleIndex:], rules[ruleIndex+1:])
		rules[len(rules)-1] = pathsAndService{}
		rules = rules[:len(rules)-1]
		m.PathMatchers[matcherIndex].PathRules = rules
	} else {
		// Update
		toPatch := pathsAndService{Service: fqnService, Paths: paths}
		if ruleIndex == -1 {
			// add new path rule set
			rules := m.PathMatchers[matcherIndex].PathRules
			m.PathMatchers[matcherIndex].PathRules = append(rules, toPatch)
		} else {
			// replace existing path rule set
			m.PathMatchers[matcherIndex].PathRules[ruleIndex] = toPatch
		}
	}
	return nil
}

func cmdAddPathRulesToPathMatcher(c *cli.Context) error {
	return patchPathRulesForPathMatcher(c, false)
}

func cmdRemovePathRulesFromPathMatcher(c *cli.Context) error {
	return patchPathRulesForPathMatcher(c, true)
}

func patchPathRulesForPathMatcher(c *cli.Context, isRemove bool) error {
	// prepare
	mtx, err := getMigrationContext(c)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	verbose := c.GlobalBool("v")
	urlMapName := c.String("url-map")
	// aquire lock
	lockObjectName := fmt.Sprintf("project-%s-region-%s-urlmap-%s-gmig-lock", mtx.config().Project, mtx.config().Region, urlMapName)
	urlMapMutex, err := gcslock.New(context.Background(), mtx.config().Bucket, lockObjectName)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	if verbose {
		log.Println("acquire global lock on:", lockObjectName, " in bucket:", mtx.config().Bucket)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute) // wait for at most 1 minute
	defer cancel()
	if err := urlMapMutex.ContextLock(ctx); err != nil {
		printError(err.Error())
		return errAbort
	}
	// release lock on return
	defer func() {
		if verbose {
			log.Println("release global lock on:", lockObjectName, " in bucket:", mtx.config().Bucket)
		}
		if err := urlMapMutex.ContextUnlock(ctx); err != nil {
			printError(err.Error())
		}
	}()
	// export
	args := []string{"compute", "url-maps", "export", urlMapName, "--region", mtx.config().Region}
	cmd := exec.Command("gcloud", args...)
	if verbose {
		log.Println(strings.Join(append([]string{"gcloud"}, args...), " "))
	}
	data, err := runCommand(cmd)
	if err != nil {
		log.Println(string(data)) // stderr
		return fmt.Errorf("failed to run gcloud command: %v", err)
	}
	// unmarshal
	urlMap := new(loadbalancerURLMap)
	if err := yaml.Unmarshal(data, urlMap); err != nil {
		printError(err.Error())
		return errAbort
	}
	serviceName := c.String("service")
	fqnService := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/backendServices/%s",
		mtx.config().Project,
		mtx.config().Region,
		serviceName)
	if err := urlMap.patchPathsAndService(
		isRemove,
		c.String("path-matcher"), fqnService, strings.Split(c.String("paths"), ","),
		true); err != nil {
		return err
	}
	// can only import from source file
	importdata, err := yaml.Marshal(urlMap)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	source := filepath.Join(os.TempDir(), "patchPathRulesForPathMatcher.yaml")
	err = ioutil.WriteFile(source, importdata, os.ModePerm)
	if err != nil {
		printError(err.Error())
		return errAbort
	}
	defer os.Remove(source)
	// import
	{
		args := []string{"compute", "url-maps", "import", urlMapName, "--source", source, "--region", mtx.config().Region}
		cmd := exec.Command("gcloud", args...)
		if verbose {
			log.Println(strings.Join(append([]string{"gcloud"}, args...), " "))
		}
		data, err := runCommand(cmd)
		if err != nil {
			log.Println(string(data)) // stderr
			return fmt.Errorf("failed to run gcloud command: %v", err)
		}
	}
	return nil
}
