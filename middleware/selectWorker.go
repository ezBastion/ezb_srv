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
	"math/rand"
	"net/http"
	"time"

	"github.com/ezbastion/ezb_srv/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func SelectWorker(c *gin.Context) {
	tr, _ := c.Get("trace")
	trace := tr.(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"middleware": "SelectWorker",
		"xtrack":     trace.Xtrack,
	})
	logg.Debug("start")

	routeType, _ := c.MustGet("routeType").(string)

	if routeType == "worker" {
		ac, _ := c.Get("action")
		action := ac.(models.EzbActions)
		workers := action.Workers
		nbW := len(workers)
		if nbW == 0 {
			logg.Error("NO WORKER FOUND FOR ", action.Name)
			c.AbortWithError(http.StatusServiceUnavailable, errors.New("#W0001"))
			return
		}
		if nbW == 1 {
			worker := workers[0]
			if !worker.Enable {
				logg.Error("ONLY ONE DISABLE WORKER ", worker.Name, " FOUND FOR ", action.Name)
				c.AbortWithError(http.StatusServiceUnavailable, errors.New("#W0002"))
				return
			}
			logg.Debug("one worker found: ", worker.Name, " (", worker.Comment, ")")
			c.Set("worker", worker)

			trace.Worker = worker.Name
			c.Set("trace", trace)
		}
		if nbW > 1 {
			var enableWorkers []models.EzbWorkers
			//var keys []int
			for _, w := range workers {
				if w.Enable {
					if checksumISok(w.Fqdn, action.Jobs.Path, action.Jobs.Checksum) {
						enableWorkers = append(enableWorkers, w)
						logg.Debug("append worker ", w.Name, " :", len(enableWorkers))

						//keys = append(keys, len(enableWorkers))
					}
				}
			}
			if len(enableWorkers) == 0 {
				logg.Error("ALL WORKERS ARE DISABLE FOR ", action.Name)
				c.AbortWithError(http.StatusServiceUnavailable, errors.New("#W0003"))
				return
			}

			var	worker = enableWorkers[randomWorker(&enableWorkers)]

			// switch on config worker algo

			logg.Debug("found ", len(enableWorkers), " worker, select ", worker.Name, " random")
			c.Set("worker", worker)

			trace.Worker = worker.Name
			c.Set("trace", trace)
			c.Next()
		}
	}
	c.Next()
}

//func orderBYjobs(keys *[]int, workers *[]models.EzbWorkers) {
//	var nbj []int
//	for _, _ = range *workers {
//		var j int // w.fqdn /healthcheck/jobs -> j int  goroutine
//		nbj = append(nbj, j)
//	}
//	sort.Ints(nbj)
//	keys = &nbj
//
//}
//func orderBYload(keys *[]int, workers *[]models.EzbWorkers) {
//	sort.Ints(*keys)
//}
func randomWorker(workers *[]models.EzbWorkers) int {
	j := rand.Intn(len(*workers) )
	if j < len(*workers) {
		return j
	} else {
		return j -1
	}
}

//func orderBYroundrobin(keys *[]int, workers *[]models.EzbWorkers) {
//	sort.Ints(*keys)
//}

func checksumISok(fqdn, path, checksum string) bool {
	/*
		connect to fqdn  /healthcheck/scripts  (cache)
		type wksScript struct {
			Name     string `json:name`
			Path     string `json:path`
			Checksum string `json:checksum`
		}
	*/
	return true
}
