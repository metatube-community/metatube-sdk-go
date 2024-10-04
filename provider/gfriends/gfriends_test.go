package gfriends

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestGfriends_GetActorInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"小澤マリア",
		"小松凛花",
		"谷あづさ",
		"若宮はずき",
	})
}

func TestGfriends_GetActorInfoByURL(t *testing.T) {
	testkit.Test(t, New, []string{
		"https://github.com/gfriends/gfriends?gfriends-id=%E5%B0%8F%E6%9D%BE%E5%87%9B%E8%8A%B1",
		//"https://github.com/gfriends/gfriends?gfriends-id=%E8%B0%B7%E3%81%82%E3%81%A5%E3%81%95",
	}, func(t *testing.T, a any) {
		require.Len(t, a.(*model.ActorInfo).Images, 14)
	})
}

func TestGfriends_SearchActor(t *testing.T) {
	testkit.Test(t, New, []string{
		"美竹すず",
	})
}
