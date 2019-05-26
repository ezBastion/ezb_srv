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

package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetToken(ezbSta, User, Password string) (string, int, string) {
	fmt.Println("GetToken")
	client := &http.Client{}
	type JsonBody struct {
		GrantType string `json:"grant_type"`
		UserName  string `json:"username"`
		Password  string `json:"password"`
	}
	jb := JsonBody{GrantType: "password", UserName: User, Password: Password}
	jsonBody, _ := json.Marshal(jb)
	req, _ := http.NewRequest("GET", ezbSta, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", 0, res.Status
	}
	if res.Body == nil {
		return "", 0, res.Status
	}
	type token struct {
		ExpireIn int `json:"expire_in"`
		// ExpireAt int    `json:"expire_at"`
		Bearer string `json:"access_token"`
		// Type     string `json:"token_type"`
		// SignKey  string `json:"sign_key"`
	}
	var t token
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&t)
	if err != nil {
		fmt.Println(err)
	}
	return t.Bearer, t.ExpireIn, ""
}
