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
	"strconv"
	"strings"

	"github.com/ezbastion/ezb_srv/cache"
	"github.com/ezbastion/ezb_srv/models"
	"github.com/ezbastion/ezb_srv/tool"
	"github.com/gin-contrib/location"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func SendAction(c *gin.Context, storage cache.Storage) {
	trace, ok := c.MustGet("trace").(models.EzbLogs)
	if !ok {
		err := "cant read context trace"
		log.WithFields(log.Fields{"controller": "worker"}).Error(err)
		c.JSON(500, err)
		return
	}

	logg := log.WithFields(log.Fields{
		"controller": "worker",
		"xtrack":     trace.Xtrack,
	})
	logg.Debug("start")
	action := c.MustGet("action").(models.EzbActions)
	exPath := c.MustGet("exPath").(string)
	tokenid := c.MustGet("tokenid").(string)
	job := c.MustGet("job").(models.EzbJobs)
	data := c.MustGet("params").(map[string]string)
	worker := c.MustGet("worker").(models.EzbWorkers)
	conf := c.MustGet("configuration").(*models.Configuration)
	body := data["body"]
	rawpath := c.Request.URL.EscapedPath()
	rawquery := c.Request.URL.RawQuery
	key := fmt.Sprintf("%x", md5.Sum([]byte(rawpath+rawquery+body)))
	var meta models.EzbParamMeta
	meta.Job = job
	var params models.EzbParams
	params.Data = data
	params.Meta = meta

	// worker.Request++
	tool.IncRequest(&worker, c)
	var respStruct interface{}
	if action.Jobs.Cache > 0 && c.Request.Method == "GET" {
		response, ok := models.GetResult(storage, key)
		if ok {
			js := strings.NewReader(string(response))
			json.NewDecoder(js).Decode(&respStruct)
			c.JSON(200, respStruct)
			return
		}
	}
	var Url *url.URL
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = "/exec"
	if err != nil {
		logg.Error(err)
		c.JSON(500, err)
		return
	}
	cert, err := tls.LoadX509KeyPair(path.Join(exPath, conf.PublicCert), path.Join(exPath, conf.PrivateKey))
	if err != nil {
		logg.Error(err)
		c.JSON(500, err.Error())
		return
	}
	client := resty.New()
	client.SetRootCertificate(path.Join(exPath, conf.CaCert))
	client.SetCertificates(cert)
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("X-Track", trace.Xtrack).
		SetHeader("X-Polling", strconv.FormatBool(action.Polling)).
		SetHeader("x-ezb-tokenid", tokenid).
		SetBody(&params).
		SetResult(&respStruct).
		Post(Url.String())
	if err != nil {
		logg.Warning(err)
		c.JSON(resp.StatusCode(), err)
		return
	}
	if resp.StatusCode() < 300 {
		var task models.EzbTasks
		err = json.Unmarshal(resp.Body(), &task)
		if action.Polling == true && err == nil {
			u := location.Get(c)
			logg.Debug("respStruct is models.EzbTasks")
			Location := fmt.Sprintf("%s://%s/tasks/%04d%s/status", u.Scheme, u.Host, worker.ID, task.UUID)
			c.Writer.Header().Set("Location", Location)
			task.StatusURL = Location
			task.LogURL = fmt.Sprintf("%s://%s/tasks/%04d%s/log", u.Scheme, u.Host, worker.ID, task.UUID)
			task.ResultURL = fmt.Sprintf("%s://%s/tasks/%04d%s/result", u.Scheme, u.Host, worker.ID, task.UUID)
			c.JSON(201, task)
		} else {
			c.JSON(resp.StatusCode(), respStruct)
		}
	} else {
		c.JSON(resp.StatusCode(), gin.H{"error": resp.String(), "StatusCode": resp.StatusCode()})
	}
	if action.Jobs.Cache > 0 && resp.StatusCode() == 200 && c.Request.Method == "GET" {
		go models.SetResult(storage, resp.Body(), key, action.Jobs.Cache)
	}

}
