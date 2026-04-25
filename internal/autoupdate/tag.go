package autoupdate

import (
	"strconv"
	"strings"
)

type tagName string

func (t tagName) isLessOrEqual(ver string) bool {
	tMaj, tMin, tPat := parseSemVer(string(t))
	vMaj, vMin, vPat := parseSemVer(ver)
	return !(tMaj > vMaj || tMin > vMin || tPat > vPat)
}

func (t tagName) String() string {
	return string(t)
}

func parseSemVer(v string) (major, minor, patch int) {
	if !strings.Contains(v, ".") {
		return
	}
	v, _ = strings.CutPrefix(v, "v")
	splits := strings.SplitN(v, ".", 3)
	curMaj, curMin, curPat := splits[0], splits[1], splits[2]
	major, err := strconv.Atoi(curMaj)
	if err != nil {
		return
	}
	minor, err = strconv.Atoi(curMin)
	if err != nil {
		return
	}
	patch, err = strconv.Atoi(curPat)
	if err != nil {
		return
	}
	return
}
