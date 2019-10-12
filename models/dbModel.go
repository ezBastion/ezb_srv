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
	"time"
)

type User struct {
	User       string   `json:"user"`
	UserSid    string   `json:"userSid"`
	UserGroups []string `json:"userGroups"`
	Signkey    string   `json:"sign_key"`
}

type EzbAccounts struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Enable      bool      `json:"enable"`
	Isadmin     bool      `json:"isadmin"`
	Comment     string    `json:"comment"`
	LastRequest time.Time `json:"lastrequest"`
	Type        string    `json:"type"`
	Real        string    `json:"real"`
	Email       string    `json:"email"`
	EzbStasID   int       `json:"stasid"`
	STA         EzbStas   `json:"sta"`
}
type EzbStas struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
	Type     int    `json:"type"` // 0:internal  1:AD  2:oAuth2
	Comment  string `json:"comment"`
	EndPoint string `json:"authorization_endpoint"`
	Issuer   string `json:"issuer"`
	Default  bool   `json:"default"`
}
type EzbActions struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Enable      bool           `json:"enable"  `
	Comment     string         `json:"comment"`
	LastRequest time.Time      `json:"lastrequest"`
	Tags        []*EzbTags     `json:"tags"`
	Jobs        EzbJobs        `json:"jobs"`
	Workers     []EzbWorkers   `json:"workers"`
	Path        string         `json:"path"`
	Query       string         `json:"query"`
	Body        string         `json:"body"`
	Constant    string         `json:"constant"`
	Deprecated  int            `json:"deprecated"`
	Anonymous   bool           `json:"anonymous"`
	Controllers EzbControllers `json:"controllers"`
	Polling     bool           `gorm:"not null;default:'0'" json:"polling"`

	// EzbAccessID      int              `gorm:"default:'0'" json:"ezbaccessid" sql:"type:int REFERENCES ezb_access(id)"`
	// Access           EzbAccess        `json:"access" gorm:"ForeignKey:ID;AssociationForeignKey:EzbAccessID;association_autoupdate:false;association_autocreate:false;association_save_reference:false;"`
	// EzbJobsID        int              ` json:"ezbjobsid" sql:"type:int REFERENCES ezb_jobs(id)"`
	// Collections      []EzbCollections `json:"collections" gorm:"many2many:ezb_actions_has_ezb_collections;"`
	// Accounts         []EzbAccounts    `json:"accounts" gorm:"many2many:ezb_accounts_has_ezb_actions;"`
	// Groups           []EzbGroups      `json:"groups" gorm:"many2many:ezb_groups_has_ezb_actions;"`
	// EzbControllersID int              `gorm:"default:'0'" json:"ezbcontrollersid" sql:"type:int REFERENCES ezb_controllers(id)"`
}

type EzbTasks struct {
	UUID       string    `json:"uuid"`
	CreateDate time.Time `json:"createddate"`
	UpdateDate time.Time `json:"updatedate"`
	Status     string    `json:"status"`
	// TokenID    string    `json:"tokenid"`
	// PID        int       `json:"pid"`
	// Parameters string    `json:"parameters"`
	StatusURL string `json:"statusurl"`
	LogURL    string `json:"logurl"`
	ResultURL string `json:"resulturl"`
}

type taksStatus int

const (
	// PENDING: the job is created but not started
	PENDING taksStatus = iota
	RUNNING
	FAILED
	FINISH
)

func TaksStatus(i int) string {
	p := [4]string{"PENDING", "RUNNING", "FAILED", "FINISH"}
	return p[i]
}

type EzbWorkers struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Enable      bool      `json:"enable"`
	Comment     string    `json:"comment"`
	Fqdn        string    `json:"fqdn"`
	Register    time.Time `json:"register"`
	LastRequest time.Time `json:"lastrequest"`
	Request     int       `json:"request"`
	// Tags        []*EzbTags `json:"tags" gorm:"many2many:ezb_workers_has_ezb_tags;"`
}

type EzbJobs struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
	Comment  string `json:"comment"`
	Checksum string `json:"checksum"`
	Path     string `json:"path"`
	Cache    int    `json:"cache"`
	Output   string `json:"output"`
}
type EzbTags struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	// Workers []*EzbWorkers `json:"workers" gorm:"many2many:ezb_workers_has_ezb_tags;"`
	// Actions []*EzbActions `json:"actions" gorm:"many2many:ezb_actions_has_ezb_tags;"`
}

type ViewApi struct {
	Account   string `json:"account"`
	Accountid int    `json:"accountid"`
	Ctrl      string `json:"ctrl"`
	Ctrlid    int    `json:"ctrlid"`
	Ctrlver   int    `json:"ctrlver"`
	Action    string `json:"action"`
	Actionid  int    `json:"actionid"`
	Job       string `json:"job"`
	Jobid     int    `json:"jobid"`
	Access    string `json:"access"`
}

type ApiPath struct {
	ID  int    `json:"id"`
	API string `json:"api"`
	RGX string `json:"regex"`
}

type EzbControllers struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Enable  bool   `json:"enable"  `
	Comment string `json:"comment"`
	Version int    `json:"version"`
	// Accounts []EzbAccounts `gorm:"many2many:ezb_accounts_has_ezb_controllers;"`
	// Groups   []EzbGroups   `gorm:"many2many:ezb_groups_has_ezb_controllers;"`
	// Actions  []EzbActions  `gorm:"ForeignKey:ID"`
}

type EzbLogs struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	Status     string    `json:"status"`
	Ipaddr     string    `json:"ipaddr"`
	Token      string    `json:"token"`
	Controller string    `json:"controller"`
	Action     string    `json:"action"`
	URL        string    `json:"url"`
	Bastion    string    `json:"bastion"`
	Worker     string    `json:"worker"`
	Issuer     string    `json:"issuer"`
	Xtrack     string    `json:"xtrack"`
	Deprecated bool      `json:"deprecated"`
	Methode    string    `json:"methode"`
	Duration   int64     `json:"duration"`
	Size       int       `json:"size"`
	Error      string    `json:"error"`
	Account    string    `json:"account"`
}
