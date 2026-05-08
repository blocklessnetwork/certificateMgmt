package model

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ParseAccountDocument maps a raw accounts collection document into Account.
// It accepts common bson key variants for interoperability with hand-inserted data.
func ParseAccountDocument(doc map[string]any) (Account, error) {
	var a Account
	if id, ok := doc["_id"]; ok {
		if oid, ok := id.(primitive.ObjectID); ok {
			a.ID = oid
		}
	}

	a.Account = stringFromFields(doc, "account", "Account", "username")
	a.Password = stringFromFields(doc, "password", "Password")
	a.TwoFASec = stringFromFields(doc, "twoFASec", "TwoFASec", "two_fa_sec")
	if a.Account == "" {
		return Account{}, fmt.Errorf("missing non-empty account field (use bson key account, Account, or username)")
	}
	return a, nil
}

func stringFromFields(doc map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := doc[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" {
				return s
			}
		case primitive.ObjectID:
			return t.Hex()
		default:
			if s := strings.TrimSpace(fmt.Sprint(v)); s != "" {
				return s
			}
		}
	}
	return ""
}
