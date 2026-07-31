package adoption

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const (
	maxMgmtCFGBytes = 64 * 1024
	maxMgmtCFGLines = 256
	maxMgmtCFGLine  = 4096
)

// parseMgmtCFG accepts only whitelisted mgmt_cfg keys that affect future
// inform identity; other controller provisioning keys are ignored.
func parseMgmtCFG(mgmtCFG string) (Store, error) {
	store := Store{}
	if len(mgmtCFG) > maxMgmtCFGBytes {
		return Store{}, errors.New("mgmt_cfg exceeds 65536 bytes")
	}
	lines := strings.Split(mgmtCFG, "\n")
	if len(lines) > maxMgmtCFGLines {
		return Store{}, errors.New("mgmt_cfg exceeds 256 lines")
	}
	for _, line := range lines {
		if len(line) > maxMgmtCFGLine {
			return Store{}, errors.New("mgmt_cfg line exceeds 4096 bytes")
		}
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		// Unknown controller keys are ignored by design. The storeFields table is
		// the policy boundary for adoption data accepted from the controller.
		if field, ok := storeFieldByMgmtKey(key); ok {
			value = strings.TrimSpace(value)
			if err := validateMgmtField(key, value); err != nil {
				return Store{}, err
			}
			field.set(&store, value)
		}
	}
	return store, nil
}

func validateMgmtField(key, value string) error {
	switch key {
	case "authkey":
		if len(value) != 32 {
			return errors.New("authkey must contain exactly 32 hexadecimal characters")
		}
		if err := validateIndividualAuthKeyText(value, false); err != nil {
			return err
		}
	case "use_aes_gcm":
		if _, err := strconv.ParseBool(value); err != nil {
			return errors.New("use_aes_gcm must be boolean")
		}
	case "cfgversion":
		if len(value) > 31 {
			return errors.New("cfgversion exceeds 31 bytes")
		}
	case "inform_url":
		if len(value) > 2048 {
			return errors.New("inform_url exceeds 2048 bytes")
		}
		parsed, err := url.Parse(value)
		if err != nil || parsed.Opaque != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			return errors.New("inform_url must be an absolute HTTP(S) URL with a host")
		}
		if parsed.User != nil {
			return errors.New("inform_url must not contain userinfo")
		}
		if parsed.RawQuery != "" || parsed.ForceQuery {
			return errors.New("inform_url must not contain a query")
		}
		if parsed.Fragment != "" || strings.Contains(value, "#") {
			return errors.New("inform_url must not contain a fragment")
		}
	}
	return nil
}
