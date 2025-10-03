package kafkadispatcher

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Representation map[string]any

func (r *Representation) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*r = nil
		return nil
	}

	var asArray []map[string]any
	if err := json.Unmarshal(data, &asArray); err == nil {
		*r = Representation{"values": asArray}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return json.Unmarshal([]byte(s), (*map[string]any)(r))
	}

	return json.Unmarshal(data, (*map[string]any)(r))
}

func (r *Representation) GetAttribute(key string) ([]string, error) {
	attributes, ok := (*r)["attributes"]
	if !ok {
		return nil, fmt.Errorf("attributes not found in representation")
	}

	attributesMap, ok := attributes.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("attributes not a map[string]any type")
	}

	attributeValue, ok := attributesMap[key]
	if !ok {
		return nil, fmt.Errorf("attribute %s not found in representation", key)
	}

	attributeSlice, ok := attributeValue.([]any)
	if !ok {
		return nil, fmt.Errorf("attribute %s is not a slice", key)
	}

	result := make([]string, len(attributeSlice))
	for i, item := range attributeSlice {
		if str, ok := item.(string); ok {
			result[i] = str
		} else {
			return nil, fmt.Errorf("attribute %s contains non-string value at index %d", key, i)
		}
	}
	return result, nil
}

// AuthDetails is embedded in Event.
type AuthDetails struct {
	RealmID   string `json:"realmId"`
	RealmName string `json:"realmName"`
	ClientID  string `json:"clientId"`
	UserID    string `json:"userId"`
	IPAddress string `json:"ipAddress"`
}

// Event represents a Keycloak event coming from Kafka.
type Event struct {
	ID        string `json:"id"`
	Time      int64  `json:"time"`
	RealmID   string `json:"realmId"`
	RealmName string `json:"realmName"`

	AuthDetails AuthDetails `json:"authDetails"`

	ResourceType         ResourceType   `json:"resourceType"`
	OperationType        OperationType  `json:"operationType"`
	ResourcePath         ResourcePath   `json:"resourcePath"`
	Representation       Representation `json:"representation"`
	Error                *string        `json:"error,omitempty"`
	Details              map[string]any `json:"details"`
	ResourceTypeAsString string         `json:"resourceTypeAsString"`
}

type ResourceType string

func (rt ResourceType) Equal(resourceType ResourceType) bool {
	return rt == resourceType
}

const (
	ResourceAuthExecution               ResourceType = "AUTH_EXECUTION"
	ResourceAuthExecutionFlow           ResourceType = "AUTH_EXECUTION_FLOW"
	ResourceAuthFlow                    ResourceType = "AUTH_FLOW"
	ResourceAuthenticatorConfig         ResourceType = "AUTHENTICATOR_CONFIG"
	ResourceAuthorizationPolicy         ResourceType = "AUTHORIZATION_POLICY"
	ResourceAuthorizationResource       ResourceType = "AUTHORIZATION_RESOURCE"
	ResourceAuthorizationResourceServer ResourceType = "AUTHORIZATION_RESOURCE_SERVER"
	ResourceAuthorizationScope          ResourceType = "AUTHORIZATION_SCOPE"
	ResourceClient                      ResourceType = "CLIENT"
	ResourceClientInitialAccessModel    ResourceType = "CLIENT_INITIAL_ACCESS_MODEL"
	ResourceClientRole                  ResourceType = "CLIENT_ROLE"
	ResourceClientRoleMapping           ResourceType = "CLIENT_ROLE_MAPPING"
	ResourceClientScope                 ResourceType = "CLIENT_SCOPE"
	ResourceClientScopeClientMapping    ResourceType = "CLIENT_SCOPE_CLIENT_MAPPING"
	ResourceClientScopeMapping          ResourceType = "CLIENT_SCOPE_MAPPING"
	ResourceClusterNode                 ResourceType = "CLUSTER_NODE"
	ResourceComponent                   ResourceType = "COMPONENT"
	ResourceCustom                      ResourceType = "CUSTOM"
	ResourceGroup                       ResourceType = "GROUP"
	ResourceGroupMembership             ResourceType = "GROUP_MEMBERSHIP"
	ResourceIdentityProvider            ResourceType = "IDENTITY_PROVIDER"
	ResourceIdentityProviderMapper      ResourceType = "IDENTITY_PROVIDER_MAPPER"
	ResourceOrganization                ResourceType = "ORGANIZATION"
	ResourceOrganizationMembership      ResourceType = "ORGANIZATION_MEMBERSHIP"
	ResourceProtocolMapper              ResourceType = "PROTOCOL_MAPPER"
	ResourceRealm                       ResourceType = "REALM"
	ResourceRealmRole                   ResourceType = "REALM_ROLE"
	ResourceRealmRoleMapping            ResourceType = "REALM_ROLE_MAPPING"
	ResourceRealmScopeMapping           ResourceType = "REALM_SCOPE_MAPPING"
	ResourceRequiredAction              ResourceType = "REQUIRED_ACTION"
	ResourceRequiredActionConfig        ResourceType = "REQUIRED_ACTION_CONFIG"
	ResourceUser                        ResourceType = "USER"
	ResourceUserFederationMapper        ResourceType = "USER_FEDERATION_MAPPER"
	ResourceUserFederationProvider      ResourceType = "USER_FEDERATION_PROVIDER"
	ResourceUserLoginFailure            ResourceType = "USER_LOGIN_FAILURE"
	ResourceUserProfile                 ResourceType = "USER_PROFILE"
	ResourceUserSession                 ResourceType = "USER_SESSION"
)

type OperationType string

func (ot OperationType) Equal(operationType OperationType) bool {
	return ot == operationType
}

const (
	OperationTypeAction = "ACTION"
	OperationTypeCreate = "CREATE"
	OperationTypeDelete = "DELETE"
	OperationTypeUpdate = "UPDATE"
)

type ResourcePath string

func (rp ResourcePath) ExtractUserID() (string, error) {
	// Extract the user ID from resourcePath ("users/<user-id>")
	const prefix = "users/"
	if !strings.HasPrefix(string(rp), prefix) {
		return "", fmt.Errorf("invalid resourcePath: %s", rp)
	}
	userID := strings.TrimPrefix(string(rp), prefix)

	return userID, nil
}

func (rp ResourcePath) ExtractOrganizationID() (string, error) {
	// Extract the organization ID from resourcePath ("organizations/<organization-id>")
	const prefix = "organizations/"
	if !strings.HasPrefix(string(rp), prefix) {
		return "", fmt.Errorf("invalid resourcePath: %s", rp)
	}

	orgID := strings.TrimPrefix(string(rp), prefix)

	return orgID, nil
}
