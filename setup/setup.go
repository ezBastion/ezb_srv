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

package setup

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	fqdn "github.com/ShowMax/go-fqdn"
	"github.com/ezbastion/ezb_srv/models"
)

var exPath string

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
}

func CheckConfig(isIntSess bool) (conf models.Configuration, err error) {
	confFile := path.Join(exPath, "conf/config.json")
	raw, err := ioutil.ReadFile(confFile)
	if err != nil {
		return conf, err
	}
	json.Unmarshal(raw, &conf)
	return conf, nil
}

func CheckFolder(isIntSess bool) {
	// var exPath string
	// if isIntSess {
	// 	exPath = "./"
	// } else {
	// ex, _ := os.Executable()
	// exPath = filepath.Dir(ex)
	// }
	// var conf models.Configuration
	// confFile := path.Join(exPath, "conf/config.json")
	if _, err := os.Stat(path.Join(exPath, "cert")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "cert"), 0600)
		if err != nil {
			return
		}
		log.Println("Make cert folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "log")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "log"), 0600)
		if err != nil {
			return
		}
		log.Println("Make log folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "conf")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "conf"), 0600)
		if err != nil {
			return
		}
		log.Println("Make conf folder.")
	}
	// c, _ := json.Marshal(conf)
	// ioutil.WriteFile(confFile, c, 0600)
	// log.Println(confFile, " saved.")

}

func Setup(isIntSess bool) error {

	_fqdn := fqdn.Get()
	quiet := true
	hostname, _ := os.Hostname()
	confFile := path.Join(exPath, "conf/config.json")
	CheckFolder(isIntSess)
	conf, err := CheckConfig(isIntSess)
	if err != nil {
		quiet = false
		conf.CacheL1 = 600
		conf.EzbDB = "https://localhost:8444/"
		conf.Listen = ":5100"
		conf.ServiceFullName = "Easy Bastion"
		conf.ServiceName = "ezb_srv"
		conf.Logger.LogLevel = "warning"
		conf.Logger.MaxSize = 10
		conf.Logger.MaxBackups = 5
		conf.Logger.MaxAge = 180
		conf.CaCert = "cert/ca.crt"
		conf.PrivateKey = "cert/ezb_srv.key"
		conf.PublicCert = "cert/ezb_srv.crt"
		conf.EzbPki = "localhost:6000"
		conf.SAN = []string{_fqdn, hostname}
	}

	_, fica := os.Stat(path.Join(exPath, conf.CaCert))
	_, fipriv := os.Stat(path.Join(exPath, conf.PrivateKey))
	_, fipub := os.Stat(path.Join(exPath, conf.PublicCert))
	if quiet == false {
		fmt.Print("\n\n")
		fmt.Println("***********")
		fmt.Println("*** PKI ***")
		fmt.Println("***********")
		fmt.Println("ezBastion nodes use elliptic curve digital signature algorithm ")
		fmt.Println("(ECDSA) to communicate.")
		fmt.Println("We need ezb_pki address and port, to request certificat pair.")
		fmt.Println("ex: 10.20.1.2:6000 pki.domain.local:6000")

		for {
			p := askForValue("ezb_pki", conf.EzbPki, `^[a-zA-Z0-9-\.]+:[0-9]{4,5}$`)
			c := askForConfirmation(fmt.Sprintf("pki address (%s) ok?", p))
			if c {
				conn, err := net.Dial("tcp", p)
				if err != nil {
					fmt.Printf("## Failed to connect to %s ##\n", p)
				} else {
					conn.Close()
					conf.EzbPki = p
					break
				}
			}
		}

		fmt.Print("\n\n")
		fmt.Println("Certificat Subject Alternative Name.")
		fmt.Printf("\nBy default using: <%s, %s> as SAN. Add more ?\n", _fqdn, hostname)
		for {
			tmp := conf.SAN

			san := askForValue("SAN (comma separated list)", strings.Join(conf.SAN, ","), `(?m)^[[:ascii:]]*,?$`)

			t := strings.Replace(san, " ", "", -1)
			tmp = strings.Split(t, ",")
			c := askForConfirmation(fmt.Sprintf("SAN list %s ok?", tmp))
			if c {
				conf.SAN = tmp
				break
			}
		}
	}

	if os.IsNotExist(fica) || os.IsNotExist(fipriv) || os.IsNotExist(fipub) {
		keyFile := path.Join(exPath, conf.PrivateKey)
		certFile := path.Join(exPath, conf.PublicCert)
		caFile := path.Join(exPath, conf.CaCert)
		request := newCertificateRequest(conf.ServiceName, 730, conf.SAN)
		generate(request, conf.EzbPki, certFile, keyFile, caFile)
	}
	if quiet == false {
		c, _ := json.Marshal(conf)
		ioutil.WriteFile(confFile, c, 0600)
		log.Println(confFile, " saved.")
	}

	return nil
}
func newCertificateRequest(commonName string, duration int, addresses []string) *x509.CertificateRequest {
	certificate := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"ezBastion"},
			CommonName:   commonName,
		},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}

	for i := 0; i < len(addresses); i++ {
		if ip := net.ParseIP(addresses[i]); ip != nil {
			certificate.IPAddresses = append(certificate.IPAddresses, ip)
		} else {
			certificate.DNSNames = append(certificate.DNSNames, addresses[i])
		}
	}

	return &certificate
}

