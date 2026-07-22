package quizUtilsHelper

import (
	"strconv"

	"github.com/randhir3-cloud/GK-Circle-v2/api/constants"
)

func IsValidCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	var num int
	var err error
	if num, err = strconv.Atoi(code); err != nil {
		return false
	}

	if num < constants.MinInvitationCode || num > constants.MaxInvitationCode {
		return false
	}

	return true
}
