package dlsite

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestDLsite_GetMovieInfoByID(t *testing.T) {
	// 测试通过ID获取电影信息
	// 注意：DLsite的ID通常是数字或字母组合，可能需要根据实际情况调整测试数据
	testkit.Test(t, New, []string{
		"RJ01398548", // 示例：同人(maniax)作品编号
		"VJ011486",   // 示例：美少女游戏分区(pro)作品编号
	})
}

func TestDLsite_SearchMovie(t *testing.T) {
	// 测试搜索功能
	testkit.Test(t, New, []string{
		"勇者ちゃんの冒険",
		"祓魔○女シャルロット",
		"RJ01379131",
	})
}
