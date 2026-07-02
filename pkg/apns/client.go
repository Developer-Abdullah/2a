package apns

import (
"fmt"
"github.com/sideshow/apns2"
"github.com/sideshow/apns2/token"
)

type Client struct {
client *apns2.Client
topic  string
}

func NewClient(authKeyPEM []byte, keyID, teamID, topic string, production bool) (*Client, error) {
authKey, err := token.AuthKeyFromBytes(authKeyPEM)
if err != nil { return nil, fmt.Errorf("failed to parse APNs key: %w", err) }

tokenData := &token.Token{ AuthKey: authKey, KeyID: keyID, TeamID: teamID }
client := apns2.NewTokenClient(tokenData)
if production { client = client.Production() } else { client = client.Development() }

return &Client{client: client, topic: topic}, nil
}

func (c *Client) Send(deviceToken string, payload []byte, isSilent bool) error {
notification := &apns2.Notification{ DeviceToken: deviceToken, Topic: c.topic, Payload: payload }
if isSilent {
notification.Priority = apns2.PriorityLow
notification.PushType = apns2.PushTypeBackground
} else {
notification.Priority = apns2.PriorityHigh
notification.PushType = apns2.PushTypeAlert
}
res, err := c.client.Push(notification)
if err != nil { return err }
if !res.Sent() { return fmt.Errorf("apns rejected push: %v %v", res.StatusCode, res.Reason) }
return nil
}
