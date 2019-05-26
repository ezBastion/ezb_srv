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
	"fmt"
	"net/url"
	"path"

	"github.com/ezbastion/ezb_srv/model"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
)

func GetXtrack(c *gin.Context) {
	trace := c.MustGet("trace").(model.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*model.Configuration)
	worker := c.MustGet("worker").(model.EzbWorkers)
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
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
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
	trace := c.MustGet("trace").(model.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*model.Configuration)
	worker := c.MustGet("worker").(model.EzbWorkers)
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
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
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
	trace := c.MustGet("trace").(model.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*model.Configuration)
	worker := c.MustGet("worker").(model.EzbWorkers)
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
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
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
	trace := c.MustGet("trace").(model.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*model.Configuration)
	worker := c.MustGet("worker").(model.EzbWorkers)
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
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
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
	trace := c.MustGet("trace").(model.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*model.Configuration)
	worker := c.MustGet("worker").(model.EzbWorkers)
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
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
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
	trace := c.MustGet("trace").(model.EzbLogs)
	logg := log.WithFields(log.Fields{
		"controller": "internal",
		"xtrack":     trace.Xtrack,
	})
	conf := c.MustGet("configuration").(*model.Configuration)
	worker := c.MustGet("worker").(model.EzbWorkers)
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
	resty.SetRootCertificate(path.Join(exPath, conf.CaCert))
	resty.SetCertificates(cert)
	resp, err := resty.R().
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
