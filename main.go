//go:generate goversioninfo

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
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ezbastion/ezb_srv/models"
	"github.com/ezbastion/ezb_srv/setup"

	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
)

var (
	exPath string
	conf   models.Configuration
)

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
}

func main() {

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}

	if !isIntSess {
		conf, err := setup.CheckConfig()
		if err == nil {
			runService(conf.ServiceName, false)
		}
		log.Fatal(err)
		return
	}
	app := cli.NewApp()
	app.Name = "ezb_srv"
	app.Version = "0.2.3"
	app.Usage = "ezBastion frontend server."

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Genarate config file and PKI certificat.",
			Action: func(c *cli.Context) error {
				err := setup.Setup()
				return err
			},
		}, {
			Name:  "debug",
			Usage: "Start ezb_srv in console.",
			Action: func(c *cli.Context) error {
				conf, _ := setup.CheckConfig()
				runService(conf.ServiceName, true)
				return nil
			},
		}, {
			Name:  "install",
			Usage: "Add ezb_srv deamon windows service.",
			Action: func(c *cli.Context) error {
				conf, _ := setup.CheckConfig()
				err = installService(conf.ServiceName, conf.ServiceFullName)
				if err != nil {
					log.Fatalf("install ezb_srv service: %v", err)
				}
				return err
			},
		}, {
			Name:  "remove",
			Usage: "Remove ezb_srv deamon windows service.",
			Action: func(c *cli.Context) error {
				conf, _ := setup.CheckConfig()
				err = removeService(conf.ServiceName)
				if err != nil {
					log.Fatalf("remove ezb_srv service: %v", err)
				}
				return err
			},
		}, {
			Name:  "start",
			Usage: "Start ezb_srv deamon windows service.",
			Action: func(c *cli.Context) error {
				conf, _ := setup.CheckConfig()
				err = startService(conf.ServiceName)
				if err != nil {
					log.Fatalf("start ezb_srv service: %v", err)
				}
				return err
			},
		}, {
			Name:  "stop",
			Usage: "Stop ezb_srv deamon windows service.",
			Action: func(c *cli.Context) error {
				conf, _ := setup.CheckConfig()
				err = controlService(conf.ServiceName, svc.Stop, svc.Stopped)
				if err != nil {
					log.Fatalf("stop ezb_srv service: %v", err)
				}
				return err
			},
		},
	}
	cli.AppHelpTemplate = fmt.Sprintf(`

	███████╗███████╗██████╗  █████╗ ███████╗████████╗██╗ ██████╗ ███╗   ██╗
	██╔════╝╚══███╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
	█████╗    ███╔╝ ██████╔╝███████║███████╗   ██║   ██║██║   ██║██╔██╗ ██║
	██╔══╝   ███╔╝  ██╔══██╗██╔══██║╚════██║   ██║   ██║██║   ██║██║╚██╗██║
	███████╗███████╗██████╔╝██║  ██║███████║   ██║   ██║╚██████╔╝██║ ╚████║
	╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			
				███████╗██████╗  ██████╗ ███╗   ██╗████████╗               
				██╔════╝██╔══██╗██╔═══██╗████╗  ██║╚══██╔══╝               
				█████╗  ██████╔╝██║   ██║██╔██╗ ██║   ██║                  
				██╔══╝  ██╔══██╗██║   ██║██║╚██╗██║   ██║                  
				██║     ██║  ██║╚██████╔╝██║ ╚████║   ██║                  
				╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝                  	

%s
INFO:
		http://www.ezbastion.com		
		support@ezbastion.com
		`, cli.AppHelpTemplate)
	app.Run(os.Args)
}
