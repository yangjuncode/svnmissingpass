package svnmissingpass

import "os"

var DefaultsvnPath = os.Getenv("APPDATA") + "\\Subversion\\auth\\svn.simple"

type TsvnPassItem struct {
	UserName string
	Repo     string
	Pass     string
}
