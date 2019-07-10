package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"

	"github.com/inabajunmr/treview/config"
	"github.com/inabajunmr/treview/github/trending"
	treview "github.com/inabajunmr/treview/service"
	"github.com/zserge/lorca"
)

type Condition struct {
	Span    string
	Langs   []string
	OnlyNew bool
}

var ui lorca.UI

func main() {
	// setup log
	usr, err := user.Current()
	if err != nil {
		os.Exit(1)
	}

	path := usr.HomeDir + "/.treview/treview.log"

	logfile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic("cannnot open test.log:" + err.Error())
	}
	log.SetOutput(logfile)

	defer logfile.Close()

	// init lorca
	ui, _ = lorca.New("", "", 1280, 800)

	defer ui.Close()

	err = ui.Bind("load", load)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = ui.Bind("reload", reloadRepositories)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = ui.Bind("updateConfig", updateConfig)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer ln.Close()

	go serveContents(ln)

	err = ui.Load(fmt.Sprintf("http://%s", ln.Addr()))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer ui.Close()

	<-ui.Done()

}

func serveContents(ln net.Listener) {
	err := http.Serve(ln, http.FileServer(FS))

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func load() {
	langs := treview.GetLangs("")
	span := trending.GetSpanByString("today")
	repos := treview.GetRepositories(span, langs, true)
	bindRepositories(repos)

	bindLangs(trending.FindLangs())
	bindConfigLangs(langs)
}

func reloadRepositories(cond Condition) {
	span := trending.GetSpanByString(cond.Span)
	repos := treview.GetRepositories(span, cond.Langs, cond.OnlyNew)
	bindRepositories(repos)
}

func updateConfig(langs []string) {
	langs = langs
	for _, v := range langs {
		log.Println("Update lang:" + v)
	}
	usr, err := user.Current()
	if err != nil {
		os.Exit(1)
	}

	path := usr.HomeDir + "/.treview"
	cpath := path + "/.config"

	config.SetLangs(cpath, langs)
}

func removeEmpty(strs []string) []string {
	result := []string{}
	for _, v := range strs {
		if v != "" {
			result = append(result, v)
		}
	}

	return result
}

func bindRepositories(repos []trending.Repository) {
	val, _ := json.Marshal(repos)
	ui.Eval("vm.repos = " + string(val[:]))
}

func bindLangs(langs []string) {
	val, _ := json.Marshal(langs)
	ui.Eval("vm.langs = " + string(val[:]))
}

func bindConfigLangs(langs []string) {
	val, _ := json.Marshal(langs)
	ui.Eval("vm.condition.Langs = " + string(val[:]))
	ui.Eval("vm.condition.CloneLangs = " + string(val[:]))
}
