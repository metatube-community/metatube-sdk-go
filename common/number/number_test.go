package number

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrim(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want string
	}{
		{"ABP-030", "ABP-030"},
		{"ABP-030-C", "ABP-030"},
		{"ABP-030-C.mp4", "ABP-030"},
		{"ABP-358_C.mkv", "ABP-358"},
		{"[]ABP-358_C.mkv", "ABP-358"},
		{"[22sht.me]ABP-358_C.mkv", "ABP-358"},
		{"ABP-030-C-c_c-C-Cd1-cd4.mp4", "ABP-030"},
		{"rctd-460ch.mp4", "rctd-460"},
		{"rctd-460-ch.mp4", "rctd-460"},
		{"rctd-460ch-ch.mp4", "rctd-460"},
		{"rctd-461-C-cD4.mp4", "rctd-461"},
		{"rctd-461-cd3.mp4", "rctd-461"},
		{"rctd-461-Cd3-C.mp4", "rctd-461"},
		{"rctd-461_Cd39-C.mp4", "rctd-461"},
		{"FC2-PPV-123456", "FC2-123456"},
		{"FC2PPV-123456", "FC2-123456"},
		{"FC2PPV_123456", "FC2-123456"},
		{"FC2_PPV_123456", "FC2-123456"},
		{"FC2-PPV_123456", "FC2-123456"},
		{"FC2-PPV-123456-C.mp4", "FC2-123456"},
		{"ssis00123.mp4", "ssis00123"},
		{"SDDE-625_uncensored_C", "SDDE-625"},
		{"SDDE-625_uncensored_C.mp4", "SDDE-625"},
		{"SDDE-625_uncensored_leak_C.mp4", "SDDE-625"},
		{"SDDE-625_uncensored_leak_C_cd1.mp4", "SDDE-625"},
		{"GIGL-677_4K.mp4", "GIGL-677"},
		{"GIGL-677_2K_h265.mp4", "GIGL-677"},
		{"GIGL-677_4K60FPS.mp4", "GIGL-677"},
		{"093021_539-FHD.mkv", "093021_539"},
		{"093021_539-1080pFHD.mkv", "093021_539"},
		{"SSIS-329_60FPS", "SSIS-329"},
		{"SSIS-329-C_60FPS", "SSIS-329"},
		{"SSIS-329-C_1080P30FPS", "SSIS-329"},
		{"SSIS-329-C_1080P30FPSFHDx264", "SSIS-329"},
		{"hhd800.com@HUNTB-269", "HUNTB-269"},
		{"sbw99.cc@iesp-653-4K.mp4", "iesp-653"},
		{"jav20s8.com@GIGL-677.mp4", "GIGL-677"},
		{"jav20s8.com@GIGL-677_4K.mp4", "GIGL-677"},
		{"133ARA-030你好.mp4", "133ARA-030"},
		{"133ARA-030 你好.mp4", "133ARA-030"},
		{"133ARA-030 hello there", "133ARA-030"},
		{"133ARA-030 hello there.mp4", "133ARA-030"},
		{"133ARA-030 - hello there.mp4", "133ARA-030"},
		{"133ARA-030-C 你好.mp4", "133ARA-030"},
		{"133ARA-030-C - 你好.mp4", "133ARA-030"},
		{"test.xxx@133ARA-030 你好", "133ARA-030"},
		{"test.xxx@133ARA-030 你好.mp4", "133ARA-030"},
		{"Tokyo Hot n9001 FHD.mp4", "n9001"},
		{"TokyoHot-n1287-HD .mp4", "n1287"},
		{"caribean-020317_001.mp4", "020317_001"},
		{"heydouga-4102-023-CD2.iso", "4102-023"},
	} {
		assert.Equal(t, unit.want, Trim(unit.orig), unit.orig)
	}
}

func TestIsUncensored(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want bool
	}{
		{"ABP-030", false},
		{"ssis00123", false},
		{"133ARA-030", false},
		{"FC2-738573", true},
		{"123456_789", true},
		{"123456-789", true},
		{"123456-01", true},
		{"xxx-av-1789", true},
		{"heydouga-1789-233", true},
		{"heyzo-1342", true},
		{"n1342", true},
		{"kb1342", true},
	} {
		assert.Equal(t, unit.want, IsUncensored(unit.orig), unit.orig)
	}
}

func TestSimilarity(t *testing.T) {
	for _, unit := range []struct {
		a, b string
	}{
		{"ABP-030", "ABP-030"},
		{"abp-030", "ABP-030"},
		{"ABS-030", "ABP-030"},
		{"AABP-030", "ABP-030"},
		{"KABP-030", "ABP-030"},
		{"ABP-030SP", "ABP-030"},
	} {
		t.Log(unit.a, unit.b, Similarity(unit.a, unit.b))
	}
}

func TestRequireFaceDetection(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want bool
	}{
		{"ABP-030", false},
		{"ssis00123", false},
		{"SIRO-030", true},
		{"133ARA-030", true},
		{"FC2-738573", true},
		{"123456_789", true},
		{"123456-01", true},
		{"xxx-av-1789", true},
		{"heydouga-1789-233", true},
		{"heyzo-1342", true},
		{"n1342", true},
		{"kb1342", true},
	} {
		assert.Equal(t, unit.want, RequireFaceDetection(unit.orig), unit.orig)
	}
}
