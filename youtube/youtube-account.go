package youtube

import (
	"github.com/elek/go-utils/kv"
	"google.golang.org/api/youtube/v3"
	"path"
)

func UpdateAccount(store kv.KV, service *youtube.Service, slug string) error {
	channel, err := GetChannel(service, slug)
	if err != nil {
		return err
	}
	return store.Put(path.Join("youtube", channel.Snippet.Title, "ytid"), []byte(channel.Id))

}
