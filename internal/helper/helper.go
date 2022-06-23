package helper

import (
	"fmt"
	"strconv"
	"strings"
)

func IncStr(str string) (string, error) {
	if str == "" {
		return "", nil
	}
	strParts := strings.Split(str, "-")
	numPart, err := strconv.ParseInt(strParts[1], 10, 64)
	if err != nil {
		return "", err
	}
	numPart += 1

	numLength := len(strParts[1]) - len(fmt.Sprintf("%v", numPart))
	strParts[1] = ""
	for i := 0; i < numLength; i++ {
		strParts[1] += "0"
	}
	return fmt.Sprintf("%v-%v%v", strParts[0], strParts[1], numPart), nil

}
