package main

import (
	"fmt"
	"github.com/yangjuncode/svnmissingpass"
)

func main() {
	svnpass := svnmissingpass.SvnMissingPass(svnmissingpass.DefaultsvnPath)

	for i, item := range svnpass {
		fmt.Println("found ", i+1, " username:", item.UserName, " repo:", item.Repo, " pass:", item.Pass)
	}

	if len(svnpass) == 0 {
		fmt.Println("not found missing svn pass!")
	}
}
