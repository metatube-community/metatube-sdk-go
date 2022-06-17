package m3u8

import (
	"errors"
	"io"
	"sort"

	"github.com/grafov/m3u8"
)

const (
	MEDIA  = m3u8.MEDIA
	MASTER = m3u8.MASTER
)

func ParseBestMediaURI(reader io.Reader) (string, m3u8.ListType, error) {
	playList, listType, err := m3u8.DecodeFrom(reader, true)
	if err != nil {
		return "", 0, err
	}
	switch listType {
	case m3u8.MEDIA:
		return "" /* as is */, MEDIA, nil
	case m3u8.MASTER:
		masterPL := playList.(*m3u8.MasterPlaylist)
		if len(masterPL.Variants) == 0 {
			return "", MASTER, errors.New("no variants")
		}
		// sort by bandwidth.
		sort.SliceStable(masterPL.Variants, func(i, j int) bool {
			return masterPL.Variants[i].Bandwidth < masterPL.Variants[j].Bandwidth
		})
		// return uri with the highest bandwidth.
		return masterPL.Variants[len(masterPL.Variants)-1].URI, MASTER, nil
	default:
		return "", 0, errors.New("bad list type")
	}
}
