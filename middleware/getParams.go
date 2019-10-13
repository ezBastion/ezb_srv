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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	s "strings"

	"github.com/ezbastion/ezb_srv/cache"
	"github.com/ezbastion/ezb_srv/models"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetParams(storage cache.Storage, conf *models.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		tr, _ := c.Get("trace")
		trace := tr.(models.EzbLogs)
		logg := log.WithFields(log.Fields{
			"middleware": "GetParams",
			"xtrack":     trace.Xtrack,
		})
		logg.Debug("start")

		rt, _ := c.Get("routeType")
		routeType := rt.(string)
		if routeType == "worker" {
			aID, _ := c.Get("matchApiID")
			matchApiID := aID.(map[int]string)
			var actionID int
			var actionRGX string
			for k, v := range matchApiID {
				actionID = k
				actionRGX = v
			}
			action, err := models.GetAction(storage, conf, actionID)
			if err != nil {
				logg.Error(err)
				c.AbortWithError(http.StatusInternalServerError, errors.New("#V0001"))
				return
			}
			params := make(map[string]string)
			escapedPath := c.Request.URL.EscapedPath()
			c.Set("action", action)
			/* tags */
			var tag []string
			for _, t := range action.Tags {
				tag = append(tag, t.Name)
			}
			jsTag, _ := json.Marshal(tag)
			params["tag"] = string(jsTag)
			/* tags */
			params["methode"] = s.ToLower(c.Request.Method)
			usr, _ := c.Get("account")
			account := usr.(models.EzbAccounts)
			params["tokenid"] = account.Name
			params["version"] = fmt.Sprintf("%d", action.Controllers.Version)
			params["constant"] = action.Constant
			// params["polling"] = strconv.FormatBool(action.Polling)
			if action.Polling == true && c.Request.Method != "POST" {
				logg.Error("Polling must use POST methode.")
				c.AbortWithError(http.StatusBadRequest, errors.New("#V0002"))
				return
			}
			/* query string */
			query := make(map[string]string)
			reQ := regexp.MustCompile(`([a-z0-9A-Z-]+)=([si]{1})`)
			var isInt = regexp.MustCompile(`(?m)^[0-9]+$`)
			for _, match := range reQ.FindAllString(action.Query, -1) {
				qmatch := reQ.FindStringSubmatch(match)
				ret, ok := c.GetQuery(qmatch[1])
				if ok {
					if qmatch[2] == "i" {
						if isInt.MatchString(ret) {
							query[match[:len(match)-2]] = ret
						}
					} else {
						query[match[:len(match)-2]] = ret
					}

				}
			}
			// if len(query) > 0 {
			jsQuery, _ := json.Marshal(query)
			params["query"] = string(jsQuery)
			// }
			/* query string */
			/* path */
			path := make(map[string]string)
			reP := regexp.MustCompile(`\{([a-z0-9A-Z-]+)\|([si]{1})\}`)
			var mVar []string
			for _, match := range reP.FindAllString(action.Path, -1) {
				// fmt.Println("match: ", match)
				vmatch := reP.FindStringSubmatch(match)
				mVar = append(mVar, vmatch[1])
			}
			reV := regexp.MustCompile(actionRGX)
			for _, match := range reV.FindAllString(escapedPath, -1) {
				vmatch := reV.FindStringSubmatch(match)
				for i, v := range mVar {
					path[v] = vmatch[i+1]
				}
			}
			jsPath, _ := json.Marshal(path)
			params["path"] = string(jsPath)
			/* path */
			/* body */

			var body interface{}
			err = c.ShouldBindJSON(&body)
			if err == nil {
				// fmt.Println("err body: ", err.Error())
				jsBody, _ := json.Marshal(body)
				// c.Set("body", string(jsBody))
				params["body"] = string(jsBody)
			}
			/* body */
			/* job */
			c.Set("job", action.Jobs)
			/* job */
			c.Set("params", params)
			// fmt.Println(params)

			trace.Deprecated = action.Deprecated > 0
			c.Set("trace", trace)
			// tool.Trace(&trace, c)
		}
		c.Next()
	}
}
