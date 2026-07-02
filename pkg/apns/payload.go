package apns

import "encoding/json"

func BuildAlertPayload(title, body string, badge int, customData map[string]interface{}) ([]byte, error) {
payload := map[string]interface{}{
"aps": map[string]interface{}{ "alert": map[string]interface{}{ "title": title, "body": body }, "badge": badge, "sound": "default" },
}
for k, v := range customData { payload[k] = v }
return json.Marshal(payload)
}

func BuildSilentPayload(customData map[string]interface{}) ([]byte, error) {
payload := map[string]interface{}{ "aps": map[string]interface{}{ "content-available": 1 } }
for k, v := range customData { payload[k] = v }
return json.Marshal(payload)
}
