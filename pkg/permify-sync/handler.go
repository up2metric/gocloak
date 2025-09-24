package permifysync

import (
	"context"
	"errors"
	"strings"

	permissions "bitbucket.org/up2metricPC/u2m-permissions"
	organizationPermissions "bitbucket.org/up2metricPC/u2m-permissions/entities/organization"
	"github.com/Nerzal/gocloak/v13"
	kafkadispatcher "github.com/Nerzal/gocloak/v13/pkg/kafka-dispatcher"
)

// Interface for Keycloak client to be able to use GetUserByEmail
type KeycloakClient interface {
	GetUserByEmail(email string) (*gocloak.User, error)
}

// PermifySync is a handler that syncs the organization memberships and roles to Permify based on Keycloak events
type PermifySync struct {
	permifyClient  *permissions.AuthzClient
	keycloakClient KeycloakClient
	dispatcher     *kafkadispatcher.Dispatcher
	realm          string
}

func NewPermifySync(permifyClient *permissions.AuthzClient, keycloakClient KeycloakClient, dispatcher *kafkadispatcher.Dispatcher, realm string) kafkadispatcher.Handler {
	return &PermifySync{
		permifyClient:  permifyClient,
		keycloakClient: keycloakClient,
		dispatcher:     dispatcher,
		realm:          realm,
	}
}

func (h *PermifySync) Handle(ctx context.Context, event kafkadispatcher.Event) error {
	if h.realm != event.RealmName {
		return nil
	}

	switch event.ResourceType {
	case kafkadispatcher.ResourceOrganizationMembership:
		if event.OperationType.Equal(kafkadispatcher.OperationTypeCreate) {
			return h.handleAddUser(ctx, event)
		} else if event.OperationType.Equal(kafkadispatcher.OperationTypeDelete) {
			return h.handleRemoveUser(ctx, event)
		}

	case kafkadispatcher.ResourceOrganization:
		if event.OperationType.Equal(kafkadispatcher.OperationTypeCreate) {
			return h.handleCreateOrganization(ctx, event)
		} else if event.OperationType.Equal(kafkadispatcher.OperationTypeDelete) {
			return h.handleDeleteOrganization(ctx, event)
		}
	}

	return nil

}

func (h *PermifySync) handleAddUser(_ context.Context, event kafkadispatcher.Event) error {
	orgID, err := event.ResourcePath.ExtractOrganizationID()
	if err != nil {
		return err
	}

	emailVal, ok := event.Details["email"]
	if !ok {
		return errors.New("email not found in event details")
	}

	email, ok := emailVal.(string)
	if !ok {
		return errors.New("email not a string")
	}

	user, err := h.keycloakClient.GetUserByEmail(email)
	if err != nil {
		return err
	}

	userID := user.ID

	if err := organizationPermissions.AddUser(h.permifyClient, orgID, *userID); err != nil {
		return err
	}

	return nil
}

func (h *PermifySync) handleRemoveUser(_ context.Context, event kafkadispatcher.Event) error {
	// Extract the organization ID and user ID from the resource path (organizations/<org-id>/members/<user-id>)
	tokens := strings.Split(string(event.ResourcePath), "/")
	if len(tokens) != 4 || tokens[0] != "organizations" || tokens[2] != "members" {
		return errors.New("invalid resource path")
	}

	orgID := tokens[1]
	userID := tokens[3]

	if err := organizationPermissions.RemoveUser(h.permifyClient, orgID, userID); err != nil {
		return err
	}

	return nil
}

func (h *PermifySync) handleCreateOrganization(_ context.Context, event kafkadispatcher.Event) error {
	// Extract the owner ID from representation's attributes
	attributes, ok := event.Representation["attributes"]
	if !ok {
		return errors.New("attributes not found in event representation")
	}

	attributesMap, ok := attributes.(map[string]any)
	if !ok {
		return errors.New("attributes not a map[string]any type")
	}

	owners, ok := attributesMap["owners"]
	if !ok {
		return errors.New("owners not found in attributes")
	}

	ownersList, ok := owners.([]string)
	if !ok {
		return errors.New("owners not a []string type")
	}

	if len(ownersList) == 0 {
		return errors.New("owners list is empty")
	}

	// There is only one owner in the list
	ownerID := ownersList[0]

	// Extract the organization ID from the resource path
	orgID, err := event.ResourcePath.ExtractOrganizationID()
	if err != nil {
		return err
	}

	// Set the owner role for the organization
	err = organizationPermissions.SetRole(h.permifyClient, orgID, organizationPermissions.ROLE_OWNER, ownerID)
	if err != nil {
		return err
	}

	return nil
}

func (h *PermifySync) handleDeleteOrganization(_ context.Context, event kafkadispatcher.Event) error {
	orgID, err := event.ResourcePath.ExtractOrganizationID()
	if err != nil {
		return err
	}

	if err := organizationPermissions.DeleteOrganization(h.permifyClient, orgID); err != nil {
		return err
	}

	return nil
}
