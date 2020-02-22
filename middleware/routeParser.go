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

package middleware

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	s "strings"

	"github.com/ezbastion/ezb_srv/models"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// use cache for sql https://github.com/goenning/go-cache-demo/blob/master/cache/cache.go
func RouteParser(c *gin.Context) {
	tr, _ := c.Get("trace")
	trace := tr.(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"middleware": "RouteParser",
		"xtrack":     trace.Xtrack,
	})
	logg.Debug("start")

	escapedPath := c.Request.URL.EscapedPath()
	path := s.Split(escapedPath, "/")
	// fmt.Printf("escapedPath: '%d'\n", len(path))
	sver := path[1]
	routeType, _ := c.MustGet("routeType").(string)
	if routeType == "worker" {
		// switch sver {
		// case "authorize":
		// 	c.Set("routeType", "authorize")
		// 	c.Next()
		// 	break
		// default:
		// 	logg.Error("NO ROUTE MATCH ", escapedPath)
		// 	c.AbortWithError(http.StatusBadRequest, errors.New("#P0001"))
		// 	return
		// }
		// if sver == "authorize" {

		// } else {
		// 	//TODO: not error but return possible api for this partial url
		// 	c.AbortWithError(http.StatusBadRequest, errors.New("bad query #P0001"))
		// 	return
		// }
		// } else {
		// c.Set("routeType", "worker")
		var version int
		var err error
		if len(sver) > 1 && s.HasPrefix(sver, "v") {
			version, err = strconv.Atoi(s.TrimLeft(sver, "v"))
			if err != nil {
				logg.Error("NO ROUTE MATCH ", escapedPath)
				c.AbortWithError(http.StatusBadRequest, errors.New("#P0002"))
				return
			}
		} else {
			logg.Error("NO ROUTE MATCH ", escapedPath)
			c.AbortWithError(http.StatusBadRequest, errors.New("#P0003"))
			return
		}

		ctrl := path[2]
		action := path[3]
		access := s.ToUpper(c.Request.Method)
		viewApis, _ := c.Get("ViewApi")
		ViewApis := viewApis.([]models.ViewApi)
		apiFound := false
		var CurrentAction []models.ViewApi
		for i := range ViewApis {
			if ViewApis[i].Ctrlver == version {
				logg.Debug("ViewApis version match:", version)
				if ViewApis[i].Ctrl == ctrl {
					logg.Debug("ViewApis ctrl match:", ctrl)
					if ViewApis[i].Action == action {
						logg.Debug("ViewApis action match:", action)
						if s.ToUpper(ViewApis[i].Access) == access {
							logg.Debug("ViewApis access match:", access)
							apiFound = true
							// CurrentAction = ViewApis[i]
							CurrentAction = append(CurrentAction, ViewApis[i])
							// break
						}
					}
				}
			}
		}
		c.Set("ViewApi", nil)
		if !apiFound {
			logg.Error("NO API MATCH ", escapedPath)
			c.AbortWithError(http.StatusForbidden, errors.New("#P0004"))
			return
		}
		logg.Debug("found ", len(CurrentAction), " api matching ver, ctrl, act")
		// fmt.Println(CurrentAction)
		// c.Set("CurrentAction", CurrentAction)
		apiPath, _ := c.Get("apiPath")
		ApiPath := apiPath.([]models.ApiPath)
		mapping := make(map[int]string)
		for _, a := range ApiPath {
			// fmt.Println(a)
			for _, ca := range CurrentAction {
				if ca.Actionid == a.ID {
					mapping[a.ID] = a.RGX
				}
			}
		}
		c.Set("apiPath", nil)
		if len(mapping) == 0 {
			logg.Error("NO API MATCH ", escapedPath)
			// TODO: test it
		}
		// var matchApiID []int
		matchApiID := make(map[int]string)
		for i, rx := range mapping {
			var regx = regexp.MustCompile(rx)
			if regx.MatchString(escapedPath) {
				// matchApiID = append(matchApiID, i)
				matchApiID[i] = rx
				// fmt.Printf("found api match %d for %s \n", i, escapedPath)
			}
		}
		if len(matchApiID) == 0 {
			logg.Error("NO API MATCH ", escapedPath)
			c.AbortWithError(http.StatusForbidden, errors.New("#P0005"))
			return
		}
		if len(matchApiID) > 1 {
			logg.Error("AMBIGUOUS REQUEST MORE THAN ONE API MATCH")
			c.AbortWithError(http.StatusConflict, errors.New("#P0006"))
			return
		}
		// subpath := path[4:]
		// // var re = regexp.MustCompile(`(?m)^\{(?P<name>[a-z0-9A-Z-]+)\|(?P<type>[is]){1}\}$`)
		// var isString = regexp.MustCompile(`(?m)^[a-z0-9A-Z-]+$`)
		// var isInt = regexp.MustCompile(`(?m)^[0-9]+$`)
		// var matchApi = fmt.Sprintf("%s .*(?P<value>/v%d/%s/%s", access, version, ctrl, action)
		// for _, str := range subpath {
		// 	fmt.Println(str)
		// 	if isInt.MatchString(str) {
		// 		matchApi = fmt.Sprintf("%s/\\{[a-z0-9A-Z-]+\\|i\\}", matchApi)
		// 	} else if isString.MatchString(str) {
		// 		matchApi = fmt.Sprintf("%s/\\{[a-z0-9A-Z-]+\\|s\\}", matchApi)
		// 	}
		// 	// if re.MatchString(str) {
		// 	// 	t := re.FindStringSubmatch(str)[2]
		// 	// 	v := re.FindStringSubmatch(str)[1]
		// 	// 	if t == "i" && isInt.MatchString(v) {
		// 	// 		matchApi = fmt.Sprintf("%s/{%s|%s}", matchApi, v, t)
		// 	// 	} else if t == "s" && isString.MatchString(v) {
		// 	// 		matchApi = fmt.Sprintf("%s/{%s|%s}", matchApi, v, t)
		// 	// 	} else {
		// 	// 		// error
		// 	// 	}
		// 	// } else {
		// 	// error
		// 	// }

		// }
		// matchApi = fmt.Sprintf("%s)", matchApi)
		/*
		   	Invoke-RestMethod  -Uri http://localhost:8443/api
		   	/v1/vmware/vms/unechaine/subfolder/533

		   id api
		   -- ---
		   26 GET http://localhost:6000/v1/vmware/vms?enable=i
		   31 GET http://localhost:6000/v1/vmware/vms/{name|s}/subfolder/{id|i}?folder=s

		*/
		// fmt.Println(matchApi)

		trace.Controller = ctrl
		trace.Action = action
		c.Set("trace", trace)
		// tool.Trace(&trace, c)
		c.Set("matchApiID", matchApiID)
		// fmt.Println("routeparser ok use api: ", matchApiID)
	}
	if routeType == "internal" {

		action := path[3]
		trace.Action = action
		if action == "log" {
			cmd := path[4]
			switch cmd {
			case "xtrack":
				c.Set("params", path[5])
				break
			case "last":
				c.Set("params", path[5])
				break
			}
		}
		// switch action {
		// case "log":
		// 	switch cmd {
		// 	case "xtrack":
		// 		ctrl.GetXtrack()
		// 		break
		// 	case "last":
		// 		ctrl.GetLog()
		// 		break
		// 	}
		// 	break
		// case "healthcheck":
		// 	switch cmd {
		// 	case "jobs":
		// 		ctrl.GetJobs()
		// 		break
		// 	case "load":
		// 		ctrl.GetLoad()
		// 		break
		// 	case "scripts":
		// 		ctrl.GetScripts()
		// 		break
		// 	case "conf":
		// 		ctrl.GetConf()
		// 		break

		// 	}
		// 	break
		// }
		c.Set("trace", trace)
	}
	c.Next()
}
