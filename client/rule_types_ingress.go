package ably

// IngressPostgresOutboxTarget is the target configuration for ingress Postgres outbox rules.
type IngressPostgresOutboxTarget struct {
	URL               string  `json:"url"`
	OutboxTableSchema string  `json:"outboxTableSchema"`
	OutboxTableName   string  `json:"outboxTableName"`
	NodesTableSchema  string  `json:"nodesTableSchema"`
	NodesTableName    string  `json:"nodesTableName"`
	SSLMode           string  `json:"sslMode"`
	SSLRootCert       *string `json:"sslRootCert,omitempty"`
	PrimarySite       string  `json:"primarySite"`
}

// IngressPostgresOutboxRulePost is the request body for creating an ingress Postgres outbox rule.
type IngressPostgresOutboxRulePost struct {
	Status   string                      `json:"status,omitempty"`
	RuleType string                      `json:"ruleType"`
	Target   IngressPostgresOutboxTarget `json:"target"`
}

// IngressPostgresOutboxRulePatch is the request body for updating an ingress Postgres outbox rule.
type IngressPostgresOutboxRulePatch struct {
	Status   string                       `json:"status,omitempty"`
	RuleType string                       `json:"ruleType"`
	Target   *IngressPostgresOutboxTarget `json:"target,omitempty"`
}

// IngressMongoDBTarget is the target configuration for ingress MongoDB rules.
type IngressMongoDBTarget struct {
	URL                      string `json:"url"`
	Database                 string `json:"database"`
	Collection               string `json:"collection"`
	Pipeline                 string `json:"pipeline"`
	FullDocument             string `json:"fullDocument,omitempty"`
	FullDocumentBeforeChange string `json:"fullDocumentBeforeChange,omitempty"`
	PrimarySite              string `json:"primarySite"`
}

// IngressMongoDBRulePost is the request body for creating an ingress MongoDB rule.
type IngressMongoDBRulePost struct {
	Status   string               `json:"status,omitempty"`
	RuleType string               `json:"ruleType"`
	Target   IngressMongoDBTarget `json:"target"`
}

// IngressMongoDBRulePatch is the request body for updating an ingress MongoDB rule.
type IngressMongoDBRulePatch struct {
	Status   string                `json:"status,omitempty"`
	RuleType string                `json:"ruleType"`
	Target   *IngressMongoDBTarget `json:"target,omitempty"`
}
