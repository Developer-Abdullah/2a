package signing
import ( "encoding/json"; "github.com/hibiken/asynq" )
const TaskIPASign = "signing:ipa_sign"
type IPASignPayload struct { JobID, VersionID, CertID, TenantID string }
func NewIPASignTask(payload IPASignPayload) (*asynq.Task, error) {
bytes, _ := json.Marshal(payload)
return asynq.NewTask(TaskIPASign, bytes), nil
}
