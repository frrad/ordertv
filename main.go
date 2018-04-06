package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type dirRule struct {
	regex *regexp.Regexp
	name  string
	score int
}

func makeDirRules(names []string) []dirRule {
	generators := []func(string) dirRule{
		simpleGen,
		genTwo,
	}

	return generate(generators, names)
}

func generate(generators []func(string) dirRule, names []string) []dirRule {
	ans := []dirRule{}
	for _, name := range names {
		for _, gen := range generators {
			ans = append(ans, gen(name))
		}
	}

	return ans
}

func makeFileRules(names []string) []dirRule {
	generators := []func(string) dirRule{
		easyFileGen,
	}

	return generate(generators, names)
}

func simpleGen(name string) dirRule {
	clean := strings.Replace(name, " ", ".", -1)

	thisRe := regexp.MustCompile("(?i)^" + clean + ".*S([0-9]+).*$")

	return dirRule{
		regex: thisRe,
		name:  name,
		score: 10,
	}
}

func easyFileGen(name string) dirRule {
	clean := strings.Replace(name, " ", ".", -1)

	thisRe := regexp.MustCompile("(?i)^" + clean + ".*S([0-9]+).*$")

	return dirRule{
		regex: thisRe,
		name:  name,
		score: 10,
	}
}

func genTwo(name string) dirRule {
	clean := strings.Replace(name, " ", ".", -1)

	thisRe := regexp.MustCompile("(?i)" + clean + ".*S([0-9]+).*$")

	return dirRule{
		regex: thisRe,
		name:  name,
		score: 5,
	}
}

func main() {

	dirRules := makeDirRules(dirNames)
	fileRules := makeFileRules(dirNames)

	path := "/data/btn-dump"
	fileDirs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d files and directories.\n", len(fileDirs))
	log.Println("Processing...")

	for _, filedir := range fileDirs {
		if filedir.IsDir() {
			name, season, err := classifyDir(filedir, dirRules)
			if err != nil {
				log.Fatalf("Error classifying dir %s: %v", filedir.Name(), err)
			}

			log.Println("Classified:", name, season)

		} else {
			name, season, ep, err := classifyFile(filedir, fileRules)
			if err != nil {
				log.Fatalf("Error classifying file %s: %v", filedir.Name(), err)
			}

			log.Println("Classified:", name, season, ep)
		}

	}
}

func classifyDir(dirInfo os.FileInfo, rules []dirRule) (string, int, error) {
	showName := ""
	season := -1

	name := dirInfo.Name()

	if name == "S02" {
		return "Gossip Girl", 2, nil
	}

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

func classifyFile(fileInfo os.FileInfo, rules []dirRule) (string, int,
	int, error) {

	showName := ""
	season := -1
	episode := -1

	name := fileInfo.Name()

	log.Printf("Processing File: %s\n", name)
	for _, rule := range rules {
		seasonStrs := rule.regex.FindStringSubmatch(name)
		if len(seasonStrs) == 0 {
			continue
		}

	}

	if showName == "" || season == -1 || episode == -1 {
		return showName, season, episode, fmt.Errorf("No rule matched file %s", name)
	}

	return showName, season, episode, fmt.Errorf("No rule matched file %s", name)

}