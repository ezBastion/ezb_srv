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
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/ezbastion/ezb_srv/models"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func GetTask(c *gin.Context) {
	tasksID := c.MustGet("tasksid").(string)
	tasksAction := c.MustGet("tasksaction").(string)
	trace := c.MustGet("trace").(models.EzbLogs)
	conf := c.MustGet("configuration").(*models.Configuration)
	exPath := c.MustGet("exPath").(string)
	tokenID := c.MustGet("tokenid").(string)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})

	logg.Debug("start")
	worker := c.MustGet("worker").(models.EzbWorkers)
	trace.Controller = "internal"
	trace.Action = "GetTaskResult"
	URL, err := url.Parse(worker.Fqdn)
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
	pemFilePath := path.Join(exPath, conf.CaCert)
	switch tasksAction {
	case "status":
		URL.Path = fmt.Sprintf("/tasks/status/%s", tasksID[4:])
		getTaskStatus(c, cert, URL.String(), pemFilePath, trace.Xtrack, tokenID)
		break
	case "log":
		URL.Path = fmt.Sprintf("/tasks/log/%s", tasksID[4:])
		getTaskLog(c, cert, URL.String(), pemFilePath, trace.Xtrack, tokenID)
		break
	case "result":
		URL.Path = fmt.Sprintf("/tasks/result/%s", tasksID[4:])
		getTaskResult(c, cert, URL.String(), pemFilePath, trace.Xtrack, tokenID)
		break
	default:
		logg.Error("bad task action name, must be result/log/status")
		c.AbortWithError(http.StatusBadRequest, errors.New("#I0003"))
		break

	}

}

func getTaskResult(c *gin.Context, cert tls.Certificate, url, pemFilePath, xtrack, tokenID string) {
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     xtrack,
	})
	client := resty.New()

	client.SetRootCertificate(pemFilePath)
	client.SetCertificates(cert)
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("X-Track", xtrack).
		SetHeader("x-ezb-tokenid", tokenID).
		Get(url)
	if err != nil {
		logg.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.Data(resp.StatusCode(), "application/json", resp.Body())
	} else {
		c.JSON(resp.StatusCode(), gin.H{"error": resp.String()})
	}
}

func getTaskStatus(c *gin.Context, cert tls.Certificate, url, pemFilePath, xtrack, tokenID string) {
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     xtrack,
	})
	var respStruct models.EzbTasks
	client := resty.New()
	client.SetRootCertificate(pemFilePath)
	client.SetCertificates(cert)
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("X-Track", xtrack).
		SetHeader("x-ezb-tokenid", tokenID).
		SetResult(&respStruct).
		Get(url)

	if err != nil {
		logg.Error("resty: ", err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		u := location.Get(c)
		worker := c.MustGet("worker").(models.EzbWorkers)
		respStruct.StatusURL = fmt.Sprintf("%s://%s/tasks/%04d%s/status", u.Scheme, u.Host, worker.ID, respStruct.UUID)
		respStruct.LogURL = fmt.Sprintf("%s://%s/tasks/%04d%s/log", u.Scheme, u.Host, worker.ID, respStruct.UUID)
		respStruct.ResultURL = fmt.Sprintf("%s://%s/tasks/%04d%s/result", u.Scheme, u.Host, worker.ID, respStruct.UUID)
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), gin.H{"error": resp.String()})
	}
}
func getTaskLog(c *gin.Context, cert tls.Certificate, url, pemFilePath, xtrack, tokenID string) {
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     xtrack,
	})
	client := resty.New()
	client.SetRootCertificate(pemFilePath)
	client.SetCertificates(cert)
	resp, err := client.R().
		SetHeader("X-Track", xtrack).
		SetHeader("x-ezb-tokenid", tokenID).
		Get(url)

	if err != nil {
		logg.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() == 200 {
		c.String(resp.StatusCode(), resp.String())
	} else {
		c.JSON(resp.StatusCode(), gin.H{"error": resp.String()})
	}
}

func GetXtrack(c *gin.Context) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*models.Configuration)
	worker := c.MustGet("worker").(models.EzbWorkers)
	exPath := c.MustGet("exPath").(string)
	param := c.MustGet("params").(string)
	trace.Controller = "internal"
	trace.Action = "getxtrack"
	var Url *url.URL
	var respStruct map[string]interface{}
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = fmt.Sprintf("/log/xtrack/%s", param)
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
		SetResult(&respStruct).
		Get(Url.String())
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
}

func GetLog(c *gin.Context) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*models.Configuration)
	worker := c.MustGet("worker").(models.EzbWorkers)
	exPath := c.MustGet("exPath").(string)
	param := c.MustGet("params").(string)
	trace.Controller = "internal"
	trace.Action = "getlog"
	var Url *url.URL
	var respStruct map[string]interface{}
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = fmt.Sprintf("/log/last/%s", param)
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
		SetResult(&respStruct).
		Get(Url.String())
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
}

func GetLoad(c *gin.Context) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*models.Configuration)
	worker := c.MustGet("worker").(models.EzbWorkers)
	exPath := c.MustGet("exPath").(string)
	trace.Controller = "internal"
	trace.Action = "getload"
	var Url *url.URL
	var respStruct map[string]interface{}
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = "/healthcheck/load"
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
		SetResult(&respStruct).
		Get(Url.String())
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
}
func GetJobs(c *gin.Context) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*models.Configuration)
	worker := c.MustGet("worker").(models.EzbWorkers)
	exPath := c.MustGet("exPath").(string)
	trace.Controller = "internal"
	trace.Action = "getjobs"
	var Url *url.URL
	var respStruct map[string]interface{}
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = "/healthcheck/jobs"
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
		SetResult(&respStruct).
		Get(Url.String())
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
}

type wksScript struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

func GetScripts(c *gin.Context) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*models.Configuration)
	worker := c.MustGet("worker").(models.EzbWorkers)
	exPath := c.MustGet("exPath").(string)
	trace.Controller = "internal"
	trace.Action = "getscripts"
	var Url *url.URL
	var respStruct []wksScript
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = "/healthcheck/scripts"
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
		SetResult(&respStruct).
		Get(Url.String())
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
}

func GetConf(c *gin.Context) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*models.Configuration)
	worker := c.MustGet("worker").(models.EzbWorkers)
	exPath := c.MustGet("exPath").(string)
	trace.Controller = "internal"
	trace.Action = "getconfig"
	var Url *url.URL
	var respStruct map[string]interface{}
	Url, err := url.Parse(worker.Fqdn)
	Url.Path = "/healthcheck/conf"
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
		SetResult(&respStruct).
		Get(Url.String())
	if err != nil {
		log.Error(err)
		c.JSON(500, err)
		return
	}
	if resp.StatusCode() < 300 {
		c.JSON(resp.StatusCode(), respStruct)
	} else {
		c.JSON(resp.StatusCode(), resp.String())
	}
}
