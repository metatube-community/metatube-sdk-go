package aylo

import (
	"strings"

	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/aylo/core"
)

const (
	// Sites
	BangBros                = "BangBros"
	BlackMaleMe             = "BlackMaleMe"
	BrazzersName            = "Brazzers"
	Bromo                   = "Bromo"
	CzechHunter             = "CzechHunter"
	Deviante                = "Deviante"
	DigitalPlayground       = "DigitalPlayground"
	FakeHub                 = "FakeHub"
	GayWire                 = "GayWire"
	LetsDoeIt               = "LetsDoeIt"
	Men                     = "Men"
	MetroHD                 = "MetroHD"
	MileHighMediaBiandTrans = "MileHighMedia_BiandTrans"
	MileHighMediaGay        = "MileHighMedia_Gay"
	MileHighMediaStraight   = "MileHighMedia_Straight"
	Mofos                   = "Mofos"
	NextDoorHobby           = "NextDoorHobby"
	PropertySex             = "PropertySex"
	RealityDudes            = "RealityDudes"
	RealityKings            = "RealityKings"
	SeanCody                = "SeanCody"
	SexyHub                 = "SexyHub"
	TransAngels             = "TransAngels"
	Tube8Vip                = "Tube8Vip"
	Twistys                 = "Twistys"
	WhyNotBi                = "WhyNotBi"
)

type Brazzers struct {
	*core.Aylo
}

func NewBrazzers() *Brazzers {
	return &Brazzers{
		Aylo: core.New(BrazzersName, strings.ToLower(BrazzersName)),
	}
}

func init() {
	// Register all sites
	//provider.Register(BangBros, NewBangBros)
	//provider.Register(BlackMaleMe, NewBlackMaleMe)
	provider.Register(BrazzersName, NewBrazzers)
	//provider.Register(Bromo, NewBromo)
	//provider.Register(CzechHunter, NewCzechHunter)
	//provider.Register(Deviante, NewDeviante)
	//provider.Register(DigitalPlayground, NewDigitalPlayground)
	//provider.Register(FakeHub, NewFakeHub)
	//provider.Register(GayWire, NewGayWire)
	//provider.Register(LetsDoeIt, NewLetsDoeIt)
	//provider.Register(Men, NewMen)
	//provider.Register(MetroHD, NewMetroHD)
	//provider.Register(MileHighMediaBiandTrans, NewMileHighMediaBiandTrans)
	//provider.Register(MileHighMediaGay, NewMileHighMediaGay)
	//provider.Register(MileHighMediaStraight, NewMileHighMediaStraight)
	//provider.Register(Mofos, NewMofos)
	//provider.Register(NextDoorHobby, NewNextDoorHobby)
	//provider.Register(PropertySex, NewPropertySex)
	//provider.Register(RealityDudes, NewRealityDudes)
	//provider.Register(RealityKings, NewRealityKings)
	//provider.Register(SeanCody, NewSeanCody)
	//provider.Register(SexyHub, NewSexyHub)
	//provider.Register(TransAngels, NewTransAngels)
	//provider.Register(Tube8Vip, NewTube8Vip)
	//provider.Register(Twistys, NewTwistys)
	//provider.Register(WhyNotBi, NewWhyNotBi)
}
