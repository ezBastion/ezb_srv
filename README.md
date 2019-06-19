#  ezBastion API frontend (ezb_srv)

The vault service, store key/value pair in a central store. It's used to not hardcode data in worker's scripts, like password or constant.



## SETUP


### 1. Download ezb_srv from [GitHub](<https://github.com/ezBastion/ezb_srv/releases/latest>)

### 2. Open an admin command prompte, like CMD or Powershell.

### 3. Run ezb_srv.exe with **init** option.

```powershell
    PS E:\ezbastion\ezb_srv> ezb_srv init
```

this commande will create folder and the default config.json file.
```json
{
    "listen": ":5002",
    "ezb_db": "https://localhost:8444/",
    "ezb_pki":"localhost:6000",
    "san":"ezbastion.public",
    "logger": {
        "loglevel": "debug",
        "maxsize": 10,
        "maxbackups": 5,
        "maxage": 180
    },    "cacheL1": 600,
    "privatekey": "cert/ezb_srv.key",
    "publiccert": "cert/ezb_srv.crt",
    "cacert": "cert/ca.crt",
    "jwtpubkey": "cert/ezb_sta.crt",
    "servicename": "ezBastion",
    "servicefullname": "Easy Bastion front"
}
```

- **loglevel**: Choose log level in debug,info,warning,error,critical.
- **maxsize**: is the maximum size in megabytes of the log file before it gets rotated. It defaults to 100 megabytes.
- **maxbackups**: MaxBackups is the maximum number of old log files to retain.
- **maxage**: MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.


### 4. Install Windows service and start it.

```powershell
    PS E:\ezbastion\ezb_srv> ezb_srv install
    PS E:\ezbastion\ezb_srv> ezb_srv start
```




## Copyright

Copyright (C) 2018 Renaud DEVERS info@ezbastion.com
<p align="center">
<a href="LICENSE"><img src="https://img.shields.io/badge/license-AGPL%20v3-blueviolet.svg?style=for-the-badge&logo=gnu" alt="License"></a></p>


Used library:

Name      | Copyright | version | url
----------|-----------|--------:|----------------------------
gin       | MIT       | 1.2     | github.com/gin-gonic/gin
cli       | MIT       | 1.20.0  | github.com/urfave/cli
gorm      | MIT       | 1.9.2   | github.com/jinzhu/gorm
logrus    | MIT       | 1.0.4   | github.com/sirupsen/logrus
go-fqdn   | Apache v2 | 0       | github.com/ShowMax/go-fqdn
jwt-go    | MIT       | 3.2.0   | github.com/dgrijalva/jwt-go
gopsutil  | BSD       | 2.15.01 | github.com/shirou/gopsutil
lumberjack| MIT       | 2.1     | github.com/natefinch/lumberjack
