package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatFloat64(in float64) string {
	var res string
	if float64(int64(in)) < in {
		res = fmt.Sprintf("%.1f", in)
		res = strings.Replace(res, ".", ",", -1)
	} else {
		res = strconv.FormatInt(int64(in), 10)
	}
	return res
}
