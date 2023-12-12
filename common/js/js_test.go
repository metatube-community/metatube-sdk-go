package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalObject(t *testing.T) {
	{
		data := struct {
			MovieSeq     string `json:"movie_seq"`
			Page         int    `json:"page"`
			Lang         string `json:"lang"`
			ProviderName string `json:"provider_name"`
		}{}
		jsCode := `var object = {movie_seq: '2969', page: 1, type: 'monthly', provider_name: 'heyzo', lang: 'ja'};
				reviews_get(object);`
		objName := "object"
		err := UnmarshalObject(jsCode, objName, &data)
		assert.NoError(t, err)
		assert.Equal(t, "2969", data.MovieSeq)
		assert.Equal(t, 1, data.Page)
		assert.Equal(t, "ja", data.Lang)
	}

	{
		data := struct {
			Comments []struct {
				Username string `json:"user_name"`
			} `json:"comments"`
		}{}
		jsCode := `var reviews = {"comments":[{"star":"\u2605?","user_name":"HEY","date":"2023-02-11 18:38:18","comment":"\u30bf\u30c4\u541b\u3001\u3044\u3064\u3082\u306f\u3082\u3063\u3068\u7a4d\u6975\u7684\u306a\u611f\u3058\u306a\u306e\u306b\u4eca\u56de\u306f\u89aa\u53cb\u306e\u5f7c\u5973\u3068\u7d61\u3080\u3068\u3044\u3046\u5f79\u67c4\u8a2d\u5b9a\u304b\u3089\u306a\u306e\u304b\u7a4d\u6975\u6027\u306b\u6b20\u3051\u307e\u3059\u306d\u3002\u306a\u307f\u3061\u3083\u3093\u306f\u7a4d\u6975\u7684\u306a\u5f79\u306a\u306e\u3067\u3001\u306a\u307f\u3061\u3083\u3093\u306b\u305f\u3059\u3051\u3063\u308c\u3066\u3044\u308b\u90e8\u5206\u304c\u5927\u304d\u3044\u304b\u306a\uff1f\u306a\u307f\u3061\u3083\u3093\u306e\u304a\u304b\u3052\u3067\u6e80\u70b9\u3067\u3059\u203c\u3067\u3082\u9732\u51fa\u597d\u304d\u306a\u3089\u3001\u30bf\u30c4\u541b\u304c\u8208\u596e\u3059\u308b\u3088\u3046\u306b\u3082\u3063\u3068\u30a8\u30c3\u30c1\u306a\u6311\u767a\u3057\u306a\u3044\u3068\u306d\u3002","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"f0cbd8148980aed6709e52907a9c8b8b","base64":"8MvYFImArt","id":"53627","eng":"0","score":{"overall":"5","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}},{"star":"\u2605?","user_name":"\u98a8\u96f2\u5150","date":"2023-01-05 11:45:15","comment":"\u306a\u3093\u304b\u7537\u512a\u304b\u3089\u306e\u8cac\u3081\u3067\u76db\u308a\u4e0a\u304c\u308a\u306b\u6b20\u3051\u308b\u306e\u304c\u6b98\u5ff5\u3002\u5973\u512a\u3055\u3093\u304c\u305f\u3060\u4e00\u751f\u61f8\u547d\u9811\u5f35\u3063\u3066\u308b\u3063\u3066\u611f\u3058\u3002\u3042\u3068\u3001\u3069\u3053\u304c\u300c\u9732\u51fa\u300d\u3060\u3088\uff57\uff57","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"ce115f66a3a91e055804734830ba43c8","base64":"zhFfZqOpHg","id":"53477","eng":"0","score":{"overall":"4","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}},{"star":"\u2605?","user_name":"\u4e16\u754c\uff11\u4f4d\u306e\u7537","date":"2023-01-03 09:06:15","comment":"\u3053\u308c\u306f\u6587\u53e5\u306a\u3057\u306e\u826f\u4f5c\u3067\u3059\u306d\u3002\u3053\u3093\u306a\u306b\u30a8\u30ed\u304f\u8feb\u3063\u3066\u304f\u308c\u305f\u3089\u3059\u3050\u6483\u6c88\u3057\u305d\u3046\u3067\u3059\u3002","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"385f9d684324bec37663651c02b8010d","base64":"OF+daEMkvs","id":"53463","eng":"0","score":{"overall":"5","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}},{"star":"\u2605?","user_name":"\u30aa\u30b8\u30b5\u30f3","date":"2023-01-01 13:07:18","comment":"\u5b89\u5ba4\u306a\u307f\u3061\u3083\u3093\u304b\u308f\u3044\u304f\u3066\u5927\u304d\u3081\u306e\u30aa\u30c3\u30d1\u30a4\u306b\u30d7\u30ea\u30f3\u3068\u3057\u305f\u304a\u5c3b\u306e\u30b9\u30ec\u30f3\u30c0\u30fc\u306a\u8eab\u4f53\u3067\u4e00\u672c\u7b4b\u306e\u7f8e\u30de\u30f3\u304c\u3044\u3044\u3067\u3059\u306d\u3002\u4f8b\u3048\u89aa\u53cb\u306e\u5f7c\u5973\u3067\u3042\u3063\u305f\u3068\u3057\u3066\u3082\u306a\u307f\u3061\u3083\u3093\u306e\u65b9\u304b\u3089\u7a4d\u6975\u7684\u306b\u8feb\u3089\u308c\u305f\u3089\u30bb\u30c3\u30af\u30b9\u3057\u306a\u3044\u8a33\u306b\u306f\u3044\u304d\u307e\u305b\u3093\u306d\u3002\u30bf\u30c4\u304f\u3093\u3068\u306e\u604b\u4eba\u540c\u58eb\u306e\u3088\u3046\u306a\u7d61\u307f\u306e\u30bb\u30c3\u30af\u30b9\u3092\u3057\u3066\u3044\u308b\u306a\u307f\u3061\u3083\u3093\u306f\u7f8e\u30dc\u30c7\u30a3\u3082\u76f8\u307e\u3063\u3066\u3068\u3066\u3082\u30ad\u30ec\u30a4\u3067\u7d20\u6575\u3067\u3057\u305f\u3002\u6587\u53e5\u306a\u3057\u306e\u6e80\u70b9\u3067\u3059\u3002","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"91cf09fbcca9303e8a4d34775f830fed","base64":"kc8J+8ypMD","id":"53455","eng":"0","score":{"overall":"5","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}},{"star":"\u2605?","user_name":"\u6c41\u5973\u512a\u547d","date":"2023-01-01 12:42:16","comment":"\u9a0e\u4e57\u4f4d\u3067\u306e\u7d20\u6674\u3089\u3057\u3044\u8170\u632f\u308a\u3002\u30bd\u30d5\u30a1\u30fc\u3067\uff13\uff10\u5206\u4f4d\u3001\u80cc\u9762\u9a0e\u4e57\u3057\u3066\u308b\u5f71\u50cf\u304c\u3001\u898b\u305f\u3044\u3067\u3059\u3002","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"215632e4acfffdabadfc20bad4823cbc","base64":"IVYy5Kz\/\/a","id":"53454","eng":"0","score":{"overall":"5","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}},{"star":"\u2605?","user_name":"tsuruman","date":"2023-01-01 07:52:02","comment":"\u5de6\u80a9\u306e\u30bf\u30c8\u30a5\u306f\u3082\u3046\u5c11\u3057\u4e0a\u624b\u304f\u6d88\u305b\u306a\u3044\u306e\u304b\u3057\u3089\u3002\u3053\u308c\u304c\u76ee\u306b\u3064\u3044\u3066\u3001\u3069\u3046\u3057\u3066\u3082\u5b89\u3063\u307d\u3044\u5973\u306b\u898b\u3048\u3061\u3083\u3046\u3002","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"bfe6cd26fc89ffbb479d6b4087dd7a2a","base64":"v+bNJvyJ\/7","id":"53453","eng":"0","score":{"overall":"4","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}},{"star":"\u2605?","user_name":"nanchun","date":"2023-01-01 02:17:23","comment":"just amazing","comment_res":"","point_flag":0,"monthly_point_flag":0,"hash":"93157bc26dcea2d1a16f28e7816dc07a","base64":"kxV7wm3Oot","id":"53452","eng":"0","score":{"overall":"5","quality":"5","content":"5","actress":"5","play":"5","price":"5"},"vote":{"yes":"0","no":"0"}}],"pages":{"curr":"1","movie_seq":"2969","showall":0,"prev":0,"next":0,"first":0,"last":0,"lang":"ja","n":[1]},"count_ja":7,"count_en":1};`
		objName := "reviews"
		err := UnmarshalObject(jsCode, objName, &data)
		assert.NoError(t, err)
		assert.Equal(t, "HEY", data.Comments[0].Username)
	}
}
