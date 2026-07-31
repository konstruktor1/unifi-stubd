package adoption

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	commandWireGuardGenerateKeyPair = "wg-genkeypair"
	commandNotificationReason       = "cmd"
)

type controllerCommand struct {
	name      string
	requestID string
}

type commandResult struct {
	name       string
	requestID  string
	privateKey string
	publicKey  string
}

// CommandLifecycleSummary contains only non-secret counters suitable for
// runtime status and controller-integration assertions.
type CommandLifecycleSummary struct {
	LastCommand     string `json:"last_command,omitempty"`
	HandledCount    int    `json:"handled_count,omitempty"`
	ResultSentCount int    `json:"result_sent_count,omitempty"`
	DuplicateCount  int    `json:"duplicate_count,omitempty"`
	Pending         bool   `json:"pending,omitempty"`
}

// CommandLifecycle keeps one controller command result in memory for retry
// and duplicate suppression. Private key material is never persisted.
type CommandLifecycle struct {
	result  commandResult
	has     bool
	pending bool
	summary CommandLifecycleSummary
}

func parseControllerCommand(response *ControllerResponse, raw map[string]json.RawMessage) {
	name := jsonExactString(raw["cmd"])
	if name != commandWireGuardGenerateKeyPair {
		response.Ignored = true
		switch {
		case isRejectOnlyControllerCommand(name):
			response.CommandName = name
			response.IgnoredReason = "controller command rejected by safety policy"
		case name == "wg-genpsk" || name == "wg-pubkey":
			response.CommandName = name
			response.IgnoredReason = "controller command has no in-process implementation"
		case name == "get-udapi-blocks":
			response.CommandName = name
			response.CommandRequestIDSet = jsonExactString(raw["request_id"]) != ""
			response.IgnoredReason = "get-udapi-blocks unavailable without a trusted local provider"
		default:
			response.IgnoredReason = "unsupported controller command ignored by safety policy"
		}
		return
	}
	response.CommandName = commandWireGuardGenerateKeyPair
	requestID := jsonString(raw["request_id"])
	response.CommandRequestIDSet = requestID != ""
	if requestID == "" {
		response.Ignored = true
		response.IgnoredReason = "wg-genkeypair command missing request_id"
		return
	}
	if commandValuePresent(raw["args"]) || commandValuePresent(raw["params"]) {
		response.Ignored = true
		response.IgnoredReason = "wg-genkeypair command arguments are not supported"
		return
	}
	response.command = &controllerCommand{name: commandWireGuardGenerateKeyPair, requestID: requestID}
}

func commandValuePresent(raw json.RawMessage) bool {
	value := strings.TrimSpace(string(raw))
	switch value {
	case "", "null", "{}", "[]":
		return false
	default:
		return true
	}
}

// Handle prepares a supported command result or requeues the cached result for
// a repeated request ID. It performs no shell execution or host mutation.
func (lifecycle *CommandLifecycle) Handle(response *ControllerResponse) error {
	if response == nil || response.command == nil {
		return nil
	}
	command := response.command
	response.CommandHandled = true
	response.CommandResultQueued = true
	if lifecycle.has && lifecycle.result.name == command.name && lifecycle.result.requestID == command.requestID {
		lifecycle.pending = true
		lifecycle.summary.DuplicateCount++
		lifecycle.summary.Pending = true
		response.CommandDuplicate = true
		return nil
	}

	result, err := newWireGuardKeyPairResult(*command)
	if err != nil {
		response.CommandHandled = false
		response.CommandResultQueued = false
		response.Ignored = true
		response.IgnoredReason = "wg-genkeypair result generation failed"
		return err
	}
	lifecycle.result = result
	lifecycle.has = true
	lifecycle.pending = true
	lifecycle.summary.LastCommand = command.name
	lifecycle.summary.HandledCount++
	lifecycle.summary.Pending = true
	return nil
}

// Attach adds the pending command result as a firmware-compatible notification
// to the next ordinary inform payload.
func (lifecycle *CommandLifecycle) Attach(payload []byte) ([]byte, bool, error) {
	if !lifecycle.pending || !lifecycle.has {
		return payload, false, nil
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(payload, &root); err != nil {
		return nil, false, fmt.Errorf("decode inform payload for controller command result: %w", err)
	}
	if root == nil {
		return nil, false, errors.New("inform payload for controller command result must be an object")
	}
	result := struct {
		EventString string `json:"event_string"`
		RequestID   string `json:"request_id"`
		Params      struct {
			PrivateKey string `json:"private_key"`
			PublicKey  string `json:"public_key"`
		} `json:"params"`
	}{EventString: lifecycle.result.name, RequestID: lifecycle.result.requestID}
	result.Params.PrivateKey = lifecycle.result.privateKey
	result.Params.PublicKey = lifecycle.result.publicKey
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, false, fmt.Errorf("encode controller command result: %w", err)
	}
	reason, err := json.Marshal(commandNotificationReason)
	if err != nil {
		return nil, false, fmt.Errorf("encode controller command notification reason: %w", err)
	}
	root["inform_as_notif"] = json.RawMessage("true")
	root["notif_reason"] = reason
	root["notif_payload"] = encoded
	attached, err := json.Marshal(root)
	if err != nil {
		return nil, false, fmt.Errorf("encode inform payload with controller command result: %w", err)
	}
	return attached, true, nil
}

// MarkResultSent records successful HTTP delivery without retaining the
// result as pending. The cached key pair remains memory-only for deduplication.
func (lifecycle *CommandLifecycle) MarkResultSent() {
	if !lifecycle.pending {
		return
	}
	lifecycle.pending = false
	lifecycle.summary.ResultSentCount++
	lifecycle.summary.Pending = false
}

// Summary returns a secret-free snapshot of command lifecycle counters.
func (lifecycle *CommandLifecycle) Summary() CommandLifecycleSummary {
	return lifecycle.summary
}

// Reset discards all memory-only command correlation, result, and summary
// state after a controller-requested local reset.
func (lifecycle *CommandLifecycle) Reset() {
	if lifecycle == nil {
		return
	}
	*lifecycle = CommandLifecycle{}
}

func newWireGuardKeyPairResult(command controllerCommand) (commandResult, error) {
	privateBytes := make([]byte, 32)
	if _, err := rand.Read(privateBytes); err != nil {
		return commandResult{}, fmt.Errorf("generate in-memory WireGuard private key: %w", err)
	}
	// WireGuard serializes the clamped X25519 scalar, while crypto/ecdh's
	// GenerateKey returns the original random bytes from PrivateKey.Bytes.
	privateBytes[0] &= 248
	privateBytes[31] &= 127
	privateBytes[31] |= 64
	privateKey, err := ecdh.X25519().NewPrivateKey(privateBytes)
	if err != nil {
		return commandResult{}, fmt.Errorf("derive in-memory WireGuard key pair: %w", err)
	}
	return commandResult{
		name:       command.name,
		requestID:  command.requestID,
		privateKey: base64.StdEncoding.EncodeToString(privateKey.Bytes()),
		publicKey:  base64.StdEncoding.EncodeToString(privateKey.PublicKey().Bytes()),
	}, nil
}
