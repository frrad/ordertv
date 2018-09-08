package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
)

type ruleset struct {
	name   string
	sRules []string
	eRules []string
}

type compiledRules struct {
	name    string
	regexps []*regexp.Regexp
}


func compileList(strings []string, name string) (*compiledRules, error) {
	regexps := make([]*regexp.Regexp, len(strings))
	for i, sRule := range strings {
		regex, err := regexp.Compile(sRule)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex: %s", sRule)
		}
		regexps[i] = regex
	}

	return &compiledRules{
		name:    name,
		regexps: regexps,
	}, nil
}

func compileRules(allRules []ruleset) ([]*compiledRules, []*compiledRules, error) {
	sRules := make([]*compiledRules, len(allRules))
	eRules := make([]*compiledRules, len(allRules))

	for i, ruleset := range allRules {
		compR, err := compileList(ruleset.sRules, ruleset.name)
		if err != nil {
			return nil, nil, err
		}
		sRules[i] = compR

		compR, err = compileList(ruleset.eRules, ruleset.name)
		if err != nil {
			return nil, nil, err
		}
		eRules[i] = compR
	}
	return sRules, eRules, nil
}

func main() {

	path := "/data/btn-dump"

	rules := []ruleset{
		{
			name:   "Butt",
			sRules: []string{""},
  			eRules: []string{""},
		},
		{
			name:   "Face",
			sRules: []string {""},
			eRules: []string {""},
		},
	}

	sCompiled, eCompiled, err := compileRules(rules)

	fileDirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d files and directories.\n", len(fileDirs))
	log.Println("Processing...")

	for _, filedir := range fileDirs {
		if filedir.IsDir() {
			name, s, err := classifyDir(filedir, sCompiled)
			if err != nil {
				log.Fatalf("Error classifying dir %s: %v", filedir.Name(), err)
			}

			log.Println("Classified:", name, s)

		} else {
			name, s, e, err := classifyFile(filedir, eCompiled)
			if err != nil {
				log.Fatalf("Error classifying file %s: %v", filedir.Name(), err)
			}

			log.Println("Classified:", name, s, e)
		}

	}
}

func classifyDir(dirInfo os.FileInfo, rules []*compiledRules) (string, int, error) {
	showName := ""
	season := -1

	name := dirInfo.Name()

	log.Printf("Processing dir: %s\n", name)
	for _, rule := range rules {
		seasonStrs := rule.regex.FindStringSubmatch(name)
		if len(seasonStrs) == 0 {
			continue
		}
		if len(seasonStrs) != 2 {
			return "", 0, fmt.Errorf("Wrong number of match groups: %v", seasonStrs)
		}

		if showName != "" && showName != rule.name {
			return "", 0, fmt.Errorf("Matched rules for two different shows: %s and %s", showName, rule.name)
		}
		showName = rule.name
		season32, err := strconv.ParseInt(seasonStrs[1], 10, 32)
		if err != nil {
			return "", 0, fmt.Errorf("Can't parse int %s", seasonStrs[1])
		}

		season = int(season32)
	}

	if showName == "" || season == -1 {
		return showName, season, fmt.Errorf("No rule matched dir %s", name)
	}

	return showName, season, nil

}

func classifyFile(fileInfo os.FileInfo, rules []dirRule, skips []*regexp.Regexp) (string, int, int, error) {

	name := fileInfo.Name()

	for _, skip := range skips {
		if skip.MatchString(name) {
			return "", -1, -1, nil
		}
	}

	showName := ""
	season := -1
	episode := -1

	log.Printf("Processing File: %s\n", name)
	for _, rule := range rules {
		epStrs := rule.regex.FindStringSubmatch(name)
		if len(epStrs) == 0 {
			continue
		}
		if len(epStrs) != 3 {
			return showName, season, episode, fmt.Errorf("Partial episode rule match %v", epStrs)

		}

		// now we have a real match ...

		fmt.Println(epStrs)

	}

	if showName == "" || season == -1 || episode == -1 {
		return showName, season, episode, fmt.Errorf("No rule matched file %s", name)
	}

	return showName, season, episode, fmt.Errorf("No rule matched file %s", name)

}
