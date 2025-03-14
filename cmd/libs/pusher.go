package libs

import (
	"github.com/pusher/pusher-http-go/v5"
)

var PusherClient *pusher.Client

func NewPusherClient(appId, key, secret, cluster string) *pusher.Client {
	return &pusher.Client{
		AppID:   appId,
		Key:     key,
		Secret:  secret,
		Cluster: cluster,
		Secure:  true,
	}
}
