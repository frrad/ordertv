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

	sCompiled, eCompiled, err := compileRules(globalRuleList)

	fileDirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d files and directories.\n", len(fileDirs))
	log.Println("Processing...")

	for _, filedir := range fileDirs {
		if filedir.IsDir() {
			name, s, match, err := classifyDir(filedir, sCompiled)
			if err != nil {
				log.Fatalf("Error classifying dir %s: %v", filedir.Name(), err)
			}

			if match {
				log.Println("Classified:", name, s)
			} else {
				log.Println("Couldn't classify:", filedir.Name())
			}

		} else {
			name, s, e, match, err := classifyFile(filedir, eCompiled)
			if err != nil {
				log.Fatalf("Error classifying file %s: %v", filedir.Name(), err)
			}

			if match {
				log.Println("Classified:", name, s, e)
			} else {
				log.Println("Couldn't Classify", filedir.Name())
			}
		}

	}
}

func classifyDir(dirInfo os.FileInfo, ruleGroups []*compiledRules) (string, int, bool, error) {
	showName := ""
	season := -1

	name := dirInfo.Name()

	for _, grp := range ruleGroups {
		for _, rule := range grp.regexps {
			seasonStrs := rule.FindStringSubmatch(name)
			if len(seasonStrs) == 0 {
				continue
			}
			if len(seasonStrs) != 2 {
				return "", -1, false, fmt.Errorf("Wrong number of match groups: %v", seasonStrs)
			}

			if showName != "" && showName != grp.name {
				return "", -1, false, fmt.Errorf("Matched rules for two different shows: %s and %s", showName, grp.name)
			}
			showName = grp.name
			season32, err := strconv.ParseInt(seasonStrs[1], 10, 32)
			if err != nil {
				return "", -1, false, fmt.Errorf("Can't parse int %s", seasonStrs[1])
			}

			season = int(season32)
		}
	}

	if showName == "" || season == -1 {
		return "", -1, false, nil
	}

	return showName, season, true, nil

}

func classifyFile(fileInfo os.FileInfo, ruleGroups []*compiledRules) (string, int, int, bool, error) {

	name := fileInfo.Name()

	showName := ""
	season := -1
	episode := -1

	for _, grp := range ruleGroups {
		for _, rule := range grp.regexps {

			epStrs := rule.FindStringSubmatch(name)
			if len(epStrs) == 0 {
				continue
			}
			if len(epStrs) != 3 {
				return showName, season, episode, false, fmt.Errorf("Partial episode rule match %v", epStrs)

			}

			// now we have a real match ...

			if showName != "" && showName != grp.name {
				return "", -1, -1, false, fmt.Errorf("mismatch showname: %s, %s", showName, grp.name)
			}
			showName = grp.name

			x, err := strconv.ParseInt(epStrs[1], 10, 32)
			if err != nil {
				return "", -1, -1, false, fmt.Errorf("Can't parse int %s", epStrs[1])
			}
			seas := int(x)

			if season != -1 && season != seas {
				return "", -1, -1, false, fmt.Errorf("mismatch season: %d, %d", season, seas)
			}
			season = seas

			x, err = strconv.ParseInt(epStrs[2], 10, 32)
			if err != nil {
				return "", -1, -1, false, fmt.Errorf("Can't parse int %s", epStrs[2])
			}
			ep := int(x)

			if episode != -1 && episode != ep {
				return "", -1, -1, false, fmt.Errorf("mismatch episode: %d, %d", episode, ep)
			}
			episode = ep

		}
	}

	if showName == "" || season == -1 || episode == -1 {
		return "", -1, -1, false, nil
	}

	return showName, season, episode, true, nil
}
