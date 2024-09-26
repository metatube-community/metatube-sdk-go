package xslist

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestXsList_GetActorInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"107",
		"24490",
		"5085",
		"8",
		"12",
		"139",
		"175",
		"15659",
	})
}

func TestXsList_SearchActor(t *testing.T) {
	testkit.Test(t, New, []string{
		"白川ゆず",
		"果梨",
		"Saki",
		"美竹すず",
	})
}
