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
	"net/url"
	s "strings"

	"github.com/ezbastion/ezb_srv/cache"
	"github.com/ezbastion/ezb_srv/models"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type formAuth struct {
	Granttype string `json:"grant_type" form:"grant_type"`
	Username  string `json:"username" form:"username" binding:"required"`
	Password  string `json:"password" form:"password"`
}

func InternalWork(storage cache.Storage, conf *models.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		tr, _ := c.Get("trace")
		trace := tr.(models.EzbLogs)
		logg := log.WithFields(log.Fields{
			"middleware": "InternalWork",
			"xtrack":     trace.Xtrack,
		})
		logg.Debug("start")
		escapedPath := c.Request.URL.EscapedPath()
		path := s.Split(escapedPath, "/")
		// sver := path[1]
		if len(path) == 2 && path[1] == "authorize" {
			logg.Debug("routeType: authorize")
			c.Set("routeType", "authorize")
			ret, err := authorize(c, storage, conf)
			if err != nil {
				logg.Error(err)
				c.AbortWithError(http.StatusInternalServerError, errors.New("#I0001"))
				return
			} else {
				trace.Controller = "internal"
				trace.Action = "authorize"
				logg.Info(ret)
			}
		}
		switch path[1] {
		case "wks":
			logg.Debug("routeType: internal")
			c.Set("routeType", "internal")
			// trace.Action = "wks"
			wksid := path[2]
			c.Set("wksid", wksid)
			break
		case "tasks":
			if len(path) == 4 {
				logg.Debug("routeType: tasks")
				c.Set("routeType", "tasks")
				c.Set("tasksid", path[2])
				c.Set("tasksaction", path[3])
			} else {
				logg.Error("bad task path url")
				c.AbortWithError(http.StatusBadRequest, errors.New("#I0002"))
			}
			break
		default:
			logg.Debug("routeType: worker")
			c.Set("routeType", "worker")
		}
		// if path[1] == "wks" {
		// 	logg.Debug("routeType: internal")
		// 	c.Set("routeType", "internal")
		// 	// trace.Action = "wks"
		// 	wksid := path[2]
		// 	c.Set("wksid", wksid)
		// } else {
		// 	logg.Debug("routeType: worker")
		// 	c.Set("routeType", "worker")
		// }
		c.Set("trace", trace)
		c.Next()
	}
}

func authorize(c *gin.Context, storage cache.Storage, conf *models.Configuration) (string, error) {
	var EndPoint string
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"middleware": "authorize",
		"xtrack":     trace.Xtrack,
	})
	var formauth = formAuth{}
	if err := c.ShouldBind(&formauth); err == nil {
		Username := formauth.Username
		tokenid := c.GetHeader("x-ezb-tokenid")
		if tokenid != "" {
			Username = tokenid
		}
		account, err := models.GetAccount(storage, conf, Username)
		if err != nil {

			logg.Error(err)
			return "", err
		}
		var endPoint = account.STA.EndPoint
		if s.ToLower(formauth.Granttype) == "password" {
			URL, err := url.Parse(endPoint)
			if err != nil {
				logg.Error(err)
				return "", err
			}
			var respStruct map[string]interface{}
			client := resty.New()
			resp, err := client.R().
				SetHeader("Accept", "application/json").
				SetHeader("X-Track", trace.Xtrack).
				SetBody(&formauth).
				SetResult(&respStruct).
				Post(URL.String())
			if err != nil {
				logg.Error(err)
				return "", err
			}
			if resp.StatusCode() < 300 {
				c.JSON(resp.StatusCode(), respStruct)
				return "return jwt", nil
			} else {
				c.JSON(resp.StatusCode(), resp.String())
				logg.Error(resp.String())
				return "", errors.New(resp.String())
			}
		} else {
			c.Redirect(http.StatusFound, endPoint)
			return "redirect with formauth to " + EndPoint, nil
		}
	} else {
		stas, err := models.GetStas(storage, conf)
		if err != nil {
			logg.Error(err)
			return "", err
		}
		for _, s := range stas {
			if s.Enable && s.Default {
				EndPoint = s.EndPoint
				c.Redirect(http.StatusFound, s.EndPoint)
				break
			}
		}
		return "redirect to " + EndPoint, nil
	}

}
