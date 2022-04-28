//go:build windows

package svnmissingpass

import (
	"bufio"
	"github.com/billgraziano/dpapi"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var DefaultsvnPath = os.Getenv("APPDATA") + "\\Subversion\\auth\\svn.simple"

type TsvnPassItem struct {
	UserName string
	Repo     string
	Pass     string
}

func SvnMissingPass(svnPath string) (missingpass []TsvnPassItem) {
	if len(svnPath) == 0 {
		svnPath = DefaultsvnPath
	}
	fileInfos, _ := ioutil.ReadDir(svnPath)
	for _, f := range fileInfos {
		filename := filepath.Join(svnPath, f.Name())
		dictValues := ReadFile(filename)
		if len(dictValues) == 0 {
			continue
		}
		username := dictValues["username"]
		repo := dictValues["svn:realmstring"]
		encryptedpasswd := dictValues["password"]
		decryptpass := TryDecryptPassword(encryptedpasswd)

		missingpass = append(missingpass, TsvnPassItem{
			UserName: username,
			Repo:     repo,
			Pass:     decryptpass,
		})

		//fmt.Println("username:", username, "repo:", repo, "pass:", decryptpass)

	}

	return
}

//============impl internal==========

const MAX_LINES = 1024

type States int

const (
	ExpectingKeyDef States = iota
	ExpectingKeyName
	ExpectingValueDef
	ExpectingValue
)

type AuthFileParser struct {
	state      States
	keyName    string
	nextLength int
	props      map[string]string
}

func NewAuthFileParser() *AuthFileParser {
	return &AuthFileParser{
		state:      ExpectingKeyDef,
		keyName:    "",
		nextLength: -1,
		props:      map[string]string{},
	}
}
func (this *AuthFileParser) tryParseNextLine(line string) bool {
	switch this.state {
	case ExpectingKeyDef:
		return this.parseKeyDef(line)
	case ExpectingKeyName:
		return this.parseKeyName(line)
	case ExpectingValueDef:
		return this.parseValueDef(line)
	case ExpectingValue:
		return this.parseValue(line)
	default:
		return false
	}
}
func (this *AuthFileParser) parseKeyDef(line string) bool {
	if !this.parseDefLine("K", line) {
		return false

	}
	this.state = ExpectingKeyName
	return true
}
func (this *AuthFileParser) parseKeyName(line string) bool {
	if !this.parseValLine(line) {
		return false
	}
	this.state = ExpectingValueDef
	return true
}

func (this *AuthFileParser) parseValueDef(line string) bool {
	if !this.parseDefLine("V", line) {
		return false

	}
	this.state = ExpectingValue
	return true
}

func (this *AuthFileParser) parseValue(line string) bool {
	if !this.parseValLine(line) {
		return false

	}
	this.state = ExpectingKeyDef
	return true
}
func (this *AuthFileParser) parseDefLine(prefix string, line string) bool {
	line = strings.TrimSpace(line)
	lineupper := strings.ToUpper(line)
	if !strings.HasPrefix(lineupper, prefix+" ") {
		return false

	}
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		return false

	}
	atoi, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	this.nextLength = atoi

	return true
}

func (this *AuthFileParser) parseValLine(line string) bool {

	if len(line) < this.nextLength {
		return false

	}
	val := line[0:this.nextLength]
	this.nextLength = -1

	if this.state == ExpectingKeyName {
		this.keyName = strings.TrimSpace(val)
		if this.keyName == "" {
			return false
		}
		if strings.Contains(this.keyName, " ") {
			return false
		}
	} else {
		this.props[this.keyName] = val
		this.keyName = ""
	}

	return true
}

func ReadFile(path string) map[string]string {
	parser := NewAuthFileParser()

	f, err := os.Open(path)
	if err != nil {
		return map[string]string{}
	}
	// remember to close the file at the end of the program
	defer f.Close()

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)

	lineNum := 1

	for scanner.Scan() {
		if lineNum > MAX_LINES {
			break
		}

		line := scanner.Text()

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		lineupper := strings.ToUpper(line)
		if parser.state == ExpectingKeyDef && lineupper == "END" {
			return parser.props
		}
		parser.tryParseNextLine(line)
		lineNum += 1
	}

	return map[string]string{}

}

func TryDecryptPassword(encrypted string) (decrypted string) {
	decrypted, _ = dpapi.Decrypt(encrypted)
	return
}
