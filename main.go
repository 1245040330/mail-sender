package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/runner"

	"github.com/n9e/mail-sender/config"
	"github.com/n9e/mail-sender/cron"
	"github.com/n9e/mail-sender/redisc"
)

var (
	vers *bool
	help *bool
	conf *string
	test *string
)

func init() {
	vers = flag.Bool("v", false, "display the version.")
	help = flag.Bool("h", false, "print this help.")
	conf = flag.String("f", "", "specify configuration file.")
	test = flag.String("t", "", "test smtp configuration.")
	flag.Parse()

	if *vers {
		fmt.Println("version:", config.Version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	runner.Init()
	fmt.Println("runner.cwd:", runner.Cwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}

func main() {
	aconf()
	pconf()

	if *test != "" {
		config.TestSMTP(strings.Split(*test, ","))
		os.Exit(0)
	}

	config.InitLogger()
	redisc.InitRedis()

	go cron.SendMails()
	http.HandleFunc("/", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	ending()

}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		_, err := fmt.Fprint(w, "Hello, World!`````````	")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	if r.URL.Path == "/api/alarm" {
		res,err:=redisc.FindMessage()
		b,err:=json.Marshal(res)
		if err != nil {
			fmt.Println("失败")
		}
		w.Header().Set("Content-Type","application/json; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)

	}else {
		http.NotFound(w, r)
		return
	}
}

func ending() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		fmt.Printf("stop signal caught, stopping... pid=%d\n", os.Getpid())
	}

	logger.Close()
	redisc.CloseRedis()
	fmt.Println("sender stopped successfully")
}

// auto detect configuration file
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}

	*conf = path.Join(runner.Cwd, "etc", "mail-sender.local.yml")
	if file.IsExist(*conf) {
		return
	}

	*conf = path.Join(runner.Cwd, "etc", "mail-sender.yml")
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for sender")
	os.Exit(1)
}

// parse configuration file
func pconf() {
	if err := config.ParseConfig(*conf); err != nil {
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	} else {
		fmt.Println("parse configuration file:", *conf)
	}
}
