package youtube

import (
	"encoding/json"
	js "github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/runner"
	"github.com/elek/opm/writer"
	"github.com/urfave/cli/v2"
	"google.golang.org/api/youtube/v3"
	"path"
	"strconv"
)

func CreateYoutubeCommand() *cli.Command {

	return &cli.Command{
		Name:        "youtube",
		Description: "Retrieve and extract information from Youtube playlist",
		Flags: []cli.Flag{

		},
		Subcommands: []*cli.Command{
			{
				Name: "update",
				Subcommands: []*cli.Command{
					{
						Name: "account",
						Action: func(c *cli.Context) error {
							store, err := runner.CreateRepo(c)
							if err != nil {
								return err
							}
							client := getClient(youtube.YoutubeReadonlyScope)

							service, err := youtube.New(client)
							if err != nil {
								return err
							}
							return UpdateAccount(store, service, c.Args().Get(0))
						},
					},
					{
						Name:        "playlist",
						Description: "Update account information and download the list of available playlists",
						Action: func(c *cli.Context) error {
							store, err := runner.CreateRepo(c)
							if err != nil {
								return err
							}
							client := getClient(youtube.YoutubeReadonlyScope)

							service, err := youtube.New(client)
							if err != nil {
								return err
							}
							return DownloadPlaylists(store, service)
						},
					},
					{
						Name:        "playlist-item",
						Aliases:     []string{"pi"},
						Description: "Update video information for each of the playlists",
						Action: func(c *cli.Context) error {
							store, err := runner.CreateRepo(c)
							if err != nil {
								return err
							}
							client := getClient(youtube.YoutubeReadonlyScope)

							service, err := youtube.New(client)
							if err != nil {
								return err
							}
							return DownloadPlaylistDetails(store, service)
						},
					},
					{
						Name:        "video",
						Description: "Update video metadata for each playlist item",
						Action: func(c *cli.Context) error {
							store, err := runner.CreateRepo(c)
							if err != nil {
								return err
							}
							client := getClient(youtube.YoutubeReadonlyScope)

							service, err := youtube.New(client)
							if err != nil {
								return err
							}
							return DownloadVideoDetails(store, service)
						},
					},
				},
			},
			{
				Name: "extract",

				Action: func(c *cli.Context) error {
					store, err := runner.CreateRepo(c)
					if err != nil {
						return err
					}
					dest, err := runner.DestDir(c)
					if err != nil {
						return err
					}
					return extract(store, dest)
				},
			},
		},
	}
}

type YoutubeVideo struct {
	Account           string
	AccountId         string
	PlaylistAccountId string
	Id                string
	Title             string
	PublishedAt       int64
	LikeCount         int
	ViewCount         int
}
type YoutubePlaylistItem struct {
	Account       string
	PlaylistId    string
	PlaylistTitle string
	VideoId       string
}

func extract(store kv.KV, dest string) error {
	videosDest, err := writer.NewWriter("youtube-video", "csv", new(YoutubeVideo))
	if err != nil {
		return err
	}
	defer videosDest.Close()

	playlistItem, err := writer.NewWriter("youtube-playlist-video", "csv", new(YoutubePlaylistItem))
	if err != nil {
		return err
	}
	defer playlistItem.Close()

	accounts, err := store.List(path.Join("youtube"))
	if err != nil {
		return err
	}

	for _, account := range accounts {
		videos, err := store.List(path.Join(account, "videos"))
		if err != nil {
			return err
		}
		accountId, err := store.Get(path.Join(account, "ytid"))

		for _, videoKey := range videos {
			video, err := js.AsJson(store.Get(videoKey))
			if err != nil {
				return err
			}
			likeCount, err := optionlStringNumber(video, "statistics", "likeCount")
			if err != nil {
				return err
			}
			viewCount, err := optionlStringNumber(video, "statistics", "viewCount")
			if err != nil {
				return err
			}

			videosDest.Write(YoutubeVideo{
				Account:           path.Base(account),
				AccountId:         string(accountId),
				PlaylistAccountId: js.MS(video, "snippet", "channelId"),
				Id:                js.MS(video, "id"),
				Title:             js.MS(video, "snippet", "title"),
				PublishedAt:       js.ME("2006-01-02T15:04:05Z", video, "snippet", "publishedAt"),
				LikeCount:         likeCount,
				ViewCount:         viewCount,
			})
		}

		playlistKeys, err := store.List(path.Join(account, "playlists"))
		for _, playlistKey := range playlistKeys {
			playlistInfo, err := js.AsJson(store.Get(path.Join(playlistKey, "info")))

			playlistItemsKey, err := store.List(path.Join(playlistKey, "videos"))
			if err != nil {
				return err
			}
			for _, playlistItemKey := range playlistItemsKey {
				playlistItemJson, err := js.AsJson(store.Get(playlistItemKey))
				if err != nil {
					return err
				}
				err = playlistItem.Write(YoutubePlaylistItem{
					Account:       path.Base(account),
					PlaylistId:    js.MS(playlistItemJson, "snippet", "playlistId"),
					PlaylistTitle: js.MS(playlistInfo, "snippet", "title"),
					VideoId:       js.MS(playlistItemJson, "snippet", "resourceId", "videoId"),
				})
				if err != nil {
					return err
				}
			}

		}

	}
	return nil
}

