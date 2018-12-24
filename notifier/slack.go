package notifier

// 	case ChannelTypeSlack:
// 		client := slack.NewClient(channel.URL.String())
// 		field := slack.Field{
// 			Title: "NEOS Content Export",
// 			Value: strings.ToUpper(user) + " published a NEOS content export to " + strings.ToUpper(workspace),
// 			Short: false,
// 		}
// 		fields := []slack.Field{field}
// 		attachment := &slack.Attachment{
// 			Fallback: strings.ToUpper(user) + " published a NEOS content export to " + strings.ToUpper(workspace),
// 			Color:    "warning",
// 			Fields:   fields,
// 		}
// 		attachments := []*slack.Attachment{attachment}
// 		msg := &slack.Message{
// 			Channel:     channel.SlackChannel,
// 			IconEmoji:   ":ghost:",
// 			UserName:    "neosproxy",
// 			Attachments: attachments,
// 		}
// 		err := client.SendMessage(msg)
// 		return err == nil
