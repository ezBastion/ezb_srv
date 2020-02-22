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

package models

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ezbastion/ezb_srv/cache"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

var storage cache.Storage
var conf Configuration
var exPath string

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
	// exPath = "./"

}

func GetViewApi(s cache.Storage, c *Configuration, token, xtrack string) ([]ViewApi, error) {
	storage = s
	var vapis []ViewApi
	var vapisFiltred []ViewApi
	logg := log.WithFields(log.Fields{
		"middleware": "GetViewApi",
		"xtrack":     xtrack,
	})
	content := storage.Get("ViewApi")
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&vapis)
		if err != nil {
			logg.Error(err)
			return vapisFiltred, err
		}
	} else {
		cert, err := tls.LoadX509KeyPair(path.Join(exPath, c.PublicCert), path.Join(exPath, c.PrivateKey))
		if err != nil {
			logg.Error("LoadX509KeyPair: ", err)
			logg.Error("PublicCert:", c.PublicCert)
			return vapisFiltred, err
		}
		client := resty.New()
		client.SetRootCertificate(path.Join(exPath, c.CaCert))
		client.SetCertificates(cert)
		_, err = client.R().
			SetResult(&vapis).
			Get(c.EzbDB + "accountactions")
		if err != nil {
			return vapisFiltred, err
		}
		if len(vapis) == 0 {
			return vapis, fmt.Errorf("VAPIS NOT FOUND")
		}
		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(vapis)
		if encErr != nil {
			return vapisFiltred, encErr
		}
		storage.Set("ViewApi", byteCtrl.Bytes(), time.Duration(c.CacheL1)*time.Second)
	}
	if token == "" {
		return vapisFiltred, nil
	}
	for i := range vapis {
		if strings.ToUpper(vapis[i].Account) == strings.ToUpper(token) {
			vapisFiltred = append(vapisFiltred, vapis[i])
		}
		if vapis[i].Accountid == 0 {
			vapisFiltred = append(vapisFiltred, vapis[i])
		}
	}

	return vapisFiltred, nil
}

func GetApiPath(s cache.Storage, c *Configuration) (apiPath []ApiPath, err error) {
	content := s.Get("apiPath")
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&apiPath)
		if err != nil {
			fmt.Println(err)
			return apiPath, err
		}
	} else {
		cert, err := tls.LoadX509KeyPair(path.Join(exPath, c.PublicCert), path.Join(exPath, c.PrivateKey))
		if err != nil {
			fmt.Println(err)
			return apiPath, err
		}
		client := resty.New()
		client.SetRootCertificate(path.Join(exPath, c.CaCert))
		client.SetCertificates(cert)
		_, err = client.R().
			SetResult(&apiPath).
			Get(c.EzbDB + "api")
		if err != nil {
			return apiPath, err
		}
		if len(apiPath) == 0 {
			return apiPath, fmt.Errorf("APIPATH NOT FOUND")
		}
		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(apiPath)
		if encErr != nil {
			return apiPath, encErr
		}
		s.Set("apiPath", byteCtrl.Bytes(), time.Duration(c.CacheL1)*time.Second)
	}
	return apiPath, nil
}

func GetAction(s cache.Storage, c *Configuration, actionID int) (action EzbActions, err error) {
	content := s.Get(fmt.Sprintf("action%d", actionID))
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&action)
		if err != nil {
			fmt.Println(err)
			return action, err
		}
	} else {
		cert, err := tls.LoadX509KeyPair(path.Join(exPath, c.PublicCert), path.Join(exPath, c.PrivateKey))
		if err != nil {
			fmt.Println(err)
			return action, err
		}
		client := resty.New()
		client.SetRootCertificate(path.Join(exPath, c.CaCert))
		client.SetCertificates(cert)
		_, err = client.R().
			SetResult(&action).
			Get(fmt.Sprintf("%sactions/%d", c.EzbDB, actionID))
		if err != nil {
			return action, err
		}
		if action.ID == 0 {
			return action, fmt.Errorf("ACTION NOT FOUND")
		}
		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(action)
		if encErr != nil {
			return action, encErr
		}
		s.Set(fmt.Sprintf("action%d", actionID), byteCtrl.Bytes(), time.Duration(c.CacheL1)*time.Second)
	}
	return action, nil
}

