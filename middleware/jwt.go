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
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/ezbastion/ezb_srv/cache"
	"github.com/ezbastion/ezb_srv/models"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Payload struct {
	JTI string `json:"jti"`
	ISS string `json:"iss"`
	SUB string `json:"sub"`
	AUD string `json:"aud"`
	EXP int    `json:"exp"`
	IAT int    `json:"iat"`
}

type Introspec struct {
	User       string   `json:"user"`
	UserSID    string   `json:"userSid"`
	UserGroups []string `json:"userGroups"`
}

func AuthJWT(storage cache.Storage, conf *models.Configuration, exPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// tr, _ := c.MustGet("trace").(models.EzbLogs)
		trace := c.MustGet("trace").(models.EzbLogs)
		var err error
		logg := log.WithFields(log.Fields{
			"middleware": "jwt",
			"xtrack":     trace.Xtrack,
		})
		logg.Debug("start")
		authHead := c.GetHeader("Authorization")
		tokenid := c.GetHeader("x-ezb-tokenid")
		logg.Debug("Authorization:", authHead)
		if authHead == "" {
			var Account models.EzbAccounts
			Account.Name = "anonymous"
			c.Set("account", Account)
			c.Set("tokenid", "anonymous")
			tr, ok := c.Get("trace")
			if ok {
				trace := tr.(models.EzbLogs)
				trace.Token = "anonymous"
				trace.Account = "none"
				trace.Issuer = "NA"
				c.Set("trace", trace)
			}
			c.Next()
			return
		}
		bearer := strings.Split(authHead, " ")
		if strings.Compare(strings.ToLower(bearer[0]), "bearer") != 0 {
			logg.Error("NOT AUTHORIZATION TYPE BEARER")
			c.AbortWithError(http.StatusForbidden, errors.New("#J0002"))
			return
		}
		tokenString := bearer[1]
		claims, err := parseJWT(c, tokenString, exPath)
		if err == nil {
			if tokenid == "" {
				tokenid = claims["sub"].(string)
			}
			Account, err := getAccount(c, tokenid, storage, conf)
			if err != nil {
				logg.Error(err)
				c.AbortWithError(http.StatusForbidden, errors.New("#J0012"))
				return
			}
			if Account.Type == "g" {
				err = Introspection(c, Account.Name, Account)
				if err != nil {
					logg.Error(err)
					c.AbortWithError(http.StatusForbidden, errors.New("#J0013"))
					return
				}
			}
			c.Set("account", Account)
			c.Set("tokenid", tokenid)
			trace.Token = tokenid
			trace.Account = Account.Name
			trace.Issuer = claims["iss"].(string)
			c.Set("trace", trace)
		} else {
			logg.Error(err)
			c.AbortWithError(http.StatusForbidden, errors.New("#J0005"))
			return
		}
		c.Next()
	}
}
func Introspection(c *gin.Context, sub string, Account models.EzbAccounts) (err error) {
	var Url *url.URL
	Url, _ = url.Parse(Account.STA.EndPoint)
	Url.Path = "/access"
	var introspec Introspec
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", c.GetHeader("Authorization")).
		SetResult(&introspec).
		Get(Url.String())
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	for _, g := range introspec.UserGroups {
		if strings.ToLower(g) == strings.ToLower(Account.Real) {
			return nil
		}
	}
	return errors.New("WRONG GROUP")
}
func getAccount(c *gin.Context, sub string, storage cache.Storage, conf *models.Configuration) (Account models.EzbAccounts, err error) {
	Account, err = models.GetAccount(storage, conf, sub)
	if err != nil {
		c.AbortWithError(http.StatusForbidden, errors.New("#J0006"))
		return Account, err
	}
	if !Account.Enable {
		c.AbortWithError(http.StatusForbidden, errors.New("#J0001"))
		return Account, err
	}
	routeType, _ := c.MustGet("routeType").(string)
	if !Account.Isadmin && routeType == "internal" {
		c.AbortWithError(http.StatusForbidden, errors.New("#J0008"))
		return Account, err
	}
	return Account, nil
}

func parseJWT(c *gin.Context, tokenString string, exPath string) (claims jwt.MapClaims, err error) {
	trace := c.MustGet("trace").(models.EzbLogs)
	logg := log.WithFields(log.Fields{
		"middleware": "jwt",
		"xtrack":     trace.Xtrack,
	})
	parts := strings.Split(tokenString, ".")
	p, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		logg.Error("Unable to parse payload: ", err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0009"))
		return nil, err
	}
	var payload Payload
	err = json.Unmarshal(p, &payload)
	if err != nil {
		logg.Error("Unable to parse payload: ", err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0011"))
		return nil, err
	}
	jwtkeyfile := fmt.Sprintf("%s.crt", payload.ISS)
	jwtpubkey := path.Join(exPath, "cert", jwtkeyfile)

	if _, err := os.Stat(jwtpubkey); os.IsNotExist(err) {
		logg.Error("Unable to load sta public certificat: ", err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0010"))
		return nil, err
	}
	key, _ := ioutil.ReadFile(jwtpubkey)
	var ecdsaKey *ecdsa.PublicKey
	if ecdsaKey, err = jwt.ParseECPublicKeyFromPEM(key); err != nil {
		logg.Error("Unable to parse ECDSA public key: ", err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0003"))
		return nil, err
	}
	methode := jwt.GetSigningMethod("ES256")
	err = methode.Verify(strings.Join(parts[0:2], "."), parts[2], ecdsaKey)
	if err != nil {
		logg.Error("Error while verifying key: ", err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0004"))
		return nil, err
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwt.ParseECPublicKeyFromPEM(key)
	})
	if err != nil {
		logg.Error(err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0007"))
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		logg.Error(err)
		c.AbortWithError(http.StatusForbidden, errors.New("#J0005"))
		return nil, err
	}

}
