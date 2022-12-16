package main

import (
	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/provider/arzon"
	"github.com/metatube-community/metatube-sdk-go/provider/fanza"
	"github.com/metatube-community/metatube-sdk-go/provider/javbus"
	"github.com/metatube-community/metatube-sdk-go/provider/sod"
	"github.com/metatube-community/metatube-sdk-go/provider/xslist"
)

func main() {
	// Allocate app engine with request timeout set to one minute.
	app := engine.Default()

	// Search actor named `ひなたまりん` from Xs/List with fallback enabled.
	app.SearchActor("ひなたまりん", xslist.Name, true)

	// Search actor named `一ノ瀬もも` from all available providers with fallback enabled.
	app.SearchActorAll("一ノ瀬もも", true)

	// Search movie named `ABP-330` from JavBus with fallback enabled.
	app.SearchMovie("ABP-330", javbus.Name, true)

	// Search movie named `SSIS-110` from all available providers with fallback enabled.
	// Option fallback will search the database for movie info if the corresponding providers
	// fail to return valid metadata.
	app.SearchMovieAll("SSIS-110", true)

	// Get actor metadata id `5085` from Xs/List with lazy enabled.
	app.GetActorInfoByProviderID(xslist.Name, "5085", true)

	// Get actor metadata from given URL with lazy enabled.
	app.GetActorInfoByURL("https://xslist.org/zh/model/15659.html", true)

	// Get movie metadata id `1252925` from ARZON with lazy enable.
	// With the lazy option set to true, it will first try to search the database and return
	// the info directly if it exists. If the lazy option is set to false, it will fetch info
	// from the given provider and update the database.
	app.GetMovieInfoByProviderID(arzon.Name, "1252925", true)

	// Get movie metadata from given URL with lazy enabled.
	app.GetMovieInfoByURL("https://www.heyzo.com/moviepages/2189/index.html", true)

	// Get actor primary image id `24490` from Xs/List.
	app.GetActorPrimaryImage(xslist.Name, "24490")

	// Get movie primary image id `hmn00268` from FANZA with aspect ratio and pos set to default.
	app.GetMoviePrimaryImage(fanza.Name, "hmn00268", -1, -1)

	// Get movie primary image id `hmn00268` from FANZA with aspect ratio set to 7:10 and pos set to center.
	app.GetMoviePrimaryImage(fanza.Name, "hmn00268", 0.70, 0.5)

	// Get movie backdrop image id `DLDSS-077` from SOD.
	app.GetMovieBackdropImage(sod.Name, "DLDSS-077")
}
