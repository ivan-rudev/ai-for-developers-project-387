package usecase

import (
	"regexp"
	"strings"
)

var (
	emailRe    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	uuidV4Re   = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	timeHHMMRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
)

func validateEmail(email string) error {
	if !emailRe.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameRequired
	}
	return nil
}

func validateDuration(minutes int) error {
	if minutes <= 0 {
		return ErrInvalidDuration
	}
	return nil
}

func isValidUUIDV4(s string) bool {
	return uuidV4Re.MatchString(s)
}

func validateTimeHHMM(s string) error {
	if !timeHHMMRe.MatchString(s) {
		return ErrInvalidTime
	}
	return nil
}
