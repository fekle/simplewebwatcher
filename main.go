// Package main is the main package
package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/fekle/simplewebwatcher/config"

	"time"

	"io"

	"github.com/codegangsta/cli"
	"github.com/everdev/mack"
	"github.com/skratchdot/open-golang/open"
)

func main() {
	app := cli.NewApp()
	app.Name = "simplewebwatcher"
	app.Usage = "adsf"
	app.Version = "0.0.1"

	// flags TODO: Allow user to specify config location - what about the application directory?
	//app.Flags = []cli.Flag{
	//	cli.StringFlag{
	//		Name:  "config, c",
	//		Value: "",
	//	},
	//}

	// commands TODO: add daemon mode?
	app.Commands = []cli.Command{
		{
			// mode to use for cronjobs
			Name:  "cron",
			Usage: "query all pages once, then quit",
			Action: func(c *cli.Context) {
				log.Println(start())
			},
		},
	}

	app.Run(os.Args)
}

func start() error {

	// set max procs to cpu count though only needed for go versions < 1.5
	runtime.GOMAXPROCS(runtime.NumCPU())

	// determine home directory
	var userHome string
	if runtime.GOOS == "windows" {
		userHome = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if userHome == "" {
			userHome = os.Getenv("USERPROFILE")
		}
	} else {
		userHome = os.Getenv("HOME")
	}

	// create paths
	homePath := filepath.Join(userHome, ".simplewebwatcher")
	configPath := filepath.Join(homePath, "config")
	logPath := filepath.Join(homePath, "log")

	// create application home directory, if not exists
	if err := os.MkdirAll(homePath, 0700); err != nil {
		return err
	}

	// chdir to the application home path
	if err := os.Chdir(homePath); err != nil {
		return err
	}

	{
		logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0700)
		if err != nil {
			log.Fatal("error opening file:", err)
		}
		defer logFile.Close()

		log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	}

	// check if Working Directory is valid
	{
		cwd, err := os.Getwd()

		if err != nil {
			return err
		}

		if cwd != homePath {
			if err := errors.New("wrong pwd"); err != nil {
				return err
			}
		}
	}

	// check if config exists
	{
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// no config found - crete new default config and exit

			log.Println("config not found, creating new default config at", configPath)

			configFile, err := os.Create(configPath)
			if err != nil {
				return err
			}

			defer configFile.Close()

			if err := config.WriteConfig(config.NewDefaultConfig(), configFile); err != nil {
				return err
			}

			log.Println("default config created - please edit")

			return nil
		}
	}

	// variable for config
	var safeConfig *config.ThreadSafeConfigWrapper

	// read config
	{
		log.Println("reading config from", configPath)
		configFile, err := os.Open(configPath)

		if err != nil {
			return err
		}

		defer configFile.Close()

		configBytes, err := ioutil.ReadAll(configFile)

		if err != nil {
			return err
		}

		tmpConfig, err := config.ReadConfig(string(configBytes))

		if err != nil {
			return err
		}

		// create and initialize new threadsafeconfig
		safeConfig = new(config.ThreadSafeConfigWrapper)
		safeConfig.Set(*tmpConfig)
	}

	// create sync waitgroup
	waitGroup := new(sync.WaitGroup)

	// iterate through configured sites and spawn a gouroutine for each one
	for i := range safeConfig.Get().Site {
		waitGroup.Add(1)
		go doCheck(safeConfig, i, waitGroup, homePath)
	}

	// wait for all checks to finish
	waitGroup.Wait()

	// write new configuration
	{

		configFile, err := os.OpenFile(configPath, os.O_RDWR, 0700)
		defer configFile.Close()
		if err != nil {
			return err
		}

		newConf := safeConfig.Get()

		log.Println("updating config file")

		if err := config.WriteConfig(&newConf, configFile); err != nil {
			return err
		}

	}

	return nil

}

func doCheck(safeConfig *config.ThreadSafeConfigWrapper, pos int, wg *sync.WaitGroup, dir string) {
	defer wg.Done()

	// copy site config
	siteConfig := safeConfig.Get().Site[pos]

	// create new http client
	webClient := &http.Client{}

	// configure request
	req, err := http.NewRequest("GET", siteConfig.URL, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// if set, configure http basic auth
	if siteConfig.Password != "" && len(siteConfig.Password) > 0 && siteConfig.Username != "" && len(siteConfig.Username) > 0 {
		req.SetBasicAuth(siteConfig.Username, siteConfig.Password)
	}

	// execute request
	resp, err := webClient.Do(req)
	if err != nil {
		log.Println(err)

		moptions := mack.AlertOptions{
			Title:         "CIS Notifier",
			Message:       "Site Error for " + siteConfig.Description + ": " + err.Error(),
			Style:         "informational",
			Buttons:       "Open",
			DefaultButton: "Open",
			Duration:      0,
		}
		if _, err = mack.AlertBox(moptions); err != nil {
			log.Println(err)
			return
		}

		return
	}
	defer resp.Body.Close()

	// read body and determine size
	body, err := ioutil.ReadAll(resp.Body)
	size := binary.Size(body)

	// determine sha1 hash
	var hash string
	{
		hasher := sha1.New()
		_, err := hasher.Write(body)

		if err != nil {
			log.Println(err)
			return
		}

		hash = hex.EncodeToString(hasher.Sum(nil))
	}

	// compare size and hash to stored data
	if size != siteConfig.LastBytes || hash != siteConfig.LastHash {
		// announce match
		log.Println(siteConfig.Description, " | ", siteConfig.LastBytes, "->", size, " | ", siteConfig.LastHash, "->", hash, " | ", "change detected")

		// set options for alert, and execute it - OSX ONLY
		moptions := mack.AlertOptions{
			Title:         "CIS Notifier",
			Message:       "Change detected for " + siteConfig.Description,
			Style:         "informational",
			Buttons:       "Open",
			DefaultButton: "Open",
			Duration:      0,
		}
		if _, err = mack.AlertBox(moptions); err != nil {
			log.Println(err)
			return
		}

		// open site in browser
		open.Run(siteConfig.URL)

		// update current site config
		siteConfig.LastBytes = size
		siteConfig.LastHash = hash
		siteConfig.LastCheck = time.Now()

		// write new site config to safe config
		safeConfig.SetSite(pos, siteConfig)

	} else {
		// announce mismatch
		log.Println(siteConfig.Description, " | ", siteConfig.LastBytes, "->", size, " | ", siteConfig.LastHash, "->", hash, " | ", "no change detected")
	}
}
