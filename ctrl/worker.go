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

package ctrl

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/ezbastion/ezb_srv/cache"
	"github.com/ezbastion/ezb_srv/models"
	"github.com/ezbastion/ezb_srv/tool"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
)

func SendAction(c *gin.Context, storage cache.Storage) {
	ac, _ := c.Get("action")
	action := ac.(models.EzbActions)
	p, _ := c.Get("params")
	params := p.(map[string]string)
	body := params["body"]
	wk, _ := c.Get("worker")
	worker := wk.(models.EzbWorkers)
	cf, _ := c.Get("configuration")
	conf := cf.(*models.Configuration)
	rawpath := c.Request.URL.EscapedPath()
	rawquery := c.Request.URL.RawQuery
	tr, _ := c.Get("trace")
	trace := tr.(models.EzbLogs)
	exPath := c.MustGet("exPath").(string)
	key := fmt.Sprintf("%x", md5.Sum([]byte(rawpath+rawquery+body)))
	logg := log.WithFields(log.Fields{
		"controller": "worker",
		"xtrack":     trace.Xtrack,
	})
	// worker.Request++
	tool.IncRequest(&worker, c)
	// var respStruct map[string]interface{}
	var respStruct interface{}
	if action.Jobs.Cache > 0 && c.Request.Method == "GET" {
		response, ok := models.GetResult(storage, key)
		if ok {
			// log.Info("found")
			js := strings.NewReader(string(response))
			json.NewDecoder(js).Decode(&respStruct)
			// respStruct["xtrack"] = trace.Xtrack
			c.JSON(200, respStruct)
			return
			// break
		}
		// log.Info("not found")
	}
	var Url *url.URL
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = "/exec"
	if err != nil {
		logg.Error(err)
		c.JSON(500, err)
		return
		// break
	}
	cert, err := tls.LoadX509KeyPair(path.Join(exPath, conf.PublicCert), path.Join(exPath, conf.PrivateKey))
	if err != nil {
		logg.Error(err)
		c.JSON(500, err.Error())
		return
		// break
	}
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
		SetHeader("Accept", "application/json").
		SetHeader("X-Track", trace.Xtrack).
		SetBody(&params).
		SetResult(&respStruct).
		Post(Url.String())
	// trace.Status = resp.Status()
	if err != nil {
		logg.Error(err)
		c.JSON(500, err)
		return
		// break
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
	// log.Info(resp.Size())
	if action.Jobs.Cache > 0 && resp.StatusCode() == 200 && c.Request.Method == "GET" {
		go models.SetResult(storage, resp.Body(), key, action.Jobs.Cache)
	}
	// break

}
