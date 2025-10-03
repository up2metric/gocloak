package permifysync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	permissions "bitbucket.org/up2metricPC/u2m-permissions"
	organizationPermissions "bitbucket.org/up2metricPC/u2m-permissions/entities/organization"
	"github.com/Nerzal/gocloak/v13"
	kafkadispatcher "github.com/Nerzal/gocloak/v13/pkg/kafka-dispatcher"
)

// Interface for Keycloak client to be able to use GetUserByEmail
type KeycloakClient interface {
	GetUserByEmail(context.Context, string) (*gocloak.User, error)
}

// PermifySync is a handler that syncs the organization memberships and roles to Permify based on Keycloak events
type PermifySync struct {
	permifyClient  *permissions.AuthzClient
	keycloakClient KeycloakClient
	realm          string
}

func NewPermifySync(permifyClient *permissions.AuthzClient, keycloakClient KeycloakClient, realm string) kafkadispatcher.Handler {
	return &PermifySync{
		permifyClient:  permifyClient,
		keycloakClient: keycloakClient,
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
		if event.OperationType.Equal(kafkadispatcher.OperationTypeDelete) {
			return h.handleDeleteOrganization(ctx, event)
		}
	}

	return nil

}

func (h *PermifySync) handleAddUser(ctx context.Context, event kafkadispatcher.Event) error {
	tokens := strings.Split(string(event.ResourcePath), "/")
	if len(tokens) != 3 || tokens[0] != "organizations" || tokens[2] != "members" {
		return errors.New("invalid resource path")
	}

	orgID := tokens[1]

	emailVal, ok := event.Details["email"]
	if !ok {
		return errors.New("email not found in event details")
	}

	email, ok := emailVal.(string)
	if !ok {
		return errors.New("email not a string")
	}

	user, err := h.keycloakClient.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	userID := user.ID

	// Add the user to the organization, thsi sets role to member
	if err := organizationPermissions.AddUser(h.permifyClient, orgID, *userID); err != nil {
		return err
	}

	// Check if the user is the owner of the organization
	owner, err := event.Representation.GetAttribute("owner")
	if err != nil {
		fmt.Println("Error getting owner attribute", err)
	}

	// If the user is the owner of the organization, set the owner role
	if err == nil && owner[0] == *userID {
		return organizationPermissions.SetRole(h.permifyClient, orgID, organizationPermissions.ROLE_OWNER, *userID)
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