func generate(certificate *x509.CertificateRequest, ezbpki, certFilename, keyFilename, caFileName string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		os.Exit(1)
	}

	derBytes, err := x509.CreateCertificateRequest(rand.Reader, certificate, priv)
	if err != nil {
		return
	}
	fmt.Println("Created Certificate Signing Request for client.")
	conn, err := net.Dial("tcp", ezbpki)
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Println("Successfully connected to Root Certificate Authority.")
	writer := bufio.NewWriter(conn)
	// Send two-byte header containing the number of ASN1 bytes transmitted.
	header := make([]byte, 2)
	binary.LittleEndian.PutUint16(header, uint16(len(derBytes)))
	_, err = writer.Write(header)
	if err != nil {
		return
	}
	// Now send the certificate request data
	_, err = writer.Write(derBytes)
	if err != nil {
		return
	}
	err = writer.Flush()
	if err != nil {
		return
	}
	fmt.Println("Transmitted Certificate Signing Request to RootCA.")
	// The RootCA will now send our signed certificate back for us to read.
	reader := bufio.NewReader(conn)
	// Read header containing the size of the ASN1 data.
	certHeader := make([]byte, 2)
	_, err = reader.Read(certHeader)
	if err != nil {
		return
	}
	certSize := binary.LittleEndian.Uint16(certHeader)
	// Now read the certificate data.
	certBytes := make([]byte, certSize)
	_, err = reader.Read(certBytes)
	if err != nil {
		return
	}
	fmt.Println("Received new Certificate from RootCA.")
	newCert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return
	}

	// Finally, the RootCA will send its own certificate back so that we can validate the new certificate.
	rootCertHeader := make([]byte, 2)
	_, err = reader.Read(rootCertHeader)
	if err != nil {
		return
	}
	rootCertSize := binary.LittleEndian.Uint16(rootCertHeader)
	// Now read the certificate data.
	rootCertBytes := make([]byte, rootCertSize)
	_, err = reader.Read(rootCertBytes)
	if err != nil {
		return
	}
	fmt.Println("Received Root Certificate from RootCA.")
	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return
	}

	err = validateCertificate(newCert, rootCert)
	if err != nil {
		return
	}
	// all good save the files
	keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("Failed to open key "+keyFilename+" for writing:", err)
		os.Exit(1)
	}
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Println("Failed to marshal priv:", err)
		os.Exit(1)
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyOut.Close()

	certOut, err := os.Create(certFilename)
	if err != nil {
		fmt.Println("Failed to open "+certFilename+" for writing:", err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	certOut.Close()

	caOut, err := os.Create(caFileName)
	if err != nil {
		fmt.Println("Failed to open "+caFileName+" for writing:", err)
		os.Exit(1)
	}
	pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: rootCertBytes})
	caOut.Close()

}
func validateCertificate(newCert *x509.Certificate, rootCert *x509.Certificate) error {
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)
	verifyOptions := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	_, err := newCert.Verify(verifyOptions)
	if err != nil {
		fmt.Println("Failed to verify chain of trust.")
		return err
	}
	fmt.Println("Successfully verified chain of trust.")

	return nil
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("\n%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
func askForValue(s, def string, pattern string) string {
	reader := bufio.NewReader(os.Stdin)
	re := regexp.MustCompile(pattern)
	for {
		fmt.Printf("%s [%s]: ", s, def)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		response = strings.TrimSpace(response)
		if response == "" {
			return def
		} else if re.MatchString(response) {
			return response
		} else {
			fmt.Printf("[%s] wrong format, must match (%s)\n", response, pattern)
		}
	}
}
