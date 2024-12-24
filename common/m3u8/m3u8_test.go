package m3u8

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/grafov/m3u8"
	"github.com/stretchr/testify/require"
)

const m3u8Sample1 = `
#EXTM3U
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=688301
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/0640_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=165135
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/0150_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=262346
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/0240_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=481677
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/0440_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=1308077
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/1240_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=1927853
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/1840_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=2650941
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/2540_vod.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1, BANDWIDTH=3477293
http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/3340_vod.m3u8
`

const m3u8Sample2 = `
#EXTM3U
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10,
0640/06400.ts
#EXTINF:10,
0640/0640533.ts
#EXTINF:10,
0640/0640534.ts
#EXTINF:10,
0640/0640535.ts
#EXT-X-ENDLIST
`

func TestParseBestMediaURI(t *testing.T) {
	buf := bytes.NewReader(unsafe.Slice(unsafe.StringData(m3u8Sample1), len(m3u8Sample1)))
	url, typ, err := ParseBestMediaURI(buf)
	require.NoError(t, err)
	require.Equal(t, "http://qthttp.apple.com.edgesuite.net/1010qwoeiuryfg/3340_vod.m3u8", url)
	require.Equal(t, m3u8.MASTER, typ)

	buf = bytes.NewReader(unsafe.Slice(unsafe.StringData(m3u8Sample2), len(m3u8Sample2)))
	url, typ, err = ParseBestMediaURI(buf)
	require.NoError(t, err)
	require.Equal(t, "", url)
	require.Equal(t, m3u8.MEDIA, typ)
}
