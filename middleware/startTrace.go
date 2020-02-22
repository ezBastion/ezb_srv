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
	"strconv"
	s "strings"
	"time"

	"github.com/ezbastion/ezb_srv/models"
	"github.com/ezbastion/ezb_srv/tool"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func StartTrace(c *gin.Context) {
	var l models.EzbLogs
	l.Date = time.Now().UTC()
	x := uuid.NewV4()
	l.Xtrack = x.String()
	l.Status = "init"
	l.URL = c.Request.RequestURI
	// b, err := os.Hostname()
	// if err != nil {
	l.Bastion = c.Request.Host
	// } else {
	// 	l.Bastion = b
	// }
	l.Ipaddr = c.Request.RemoteAddr
	l.Methode = s.ToUpper(c.Request.Method)
	c.Set("trace", l)
	logg := log.WithFields(log.Fields{
		"middleware": "trace",
		"xtrack":     x.String(),
	})
	logg.Info("start session")
	// tool.Trace(&l, c)
	c.Writer.Header().Set("X-Track", l.Xtrack)
	c.Next()
	tr, _ := c.Get("trace")
	trace := tr.(models.EzbLogs)
	if len(c.Errors) > 0 {
		logg.Error("end session with error ", c.Errors.Last().Error())
		trace.Error = c.Errors.Last().Error()
	} else {
		logg.Info("end session normally")
	}
	end := time.Now().UTC()
	duration := end.Sub(trace.Date)
	trace.Duration = duration.Nanoseconds()
	trace.Size = c.Writer.Size()
	trace.Status = strconv.Itoa(c.Writer.Status())
	if trace.Action != "authorize" {
		tool.Trace(&trace, c)
	}
}
