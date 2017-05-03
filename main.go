package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	ldap "gopkg.in/ldap.v2"
	yaml "gopkg.in/yaml.v1"
)

type Config struct {
	Host        string
	Port        int
	RootDN      string
	RootPWD     string
	Subfix      string
	CompanyName string
}

var config = Config{}

func main() {

	configFilePath := flag.String(`path`, `./conf.yaml`, `配置文件路径`)

	flag.Parse()

	_, err := os.Stat(*configFilePath)
	if !(err == nil || os.IsExist(err)) {
		panic(`配置文件不存在`)
	}
	data, _ := ioutil.ReadFile(*configFilePath)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Panic(`配置文件不正确`)
	}

	log.Printf(`config: [%s]`, config)

	log.Print(`will dial ldap server`)
	// 链接ldap服务器
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.Print(`did dial ldap server`)
	log.Print(`will bind database`)
	// 绑定config数据库
	err = l.Bind(config.RootDN, config.RootPWD)
	if err != nil {
		panic(err)
	}
	log.Print(`did bind database`)

	log.Print(`1. create root entry`)

	aq := ldap.NewAddRequest(config.Subfix)
	aq.Attribute(`objectClass`, []string{`organization`, `dcObject`, `top`})
	aq.Attribute(`o`, []string{`root created by dolores ldap init tool`})
	checkError(l.Add(aq))

	log.Print(`2. create unit`)
	aq = ldap.NewAddRequest(`ou=unit,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`unit`})
	checkError(l.Add(aq))

	log.Print(`2.1 create organization`)
	aq = ldap.NewAddRequest(`oid=1,ou=unit,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organization`, `unitExtended`, `top`})
	aq.Attribute(`o`, []string{config.CompanyName})
	aq.Attribute(`oid`, []string{`1`})

	checkError(l.Add(aq))

	log.Print(`3. create person`)
	aq = ldap.NewAddRequest(`ou=person,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`person`})
	checkError(l.Add(aq))

	log.Print(`4. create permission`)
	aq = ldap.NewAddRequest(`ou=permission,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`permission`})
	checkError(l.Add(aq))

	log.Print(`5. create role`)
	aq = ldap.NewAddRequest(`ou=role,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`role`})
	checkError(l.Add(aq))

	log.Print(`6. create type`)
	aq = ldap.NewAddRequest(`ou=type,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`type`})
	checkError(l.Add(aq))

	log.Print(`6.1 create unit type`)
	aq = ldap.NewAddRequest(`ou=unit,ou=type,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`type`})
	checkError(l.Add(aq))

	log.Print(`6.2 create person type`)
	aq = ldap.NewAddRequest(`ou=person,ou=type,` + config.Subfix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`type`})
	checkError(l.Add(aq))

	log.Print(`initial done ~`)
}

func checkError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
