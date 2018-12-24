package config

// import "fmt"

// // GetChannelType will return a channel's type
// func (c *Channel) GetChannelType() ChannelType {
// 	return c.Type
// }

// // GetWorkspacesForChannelIdentifier will return all workspaces for a given channel identifier
// func (c *Config) GetWorkspacesForChannelIdentifier(channelIdentifier string) (workspaces []string, e error) {
// 	workspaces = []string{}
// 	for _, hook := range c.Callbacks.NotifyOnUpdateHooks {
// 		if hook.Channel == channelIdentifier {
// 			// channel
// 			channel, ok := c.Channels[hook.Channel]
// 			if !ok {
// 				e = fmt.Errorf("channel %s not configured for callback %s", hook.Channel, channelIdentifier)
// 				return
// 			}

// 			if channel.GetChannelType() != ChannelTypeFoomo {
// 				e = fmt.Errorf("unexpected channel type %s for callback %s", channel.GetChannelType(), channelIdentifier)
// 				return
// 			}

// 			workspaces = append(workspaces, hook.Workspace)
// 		}
// 	}

// 	if len(workspaces) == 0 {
// 		e = fmt.Errorf("callback %s not configured", channelIdentifier)
// 	}

// 	return
// }