func optionlStringNumber(video map[string]interface{}, keys ...string) (int, error) {
	raw := js.MS(video, keys...)
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}

func DownloadVideoDetails(store kv.KV, service *youtube.Service) error {

	idsBuffer := make([]string, 0)

	accounts, err := store.List(path.Join("youtube"))
	if err != nil {
		return err
	}

	for _, account := range accounts {
		playlists, err := store.List(path.Join(account, "playlists"))
		if err != nil {
			return err
		}
		for _, playlist := range playlists {
			videos, err := store.List(path.Join(playlist, "videos"))
			if err != nil {
				return err
			}

			for _, video := range videos {
				idsBuffer = append(idsBuffer, path.Base(video))
				if len(idsBuffer) > 49 {
					err = DownloadVideoDetailsBatch(store, service, path.Base(account), idsBuffer)
					if err != nil {
						return err
					}
					idsBuffer = make([]string, 0)

				}
			}

		}
		if len(idsBuffer) > 0 {
			err = DownloadVideoDetailsBatch(store, service, path.Base(account), idsBuffer)
			if err != nil {
				return err
			}
			idsBuffer = make([]string, 0)

		}
	}

	return nil
}
func DownloadVideoDetailsBatch(store kv.KV, service *youtube.Service, account string, ids []string) error {
	videos, err := GetVideos(service, ids)
	if err != nil {
		return err
	}
	for _, video := range videos {
		jsonVideo, err := json.MarshalIndent(video, "", "  ")
		if err != nil {
			return err
		}
		err = store.Put(path.Join("youtube", account, "videos", video.Id), jsonVideo)
		if err != nil {
			return nil
		}
	}
	return nil
}

func DownloadPlaylistDetails(store kv.KV, service *youtube.Service) error {
	accounts, err := store.List(path.Join("youtube"))
	if err != nil {
		return err
	}

	for _, account := range accounts {
		playlists, err := store.List(path.Join(account, "playlists"))
		if err != nil {
			return err
		}
		for _, playlist := range playlists {
			println(playlist)
			playlistId := path.Base(playlist)
			playlistItems, err := GetVideosForPlaylist(service, playlistId)
			if err != nil {
				return err
			}
			for _, playListItem := range playlistItems {
				jsonPlaylistItem, err := json.MarshalIndent(playListItem, "", "  ")
				if err != nil {
					return err
				}
				videoId := playListItem.Snippet.ResourceId.VideoId
				err = store.Put(path.Join("youtube", path.Base(account), "playlists", playlistId, "videos", videoId), jsonPlaylistItem)
				if err != nil {
					return nil
				}
			}
		}
	}
	return nil
}

func DownloadPlaylists(store kv.KV, service *youtube.Service) error {
	accounts, err := store.List(path.Join("youtube"))
	if err != nil {
		return err
	}

	for _, account := range accounts {
		if store.Contains(path.Join(account, "ytid")) {
			rawId, err := store.Get(path.Join(account, "ytid"))
			if err != nil {
				return err
			}
			id := string(rawId)
			playlists, err := GetPlaylists(service, id)
			if err != nil {
				return err
			}
			for _, playlist := range playlists {
				jsonPlaylist, err := json.MarshalIndent(playlist, "", "  ")
				if err != nil {
					return err
				}
				err = store.Put(path.Join("youtube", path.Base(account), "playlists", playlist.Id, "info"), jsonPlaylist)
				if err != nil {
					return nil
				}
			}
		}
	}
	return nil
}
