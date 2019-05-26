// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"io"
	"os/signal"
	"strings"
	"time"

	"github.com/ezbastion/ezb_srv/cache"
	"github.com/ezbastion/ezb_srv/cache/memory"
	"github.com/ezbastion/ezb_srv/ctrl"
	"github.com/ezbastion/ezb_srv/middleware"
	"github.com/ezbastion/ezb_srv/setup"
	"github.com/ezbastion/ezb_db/Middleware"

	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var storage cache.Storage
var exPath string

func mainGin(serverchan *chan bool) {

	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)

	conf, err := setup.CheckConfig(true)

	if err != nil {
		panic(err)
	}

	storage = memory.NewStorage()

	if conf.Listen == "" {
		conf.Listen = "localhost:6000"

	}

	/* log */
	outlog := true
	gin.DisableConsoleColor()
	log.SetFormatter(&log.JSONFormatter{})
	switch conf.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		break
	case "info":
		log.SetLevel(log.InfoLevel)
		break
	case "warning":
		log.SetLevel(log.WarnLevel)
		break
	case "error":
		log.SetLevel(log.ErrorLevel)
		break
	case "critical":
		log.SetLevel(log.FatalLevel)
		break
	default:
		outlog = false
	}
	if outlog {
		if _, err := os.Stat(path.Join(exPath, "log")); os.IsNotExist(err) {
			err = os.MkdirAll(path.Join(exPath, "log"), 0600)
			if err != nil {
				log.Println(err)
			}
		}
		t := time.Now().UTC()
		l := fmt.Sprintf("log/ezb_srv-%d%d.log", t.Year(), t.YearDay())
		f, _ := os.OpenFile(path.Join(exPath, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		log.SetOutput(io.MultiWriter(f))
		ti := time.NewTicker(1 * time.Minute)
		defer ti.Stop()
		go func() {
			for range ti.C {
				t := time.Now().UTC()
				l := fmt.Sprintf("log/ezb_srv-%d%d.log", t.Year(), t.YearDay())
				f, _ := os.OpenFile(path.Join(exPath, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				defer f.Close()
				log.SetOutput(io.MultiWriter(f))
			}
		}()
	}
	/* log */
	/* gin */
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	r.Use(Middleware.AddHeaders)
	r.Use(middleware.LoadConfig(&conf, exPath))
	r.Use(middleware.StartTrace)
	r.Use(middleware.InternalWork(storage, &conf))
	r.Use(middleware.AuthJWT(storage, &conf, exPath))
	r.Use(middleware.Store(storage, &conf))
	r.Use(middleware.RouteParser)
	r.Use(middleware.GetParams(storage, &conf))
	r.Use(middleware.SelectWorker)
	r.OPTIONS("*a", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})
	r.GET("*a", sendAction)
	r.POST("*a", sendAction)
	r.PUT("*a", sendAction)
	r.DELETE("*a", sendAction)
	r.PATCH("*a", sendAction)

	// log.Fatal(http.ListenAndServe(conf.Listen, r))
	srv := &http.Server{
		Addr:    conf.Listen,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Info("listen: %s\n", err)
		}
	}()
	/* gin */

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Info("Server exiting")
}

func sendAction(c *gin.Context) {

	escapedPath := c.Request.URL.EscapedPath()
	path := strings.Split(escapedPath, "/")
	routeType := c.MustGet("routeType").(string)
	switch routeType {
	case "worker":
		ctrl.SendAction(c, storage)
		break
	case "internal":
		action := path[3]
		cmd := path[4]
		switch action {
		case "log":
			switch cmd {
			case "xtrack":
				ctrl.GetXtrack(c)
				break
			case "last":
				ctrl.GetLog(c)
				break
			}
			break
		case "healthcheck":
			switch cmd {
			case "jobs":
				ctrl.GetJobs(c)
				break
			case "load":
				ctrl.GetLoad(c)
				break
			case "scripts":
				ctrl.GetScripts(c)
				break
			case "conf":
				ctrl.GetConf(c)
				break
			}
			break
		}
		break

	}
}
