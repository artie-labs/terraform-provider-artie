package tfmodels

import (
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeOnlyOr returns a write-only secret's value when it's set, and the plain attribute's value otherwise.
func writeOnlyOr(writeOnly, plain types.String) string {
	if writeOnly.IsNull() {
		return plain.ValueString()
	}
	return writeOnly.ValueString()
}

// CopyWriteOnlySecrets copies each `_wo` secret from config into dst, since Terraform leaves them out of the plan. dst
// should only be used to build the API request: its secrets must never reach state.
func CopyWriteOnlySecrets(dst *Connector, config Connector) {
	copyConfigFields(dst, config, "_wo")
}

// copyConfigFields copies the fields whose tfsdk tag ends with suffix from src to dst, in each config block both set.
func copyConfigFields(dst *Connector, src Connector, suffix string) {
	dstValue := reflect.ValueOf(dst).Elem()
	srcValue := reflect.ValueOf(src)
	for i := range dstValue.NumField() {
		dstBlock := dstValue.Field(i)
		srcBlock := srcValue.Field(i)
		if dstBlock.Kind() != reflect.Pointer {
			continue
		}
		if dstBlock.IsNil() {
			continue
		}
		if srcBlock.IsNil() {
			continue
		}
		if dstBlock.Elem().Kind() != reflect.Struct {
			continue
		}
		blockType := dstBlock.Elem().Type()
		for j := range blockType.NumField() {
			if strings.HasSuffix(blockType.Field(j).Tag.Get("tfsdk"), suffix) {
				dstBlock.Elem().Field(j).Set(srcBlock.Elem().Field(j))
			}
		}
	}
}

// SecretMissing reports whether neither a secret nor its write-only alternative is set. Unknown counts as set.
func SecretMissing(plain, writeOnly types.String) bool {
	if !IsKnownAndEmpty(plain) {
		return false
	}
	return IsKnownAndEmpty(writeOnly)
}

// SecretSet reports whether a secret or its write-only alternative is set to a known, non-empty value.
func SecretSet(plain, writeOnly types.String) bool {
	if IsKnownAndNonEmpty(plain) {
		return true
	}
	return IsKnownAndNonEmpty(writeOnly)
}

// AllNull reports whether every value is null.
func AllNull(values []types.String) bool {
	for _, value := range values {
		if !value.IsNull() {
			return false
		}
	}
	return true
}
