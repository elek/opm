package youtube

import (
	"github.com/pkg/errors"
	"google.golang.org/api/youtube/v3"
	"strings"
)

func GetChannel(service *youtube.Service, name string) (*youtube.Channel, error) {
	call := service.Channels.List("id,snippet")
	call.Id(name)
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Items) > 0 {
		return resp.Items[0], nil
	}

	call = service.Channels.List("id,snippet")
	call.ForUsername(name)
	resp, err = call.Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Items) > 0 {
		return resp.Items[0], nil
	}
	return nil, errors.New("No channel with such name or id " + name)
}
func GetVideosForPlaylist(service *youtube.Service, id string) ([]*youtube.PlaylistItem, error) {
	return getVideosForPlaylistPage(service, id, "")
}

func getVideosForPlaylistPage(service *youtube.Service, id string, pageToken string) ([]*youtube.PlaylistItem, error) {
	result := make([]*youtube.PlaylistItem, 0)
	call := service.PlaylistItems.List("id,snippet")
	call.PlaylistId(id)
	call.MaxResults(50)
	if len(pageToken) > 0 {
		call.PageToken(pageToken)
	}
	response, err := call.Do()
	if err != nil {
		return nil, err
	}
	for _, video := range response.Items {
		result = append(result, video)
	}
	if len(response.NextPageToken) > 0 {
		nextPageResult, err := getVideosForPlaylistPage(service, id, response.NextPageToken)
		if err != nil {
			return nil, err
		}
		result = append(result, nextPageResult...)
	}
	return result, nil
}

func GetVideos(service *youtube.Service, ids []string) ([]*youtube.Video, error) {
	call := service.Videos.List("id,snippet,statistics")
	query := strings.Join(ids, ",")
	call.Id(query)
	call.MaxResults(50)
	response, err := call.Do()
	if err != nil {
		return nil, errors.Wrap(err, "Error on getting videos "+query)
	}

	return response.Items, nil
}

func GetPlaylists(service *youtube.Service, id string) ([]*youtube.Playlist, error) {
	return getPlaylistsPage(service, id, "")
}

func getPlaylistsPage(service *youtube.Service, id string, pageToken string) ([]*youtube.Playlist, error) {
	pls := make([]*youtube.Playlist, 0)
	call := service.Playlists.List("snippet,contentDetails")
	call.ChannelId(id)
	if len(pageToken) > 0 {
		call.PageToken(pageToken)
	}
	call.MaxResults(100)
	response, err := call.Do()
	if err != nil {
		return nil, errors.Wrap(err, "Can't download playlists for channel: "+id)
	}
	for _, playlist := range response.Items {
		pls = append(pls, playlist)
	}
	if len(response.NextPageToken) > 0 {
		pl2, err := getPlaylistsPage(service, id, response.NextPageToken)
		if err != nil {
			return pls, errors.Wrap(err, "Can't download playlist "+id)
		}
		pls = append(pls, pl2...)
	}
	return pls, nil
}
