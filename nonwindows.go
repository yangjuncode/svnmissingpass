//go:build !windows

package svnmissingpass

func SvnMissingPass(svnPath string) (missingpass []TsvnPassItem) {
	//empty implementation
	return []TsvnPassItem{}
}
