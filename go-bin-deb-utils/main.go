// go-bin-deb-utils is a cli tool to generate debian package and repos.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {

	flag.Parse()
	action := flag.Arg(0)

	// basic arg parsing
	var reposlug string
	var email string
	var ghToken string
	var version string
	var archs string
	var out string

	flag.StringVar(&reposlug, "repo", "", "The repo slug such USER/REPO.")
	flag.StringVar(&ghToken, "ghToken", "", "The ghToken to write on your repository.")
	flag.StringVar(&email, "email", "", "Your gh email.")
	flag.StringVar(&version, "version", "", "The package version.")
	flag.StringVar(&archs, "archs", "386,amd64", "The archs to build.")
	flag.StringVar(&out, "out", "", "The out build directory.")
	push := flag.Bool("push", false, "Push the new assets")
	flag.CommandLine.Parse(os.Args[2:])

	// os.Env fallback
	if email == "" {
		email = os.Getenv("EMAIL")
	}
	if email == "" {
		email = os.Getenv("MYEMAIL")
	}
	if reposlug == "" {
		reposlug = os.Getenv("REPO")
	}
	if ghToken == "" {
		ghToken = os.Getenv("GH_TOKEN")
	}

	// ci fallback
	// todo: make use of pre defined ci env
	if isTravis() {
		if version == "" {
			version = os.Getenv("TRAVIS_TAG")
		}
		if out == "" {
			out = os.Getenv("TRAVIS_BUILD_DIR")
		}
	}
	if isVagrant() {
		if version == "" {
			version = os.Getenv("VERSION")
		}
		if out == "" {
			out = os.Getenv("BUILD_DIR")
		}
	}

	// integrity check
	requireArg(reposlug, "repo", "REPO")
	requireArg(ghToken, "ghToken", "GH_TOKEN")
	requireArg(email, "email", "EMAIL", "MYEMAIL")
	if isTravis() {
		requireArg(version, "version", "TRAVIS_TAG")
		requireArg(out, "out", "TRAVIS_BUILD_DIR")
	} else if isVagrant() {
		requireArg(version, "version", "VERSION")
		requireArg(out, "out", "BUILD_DIR")
	} else {
		panic("nop, no such ci system...")
	}

	// execute some common setup, in case.
	alwaysHide[ghToken] = "$GH_TOKEN"
	os.RemoveAll(out)
	os.MkdirAll(out, os.ModePerm)
	if version == "LAST" {
		version = latestGhRelease(reposlug)
	}

	// execute the action
	if action == "create-packages" {
		CreatePackage(reposlug, ghToken, email, version, archs, out, *push)

	} else if action == "setup-repository" {
		SetupRepo(reposlug, ghToken, email, version, archs, out, *push)
	}
}

func requireArg(val, n string, env ...string) {
	if val == "" {
		log.Printf("missing argument -%v or env %q\n", n, env)
		os.Exit(1)
	}
}

func isTravis() bool {
	fmt.Println(`os.Getenv("CI")`, os.Getenv("CI"))
	fmt.Println(`os.Getenv("TRAVIS")`, os.Getenv("TRAVIS"))
	return os.Getenv("CI") == "TRUE" && os.Getenv("TRAVIS") == "TRUE"
}

func isVagrant() bool {
	fmt.Println(os.Getenv("VAGRANT_CWD"))
	_, s := os.Stat("/vagrant/")
	return !os.IsNotExist(s)
}

func latestGhRelease(repo string) string {
	ret := ""
	u := fmt.Sprintf(`https://api.github.com/repos/%v/releases/latest`, repo)
	fmt.Println("url", u)
	r := getURL(u)
	k := map[string]interface{}{}
	json.Unmarshal(r, &k)

	if x, ok := k["tag_name"]; ok {
		ret = x.(string)
	} else {
		panic("latest version not found")
	}
	return ret
}