func GetAccount(s cache.Storage, c *Configuration, userID string) (account EzbAccounts, err error) {
	content := s.Get(fmt.Sprintf("account%s", userID))
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&account)
		if err != nil {
			fmt.Println(err)
			return account, err
		}
	} else {
		cert, err := tls.LoadX509KeyPair(path.Join(exPath, c.PublicCert), path.Join(exPath, c.PrivateKey))
		if err != nil {
			fmt.Println(err)
			return account, err
		}
		client := resty.New()
		client.SetRootCertificate(path.Join(exPath, c.CaCert))
		client.SetCertificates(cert)
		_, err = client.R().
			SetResult(&account).
			Get(fmt.Sprintf("%saccounts/%s", c.EzbDB, userID))
		if err != nil {
			return account, err
		}
		if account.ID == 0 {
			return account, fmt.Errorf("ACCOUNT NOT FOUND %s", userID)
		}

		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(account)
		if encErr != nil {
			return account, encErr
		}
		s.Set(fmt.Sprintf("account%s", userID), byteCtrl.Bytes(), time.Duration(c.CacheL1)*time.Second)
	}
	return account, nil
}

func GetStas(s cache.Storage, c *Configuration) (stas []EzbStas, err error) {
	content := s.Get("stas")
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&stas)
		if err != nil {
			fmt.Println(err)
			return stas, err
		}
	} else {
		cert, err := tls.LoadX509KeyPair(path.Join(exPath, c.PublicCert), path.Join(exPath, c.PrivateKey))
		if err != nil {
			fmt.Println(err)
			return stas, err
		}
		client := resty.New()
		client.SetRootCertificate(path.Join(exPath, c.CaCert))
		client.SetCertificates(cert)
		_, err = client.R().
			SetResult(&stas).
			Get(fmt.Sprintf("%sstas", c.EzbDB))
		if err != nil {
			return stas, err
		}
		if len(stas) == 0 {
			return stas, fmt.Errorf("STA NOT FOUND")
		}
		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(stas)
		if encErr != nil {
			return stas, encErr
		}
		s.Set("stas", byteCtrl.Bytes(), time.Duration(c.CacheL1)*time.Second)
	}
	return stas, nil
}

func GetWorkers(s cache.Storage, c *Configuration) (workers []EzbWorkers, err error) {
	content := s.Get("workers")
	if content != nil {
		byteCtrl := bytes.NewBuffer(content)
		dec := gob.NewDecoder(byteCtrl)
		err := dec.Decode(&workers)
		if err != nil {
			fmt.Println(err)
			return workers, err
		}
	} else {
		cert, err := tls.LoadX509KeyPair(path.Join(exPath, c.PublicCert), path.Join(exPath, c.PrivateKey))
		if err != nil {
			fmt.Println(err)
			return workers, err
		}
		client := resty.New()
		client.SetRootCertificate(path.Join(exPath, c.CaCert))
		client.SetCertificates(cert)
		_, err = client.R().
			SetResult(&workers).
			Get(fmt.Sprintf("%sworkers", c.EzbDB))
		if err != nil {
			return workers, err
		}
		if len(workers) == 0 {
			return workers, fmt.Errorf("WORKERS NOT FOUND")
		}
		var byteCtrl bytes.Buffer
		enc := gob.NewEncoder(&byteCtrl)
		encErr := enc.Encode(workers)
		if encErr != nil {
			return workers, encErr
		}
		s.Set("workers", byteCtrl.Bytes(), time.Duration(c.CacheL1)*time.Second)
	}
	return workers, nil
}

func GetResult(s cache.Storage, key string) (result []byte, ok bool) {
	// fmt.Println("getresult")
	result = s.Get(key)
	ok = false
	if result != nil {
		// byteCtrl := bytes.NewBuffer(content)
		// dec := gob.NewDecoder(byteCtrl)
		// err := dec.Decode(&result)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return result, ok
		// }
		ok = true
		return result, ok
	}
	return result, ok
}

func SetResult(s cache.Storage, result []byte, key string, ttk int) error {
	// fmt.Println("SetResult")
	// fmt.Println(result)
	// var byteCtrl bytes.Buffer
	// enc := gob.NewEncoder(&byteCtrl)
	// encErr := enc.Encode(result)
	// if encErr != nil {
	// 	fmt.Println(encErr)
	// 	return encErr
	// }
	// bodyBytes, err := ioutil.ReadAll(result)
	// s.Set(key, byteCtrl.Bytes(), time.Duration(ttk)*time.Second)
	s.Set(key, result, time.Duration(ttk)*time.Second)
	return nil
}
