package ably_control

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ably App
type AblyApp struct {
	AccountID types.String `tfsdk:"account_id"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	TLSOnly   types.Bool   `tfsdk:"tls_only"`
}

// Ably Namespace
type AblyNamespace struct {
	AppID            types.String `tfsdk:"app_id"`
	ID               types.String `tfsdk:"id"`
	Authenticated    types.Bool   `tfsdk:"authenticated"`
	Persisted        types.Bool   `tfsdk:"persisted"`
	PersistLast      types.Bool   `tfsdk:"persist_last"`
	PushEnabled      types.Bool   `tfsdk:"push_enabled"`
	TlsOnly          types.Bool   `tfsdk:"tls_only"`
	ExposeTimeserial types.Bool   `tfsdk:"expose_timeserial"`
}
