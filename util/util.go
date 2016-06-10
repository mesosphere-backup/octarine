package util

import (
	"os"
)

const DcosDomain string = ".mydcos.directory"
const MaxPortLength int = 5

func RmIfExist(s string) error {
	if _, err := os.Stat(s); err == nil {
		if err := os.Remove(s); err != nil {
			return err
		}
	}
	return nil
}
