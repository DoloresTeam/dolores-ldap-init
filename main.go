package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	ldap "gopkg.in/ldap.v2"
	yaml "gopkg.in/yaml.v2"
)

// Config ...
type Config struct {
	Host        string
	Port        int
	RootDN      string
	RootPWD     string
	CRootDN     string
	CRootPWD    string
	Subffix     string
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

	// 链接ldap服务器
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		panic(err)
	}
	defer l.Close()

	err = l.Bind(config.CRootDN, config.CRootPWD)
	if err != nil {
		panic(err)
	}

	sq := ldap.NewSearchRequest(`cn=schema,cn=config`,
		ldap.ScopeBaseObject, ldap.DerefAlways, 0, 0, true, `(cn=dolores)`, []string{`objectClass`}, nil)

	sr, _ := l.Search(sq)
	if len(sr.Entries) == 0 {
		aq := ldap.NewAddRequest(`cn=dolores,cn=schema,cn=config`)
		if len(sr.Entries) == 0 {
			aq.Attribute(`objectClass`, []string{`olcSchemaConfig`})
			checkError(l.Add(aq))
		}
	}

	mq := ldap.NewModifyRequest(`cn={1}dolores,cn=schema,cn=config`) // config数据库默认会加上序号 {0}是core
	mq.Replace(`olcAttributeTypes`, []string{
		`( 0.9.3.2.8.0.1 NAME 'id' DESC 'Kevin.Gong unit id' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.2 NAME 'thirdAccount' DESC 'Kevin.Gong: thrid party account id' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.3 NAME 'thirdPassword' DESC 'Kevin.Gong: password of im' EQUALITY octetStringMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.40{256} )`,
		`( 0.9.3.2.8.0.4 NAME 'rbacType' DESC 'Kevin.Gong: rbace tpe for dolores.' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.5 NAME 'rbacRole' DESC 'Kevin.Gong: rbac alc of role' EQUALITY caseExactMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15 USAGE userApplications )`,
		`( 0.9.3.2.8.0.6 NAME ( 'upid' 'unitpermissionIdentifier' ) DESC 'Kevin.Gong: unit permission ids' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.7 NAME ( 'ppid' 'personpermissionIdentifier' ) DESC 'Kevin.Gong: person permission ids' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.8 NAME 'unitID' DESC 'Kevin.Gong unit id' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.9 NAME 'gender' DESC 'Kevin.Gong gender of member' EQUALITY numericStringMatch SUBSTR numericStringSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.36{16} )`,
		`( 0.9.3.2.8.0.10 NAME 'action' DESC 'Kevin.Gong action for audit unit&member update' EQUALITY numericStringMatch SUBSTR numericStringSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.36{16} )`,
		`( 0.9.3.2.8.0.11 NAME 'category' DESC 'Kevin.Gong audit category unit or member' EQUALITY numericStringMatch SUBSTR numericStringSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.36{16} )`,
		`( 0.9.3.2.8.0.12 NAME ( 'mid' ) DESC 'Kevin.Gong: members ids for audit' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
		`( 0.9.3.2.8.0.13 NAME 'auditContent' DESC 'Kevin.Gong: audit content' EQUALITY caseIgnoreMatch SUBSTR caseIgnoreSubstringsMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15{256} )`,
	})
	mq.Replace(`olcObjectClasses`, []string{
		`( 0.9.3.2.8.1.1 NAME 'member' DESC 'Kevin.Gong: member for dolores' STRUCTURAL MUST (id $ rbacRole $ rbacType $ unitID $ name $ telephoneNumber $ userPassword ) MAY ( labeledURI $ gender $ thirdAccount $ thirdPassword $ email $ title $ cn ) )`,
		`( 0.9.3.2.8.1.2 NAME 'unit' DESC 'Kevin.Gong: unit extended for dolores.' AUXILIARY MUST ( id $ rbacType ))`,
		`( 0.9.3.2.8.1.3 NAME 'permission' DESC 'Kevin.Gong: permission for dolores.' SUP top STRUCTURAL MUST ( id $ rbacType ) MAY ( cn $ description ) )`,
		`( 0.9.3.2.8.1.4 NAME 'role' DESC 'Kevin.Gong: role for dolores.' SUP top STRUCTURAL MUST ( cn $ id $ upid $ ppid ) MAY description )`,
		`( 0.9.3.2.8.1.5 NAME 'doloresType' DESC 'Kevin.Gong: deparment & person type for dolores.' SUP top STRUCTURAL MUST ( id $ cn ) MAY ( description ) )`,
		`( 0.9.3.2.8.1.6 NAME 'audit' DESC 'Kevin.Gong: audit unit&member changes.' SUP top STRUCTURAL MUST ( action $ category $ mid $ auditContent ) )`,
	})

	checkError(l.Modify(mq))
	// 修改mdb数据库acl

	mq = ldap.NewModifyRequest(`olcDatabase={1}mdb,cn=config`)

	mq.Replace(`olcAccess`, []string{`to * by self read`})

	err = l.Modify(mq)
	checkError(err)

	// 为mdb 添加overlay
	sq = ldap.NewSearchRequest(`cn=config`,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, true, `(olcOverlay=unique)`, []string{`objectClass`}, nil)

	sr, _ = l.Search(sq)
	if len(sr.Entries) == 0 {
		aq := ldap.NewAddRequest(`olcOverlay={0}unique,olcDatabase={1}mdb,cn=config`)
		aq.Attribute(`objectClass`, []string{`olcUniqueConfig`})
		aq.Attribute(`olcOverlay`, []string{`unique`})
		// 手机号码必须唯一
		aq.Attribute(`olcUniqueURI`, []string{`ldap:///ou=member,dc=dolores,dc=store?telephoneNumber`})
		checkError(l.Add(aq))
	}

	// 绑定 mdb 数据库
	err = l.Bind(config.RootDN, config.RootPWD)
	if err != nil {
		panic(err)
	}

	log.Print(`1. create root entry`)
	aq := ldap.NewAddRequest(config.Subffix)
	aq.Attribute(`objectClass`, []string{`organization`, `dcObject`, `top`})
	aq.Attribute(`o`, []string{`root created by dolores ldap init tool`})
	checkError(l.Add(aq))

	log.Print(`2. create unit`)
	aq = ldap.NewAddRequest(`ou=unit,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`unit`})
	checkError(l.Add(aq))

	log.Print(`2.1 create organization`)
	aq = ldap.NewAddRequest(`o=1,ou=unit,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organization`, `top`})
	aq.Attribute(`description`, []string{config.CompanyName})

	checkError(l.Add(aq))

	log.Print(`3. create member`)
	aq = ldap.NewAddRequest(`ou=member,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`member`})
	checkError(l.Add(aq))

	log.Print(`4. create permission`)
	aq = ldap.NewAddRequest(`ou=permission,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`permission`})
	checkError(l.Add(aq))

	log.Print(`4.1 create unit permission`)
	aq = ldap.NewAddRequest(`ou=unit,ou=permission,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`unit`})
	checkError(l.Add(aq))

	log.Print(`4.2 create member permission`)
	aq = ldap.NewAddRequest(`ou=member,ou=permission,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`member`})
	checkError(l.Add(aq))

	log.Print(`5. create role`)
	aq = ldap.NewAddRequest(`ou=role,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`role`})
	checkError(l.Add(aq))

	log.Print(`6. create type`)
	aq = ldap.NewAddRequest(`ou=type,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`type`})
	checkError(l.Add(aq))

	log.Print(`6.1 create unit type`)
	aq = ldap.NewAddRequest(`ou=unit,ou=type,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`type`})
	checkError(l.Add(aq))

	log.Print(`6.2 create member type`)
	aq = ldap.NewAddRequest(`ou=member,ou=type,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`type`})
	checkError(l.Add(aq))

	log.Print(`7 create audit`)
	aq = ldap.NewAddRequest(`ou=audit,` + config.Subffix)
	aq.Attribute(`objectClass`, []string{`organizationalUnit`, `top`})
	aq.Attribute(`ou`, []string{`audit`})
	checkError(l.Add(aq))

	log.Print(`initial done ~`)
}

func checkError(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
